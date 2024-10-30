package tcp

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/ormanli/form3-te/internal/app/simulator"
)

type Service interface {
	Process(amount int) error
}

type Transport struct {
	service          Service
	cfg              simulator.Config
	listener         net.Listener
	stopHandlingChan chan struct{}
	wg               *sync.WaitGroup
}

func NewTransport(cfg simulator.Config, service Service) *Transport {
	return &Transport{
		cfg:              cfg,
		service:          service,
		stopHandlingChan: make(chan struct{}),
		wg:               &sync.WaitGroup{},
	}
}

func (t *Transport) Start(ctx context.Context) error {
	slog.Info("Server started", "port", t.cfg.ServerPort)

	var err error

	t.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", t.cfg.ServerHost, t.cfg.ServerPort))
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := t.listener.Accept()
			if err != nil {
				slog.Error("Failed to accept connection", "error", err)
				continue
			}

			go t.handleConnection(conn) // TODO goroutine pool
		}
	}
}

func (t *Transport) Stop(ctx context.Context) error {
	defer slog.Info("Server stopped")

	<-ctx.Done()

	close(t.stopHandlingChan)

	t.wg.Wait()

	return t.listener.Close()
}

var defaultCancelledResponse = response{
	status: Rejected,
	reason: "Cancelled",
}

func (t *Transport) handleConnection(conn net.Conn) {
	t.wg.Add(1)
	defer t.wg.Done()

	defer conn.Close()

	slog.Info("Handling connection", "remote", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		request := scanner.Text()

		select {
		case <-t.stopHandlingChan:
			fmt.Fprintf(conn, "%s\n", defaultCancelledResponse)

			return
		default:
			response := t.handleRequest(request)
			fmt.Fprintf(conn, "%s\n", response)
			slog.Debug("Handling request", "request", request, "response", response) // TODO update to make traceable
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("Error reading from connection", "error", err)
	}
}

func (t *Transport) handleRequest(s string) response {
	r, err := parseRequest(s)
	if err != nil {
		return response{
			status: Rejected,
			reason: err.Error(),
		}
	}

	err = t.service.Process(r.amount)
	if err != nil {
		return response{
			status: Rejected,
			reason: err.Error(),
		}
	}

	return response{
		status: Accepted,
		reason: "Transaction processed",
	}
}

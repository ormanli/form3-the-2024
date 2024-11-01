package tcp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/benbjohnson/clock"

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
	wg               sync.WaitGroup
	clock            clock.Clock
}

func NewTransport(cfg simulator.Config, service Service, clock clock.Clock) *Transport {
	return &Transport{
		cfg:              cfg,
		service:          service,
		stopHandlingChan: make(chan struct{}),
		wg:               sync.WaitGroup{},
		clock:            clock,
	}
}

func (t *Transport) Start(ctx context.Context) error {
	var err error
	t.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", t.cfg.ServerHost, t.cfg.ServerPort))
	if err != nil {
		return err
	}

	defer slog.Info("Server stopped")

	slog.Info("Server started", "port", t.cfg.ServerPort)

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		for {
			select {
			case <-ctx.Done():
				t.listener.Close()
				return
			default:
				conn, err := t.listener.Accept()
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						return
					}
					slog.Error("Failed to accept connection", "error", err)
					continue
				}

				t.wg.Add(1)
				go t.handleConnection(conn)
			}
		}
	}()

	t.waitForGracefulShutdown(ctx)

	t.wg.Wait()

	return nil
}

func (t *Transport) waitForGracefulShutdown(ctx context.Context) {
	<-ctx.Done()

	t.clock.Sleep(t.cfg.ServerGracefulShutdownTimeout)

	close(t.stopHandlingChan)
}

var defaultCancelledResponse = response{
	status: Rejected,
	reason: "Cancelled",
}

func (t *Transport) handleConnection(conn net.Conn) {
	defer t.wg.Done()

	defer conn.Close()

	slog.Debug("Handling connection", "remote", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		request := scanner.Text()

		responseChan := make(chan response, 1)

		go func() {
			select {
			case <-t.stopHandlingChan:
				return
			case responseChan <- t.handleRequest(request):
			}
		}()

		select {
		case <-t.stopHandlingChan:
			fmt.Fprintf(conn, "%s\n", defaultCancelledResponse)
			slog.Debug("Handling request", "request", request, "response", defaultCancelledResponse)
			return
		case response := <-responseChan:
			fmt.Fprintf(conn, "%s\n", response)
			slog.Debug("Handling request", "request", request, "response", response)
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

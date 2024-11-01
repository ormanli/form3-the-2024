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

// Service defines the interface for processing requests.
type Service interface {
	Process(amount int) error
}

// Transport manages TCP connections and handles incoming requests.
type Transport struct {
	service          Service
	cfg              simulator.Config
	listener         net.Listener
	stopHandlingChan chan struct{}
	wg               sync.WaitGroup
	clock            clock.Clock
}

// NewTransport creates a new Transport instance.
func NewTransport(cfg simulator.Config, service Service, clock clock.Clock) *Transport {
	return &Transport{
		cfg:              cfg,
		service:          service,
		stopHandlingChan: make(chan struct{}),
		wg:               sync.WaitGroup{},
		clock:            clock,
	}
}

// Start initializes the TCP server and starts accepting connections.
// It will block until context is cancelled and grace period is finished.
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
	}()

	t.waitForGracefulShutdown(ctx)

	return nil
}

// waitForGracefulShutdown waits for a graceful shutdown signal, sleeps until shutdown timeout and then closes the channel to stop handling connections.
func (t *Transport) waitForGracefulShutdown(ctx context.Context) {
	<-ctx.Done()

	slog.Info("Server graceful shutdown started")

	err := t.listener.Close()
	if err != nil {
		slog.Error("Error closing listener", "error", err)
	}

	t.clock.Sleep(t.cfg.ServerGracefulShutdownTimeout)

	close(t.stopHandlingChan)

	t.wg.Wait()
}

var defaultCancelledResponse = response{
	status: Rejected,
	reason: "Cancelled",
}

// handleConnection manages the lifecycle of a single TCP connection, reading requests and sending responses.
func (t *Transport) handleConnection(conn net.Conn) {
	defer t.wg.Done()

	defer conn.Close() //nolint:errcheck

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
			writeResponse(conn, request, defaultCancelledResponse)
			return
		case response := <-responseChan:
			writeResponse(conn, request, response)
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("Error reading from connection", "error", err)
	}
}

// handleRequest processes an incoming request and returns a corresponding response.
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

// writeResponse sends a response back to the client over the provided connection.
func writeResponse(conn net.Conn, request string, r response) {
	_, err := fmt.Fprintf(conn, "%s\n", r)
	if err != nil {
		slog.Error("Failed to write response", "error", err, "request", request, "response", r)
		return
	}
	slog.Debug("Handling request", "request", request, "response", r)
}

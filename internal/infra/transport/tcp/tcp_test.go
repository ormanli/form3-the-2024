package tcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ormanli/form3-te/internal/app/simulator"
)

func Test_Behaviour(t *testing.T) {
	tests := []struct {
		name               string
		prepareMockService func(*MockService)
		run                func(*testing.T, net.Conn)
	}{
		{
			name: "Valid input",
			prepareMockService: func(mockService *MockService) {
				mockService.EXPECT().
					Process(1).
					Return(nil)
			},
			run: func(t *testing.T, conn net.Conn) {
				_, err := conn.Write([]byte("PAYMENT|1\n"))
				require.NoError(t, err)

				out := make([]byte, 1024)

				_, err = conn.Read(out)
				require.NoError(t, err)
				require.Contains(t, string(out), "RESPONSE|ACCEPTED|Transaction processed")
			},
		},
		{
			name:               "Invalid amount",
			prepareMockService: func(mockService *MockService) {},
			run: func(t *testing.T, conn net.Conn) {
				_, err := conn.Write([]byte("PAYMENT|A\n"))
				require.NoError(t, err)

				out := make([]byte, 1024)

				_, err = conn.Read(out)
				require.NoError(t, err)
				require.Contains(t, string(out), "RESPONSE|REJECTED|Invalid amount")
			},
		},
		{
			name:               "Invalid request",
			prepareMockService: func(mockService *MockService) {},
			run: func(t *testing.T, conn net.Conn) {
				_, err := conn.Write([]byte("CHECKOUT|1\n"))
				require.NoError(t, err)

				out := make([]byte, 1024)

				_, err = conn.Read(out)
				require.NoError(t, err)
				require.Contains(t, string(out), "RESPONSE|REJECTED|Invalid request")
			},
		},
		{
			name: "Downstream service failed",
			prepareMockService: func(mockService *MockService) {
				mockService.EXPECT().
					Process(1).
					Return(errors.New("service failure"))
			},
			run: func(t *testing.T, conn net.Conn) {
				_, err := conn.Write([]byte("PAYMENT|1\n"))
				require.NoError(t, err)

				out := make([]byte, 1024)

				_, err = conn.Read(out)
				require.NoError(t, err)
				require.Contains(t, string(out), "RESPONSE|REJECTED|Service failure")
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			mockService := NewMockService(t)
			test.prepareMockService(mockService)

			ctx, cncl := context.WithCancel(context.Background())

			port, err := getFreePort()
			require.NoError(t, err)

			cfg := simulator.Config{
				ServerPort: port,
				ServerHost: "localhost",
			}

			transport := NewTransport(cfg, mockService, clock.New())
			go transport.Start(ctx) //nolint:errcheck

			defer cncl()

			conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
			require.NoError(t, err)
			defer conn.Close() //nolint:errcheck

			test.run(t, conn)
		})
	}
}

func Test_GracefulShutdown(t *testing.T) {
	tests := []struct {
		name               string
		prepareMockService func(*MockService, chan struct{})
		run                func(*testing.T, int, *contextAndCancel, *clock.Mock)
	}{
		{
			name:               "Don't Accept New Connection During Grace Period",
			prepareMockService: func(*MockService, chan struct{}) {},
			run: func(t *testing.T, port int, contextAndCancel *contextAndCancel, mockClock *clock.Mock) {
				contextAndCancel.cncl()

				mockClock.Add(time.Second)

				conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
				assert.ErrorContains(t, err, "connect: connection refused")
				assert.Nil(t, conn)
			},
		},
		{
			name: "Accept Request From Existing Connection During Grace Period",
			prepareMockService: func(mockService *MockService, _ chan struct{}) {
				mockService.EXPECT().
					Process(1).
					Return(nil)

				mockService.EXPECT().
					Process(2).
					Return(nil)
			},
			run: func(t *testing.T, port int, contextAndCancel *contextAndCancel, mockClock *clock.Mock) {
				conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port)) // TODO add second connection
				require.NoError(t, err)
				defer conn.Close() //nolint:errcheck

				_, err = conn.Write([]byte("PAYMENT|1\n"))
				require.NoError(t, err)

				firstResponse := make([]byte, 1024)
				_, err = conn.Read(firstResponse)
				require.NoError(t, err)
				require.Contains(t, string(firstResponse), "RESPONSE|ACCEPTED|Transaction processed")

				contextAndCancel.cncl()

				mockClock.Add(time.Second)

				_, err = conn.Write([]byte("PAYMENT|2\n"))
				require.NoError(t, err)

				secondResponse := make([]byte, 1024)
				_, err = conn.Read(secondResponse)
				require.NoError(t, err)
				require.Contains(t, string(secondResponse), "RESPONSE|ACCEPTED|Transaction processed")
			},
		},
		{
			name: "Request Not Processed During Grace Period",
			prepareMockService: func(mockService *MockService, c chan struct{}) {
				mockService.EXPECT().
					Process(1).
					RunAndReturn(func(int) error {
						<-c
						return nil
					})
			},
			run: func(t *testing.T, port int, contextAndCancel *contextAndCancel, mockClock *clock.Mock) {
				conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
				require.NoError(t, err)
				defer conn.Close() //nolint:errcheck

				_, err = conn.Write([]byte("PAYMENT|1\n"))
				require.NoError(t, err)

				contextAndCancel.cncl()

				for i := 0; i < 10; i++ {
					mockClock.Add(time.Second)
				}

				response := make([]byte, 1024)
				_, err = conn.Read(response)
				require.NoError(t, err)
				require.Contains(t, string(response), "RESPONSE|REJECTED|Cancelled")
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			startCtx, startCncl := context.WithCancel(context.Background())
			c := &contextAndCancel{
				ctx:  startCtx,
				cncl: startCncl,
			}

			stop := make(chan struct{})

			mockService := NewMockService(t)
			test.prepareMockService(mockService, stop)

			port, err := getFreePort()
			require.NoError(t, err)

			cfg := simulator.Config{
				ServerPort:                    port,
				ServerHost:                    "localhost",
				ServerGracefulShutdownTimeout: time.Second,
			}

			mockClock := clock.NewMock()

			transport := NewTransport(cfg, mockService, mockClock)
			go transport.Start(startCtx) //nolint:errcheck

			test.run(t, port, c, mockClock)

			mockClock.WaitForAllTimers()
			close(stop)
		})
	}
}

var (
	freePortMu     sync.Mutex
	allocatedPorts = make(map[int]struct{})
)

// getFreePort returns a free port number.
func getFreePort() (int, error) {
	freePortMu.Lock()
	defer freePortMu.Unlock()

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close() //nolint:errcheck

	port := l.Addr().(*net.TCPAddr).Port //nolint:forcetypeassert

	if _, exists := allocatedPorts[port]; exists {
		return getFreePort()
	}

	allocatedPorts[port] = struct{}{}

	return port, nil
}

type contextAndCancel struct {
	ctx  context.Context
	cncl context.CancelFunc
}

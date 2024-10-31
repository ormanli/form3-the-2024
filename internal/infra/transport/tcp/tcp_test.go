package tcp

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ormanli/form3-te/internal/app/simulator"
)

func Test_Behaviour(t *testing.T) {
	tests := []struct {
		name               string
		prepareMockService func(*MockService)
		cfg                simulator.Config
		flow               func(*testing.T, net.Conn)
	}{
		{
			name: "Valid input",
			prepareMockService: func(mockService *MockService) {
				mockService.EXPECT().
					Process(1).
					Return(nil)
			},
			cfg: simulator.Config{
				ServerHost:                    "localhost",
				ServerGracefulShutdownTimeout: time.Second,
			},
			flow: func(t *testing.T, conn net.Conn) {
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
			cfg: simulator.Config{
				ServerHost:                    "localhost",
				ServerGracefulShutdownTimeout: time.Second,
			},
			flow: func(t *testing.T, conn net.Conn) {
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
			cfg: simulator.Config{
				ServerHost:                    "localhost",
				ServerGracefulShutdownTimeout: time.Second,
			},
			flow: func(t *testing.T, conn net.Conn) {
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
			cfg: simulator.Config{
				ServerHost:                    "localhost",
				ServerGracefulShutdownTimeout: time.Second,
			},
			flow: func(t *testing.T, conn net.Conn) {
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
			mockService := NewMockService(t)
			test.prepareMockService(mockService)

			ctx, cncl := context.WithCancel(context.Background())

			transport := NewTransport(test.cfg, mockService)
			go transport.Stop(ctx)
			go transport.Start(ctx)

			defer cncl()

			require.Eventually(t, func() bool {
				return transport.listener != nil
			}, time.Second, time.Millisecond)

			conn, err := net.Dial(transport.listener.Addr().Network(), transport.listener.Addr().String())
			require.NoError(t, err)
			defer conn.Close()

			test.flow(t, conn)
		})
	}
}

func Test_GracefulShutdown_DontAcceptNewConnection(t *testing.T) {
	mockService := NewMockService(t)

	cfg := simulator.Config{
		ServerHost:                    "localhost",
		ServerGracefulShutdownTimeout: time.Second,
	}

	ctx, cncl := context.WithCancel(context.Background())

	transport := NewTransport(cfg, mockService)
	go transport.Stop(ctx)
	go transport.Start(ctx)

	require.Eventually(t, func() bool {
		return transport.listener != nil
	}, time.Second, time.Millisecond)

	cncl()

	conn, err := net.Dial(transport.listener.Addr().Network(), transport.listener.Addr().String())
	require.ErrorContains(t, err, "connect: connection refused")
	require.Nil(t, conn)
}

func Test_GracefulShutdown_AcceptRequestFromExistingConnection(t *testing.T) {
	mockService := NewMockService(t)
	mockService.EXPECT().
		Process(1).
		Return(nil)

	mockService.EXPECT().
		Process(2).
		Return(nil)

	cfg := simulator.Config{
		ServerHost:                    "localhost",
		ServerGracefulShutdownTimeout: time.Second,
	}

	startCtx, startCncl := context.WithCancel(context.Background())
	stopCtx, stopCncl := context.WithCancel(context.Background())

	transport := NewTransport(cfg, mockService)
	go transport.Stop(stopCtx)
	go transport.Start(startCtx)

	require.Eventually(t, func() bool {
		return transport.listener != nil
	}, time.Second, time.Millisecond)

	conn, err := net.Dial(transport.listener.Addr().Network(), transport.listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	_, err = conn.Write([]byte("PAYMENT|1\n"))
	require.NoError(t, err)

	out := make([]byte, 1024)

	_, err = conn.Read(out)
	require.NoError(t, err)
	require.Contains(t, string(out), "RESPONSE|ACCEPTED|Transaction processed")

	startCncl()

	_, err = conn.Write([]byte("PAYMENT|2\n"))
	require.NoError(t, err)

	_, err = conn.Read(out) // todo new slice
	require.NoError(t, err)
	require.Contains(t, string(out), "RESPONSE|ACCEPTED|Transaction processed")

	stopCncl()
}

func Test_GracefulShutdown_RequestNotProcessed(t *testing.T) {
	mockService := NewMockService(t)
	mockService.EXPECT().
		Process(1).
		After(time.Second).
		Return(nil)

	cfg := simulator.Config{
		ServerHost:                    "localhost",
		ServerGracefulShutdownTimeout: time.Second,
	}

	ctx, cncl := context.WithCancel(context.Background())
	stopCtx, stopCncl := context.WithCancel(context.Background())

	transport := NewTransport(cfg, mockService)
	go transport.Stop(stopCtx)
	go transport.Start(ctx)

	require.Eventually(t, func() bool {
		return transport.listener != nil
	}, time.Second, time.Millisecond)

	conn, err := net.Dial(transport.listener.Addr().Network(), transport.listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	_, err = conn.Write([]byte("PAYMENT|1\n"))
	require.NoError(t, err)

	cncl()

	stopCncl()

	out := make([]byte, 1024)

	_, err = conn.Read(out)
	require.NoError(t, err)
	require.Contains(t, string(out), "RESPONSE|REJECTED|Cancelled")
}

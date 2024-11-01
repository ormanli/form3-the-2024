package internal

import (
	"context"

	"github.com/benbjohnson/clock"

	"github.com/ormanli/form3-te/internal/app/simulator"
	"github.com/ormanli/form3-te/internal/infra/logging"
	"github.com/ormanli/form3-te/internal/infra/transport/tcp"
)

// Run starts application with the passed configuration.
func Run(ctx context.Context, cfg simulator.Config) error {
	logging.Setup(cfg)

	tcpTransport := tcp.NewTransport(cfg, simulator.NewValidationService(simulator.NewDummyService(cfg)), clock.New())

	return tcpTransport.Start(ctx)
}

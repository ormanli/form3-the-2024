package logging

import (
	"log/slog"

	"github.com/ormanli/form3-te/internal/app/simulator"
)

// Setup setups logger configuration.
func Setup(cfg simulator.Config) {
	if cfg.InitDebug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Initializing debug level logging")
	}
}

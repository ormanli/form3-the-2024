package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"

	"github.com/ormanli/form3-te/internal"
	"github.com/ormanli/form3-te/internal/app/simulator"
)

func main() {
	code := 0
	defer func() {
		os.Exit(code)
	}()

	ctx, cncl := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cncl()

	var c simulator.Config

	err := envconfig.Process("app", &c)
	if err != nil {
		slog.Error("Can't process configuration", "error", err.Error())
		code = 1
		return
	}

	err = internal.Run(ctx, c)
	if err != nil {
		slog.Error("Run failed", "error", err.Error())
		code = 1
		return
	}
}

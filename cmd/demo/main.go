package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"github.com/zerok/gograz-demo-otel/internal/backend"
	"github.com/zerok/gograz-demo-otel/internal/frontend"
)

var logger zerolog.Logger

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	ctx = logger.WithContext(ctx)
	app := cli.App{
		Commands: []*cli.Command{
			{
				Name: "frontend",
				Action: func(c *cli.Context) error {
					fe := frontend.New()
					return fe.ListenAndServe(c.Context)
				},
			},
			{
				Name: "backend",
				Action: func(c *cli.Context) error {
					be := backend.New()
					return be.ListenAndServe(c.Context)
				},
			},
		},
	}
	if err := app.RunContext(ctx, os.Args); err != nil {
		logger.Error().Err(err).Msg("command failed")
		os.Exit(1)
	}
}

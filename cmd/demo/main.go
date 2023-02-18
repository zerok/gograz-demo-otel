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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var logger zerolog.Logger

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	ctx = logger.WithContext(ctx)
	tp, err := setupTracing(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to start tracing")
	}
	defer tp.Shutdown(context.Background())
	app := cli.App{
		Commands: []*cli.Command{
			{
				Name: "frontend",
				Action: func(c *cli.Context) error {
					fe := frontend.New(tp)
					return fe.ListenAndServe(c.Context)
				},
			},
			{
				Name: "backend",
				Action: func(c *cli.Context) error {
					be := backend.New(tp)
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

func setupTracing(ctx context.Context) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, err
	}
	provider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return provider, nil
}

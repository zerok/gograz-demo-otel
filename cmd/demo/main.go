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
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	ctx = logger.WithContext(ctx)

	// Setting up tracing. Note that we need to shut down the tracer provider
	// explicitly in order to make sure that all traces are flushed before the
	// application is shut down.
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

func setupTracing(ctx context.Context) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, err
	}
	provider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))

	// The tracer provider can be registered globally so that we don't have to
	// pass it explicitly into the backend and frontend services.
	otel.SetTracerProvider(provider)

	// We also would like traces coming from an external source to be continued
	// if possible. This is what propagation is doing.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return provider, nil
}

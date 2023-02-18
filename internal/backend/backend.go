package backend

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type Backend struct {
	tracer trace.Tracer
}

func New(tp *sdktrace.TracerProvider) *Backend {
	tracer := tp.Tracer("backend")
	return &Backend{
		tracer: tracer,
	}
}

func (fe *Backend) ListenAndServe(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	srv := http.Server{}
	srv.Addr = ":8080"
	srv.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}
	srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))
		ctx, span := fe.tracer.Start(ctx, "backend-handler")
		defer span.End()
		fmt.Fprint(w, "hello from the backend")
		span.SetStatus(codes.Ok, "")
	})
	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()
	logger.Info().Msgf("Listening on %s", srv.Addr)
	return srv.ListenAndServe()
}

package frontend

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type Frontend struct {
	tracer trace.Tracer
}

func New(tp *sdktrace.TracerProvider) *Frontend {
	tracer := tp.Tracer("frontend")
	return &Frontend{
		tracer: tracer,
	}
}

func (fe *Frontend) ListenAndServe(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	srv := http.Server{}
	srv.Addr = ":8080"
	srv.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}
	srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// We have to extract the parent span from the headers. This could be
		// done automatically by a HTTP server that supports this kind of
		// operation explicitly
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))

		ctx, span := fe.tracer.Start(ctx, "frontend-handler")
		defer span.End()

		logger := zerolog.Ctx(ctx)

		// Now, let's do a request to the backend service with a timeout of 3
		// seconds.
		tctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()
		req, _ := http.NewRequestWithContext(tctx, http.MethodGet, "http://backend:8080", nil)
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to retrieve data from backend")
			http.Error(w, "backend request failed", http.StatusInternalServerError)
			span.SetStatus(codes.Error, "failed to request backend data")
			span.RecordError(err)
			return
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to read response from backend")
			http.Error(w, "backend request failed", http.StatusInternalServerError)
			span.SetStatus(codes.Error, "failed to read backend response")
			span.RecordError(err)
			return
		}
		w.Write(data)
		span.SetStatus(codes.Ok, "")
	})
	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()
	logger.Info().Msgf("Listening on %s", srv.Addr)
	return srv.ListenAndServe()
}

package frontend

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type Frontend struct{}

func New() *Frontend {
	return &Frontend{}
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
		logger := zerolog.Ctx(ctx)
		tctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()
		req, _ := http.NewRequestWithContext(tctx, http.MethodGet, "http://backend:8080", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to retrieve data from backend")
			http.Error(w, "backend request failed", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to read response from backend")
			http.Error(w, "backend request failed", http.StatusInternalServerError)
			return
		}
		w.Write(data)
	})
	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()
	logger.Info().Msgf("Listening on %s", srv.Addr)
	return srv.ListenAndServe()
}

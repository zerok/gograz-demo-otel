package backend

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/rs/zerolog"
)

type Backend struct{}

func New() *Backend {
	return &Backend{}
}

func (fe *Backend) ListenAndServe(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	srv := http.Server{}
	srv.Addr = ":8080"
	srv.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}
	srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello from the backend")
	})
	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()
	logger.Info().Msgf("Listening on %s", srv.Addr)
	return srv.ListenAndServe()
}

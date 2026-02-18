package httpentry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/imidll/selfshop.ms-catalog/internal/platform/config"
	"github.com/imidll/selfshop.ms-catalog/internal/platform/logger"
)

func NewRouter(logr *logger.T, conf *config.T) *chi.Mux {
	r := chi.NewMux()
	r.Use(
		middleware.RequestID,
		requestRecover(logr),
		requestTimeout(logr, conf.Entry.HTTP.RequestTimeout),
		requestLogging(logr, conf),
	)
	r.Get("/health/alive", aliveHandler)
	r.Get("/health/ready", readyHandler)
	return r
}

func NewServer(lc fx.Lifecycle, logr *logger.T, conf *config.T, rout *chi.Mux) (*http.Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Entry.HTTP.Port))
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}
	serv := &http.Server{
		Addr:         l.Addr().String(),
		Handler:      rout,
		WriteTimeout: conf.Entry.HTTP.WriteTimeout,
		IdleTimeout:  conf.Entry.HTTP.IdleTimeout,
		ReadTimeout:  conf.Entry.HTTP.ReadTimeout,
	}
	logr = logr.Named("httpentry")
	lc.Append(fx.StartStopHook(
		func(ctx context.Context) error { return startHTTPServer(ctx, serv, logr, l) },
		func(ctx context.Context) error { return graceHTTPServer(ctx, serv, logr) },
	))
	return serv, nil
}

func startHTTPServer(parent context.Context, serv *http.Server, logr *logger.T, l net.Listener) error {
	serv.Addr = l.Addr().String()
	logr.Info("server listening", zap.String("addr", serv.Addr))

	errCh := make(chan error, 1)
	go func() {
		if err := serv.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-parent.Done():
		return fmt.Errorf("start cancelled: %w", parent.Err())
	case err := <-errCh:
		return fmt.Errorf("server failed to start: %w", err)
	case <-time.After(50 * time.Millisecond):
		logr.Info("server started successfully")
		return nil
	}
}

func graceHTTPServer(parent context.Context, serv *http.Server, logr *logger.T) error {
	const gracefulShutdownTimeout = 15 * time.Second
	dead, gracefulCancel := context.WithTimeout(parent, gracefulShutdownTimeout)
	defer gracefulCancel()
	logr.Info("initiating graceful shutdown", zap.Duration("timeout", gracefulShutdownTimeout))

	if err := serv.Shutdown(dead); err != nil {
		logr.Error("graceful shutdown failed", zap.Error(err))
		return fmt.Errorf("shutdown: %w", err)
	}
	logr.Info("server stopped successfully")
	return nil
}

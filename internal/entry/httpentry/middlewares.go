package httpentry

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/imidll/selfshop.ms-catalog/internal/platform/config"
	"github.com/imidll/selfshop.ms-catalog/internal/platform/logger"
)

const slowRequestThreshold = 800 * time.Millisecond

func requestLogging(logr *logger.T, conf *config.T) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if path == "/health/alive" ||
				path == "/health/ready" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			dur := time.Since(start)
			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}

			msg := "request completed"
			level := zap.InfoLevel

			switch {
			case status >= 500:
				msg = "server error"
				level = zap.ErrorLevel
			case status >= 400:
				msg = "client error"
				level = zap.WarnLevel
			case dur > slowRequestThreshold:
				msg = "request too slow"
				level = zap.WarnLevel
			}

			if conf.Debug || status >= 400 ||
				dur > slowRequestThreshold {
				logr.Log(level, msg,
					zap.String("req_id", middleware.GetReqID(r.Context())),
					zap.String("method", r.Method),
					zap.String("path", path),
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("user_agent", r.UserAgent()),
					zap.Duration("duration", dur),
					zap.Int("status", status),
					zap.Int("bytes", ww.BytesWritten()),
				)
			}
		})
	}
}

func requestTimeout(logr *logger.T, t time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), t)
			defer cancel()

			r = r.WithContext(ctx)
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(ww, r)
				close(done)
			}()

			select {
			case <-done:
				// запрос завершился вовремя
			case <-ctx.Done():
				ww.Header().Set("Content-Type", "application/json")
				ww.WriteHeader(http.StatusRequestTimeout) // 408
				_, _ = ww.Write([]byte(`{"error":"request timeout"}`))
				logr.Warn("request timeout",
					zap.String("req_id", middleware.GetReqID(r.Context())),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("remote_addr", r.RemoteAddr),
					zap.Duration("timeout", t),
					zap.Duration("actual_duration", time.Since(start)),
				)
			}
		})
	}
}

func requestRecover(logr *logger.T) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logr.Error("panic occurred",
						zap.String("req_id", middleware.GetReqID(r.Context())),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.Any("panic", rec),
						zap.Stack("stack"),
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

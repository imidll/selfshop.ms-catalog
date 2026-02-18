package httpentry

import (
	"context"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxtest"

	"github.com/imidll/selfshop.ms-catalog/internal/platform/config"
	"github.com/imidll/selfshop.ms-catalog/internal/platform/logger"
)

func newTestLogger() *logger.T {
	conf := &config.T{
		App: config.App{
			Name:    "test",
			Runmode: config.AppRunmodeDev,
		},
		Log: config.Log{
			MinLevel: config.LogMinLevelDebug,
			Format:   config.LogFormatConsole,
		},
	}
	return logger.MustNew(conf)
}

func newTestConfig(port uint16) *config.T {
	return &config.T{
		App: config.App{
			Name:    "test",
			Runmode: config.AppRunmodeDev,
		},
		Log: config.Log{
			MinLevel: config.LogMinLevelDebug,
			Format:   config.LogFormatConsole,
		},
		Entry: config.Entry{
			HTTP: config.HTTP{
				Port:           port,
				WriteTimeout:   20 * time.Second,
				ReadTimeout:    10 * time.Second,
				IdleTimeout:    90 * time.Second,
				RequestTimeout: 15 * time.Second,
			},
		},
	}
}

// ---------------------------------------------------------------------------
// graceHTTPServer
// ---------------------------------------------------------------------------

func TestGraceHTTPServer_ReturnsErrorWhenShutdownFails(t *testing.T) {
	// Arrange
	logr := newTestLogger()

	handlerReady := make(chan struct{})
	slow := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		var once sync.Once
		once.Do(func() { close(handlerReady) })
		<-r.Context().Done()
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	serv := &http.Server{Addr: ln.Addr().String(), Handler: slow}
	_ = startHTTPServer(context.Background(), serv, logr, ln)

	go func() { _, _ = http.Get("http://" + ln.Addr().String() + "/") }() //nolint:bodyclose
	<-handlerReady

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	err = graceHTTPServer(ctx, serv, logr)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "shutdown")
}

// ---------------------------------------------------------------------------
// startHTTPServer
// ---------------------------------------------------------------------------

func TestStartHTTPServer_StartsSuccessfully(t *testing.T) {
	// Arrange
	logr := newTestLogger()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	serv := &http.Server{Addr: ln.Addr().String()}

	// Act
	err = startHTTPServer(context.Background(), serv, logr, ln)

	// Assert
	require.NoError(t, err)
	t.Cleanup(func() { _ = serv.Shutdown(context.Background()) })
}

func TestStartHTTPServer_ReturnsErrorOnCancelledContext(t *testing.T) {
	// Arrange
	logr := newTestLogger()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	serv := &http.Server{Addr: ln.Addr().String()}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	err = startHTTPServer(ctx, serv, logr, ln)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start cancelled")
}

func TestStartHTTPServer_ReturnsErrorWhenServeFails(t *testing.T) {
	// Arrange
	logr := newTestLogger()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	require.NoError(t, ln.Close())
	serv := &http.Server{Addr: ln.Addr().String()}

	// Act
	err = startHTTPServer(context.Background(), serv, logr, ln)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server failed to start")
}

// ---------------------------------------------------------------------------
// NewServer
// ---------------------------------------------------------------------------

func TestNewServer_StartsAndStops(t *testing.T) {
	// Arrange
	lc := fxtest.NewLifecycle(t)
	logr := newTestLogger()
	conf := newTestConfig(0)
	rout := NewRouter(logr, conf)

	// Act
	serv, err := NewServer(lc, logr, conf, rout)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, serv)
	lc.RequireStart()
	lc.RequireStop()
}

func TestNewServer_ReturnsErrorOnOccupiedPort(t *testing.T) {
	// Arrange
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = occupied.Close() })

	lc := fxtest.NewLifecycle(t)
	logr := newTestLogger()
	conf := newTestConfig(uint16(occupied.Addr().(*net.TCPAddr).Port))
	rout := NewRouter(logr, conf)

	// Act
	serv, err := NewServer(lc, logr, conf, rout)

	// Assert
	require.Error(t, err)
	assert.Nil(t, serv)
	assert.Contains(t, err.Error(), "listen")
}

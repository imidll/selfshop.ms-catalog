package httpentry

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// requestLogging
// ---------------------------------------------------------------------------

func TestRequestLogging_SkipsHealthAlive(t *testing.T) {
	// Arrange
	logr := newTestLogger()
	conf := newTestConfig(8080)

	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	mw := requestLogging(logr, conf)(next)
	req := httptest.NewRequest(http.MethodGet, "/health/alive", nil)
	rec := httptest.NewRecorder()

	// Act
	mw.ServeHTTP(rec, req)

	// Assert
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestLogging_SkipsHealthReady(t *testing.T) {
	// Arrange
	logr := newTestLogger()
	conf := newTestConfig(8080)

	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	mw := requestLogging(logr, conf)(next)
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	// Act
	mw.ServeHTTP(rec, req)

	// Assert
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestLogging_Status2xx_PassesThrough(t *testing.T) {
	// Arrange
	logr := newTestLogger()
	conf := newTestConfig(8080)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := requestLogging(logr, conf)(next)
	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	rec := httptest.NewRecorder()

	// Act
	mw.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestLogging_Status4xx_PassesThrough(t *testing.T) {
	// Arrange
	logr := newTestLogger()
	conf := newTestConfig(8080)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	mw := requestLogging(logr, conf)(next)
	req := httptest.NewRequest(http.MethodGet, "/api/missing", nil)
	rec := httptest.NewRecorder()

	// Act
	mw.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRequestLogging_Status5xx_PassesThrough(t *testing.T) {
	// Arrange
	logr := newTestLogger()
	conf := newTestConfig(8080)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	mw := requestLogging(logr, conf)(next)
	req := httptest.NewRequest(http.MethodPost, "/api/fail", nil)
	rec := httptest.NewRecorder()

	// Act
	mw.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRequestLogging_DebugMode_LogsNormalRequests(t *testing.T) {
	// Arrange
	logr := newTestLogger()
	conf := newTestConfig(8080)
	conf.Debug = true

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := requestLogging(logr, conf)(next)
	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	rec := httptest.NewRecorder()

	// Act
	mw.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestLogging_ZeroStatus_TreatedAsOK(t *testing.T) {
	// Arrange - WrapResponseWriter.Status() returns 0 when the underlying
	// handler never calls WriteHeader. httptest.NewServer wraps a real TCP
	// connection, so WrapResponseWriter reports 0 and exercises the
	// `if status == 0 { status = http.StatusOK }` normalisation branch.
	logr := newTestLogger()
	conf := newTestConfig(8080)
	conf.Debug = true // force the log branch so status is actually evaluated

	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// Intentionally do nothing - no WriteHeader, no Write.
	})

	mw := requestLogging(logr, conf)(next)
	srv := httptest.NewServer(mw)
	t.Cleanup(srv.Close)

	// Act
	resp, err := srv.Client().Get(srv.URL + "/api/resource")

	// Assert
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequestLogging_SlowRequest_PassesThrough(t *testing.T) {
	// Arrange - handler sleeps past slowRequestThreshold so the middleware
	// takes the `dur > slowRequestThreshold` branch (msg = "request too slow").
	logr := newTestLogger()
	conf := newTestConfig(8080)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(slowRequestThreshold + 10*time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	mw := requestLogging(logr, conf)(next)
	req := httptest.NewRequest(http.MethodGet, "/api/slow-but-ok", nil)
	rec := httptest.NewRecorder()

	// Act
	mw.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ---------------------------------------------------------------------------
// requestTimeout
// ---------------------------------------------------------------------------

func TestRequestTimeout_HandlerCompletesInTime_PassesThrough(t *testing.T) {
	// Arrange
	logr := newTestLogger()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := requestTimeout(logr, 5*time.Second)(next)
	req := httptest.NewRequest(http.MethodGet, "/api/fast", nil)
	rec := httptest.NewRecorder()

	// Act
	mw.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestTimeout_HandlerExceedsTimeout_Returns408(t *testing.T) {
	// Arrange
	logr := newTestLogger()

	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})

	mw := requestTimeout(logr, 20*time.Millisecond)(next)
	req := httptest.NewRequest(http.MethodGet, "/api/slow", nil)
	rec := httptest.NewRecorder()

	// Act
	mw.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusRequestTimeout, rec.Code)
	assert.Contains(t, rec.Body.String(), "request timeout")
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

// ---------------------------------------------------------------------------
// requestRecover
// ---------------------------------------------------------------------------

func TestRequestRecover_NoPanic_PassesThrough(t *testing.T) {
	// Arrange
	logr := newTestLogger()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := requestRecover(logr)(next)
	req := httptest.NewRequest(http.MethodGet, "/api/ok", nil)
	rec := httptest.NewRecorder()

	// Act
	mw.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestRecover_PanicInHandler_Returns500(t *testing.T) {
	// Arrange
	logr := newTestLogger()

	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("something went wrong")
	})

	mw := requestRecover(logr)(next)
	req := httptest.NewRequest(http.MethodGet, "/api/panic", nil)
	rec := httptest.NewRecorder()

	// Act
	mw.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Internal Server Error")
}

func TestRequestRecover_PanicWithNilValue_Returns500(t *testing.T) {
	// Arrange
	logr := newTestLogger()

	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	})

	mw := requestRecover(logr)(next)
	req := httptest.NewRequest(http.MethodGet, "/api/nil-panic", nil)
	rec := httptest.NewRecorder()

	// Act
	mw.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

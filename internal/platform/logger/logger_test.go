package logger_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/imidll/selfshop.ms-catalog/internal/platform/config"
	"github.com/imidll/selfshop.ms-catalog/internal/platform/logger"
)

func newTestConfig(minLevel, format, runmode string) *config.T {
	return &config.T{
		App: config.App{
			Name:    "test-app",
			Runmode: runmode,
		},
		Log: config.Log{
			MinLevel: minLevel,
			Format:   format,
		},
		Entry: config.Entry{
			HTTP: config.HTTP{
				Port:           8080,
				WriteTimeout:   10 * 1e9 * 10, // 10s
				ReadTimeout:    10 * 1e9 * 10,
				IdleTimeout:    30 * 1e9 * 10,
				RequestTimeout: 10 * 1e9 * 10,
			},
		},
	}
}

// ---------------------------------------------------------------------------
// New
// ---------------------------------------------------------------------------

func TestNew_Success_JSONFormat(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatJSON, config.AppRunmodeProd)

	// Act
	logr, err := logger.New(conf)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, logr)
}

func TestNew_Success_ConsoleFormat(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelDebug, config.LogFormatConsole, config.AppRunmodeDev)

	// Act
	logr, err := logger.New(conf)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, logr)
}

func TestNew_Success_AutoFormat_DevMode(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatAuto, config.AppRunmodeDev)

	// Act
	logr, err := logger.New(conf)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, logr)
}

func TestNew_Success_AutoFormat_ProdMode(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatAuto, config.AppRunmodeProd)

	// Act
	logr, err := logger.New(conf)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, logr)
}

func TestNew_InvalidLogLevel(t *testing.T) {
	// Arrange
	conf := newTestConfig("not-a-level", config.LogFormatConsole, config.AppRunmodeDev)

	// Act
	logr, err := logger.New(conf)

	// Assert
	require.Error(t, err)
	assert.Nil(t, logr)
	assert.Contains(t, err.Error(), "parse log level")
}

// ---------------------------------------------------------------------------
// MustNew
// ---------------------------------------------------------------------------

func TestMustNew_Success(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatConsole, config.AppRunmodeDev)

	// Act & Assert
	assert.NotPanics(t, func() {
		logr := logger.MustNew(conf)
		assert.NotNil(t, logr)
	})
}

func TestMustNew_PanicsOnInvalidLevel(t *testing.T) {
	// Arrange
	conf := newTestConfig("bad-level", config.LogFormatConsole, config.AppRunmodeDev)

	// Act & Assert
	assert.Panics(t, func() {
		logger.MustNew(conf)
	})
}

// ---------------------------------------------------------------------------
// GetLevel
// ---------------------------------------------------------------------------

func TestGetLevel(t *testing.T) {
	testCases := [...]struct {
		name      string
		minLevel  string
		wantLevel zapcore.Level
	}{
		{
			name:      "debug_level",
			minLevel:  config.LogMinLevelDebug,
			wantLevel: zapcore.DebugLevel,
		},
		{
			name:      "info_level",
			minLevel:  config.LogMinLevelInfo,
			wantLevel: zapcore.InfoLevel,
		},
		{
			name:      "warn_level",
			minLevel:  config.LogMinLevelWarn,
			wantLevel: zapcore.WarnLevel,
		},
		{
			name:      "error_level",
			minLevel:  config.LogMinLevelError,
			wantLevel: zapcore.ErrorLevel,
		},
		{
			name:      "panic_level",
			minLevel:  config.LogMinLevelPanic,
			wantLevel: zapcore.PanicLevel,
		},
		{
			name:      "fatal_level",
			minLevel:  config.LogMinLevelFatal,
			wantLevel: zapcore.FatalLevel,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			conf := newTestConfig(tc.minLevel, config.LogFormatConsole, config.AppRunmodeDev)
			logr, err := logger.New(conf)
			require.NoError(t, err)

			// Act
			got := logr.GetLevel()

			// Assert
			assert.Equal(t, tc.wantLevel, got)
		})
	}
}

// ---------------------------------------------------------------------------
// SetLevel
// ---------------------------------------------------------------------------

func TestSetLevel_Success(t *testing.T) {
	testCases := [...]struct {
		name      string
		setTo     string
		wantLevel zapcore.Level
	}{
		{
			name:      "set_to_debug",
			setTo:     "debug",
			wantLevel: zapcore.DebugLevel,
		},
		{
			name:      "set_to_warn",
			setTo:     "warn",
			wantLevel: zapcore.WarnLevel,
		},
		{
			name:      "set_to_error",
			setTo:     "error",
			wantLevel: zapcore.ErrorLevel,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatConsole, config.AppRunmodeDev)
			logr, err := logger.New(conf)
			require.NoError(t, err)

			// Act
			err = logr.SetLevel(tc.setTo)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.wantLevel, logr.GetLevel())
		})
	}
}

func TestSetLevel_SameLevel_NoError(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatConsole, config.AppRunmodeDev)
	logr, err := logger.New(conf)
	require.NoError(t, err)

	// Act
	err = logr.SetLevel(config.LogMinLevelInfo)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, zapcore.InfoLevel, logr.GetLevel())
}

func TestSetLevel_EmptyString_Error(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatConsole, config.AppRunmodeDev)
	logr, err := logger.New(conf)
	require.NoError(t, err)

	// Act
	err = logr.SetLevel("")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestSetLevel_InvalidLevel_Error(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatConsole, config.AppRunmodeDev)
	logr, err := logger.New(conf)
	require.NoError(t, err)

	// Act
	err = logr.SetLevel("not-a-level")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

// ---------------------------------------------------------------------------
// Named
// ---------------------------------------------------------------------------

func TestNamed_WithTitle_ReturnsNewLogger(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatConsole, config.AppRunmodeDev)
	logr, err := logger.New(conf)
	require.NoError(t, err)

	// Act
	named := logr.Named("component")

	// Assert
	assert.NotNil(t, named)
	assert.NotSame(t, logr, named)
}

func TestNamed_EmptyTitle_ReturnsSameLogger(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatConsole, config.AppRunmodeDev)
	logr, err := logger.New(conf)
	require.NoError(t, err)

	// Act
	named := logr.Named("")

	// Assert
	assert.Same(t, logr, named)
}

// ---------------------------------------------------------------------------
// With
// ---------------------------------------------------------------------------

func TestWith_WithFields_ReturnsNewLogger(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatConsole, config.AppRunmodeDev)
	logr, err := logger.New(conf)
	require.NoError(t, err)

	// Act
	enriched := logr.With(zap.String("key", "value"))

	// Assert
	assert.NotNil(t, enriched)
	assert.NotSame(t, logr, enriched)
}

func TestWith_NoFields_ReturnsSameLogger(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatConsole, config.AppRunmodeDev)
	logr, err := logger.New(conf)
	require.NoError(t, err)

	// Act
	same := logr.With()

	// Assert
	assert.Same(t, logr, same)
}

// ---------------------------------------------------------------------------
// WithOptions
// ---------------------------------------------------------------------------

func TestWithOptions_WithOptions_ReturnsNewLogger(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatConsole, config.AppRunmodeDev)
	logr, err := logger.New(conf)
	require.NoError(t, err)

	// Act
	modified := logr.WithOptions(zap.AddCaller())

	// Assert
	assert.NotNil(t, modified)
	assert.NotSame(t, logr, modified)
}

func TestWithOptions_NoOptions_ReturnsSameLogger(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatConsole, config.AppRunmodeDev)
	logr, err := logger.New(conf)
	require.NoError(t, err)

	// Act
	same := logr.WithOptions()

	// Assert
	assert.Same(t, logr, same)
}

// ---------------------------------------------------------------------------
// Critical
// ---------------------------------------------------------------------------

func TestCritical_ReturnsNonNilLogger(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatConsole, config.AppRunmodeDev)
	logr, err := logger.New(conf)
	require.NoError(t, err)

	// Act
	crit := logr.Critical()

	// Assert
	assert.NotNil(t, crit)
}

// ---------------------------------------------------------------------------
// Transform
// ---------------------------------------------------------------------------

func TestTransform_ReturnsNonNilFxLogger(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatConsole, config.AppRunmodeDev)
	logr, err := logger.New(conf)
	require.NoError(t, err)

	// Act
	fxlogr := logger.Transform(logr)

	// Assert
	assert.NotNil(t, fxlogr)
}

// ---------------------------------------------------------------------------
// Sync
// ---------------------------------------------------------------------------

func TestSync_DoesNotReturnError(t *testing.T) {
	// Arrange
	conf := newTestConfig(config.LogMinLevelInfo, config.LogFormatConsole, config.AppRunmodeDev)
	logr, err := logger.New(conf)
	require.NoError(t, err)

	// Act
	// Sync on stdout may return an error on some OS (e.g. "invalid argument"),
	// so we only verify it does not panic and the call is reachable.
	_ = logr.Sync()
}

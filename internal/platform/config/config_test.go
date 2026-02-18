package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/imidll/selfshop.ms-catalog/internal/platform/config"
)

func newTestDevDefaultValues() map[string]any {
	return map[string]any{
		"app.name":    "test-app",
		"app.runmode": config.AppRunmodeDev,

		"log.min_level": config.LogMinLevelInfo,
		"log.format":    config.LogFormatConsole,

		"entry.http.port":            uint16(8080),
		"entry.http.write_timeout":   15 * time.Second,
		"entry.http.read_timeout":    10 * time.Second,
		"entry.http.idle_timeout":    60 * time.Second,
		"entry.http.request_timeout": 30 * time.Second,

		"debug": false,
	}
}

// ---------------------------------------------------------------------------
// New
// ---------------------------------------------------------------------------

func TestNew_Success(t *testing.T) {
	// Arrange
	defaults := newTestDevDefaultValues()

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, conf)
}

func TestNew_EnvOverridesDefaultValues(t *testing.T) {
	// Arrange
	t.Setenv("INIT__APP__RUNMODE", config.AppRunmodeDev)
	t.Setenv("INIT__LOG__MIN_LEVEL", config.LogMinLevelWarn)
	defaults := newTestDevDefaultValues()

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, config.LogMinLevelWarn, conf.Log.MinLevel)
}

func TestNew_UnmarshalError(t *testing.T) {
	// Arrange
	t.Setenv("INIT__ENTRY__HTTP__PORT", "not-a-number")
	defaults := newTestDevDefaultValues()

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.Error(t, err)
	assert.Nil(t, conf)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestNew_ValidationError_MissingRequiredField(t *testing.T) {
	// Arrange - omit app.name so validation fails
	defaults := newTestDevDefaultValues()
	delete(defaults, "app.name")

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.Error(t, err)
	assert.Nil(t, conf)
	assert.Contains(t, err.Error(), "validate")
}

func TestNew_NilDefaultValues_ValidationError(t *testing.T) {
	// Arrange
	// Passing nil means no default values - required fields will be absent.

	// Act
	conf, err := config.New(nil)

	// Assert
	require.Error(t, err)
	assert.Nil(t, conf)
}

func TestNew_EmptyDefaultValues_ValidationError(t *testing.T) {
	// Arrange
	defaults := map[string]any{}

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.Error(t, err)
	assert.Nil(t, conf)
}

// ---------------------------------------------------------------------------
// MustNew
// ---------------------------------------------------------------------------

func TestMustNew_Success_DoesNotPanic(t *testing.T) {
	// Arrange
	defaults := newTestDevDefaultValues()

	// Act & Assert
	assert.NotPanics(t, func() {
		conf := config.MustNew(defaults)
		assert.NotNil(t, conf)
	})
}

func TestMustNew_PanicsOnValidationError(t *testing.T) {
	// Arrange - missing required fields
	defaults := map[string]any{}

	// Act & Assert
	assert.Panics(t, func() {
		config.MustNew(defaults)
	})
}

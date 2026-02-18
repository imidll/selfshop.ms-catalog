package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/imidll/selfshop.ms-catalog/internal/platform/config"
)

func newTestProdDefaultValues() map[string]any {
	return map[string]any{
		"app.name":    "test-app",
		"app.runmode": config.AppRunmodeProd,

		"log.min_level": config.LogMinLevelInfo,
		"log.format":    config.LogFormatJSON,

		"entry.http.port":            uint16(8080),
		"entry.http.write_timeout":   15 * time.Second,
		"entry.http.read_timeout":    10 * time.Second,
		"entry.http.idle_timeout":    60 * time.Second,
		"entry.http.request_timeout": 30 * time.Second,

		"debug": false,
	}
}

// ---------------------------------------------------------------------------
// struct validation (required / oneof / range)
// ---------------------------------------------------------------------------

func TestValidate_InvalidAppRunmode(t *testing.T) {
	// Arrange
	defaults := newTestDevDefaultValues()
	defaults["app.runmode"] = "staging"

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.Error(t, err)
	assert.Nil(t, conf)
	assert.Contains(t, err.Error(), "validate")
}

func TestValidate_InvalidLogMinLevel(t *testing.T) {
	// Arrange
	defaults := newTestDevDefaultValues()
	defaults["log.min_level"] = "trace"

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.Error(t, err)
	assert.Nil(t, conf)
	assert.Contains(t, err.Error(), "validate")
}

func TestValidate_InvalidLogFormat(t *testing.T) {
	// Arrange
	defaults := newTestDevDefaultValues()
	defaults["log.format"] = "pretty"

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.Error(t, err)
	assert.Nil(t, conf)
	assert.Contains(t, err.Error(), "validate")
}

func TestValidate_PortBelowMin(t *testing.T) {
	// Arrange
	defaults := newTestDevDefaultValues()
	defaults["entry.http.port"] = uint16(80)

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.Error(t, err)
	assert.Nil(t, conf)
	assert.Contains(t, err.Error(), "validate")
}

func TestValidate_WriteLessThanRead_Timeout(t *testing.T) {
	// Arrange - write_timeout < read_timeout violates gtefield=ReadTimeout
	defaults := newTestDevDefaultValues()
	defaults["entry.http.read_timeout"] = 20 * time.Second
	defaults["entry.http.write_timeout"] = 10 * time.Second

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.Error(t, err)
	assert.Nil(t, conf)
	assert.Contains(t, err.Error(), "validate")
}

func TestValidate_IdleLessThanRequest_Timeout(t *testing.T) {
	// Arrange - idle_timeout < request_timeout violates gtefield=RequestTimeout
	defaults := newTestDevDefaultValues()
	defaults["entry.http.request_timeout"] = 60 * time.Second
	defaults["entry.http.idle_timeout"] = 30 * time.Second

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.Error(t, err)
	assert.Nil(t, conf)
	assert.Contains(t, err.Error(), "validate")
}

// ---------------------------------------------------------------------------
// prod-specific business rules
// ---------------------------------------------------------------------------

func TestValidate_Prod_DebugTrue_Error(t *testing.T) {
	// Arrange
	defaults := newTestProdDefaultValues()
	defaults["debug"] = true

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.Error(t, err)
	assert.Nil(t, conf)
	assert.Contains(t, err.Error(), "debug must be false in prod mode")
}

func TestValidate_Prod_LogLevelDebug_Error(t *testing.T) {
	// Arrange
	defaults := newTestProdDefaultValues()
	defaults["log.min_level"] = config.LogMinLevelDebug

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.Error(t, err)
	assert.Nil(t, conf)
	assert.Contains(t, err.Error(), "log min level must be info or higher in prod")
}

func TestValidate_Prod_LogFormatConsole_Error(t *testing.T) {
	// Arrange
	defaults := newTestProdDefaultValues()
	defaults["log.format"] = config.LogFormatConsole

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.Error(t, err)
	assert.Nil(t, conf)
	assert.Contains(t, err.Error(), "log format must be json in prod")
}

func TestValidate_Prod_MultipleViolations_AllReported(t *testing.T) {
	// Arrange - trigger all three prod-specific rules at once
	defaults := newTestProdDefaultValues()
	defaults["debug"] = true
	defaults["log.min_level"] = config.LogMinLevelDebug
	defaults["log.format"] = config.LogFormatConsole

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.Error(t, err)
	assert.Nil(t, conf)
	assert.Contains(t, err.Error(), "debug must be false in prod mode")
	assert.Contains(t, err.Error(), "log min level must be info or higher in prod")
	assert.Contains(t, err.Error(), "log format must be json in prod")
}

func TestValidate_Prod_ValidConfig_NoError(t *testing.T) {
	// Arrange
	defaults := newTestProdDefaultValues()

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, conf)
}

func TestValidate_Dev_DebugTrue_NoError(t *testing.T) {
	// Arrange - prod rules must NOT apply in dev mode
	defaults := newTestDevDefaultValues()
	defaults["debug"] = true
	defaults["log.min_level"] = config.LogMinLevelDebug
	defaults["log.format"] = config.LogFormatConsole

	// Act
	conf, err := config.New(defaults)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, conf)
}

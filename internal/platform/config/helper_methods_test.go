package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/imidll/selfshop.ms-catalog/internal/platform/config"
)

// ---------------------------------------------------------------------------
// IsDevelopment
// ---------------------------------------------------------------------------

func TestIsDevelopment_DevRunmode_ReturnsTrue(t *testing.T) {
	// Arrange
	conf := &config.T{App: config.App{Runmode: config.AppRunmodeDev}}

	// Act
	got := conf.IsDevelopment()

	// Assert
	assert.True(t, got)
}

func TestIsDevelopment_ProdRunmode_ReturnsFalse(t *testing.T) {
	// Arrange
	conf := &config.T{App: config.App{Runmode: config.AppRunmodeProd}}

	// Act
	got := conf.IsDevelopment()

	// Assert
	assert.False(t, got)
}

// ---------------------------------------------------------------------------
// IsProduction
// ---------------------------------------------------------------------------

func TestIsProduction_ProdRunmode_ReturnsTrue(t *testing.T) {
	// Arrange
	conf := &config.T{App: config.App{Runmode: config.AppRunmodeProd}}

	// Act
	got := conf.IsProduction()

	// Assert
	assert.True(t, got)
}

func TestIsProduction_DevRunmode_ReturnsFalse(t *testing.T) {
	// Arrange
	conf := &config.T{App: config.App{Runmode: config.AppRunmodeDev}}

	// Act
	got := conf.IsProduction()

	// Assert
	assert.False(t, got)
}

// ---------------------------------------------------------------------------
// GetLogFormat
// ---------------------------------------------------------------------------

func TestGetLogFormat(t *testing.T) {
	testCases := [...]struct {
		name       string
		format     string
		runmode    string
		wantFormat string
	}{
		{
			name:       "json_format_explicit",
			format:     config.LogFormatJSON,
			runmode:    config.AppRunmodeDev,
			wantFormat: config.LogFormatJSON,
		},
		{
			name:       "console_format_explicit",
			format:     config.LogFormatConsole,
			runmode:    config.AppRunmodeProd,
			wantFormat: config.LogFormatConsole,
		},
		{
			name:       "auto_format_in_prod_resolves_to_json",
			format:     config.LogFormatAuto,
			runmode:    config.AppRunmodeProd,
			wantFormat: config.LogFormatJSON,
		},
		{
			name:       "auto_format_in_dev_resolves_to_console",
			format:     config.LogFormatAuto,
			runmode:    config.AppRunmodeDev,
			wantFormat: config.LogFormatConsole,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			conf := &config.T{
				App: config.App{Runmode: tc.runmode},
				Log: config.Log{Format: tc.format},
			}

			// Act
			got := conf.GetLogFormat()

			// Assert
			assert.Equal(t, tc.wantFormat, got)
		})
	}
}

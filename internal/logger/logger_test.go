package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "with config",
			config: &Config{
				Level:  "debug",
				Format: "json",
				Output: "stdout",
			},
		},
		{
			name:   "with nil config",
			config: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.config)
			if logger == nil {
				t.Error("Expected logger to be non-nil")
			}
			if logger.Logger == nil {
				t.Error("Expected underlying logrus.Logger to be non-nil")
			}
		})
	}
}

func TestNewDefault(t *testing.T) {
	logger := NewDefault()
	if logger == nil {
		t.Error("Expected logger to be non-nil")
	}

	// Check that default values are applied
	if logger.Logger.Level != logrus.InfoLevel {
		t.Errorf("Expected log level to be Info, got %v", logger.Logger.Level)
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name          string
		configLevel   string
		expectedLevel logrus.Level
	}{
		{
			name:          "debug level",
			configLevel:   "debug",
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "info level",
			configLevel:   "info",
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "warn level",
			configLevel:   "warn",
			expectedLevel: logrus.WarnLevel,
		},
		{
			name:          "error level",
			configLevel:   "error",
			expectedLevel: logrus.ErrorLevel,
		},
		{
			name:          "invalid level falls back to info",
			configLevel:   "invalid",
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "empty level falls back to info",
			configLevel:   "",
			expectedLevel: logrus.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Level:  tt.configLevel,
				Format: "text",
				Output: "stdout",
			}
			logger := New(config)

			if logger.Logger.Level != tt.expectedLevel {
				t.Errorf("Expected log level %v, got %v", tt.expectedLevel, logger.Logger.Level)
			}
		})
	}
}

func TestJSONFormatter(t *testing.T) {
	var buf bytes.Buffer

	config := &Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	logger := New(config)
	logger.Logger.SetOutput(&buf)

	logger.Info("test message")

	output := buf.String()
	if output == "" {
		t.Fatal("Expected output, got empty string")
	}

	// Parse JSON to verify it's valid
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}

	// Check expected fields
	if logEntry["message"] != "test message" {
		t.Errorf("Expected message 'test message', got %v", logEntry["message"])
	}
	if logEntry["level"] != "info" {
		t.Errorf("Expected level 'info', got %v", logEntry["level"])
	}
	if _, exists := logEntry["timestamp"]; !exists {
		t.Error("Expected timestamp field to exist")
	}
}

func TestTextFormatter(t *testing.T) {
	var buf bytes.Buffer

	config := &Config{
		Level:  "info",
		Format: "text", // or any non-json value
		Output: "stdout",
	}
	logger := New(config)
	logger.Logger.SetOutput(&buf)

	logger.Info("test message")

	output := buf.String()
	if output == "" {
		t.Fatal("Expected output, got empty string")
	}

	// Text format should contain the message
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected output to contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, "level=info") {
		t.Errorf("Expected output to contain 'level=info', got: %s", output)
	}
}

func TestWithField(t *testing.T) {
	var buf bytes.Buffer

	config := &Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	logger := New(config)
	logger.Logger.SetOutput(&buf)

	entry := logger.WithField("key", "value")
	entry.Info("test message")

	output := buf.String()
	if output == "" {
		t.Fatal("Expected output, got empty string")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}

	if logEntry["key"] != "value" {
		t.Errorf("Expected field 'key' to be 'value', got %v", logEntry["key"])
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer

	config := &Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	logger := New(config)
	logger.Logger.SetOutput(&buf)

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	entry := logger.WithFields(fields)
	entry.Info("test message")

	output := buf.String()
	if output == "" {
		t.Fatal("Expected output, got empty string")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}

	if logEntry["key1"] != "value1" {
		t.Errorf("Expected field 'key1' to be 'value1', got %v", logEntry["key1"])
	}
	if logEntry["key2"] != float64(42) { // JSON numbers are float64
		t.Errorf("Expected field 'key2' to be 42, got %v", logEntry["key2"])
	}
	if logEntry["key3"] != true {
		t.Errorf("Expected field 'key3' to be true, got %v", logEntry["key3"])
	}
}

func TestWithPlugin(t *testing.T) {
	var buf bytes.Buffer

	config := &Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	logger := New(config)
	logger.Logger.SetOutput(&buf)

	entry := logger.WithPlugin("test-plugin")
	entry.Info("test message")

	output := buf.String()
	if output == "" {
		t.Fatal("Expected output, got empty string")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}

	if logEntry["plugin"] != "test-plugin" {
		t.Errorf("Expected field 'plugin' to be 'test-plugin', got %v", logEntry["plugin"])
	}
}

func TestWithComponent(t *testing.T) {
	var buf bytes.Buffer

	config := &Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	logger := New(config)
	logger.Logger.SetOutput(&buf)

	entry := logger.WithComponent("test-component")
	entry.Info("test message")

	output := buf.String()
	if output == "" {
		t.Fatal("Expected output, got empty string")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}

	if logEntry["component"] != "test-component" {
		t.Errorf("Expected field 'component' to be 'test-component', got %v", logEntry["component"])
	}
}

func TestOutputConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{
			name:   "stdout output",
			output: "stdout",
		},
		{
			name:   "stderr output",
			output: "stderr",
		},
		{
			name:   "default output (empty)",
			output: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Level:  "info",
				Format: "text",
				Output: tt.output,
			}
			logger := New(config)

			// Just verify the logger was created successfully
			// The actual output stream testing would require more complex setup
			if logger == nil {
				t.Error("Expected logger to be created successfully")
			}
		})
	}
}

func TestLoggerMethods(t *testing.T) {
	var buf bytes.Buffer

	config := &Config{
		Level:  "debug", // Set to debug to capture all levels
		Format: "json",
		Output: "stdout",
	}
	logger := New(config)
	logger.Logger.SetOutput(&buf)

	// Test different log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 4 {
		t.Errorf("Expected 4 log lines, got %d", len(lines))
	}

	// Verify each line contains the expected message
	expectedMessages := []string{"debug message", "info message", "warn message", "error message"}
	for i, line := range lines {
		if !strings.Contains(line, expectedMessages[i]) {
			t.Errorf("Expected line %d to contain '%s', got: %s", i, expectedMessages[i], line)
		}
	}
}

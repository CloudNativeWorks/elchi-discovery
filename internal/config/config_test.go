package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear environment variables
	clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check default values
	if cfg.DiscoveryInterval != 30 {
		t.Errorf("Expected DiscoveryInterval = 30, got %d", cfg.DiscoveryInterval)
	}
	if cfg.ClusterName != "" {
		t.Errorf("Expected ClusterName = '', got %s", cfg.ClusterName)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Expected Log.Level = 'info', got %s", cfg.Log.Level)
	}
	if cfg.Log.Format != "text" {
		t.Errorf("Expected Log.Format = 'text', got %s", cfg.Log.Format)
	}
	if cfg.Log.Output != "stdout" {
		t.Errorf("Expected Log.Output = 'stdout', got %s", cfg.Log.Output)
	}
	if cfg.Elchi.Token != "" {
		t.Errorf("Expected Elchi.Token = '', got %s", cfg.Elchi.Token)
	}
	if cfg.Elchi.APIEndpoint != "" {
		t.Errorf("Expected Elchi.APIEndpoint = '', got %s", cfg.Elchi.APIEndpoint)
	}
	if cfg.Elchi.InsecureSkipVerify {
		t.Error("Expected Elchi.InsecureSkipVerify = false, got true")
	}
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// Clear environment variables first
	clearEnvVars()

	// Set environment variables
	os.Setenv("DISCOVERY_INTERVAL", "60")
	os.Setenv("CLUSTER_NAME", "test-cluster")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("LOG_OUTPUT", "stderr")
	os.Setenv("ELCHI_TOKEN", "test-token")
	os.Setenv("ELCHI_API_ENDPOINT", "https://api.example.com")
	os.Setenv("ELCHI_INSECURE_SKIP_VERIFY", "true")

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check environment variable values
	if cfg.DiscoveryInterval != 60 {
		t.Errorf("Expected DiscoveryInterval = 60, got %d", cfg.DiscoveryInterval)
	}
	if cfg.ClusterName != "test-cluster" {
		t.Errorf("Expected ClusterName = 'test-cluster', got %s", cfg.ClusterName)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Expected Log.Level = 'debug', got %s", cfg.Log.Level)
	}
	if cfg.Log.Format != "json" {
		t.Errorf("Expected Log.Format = 'json', got %s", cfg.Log.Format)
	}
	if cfg.Log.Output != "stderr" {
		t.Errorf("Expected Log.Output = 'stderr', got %s", cfg.Log.Output)
	}
	if cfg.Elchi.Token != "test-token" {
		t.Errorf("Expected Elchi.Token = 'test-token', got %s", cfg.Elchi.Token)
	}
	if cfg.Elchi.APIEndpoint != "https://api.example.com" {
		t.Errorf("Expected Elchi.APIEndpoint = 'https://api.example.com', got %s", cfg.Elchi.APIEndpoint)
	}
	if !cfg.Elchi.InsecureSkipVerify {
		t.Error("Expected Elchi.InsecureSkipVerify = true, got false")
	}
}

func TestLoad_ConfigFile(t *testing.T) {
	// Clear environment variables
	clearEnvVars()

	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
discovery_interval: 45
cluster_name: file-cluster
log:
  level: warn
  format: json
  output: stderr
elchi:
  token: file-token
  api_endpoint: https://file-api.example.com
  insecure_skip_verify: true
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set config path environment variable
	os.Setenv("ELCHI_CONFIG", configPath)
	defer os.Unsetenv("ELCHI_CONFIG")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check config file values
	if cfg.DiscoveryInterval != 45 {
		t.Errorf("Expected DiscoveryInterval = 45, got %d", cfg.DiscoveryInterval)
	}
	if cfg.ClusterName != "file-cluster" {
		t.Errorf("Expected ClusterName = 'file-cluster', got %s", cfg.ClusterName)
	}
	if cfg.Log.Level != "warn" {
		t.Errorf("Expected Log.Level = 'warn', got %s", cfg.Log.Level)
	}
	if cfg.Elchi.Token != "file-token" {
		t.Errorf("Expected Elchi.Token = 'file-token', got %s", cfg.Elchi.Token)
	}
}

func TestLoad_EnvironmentOverridesFile(t *testing.T) {
	// Clear environment variables
	clearEnvVars()

	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
discovery_interval: 45
cluster_name: file-cluster
log:
  level: warn
elchi:
  token: file-token
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set both file and environment variables
	os.Setenv("ELCHI_CONFIG", configPath)
	os.Setenv("CLUSTER_NAME", "env-cluster")
	os.Setenv("LOG_LEVEL", "debug")

	defer func() {
		os.Unsetenv("ELCHI_CONFIG")
		os.Unsetenv("CLUSTER_NAME")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Environment should override file
	if cfg.ClusterName != "env-cluster" {
		t.Errorf("Expected ClusterName = 'env-cluster' (env override), got %s", cfg.ClusterName)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Expected Log.Level = 'debug' (env override), got %s", cfg.Log.Level)
	}

	// File values should be used where env is not set
	if cfg.DiscoveryInterval != 45 {
		t.Errorf("Expected DiscoveryInterval = 45 (from file), got %d", cfg.DiscoveryInterval)
	}
	if cfg.Elchi.Token != "file-token" {
		t.Errorf("Expected Elchi.Token = 'file-token' (from file), got %s", cfg.Elchi.Token)
	}
}

func TestLoad_InvalidConfigFile(t *testing.T) {
	// Clear environment variables
	clearEnvVars()

	// Create temporary invalid config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidContent := `invalid yaml content [[[`

	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set config path environment variable
	os.Setenv("ELCHI_CONFIG", configPath)
	defer os.Unsetenv("ELCHI_CONFIG")

	_, err = Load()
	if err == nil {
		t.Error("Expected error for invalid YAML file")
	}
}

func TestLoad_NonExistentConfigFile(t *testing.T) {
	// Clear environment variables
	clearEnvVars()

	// Set non-existent config path
	os.Setenv("ELCHI_CONFIG", "/non/existent/config.yaml")
	defer os.Unsetenv("ELCHI_CONFIG")

	_, err := Load()
	if err == nil {
		t.Error("Expected error for non-existent config file")
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "environment variable set",
			envKey:       "TEST_ENV_VAR",
			envValue:     "env-value",
			defaultValue: "default-value",
			expected:     "env-value",
		},
		{
			name:         "environment variable not set",
			envKey:       "NON_EXISTENT_VAR",
			envValue:     "",
			defaultValue: "default-value",
			expected:     "default-value",
		},
		{
			name:         "environment variable set to empty",
			envKey:       "EMPTY_VAR",
			envValue:     "",
			defaultValue: "default-value",
			expected:     "default-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnvOrDefault(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvOrDefault(%s, %s) = %s, expected %s", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvOrDefaultInt(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue int
		expected     int
	}{
		{
			name:         "valid integer",
			envKey:       "TEST_INT_VAR",
			envValue:     "42",
			defaultValue: 10,
			expected:     42,
		},
		{
			name:         "invalid integer",
			envKey:       "TEST_INVALID_INT",
			envValue:     "not-a-number",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "environment variable not set",
			envKey:       "NON_EXISTENT_INT",
			envValue:     "",
			defaultValue: 10,
			expected:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnvOrDefaultInt(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvOrDefaultInt(%s, %d) = %d, expected %d", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvOrDefaultBool(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "true value",
			envKey:       "TEST_BOOL_VAR",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "false value",
			envKey:       "TEST_BOOL_VAR",
			envValue:     "false",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "1 value",
			envKey:       "TEST_BOOL_VAR",
			envValue:     "1",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "0 value",
			envKey:       "TEST_BOOL_VAR",
			envValue:     "0",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "invalid boolean",
			envKey:       "TEST_INVALID_BOOL",
			envValue:     "maybe",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "environment variable not set",
			envKey:       "NON_EXISTENT_BOOL",
			envValue:     "",
			defaultValue: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			} else {
				os.Unsetenv(tt.envKey)
			}

			result := getEnvOrDefaultBool(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvOrDefaultBool(%s, %t) = %t, expected %t", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

// Helper function to clear all relevant environment variables
func clearEnvVars() {
	envVars := []string{
		"DISCOVERY_INTERVAL",
		"CLUSTER_NAME",
		"LOG_LEVEL",
		"LOG_FORMAT",
		"LOG_OUTPUT",
		"ELCHI_TOKEN",
		"ELCHI_API_ENDPOINT",
		"ELCHI_INSECURE_SKIP_VERIFY",
		"ELCHI_CONFIG",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

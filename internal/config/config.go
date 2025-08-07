package config

import (
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type ElchiConfig struct {
	Token              string `yaml:"token"`
	APIEndpoint        string `yaml:"api_endpoint"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

type Config struct {
	Elchi             ElchiConfig `yaml:"elchi"`
	Log               LogConfig   `yaml:"log"`
	DiscoveryInterval int         `yaml:"discovery_interval"`
	ClusterName       string      `yaml:"cluster_name"`
}

func Load() (*Config, error) {
	// Start with defaults
	config := &Config{
		DiscoveryInterval: 30,
		ClusterName:       "",
		Log: LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
		Elchi: ElchiConfig{
			Token:              "",
			APIEndpoint:        "",
			InsecureSkipVerify: false,
		},
	}

	// Load config file if exists (overwrites defaults)
	configPath := getConfigPath()
	if configPath != "" {
		if err := loadConfigFile(config, configPath); err != nil {
			return nil, err
		}
	}

	// Apply environment variables (overwrites file config)
	applyEnvironmentVariables(config)

	return config, nil
}

func applyEnvironmentVariables(config *Config) {
	if val := os.Getenv("DISCOVERY_INTERVAL"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			config.DiscoveryInterval = intVal
		}
	}

	if val := os.Getenv("CLUSTER_NAME"); val != "" {
		config.ClusterName = val
	}

	if val := os.Getenv("LOG_LEVEL"); val != "" {
		config.Log.Level = val
	}

	if val := os.Getenv("LOG_FORMAT"); val != "" {
		config.Log.Format = val
	}

	if val := os.Getenv("LOG_OUTPUT"); val != "" {
		config.Log.Output = val
	}

	if val := os.Getenv("ELCHI_TOKEN"); val != "" {
		config.Elchi.Token = val
	}

	if val := os.Getenv("ELCHI_API_ENDPOINT"); val != "" {
		config.Elchi.APIEndpoint = val
	}

	if val := os.Getenv("ELCHI_INSECURE_SKIP_VERIFY"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			config.Elchi.InsecureSkipVerify = boolVal
		}
	}
}

func getConfigPath() string {
	if path := os.Getenv("ELCHI_CONFIG"); path != "" {
		return path
	}

	if home, err := os.UserHomeDir(); err == nil {
		configPath := filepath.Join(home, ".elchi", "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	if _, err := os.Stat("config.yaml"); err == nil {
		return "config.yaml"
	}

	return ""
}

func loadConfigFile(config *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvOrDefaultBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

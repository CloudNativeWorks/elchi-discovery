package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/CloudNativeWorks/elchi-discovery/api"
	"github.com/CloudNativeWorks/elchi-discovery/discovery"
	"github.com/CloudNativeWorks/elchi-discovery/internal/config"
	elchiContext "github.com/CloudNativeWorks/elchi-discovery/internal/context"
	"github.com/CloudNativeWorks/elchi-discovery/internal/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestRunDiscovery(t *testing.T) {
	// Create test server
	var receivedRequests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequests++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create fake Kubernetes client with test data
	client := fake.NewSimpleClientset()

	// Add test nodes
	testNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Status: v1.NodeStatus{
			NodeInfo: v1.NodeSystemInfo{
				KubeletVersion: "v1.28.2",
			},
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionTrue,
				},
			},
			Addresses: []v1.NodeAddress{
				{
					Type:    v1.NodeInternalIP,
					Address: "192.168.1.10",
				},
			},
		},
	}
	_, err := client.CoreV1().Nodes().Create(context.TODO(), testNode, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test node: %v", err)
	}

	// Create test config
	cfg := &config.Config{
		ClusterName: "test-cluster",
		Elchi: config.ElchiConfig{
			APIEndpoint: server.URL,
			Token:       "96688e4c-6737-4230-9591-6a3332115871--683b2148ff7e3ae67d825cfa",
		},
		Log: config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
	}

	// Create logger
	loggerCfg := &logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	}
	log := logger.New(loggerCfg)

	// Create discovery service
	discoveryService := discovery.NewService(client, cfg.ClusterName)

	// Create API client
	apiClient := api.NewClient(cfg, log)

	// Run discovery with config in context
	ctx := elchiContext.WithConfig(context.Background(), cfg)
	runDiscovery(ctx, log, discoveryService, apiClient)

	// Verify that API was called
	if receivedRequests != 1 {
		t.Errorf("Expected 1 API request, got %d", receivedRequests)
	}
}

func TestRunDiscovery_NoAPIEndpoint(t *testing.T) {
	// Create fake Kubernetes client
	client := fake.NewSimpleClientset()

	// Create test config without API endpoint
	cfg := &config.Config{
		ClusterName: "test-cluster",
		Elchi: config.ElchiConfig{
			APIEndpoint: "", // No endpoint
		},
		Log: config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
	}

	// Create logger
	loggerCfg := &logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	}
	log := logger.New(loggerCfg)

	// Create discovery service
	discoveryService := discovery.NewService(client, cfg.ClusterName)

	// Create API client
	apiClient := api.NewClient(cfg, log)

	// Run discovery (should not fail even without API endpoint)
	ctx := elchiContext.WithConfig(context.Background(), cfg)
	runDiscovery(ctx, log, discoveryService, apiClient)

	// Test passes if no panic or error occurs
}

func TestRunDiscovery_APIFailure(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create fake Kubernetes client
	client := fake.NewSimpleClientset()

	// Create test config
	cfg := &config.Config{
		ClusterName: "test-cluster",
		Elchi: config.ElchiConfig{
			APIEndpoint: server.URL,
		},
		Log: config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
	}

	// Create logger
	loggerCfg := &logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	}
	log := logger.New(loggerCfg)

	// Create discovery service
	discoveryService := discovery.NewService(client, cfg.ClusterName)

	// Create API client
	apiClient := api.NewClient(cfg, log)

	// Run discovery (should not fail even with API error)
	ctx := elchiContext.WithConfig(context.Background(), cfg)
	runDiscovery(ctx, log, discoveryService, apiClient)

	// Test passes if no panic occurs (API failure should be logged but not fatal)
}

func TestGetKubernetesClient_OutsideCluster(t *testing.T) {
	// This test will fail in a real Kubernetes cluster, but should work in test environment

	// Save original environment
	originalKubeconfigPath := os.Getenv("KUBECONFIG")
	originalHome := os.Getenv("HOME")

	// Clear Kubernetes config environment variables
	os.Unsetenv("KUBECONFIG")
	os.Setenv("HOME", "/non-existent-home")

	defer func() {
		// Restore environment
		if originalKubeconfigPath != "" {
			os.Setenv("KUBECONFIG", originalKubeconfigPath)
		}
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()

	_, err := getKubernetesClient()
	if err == nil {
		t.Error("Expected error when running outside cluster without kubeconfig")
	}
}

func TestMainIntegration(t *testing.T) {
	// This is a basic smoke test to ensure main components can be initialized
	// without actually running the full main function

	// Test that we can create all the necessary components
	cfg := &config.Config{
		ClusterName:       "test-cluster",
		DiscoveryInterval: 30,
		Log: config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
		Elchi: config.ElchiConfig{
			APIEndpoint: "https://api.example.com",
			Token:       "96688e4c-6737-4230-9591-6a3332115871--683b2148ff7e3ae67d825cfa",
		},
	}

	// Create logger
	loggerCfg := &logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	}
	log := logger.New(loggerCfg)
	if log == nil {
		t.Fatal("Failed to create logger")
	}

	// Create fake client for testing
	client := fake.NewSimpleClientset()

	// Create discovery service
	discoveryService := discovery.NewService(client, cfg.ClusterName)
	if discoveryService == nil {
		t.Fatal("Failed to create discovery service")
	}

	// Create API client
	apiClient := api.NewClient(cfg, log)
	if apiClient == nil {
		t.Fatal("Failed to create API client")
	}

	// Test that discovery works
	ctx := context.Background()
	result, err := discoveryService.DiscoverNodes(ctx)
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected discovery result to be non-nil")
	}

	if result.ClusterInfo.Name != "test-cluster" {
		t.Errorf("Expected cluster name 'test-cluster', got %s", result.ClusterInfo.Name)
	}

	// Test that we can send to API (will fail but shouldn't panic)
	err = apiClient.SendDiscoveryResult(result)
	if err == nil {
		t.Error("Expected error when sending to fake API endpoint")
	}
}

func TestConfigurationLoading(t *testing.T) {
	// Test configuration loading scenarios

	// Clear relevant environment variables
	envVars := []string{
		"CLUSTER_NAME",
		"DISCOVERY_INTERVAL",
		"LOG_LEVEL",
		"ELCHI_TOKEN",
		"ELCHI_API_ENDPOINT",
	}

	originalVars := make(map[string]string)
	for _, envVar := range envVars {
		originalVars[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar)
	}

	defer func() {
		// Restore environment variables
		for envVar, originalValue := range originalVars {
			if originalValue != "" {
				os.Setenv(envVar, originalValue)
			}
		}
	}()

	// Test loading with no environment variables
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Config loading failed: %v", err)
	}

	// Verify defaults
	if cfg.DiscoveryInterval != 30 {
		t.Errorf("Expected default discovery interval 30, got %d", cfg.DiscoveryInterval)
	}

	if cfg.Log.Level != "info" {
		t.Errorf("Expected default log level 'info', got %s", cfg.Log.Level)
	}

	// Test with environment variables
	os.Setenv("CLUSTER_NAME", "env-test-cluster")
	os.Setenv("DISCOVERY_INTERVAL", "60")
	os.Setenv("LOG_LEVEL", "debug")

	cfg2, err := config.Load()
	if err != nil {
		t.Fatalf("Config loading with env vars failed: %v", err)
	}

	if cfg2.ClusterName != "env-test-cluster" {
		t.Errorf("Expected cluster name from env 'env-test-cluster', got %s", cfg2.ClusterName)
	}

	if cfg2.DiscoveryInterval != 60 {
		t.Errorf("Expected discovery interval from env 60, got %d", cfg2.DiscoveryInterval)
	}

	if cfg2.Log.Level != "debug" {
		t.Errorf("Expected log level from env 'debug', got %s", cfg2.Log.Level)
	}
}

// Benchmark tests
func BenchmarkRunDiscovery(b *testing.B) {
	// Create fake client with multiple nodes
	client := fake.NewSimpleClientset()

	// Add multiple test nodes
	for i := 0; i < 10; i++ {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-" + string(rune(i+'0')),
			},
			Status: v1.NodeStatus{
				NodeInfo: v1.NodeSystemInfo{
					KubeletVersion: "v1.28.2",
				},
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
				Addresses: []v1.NodeAddress{
					{
						Type:    v1.NodeInternalIP,
						Address: "192.168.1." + string(rune(i+10)),
					},
				},
			},
		}
		_, err := client.CoreV1().Nodes().Create(context.TODO(), node, metav1.CreateOptions{})
		if err != nil {
			b.Fatalf("Failed to create test node: %v", err)
		}
	}

	cfg := &config.Config{
		ClusterName: "benchmark-cluster",
		Log: config.LogConfig{
			Level:  "error", // Reduce log noise during benchmark
			Format: "text",
			Output: "stdout",
		},
		Elchi: config.ElchiConfig{
			APIEndpoint: "", // No API calls for benchmark
		},
	}

	loggerCfg := &logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	}
	log := logger.New(loggerCfg)
	discoveryService := discovery.NewService(client, cfg.ClusterName)
	apiClient := api.NewClient(cfg, log)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runDiscovery(ctx, log, discoveryService, apiClient)
	}
}

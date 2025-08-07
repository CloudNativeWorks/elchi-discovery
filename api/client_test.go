package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CloudNativeWorks/elchi-discovery/discovery"
	"github.com/CloudNativeWorks/elchi-discovery/internal/config"
	"github.com/CloudNativeWorks/elchi-discovery/internal/logger"
)

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		Elchi: config.ElchiConfig{
			InsecureSkipVerify: true,
		},
	}
	log := logger.NewDefault()

	client := NewClient(cfg, log)

	if client.config != cfg {
		t.Error("Expected config to be set")
	}
	if client.logger != log {
		t.Error("Expected logger to be set")
	}
	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
}

func TestSendDiscoveryResult_NoEndpoint(t *testing.T) {
	cfg := &config.Config{
		Elchi: config.ElchiConfig{
			APIEndpoint: "", // No endpoint configured
		},
	}
	log := logger.NewDefault()
	client := NewClient(cfg, log)

	result := &discovery.DiscoveryResult{
		Timestamp: time.Now(),
		ClusterInfo: discovery.ClusterInfo{
			Name:    "test-cluster",
			Version: "v1.28.2",
		},
		NodeCount: 1,
		Nodes: []discovery.NodeInfo{
			{
				Name:    "node1",
				Status:  "Ready",
				Version: "v1.28.2",
				Addresses: map[string]string{
					"InternalIP": "192.168.1.10",
				},
			},
		},
		Duration: "100ms",
	}

	err := client.SendDiscoveryResult(result)
	if err != nil {
		t.Errorf("Expected no error when endpoint is not configured, got %v", err)
	}
}

func TestSendDiscoveryResult_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header, got %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		Elchi: config.ElchiConfig{
			APIEndpoint: server.URL,
			Token:       "test-token",
		},
	}
	log := logger.NewDefault()
	client := NewClient(cfg, log)

	result := &discovery.DiscoveryResult{
		Timestamp: time.Now(),
		ClusterInfo: discovery.ClusterInfo{
			Name:    "test-cluster",
			Version: "v1.28.2",
		},
		NodeCount: 1,
		Nodes: []discovery.NodeInfo{
			{
				Name:    "node1",
				Status:  "Ready",
				Version: "v1.28.2",
				Addresses: map[string]string{
					"InternalIP": "192.168.1.10",
				},
			},
		},
		Duration: "100ms",
	}

	err := client.SendDiscoveryResult(result)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSendDiscoveryResult_HTTPError(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &config.Config{
		Elchi: config.ElchiConfig{
			APIEndpoint: server.URL,
		},
	}
	log := logger.NewDefault()
	client := NewClient(cfg, log)

	result := &discovery.DiscoveryResult{
		Timestamp: time.Now(),
		ClusterInfo: discovery.ClusterInfo{
			Name:    "test-cluster",
			Version: "v1.28.2",
		},
		NodeCount: 0,
		Nodes:     []discovery.NodeInfo{},
		Duration:  "100ms",
	}

	err := client.SendDiscoveryResult(result)
	if err == nil {
		t.Error("Expected error for HTTP 500, got nil")
	}
}

func TestSendDiscoveryResult_InvalidURL(t *testing.T) {
	cfg := &config.Config{
		Elchi: config.ElchiConfig{
			APIEndpoint: "invalid-url",
		},
	}
	log := logger.NewDefault()
	client := NewClient(cfg, log)

	result := &discovery.DiscoveryResult{
		Timestamp: time.Now(),
		ClusterInfo: discovery.ClusterInfo{
			Name:    "test-cluster",
			Version: "v1.28.2",
		},
		NodeCount: 0,
		Nodes:     []discovery.NodeInfo{},
		Duration:  "100ms",
	}

	err := client.SendDiscoveryResult(result)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestSendDiscoveryResult_WithoutToken(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Errorf("Expected no Authorization header, got %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		Elchi: config.ElchiConfig{
			APIEndpoint: server.URL,
			Token:       "", // No token
		},
	}
	log := logger.NewDefault()
	client := NewClient(cfg, log)

	result := &discovery.DiscoveryResult{
		Timestamp: time.Now(),
		ClusterInfo: discovery.ClusterInfo{
			Name:    "test-cluster",
			Version: "v1.28.2",
		},
		NodeCount: 0,
		Nodes:     []discovery.NodeInfo{},
		Duration:  "100ms",
	}

	err := client.SendDiscoveryResult(result)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSendDiscoveryResult_ServerUnavailable(t *testing.T) {
	cfg := &config.Config{
		Elchi: config.ElchiConfig{
			APIEndpoint: "http://localhost:12345", // Non-existent server
		},
	}
	log := logger.NewDefault()
	client := NewClient(cfg, log)

	result := &discovery.DiscoveryResult{
		Timestamp: time.Now(),
		ClusterInfo: discovery.ClusterInfo{
			Name:    "test-cluster",
			Version: "v1.28.2",
		},
		NodeCount: 0,
		Nodes:     []discovery.NodeInfo{},
		Duration:  "100ms",
	}

	err := client.SendDiscoveryResult(result)
	if err == nil {
		t.Error("Expected error for unavailable server, got nil")
	}
}

func TestSendDiscoveryResult_InsecureSkipVerify(t *testing.T) {
	cfg := &config.Config{
		Elchi: config.ElchiConfig{
			APIEndpoint:        "https://example.com",
			InsecureSkipVerify: true,
		},
	}
	log := logger.NewDefault()
	client := NewClient(cfg, log)

	// Verify that the client was created with insecure transport
	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}

	// The test just verifies the client is created correctly
	// Actually testing TLS skip would require a more complex setup
}

func TestClient_Timeout(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Small delay for timeout test
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		Elchi: config.ElchiConfig{
			APIEndpoint: server.URL,
		},
	}
	log := logger.NewDefault()
	client := NewClient(cfg, log)
	
	// Set a very short timeout for testing
	client.httpClient.Timeout = 50 * time.Millisecond

	result := &discovery.DiscoveryResult{
		Timestamp:   time.Now(),
		ClusterInfo: discovery.ClusterInfo{Name: "test", Version: "v1.28.2"},
		NodeCount:   0,
		Nodes:       []discovery.NodeInfo{},
		Duration:    "100ms",
	}

	err := client.SendDiscoveryResult(result)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}
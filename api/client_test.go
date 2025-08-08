package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestExtractProjectFromToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "valid token format",
			token:    "96688e4c-6737-4230-9591-6a3332115871--683b2148ff7e3ae67d825cfa",
			expected: "683b2148ff7e3ae67d825cfa",
		},
		{
			name:     "invalid token format - no separator",
			token:    "96688e4c-6737-4230-9591-6a3332115871",
			expected: "",
		},
		{
			name:     "invalid token format - empty",
			token:    "",
			expected: "",
		},
		{
			name:     "token format with multiple separators",
			token:    "uuid--project--extra",
			expected: "project--extra", // Split returns everything after first --
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractProjectFromToken(tt.token)
			if result != tt.expected {
				t.Errorf("extractProjectFromToken(%s) = %s, expected %s", tt.token, result, tt.expected)
			}
		})
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

func TestSendDiscoveryResult_InvalidToken(t *testing.T) {
	cfg := &config.Config{
		Elchi: config.ElchiConfig{
			APIEndpoint: "https://api.example.com",
			Token:       "invalid-token-format", // Missing -- separator
		},
	}
	log := logger.NewDefault()
	client := NewClient(cfg, log)

	result := &discovery.DiscoveryResult{
		Timestamp:   time.Now(),
		ClusterInfo: discovery.ClusterInfo{Name: "test", Version: "v1.28.2"},
		NodeCount:   0,
		Nodes:       []discovery.NodeInfo{},
		Duration:    "100ms",
	}

	err := client.SendDiscoveryResult(result)
	if err == nil {
		t.Error("Expected error for invalid token format")
	}
	if !strings.Contains(err.Error(), "invalid token format") {
		t.Errorf("Expected 'invalid token format' error, got: %v", err)
	}
}

func TestSendDiscoveryResult_Success(t *testing.T) {
	// Create test server
	var receivedPayload DiscoveryPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("from-elchi") != "yes" {
			t.Errorf("Expected from-elchi header 'yes', got %s", r.Header.Get("from-elchi"))
		}
		if r.Header.Get("Authorization") != "Bearer 96688e4c-6737-4230-9591-6a3332115871--683b2148ff7e3ae67d825cfa" {
			t.Errorf("Expected Authorization header, got %s", r.Header.Get("Authorization"))
		}

		// Decode the payload to verify structure
		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Send success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"result":  receivedPayload,
			"message": "Discovery processed successfully",
		})
	}))
	defer server.Close()

	cfg := &config.Config{
		Elchi: config.ElchiConfig{
			APIEndpoint: server.URL,
			Token:       "96688e4c-6737-4230-9591-6a3332115871--683b2148ff7e3ae67d825cfa",
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

	// Verify payload structure
	if receivedPayload.Project != "683b2148ff7e3ae67d825cfa" {
		t.Errorf("Expected project '683b2148ff7e3ae67d825cfa', got %s", receivedPayload.Project)
	}
	if receivedPayload.Data == nil {
		t.Error("Expected data to be present")
	}
	if receivedPayload.Data.ClusterInfo.Name != "test-cluster" {
		t.Errorf("Expected cluster name 'test-cluster', got %s", receivedPayload.Data.ClusterInfo.Name)
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
			Token:       "96688e4c-6737-4230-9591-6a3332115871--683b2148ff7e3ae67d825cfa",
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
			Token:       "96688e4c-6737-4230-9591-6a3332115871--683b2148ff7e3ae67d825cfa",
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
	cfg := &config.Config{
		Elchi: config.ElchiConfig{
			APIEndpoint: "https://api.example.com",
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
	if err == nil {
		t.Error("Expected error for missing token, got nil")
	}
	if !strings.Contains(err.Error(), "invalid token format") {
		t.Errorf("Expected 'invalid token format' error, got: %v", err)
	}
}

func TestSendDiscoveryResult_ServerUnavailable(t *testing.T) {
	cfg := &config.Config{
		Elchi: config.ElchiConfig{
			APIEndpoint: "http://localhost:12345", // Non-existent server
			Token:       "96688e4c-6737-4230-9591-6a3332115871--683b2148ff7e3ae67d825cfa",
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
			Token:              "96688e4c-6737-4230-9591-6a3332115871--683b2148ff7e3ae67d825cfa",
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
			Token:       "96688e4c-6737-4230-9591-6a3332115871--683b2148ff7e3ae67d825cfa",
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

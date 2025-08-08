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

func TestSendDiscoveryResult_APIFailureResponse(t *testing.T) {
	// Create test server that returns failure response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send failure response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"result":  nil,
			"error":   "Invalid cluster configuration",
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
		NodeCount: 0,
		Nodes:     []discovery.NodeInfo{},
		Duration:  "100ms",
	}

	err := client.SendDiscoveryResult(result)
	if err == nil {
		t.Error("Expected error for API failure response, got nil")
	}
	if !strings.Contains(err.Error(), "Invalid cluster configuration") {
		t.Errorf("Expected error to contain 'Invalid cluster configuration', got: %v", err)
	}
}

func TestSendDiscoveryResult_SuccessResponse(t *testing.T) {
	// Create test server that returns success response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"result":  map[string]string{"processed": "yes"},
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
		t.Errorf("Expected no error for successful API response, got %v", err)
	}
}

func TestSendDiscoveryResult_InvalidJSONResponse(t *testing.T) {
	// Create test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json response"))
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
		Timestamp:   time.Now(),
		ClusterInfo: discovery.ClusterInfo{Name: "test", Version: "v1.28.2"},
		NodeCount:   0,
		Nodes:       []discovery.NodeInfo{},
		Duration:    "100ms",
	}

	err := client.SendDiscoveryResult(result)
	// Should not error - invalid JSON is handled gracefully with a warning
	if err != nil {
		t.Errorf("Expected no error for invalid JSON response (should be handled gracefully), got %v", err)
	}
}

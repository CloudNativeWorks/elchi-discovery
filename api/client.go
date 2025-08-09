package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/CloudNativeWorks/elchi-discovery/discovery"
	"github.com/CloudNativeWorks/elchi-discovery/internal/config"
	"github.com/CloudNativeWorks/elchi-discovery/internal/logger"
)

type Client struct {
	httpClient *http.Client
	config     *config.Config
	logger     *logger.Logger
	// initialCompleted is used to send initial:false after success is received
	initialCompleted atomic.Bool
}

// DiscoveryPayload wraps the discovery result with project information
type DiscoveryPayload struct {
	Project string                     `json:"project"`
	Data    *discovery.DiscoveryResult `json:"data"`
}

// APIResponse represents the response from the API
type APIResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
	Message string      `json:"message"`
	Error   string      `json:"error"`
}

// extractProjectFromToken extracts project ID from token format: "uuid--project"
func extractProjectFromToken(token string) string {
	parts := strings.SplitN(token, "--", 2) // Split only on first occurrence
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func NewClient(cfg *config.Config, log *logger.Logger) *Client {
	// Create HTTP client with custom transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.Elchi.InsecureSkipVerify,
		},
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	return &Client{
		httpClient: httpClient,
		config:     cfg,
		logger:     log,
	}
}

func (c *Client) SendDiscoveryResult(result *discovery.DiscoveryResult) error {
	return c.sendDiscoveryResult(result, true)
}

func (c *Client) GetDiscoveryPayload(result *discovery.DiscoveryResult) (*DiscoveryPayload, error) {
	// Extract project ID from token
	projectID := extractProjectFromToken(c.config.Elchi.Token)
	if projectID == "" {
		return nil, fmt.Errorf("invalid token format: expected 'uuid--project' format")
	}

	// Create payload with project information
	return &DiscoveryPayload{
		Project: projectID,
		Data:    result,
	}, nil
}

func (c *Client) sendDiscoveryResult(result *discovery.DiscoveryResult, shouldSend bool) error {
	// Check if API endpoint is configured and shouldSend is true
	if c.config.Elchi.APIEndpoint == "" || !shouldSend {
		c.logger.Debug("No API endpoint configured, skipping send")
		return nil
	}

	// Get payload using shared method
	payload, err := c.GetDiscoveryPayload(result)
	if err != nil {
		return err
	}

	c.logger.Debug("Successfully extracted project from token", map[string]interface{}{
		"project_id": payload.Project,
	})

	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal discovery payload: %w", err)
	}

	// Create JSON preview for logging
	previewLen := 200
	if len(jsonData) < previewLen {
		previewLen = len(jsonData)
	}
	preview := string(jsonData[:previewLen])
	if len(jsonData) > 200 {
		preview += "..."
	}

	c.logger.Debug("Sending discovery payload to API", map[string]interface{}{
		"endpoint":     c.config.Elchi.APIEndpoint,
		"project":      payload.Project,
		"payload_size": len(jsonData),
		"json_preview": preview,
	})

	// Create request
	req, err := http.NewRequest("POST", c.config.Elchi.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("from-elchi", "yes")
	if c.initialCompleted.Load() {
		req.Header.Set("initial", "false")
	} else {
		// Send initial:true until success is received
		req.Header.Set("initial", "true")
	}
	if c.config.Elchi.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Elchi.Token))
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code first
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Try to parse error response
		var apiResponse APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err == nil && apiResponse.Error != "" {
			c.logger.WithFields(map[string]interface{}{
				"status_code": resp.StatusCode,
				"endpoint":    c.config.Elchi.APIEndpoint,
				"project":     payload.Project,
				"error":       apiResponse.Error,
			}).Error("API returned error response")
			return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, apiResponse.Error)
		} else {
			c.logger.WithFields(map[string]interface{}{
				"status_code": resp.StatusCode,
				"endpoint":    c.config.Elchi.APIEndpoint,
				"project":     payload.Project,
			}).Error("API returned non-success HTTP status")
			return fmt.Errorf("API returned non-success status: %d", resp.StatusCode)
		}
	}

	// Parse successful response body
	var apiResponse APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		c.logger.WithFields(map[string]interface{}{
			"status_code": resp.StatusCode,
			"endpoint":    c.config.Elchi.APIEndpoint,
			"project":     payload.Project,
			"error":       err.Error(),
		}).Warn("Failed to parse API response, but HTTP status indicates success")
		return nil
	}

	// Log based on response success
	if apiResponse.Success {
		// After success, initial:false will be sent
		c.initialCompleted.Store(true)
		c.logger.WithFields(map[string]interface{}{
			"status_code": resp.StatusCode,
			"endpoint":    c.config.Elchi.APIEndpoint,
			"project":     payload.Project,
			"message":     apiResponse.Message,
		}).Info("Discovery result processed successfully by API")
	} else {
		c.logger.WithFields(map[string]interface{}{
			"status_code": resp.StatusCode,
			"endpoint":    c.config.Elchi.APIEndpoint,
			"project":     payload.Project,
			"error":       apiResponse.Error,
		}).Error("API reported processing error for discovery result")

		// Return error if API explicitly reported failure
		return fmt.Errorf("API processing failed: %s", apiResponse.Error)
	}

	return nil
}

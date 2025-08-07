package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/CloudNativeWorks/elchi-discovery/discovery"
	"github.com/CloudNativeWorks/elchi-discovery/internal/config"
	"github.com/CloudNativeWorks/elchi-discovery/internal/logger"
)

type Client struct {
	httpClient *http.Client
	config     *config.Config
	logger     *logger.Logger
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
		Timeout:   30 * time.Second,
	}

	return &Client{
		httpClient: httpClient,
		config:     cfg,
		logger:     log,
	}
}

func (c *Client) SendDiscoveryResult(result *discovery.DiscoveryResult) error {
	// Check if API endpoint is configured
	if c.config.Elchi.APIEndpoint == "" {
		c.logger.Debug("No API endpoint configured, skipping send")
		return nil
	}

	// Marshal result to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal discovery result: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.config.Elchi.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.config.Elchi.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Elchi.Token))
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API returned non-success status: %d", resp.StatusCode)
	}

	c.logger.WithFields(map[string]interface{}{
		"status_code": resp.StatusCode,
		"endpoint":    c.config.Elchi.APIEndpoint,
	}).Info("Discovery result sent to API successfully")

	return nil
}

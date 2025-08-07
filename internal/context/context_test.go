package context

import (
	"context"
	"testing"

	"github.com/CloudNativeWorks/elchi-discovery/internal/config"
)

func TestWithConfig(t *testing.T) {
	// Create a test config
	testConfig := &config.Config{
		ClusterName:       "test-cluster",
		DiscoveryInterval: 60,
		Log: config.LogConfig{
			Level:  "debug",
			Format: "json",
			Output: "stdout",
		},
		Elchi: config.ElchiConfig{
			Token:              "test-token",
			APIEndpoint:        "https://api.test.com",
			InsecureSkipVerify: true,
		},
	}

	// Create context with config
	baseCtx := context.Background()
	ctx := WithConfig(baseCtx, testConfig)

	// Verify the context is not nil
	if ctx == nil {
		t.Fatal("Expected context to be non-nil")
	}

	// Verify we can retrieve the config
	retrievedConfig := GetConfig(ctx)
	if retrievedConfig == nil {
		t.Fatal("Expected to retrieve config from context")
	}

	// Verify the config values match
	if retrievedConfig.ClusterName != testConfig.ClusterName {
		t.Errorf("Expected cluster name %s, got %s", testConfig.ClusterName, retrievedConfig.ClusterName)
	}
	if retrievedConfig.DiscoveryInterval != testConfig.DiscoveryInterval {
		t.Errorf("Expected discovery interval %d, got %d", testConfig.DiscoveryInterval, retrievedConfig.DiscoveryInterval)
	}
	if retrievedConfig.Log.Level != testConfig.Log.Level {
		t.Errorf("Expected log level %s, got %s", testConfig.Log.Level, retrievedConfig.Log.Level)
	}
	if retrievedConfig.Elchi.Token != testConfig.Elchi.Token {
		t.Errorf("Expected token %s, got %s", testConfig.Elchi.Token, retrievedConfig.Elchi.Token)
	}
}

func TestGetConfig_NoConfig(t *testing.T) {
	// Create context without config
	ctx := context.Background()

	// Try to get config
	retrievedConfig := GetConfig(ctx)

	// Should return nil when no config is set
	if retrievedConfig != nil {
		t.Error("Expected nil when no config is set in context")
	}
}

func TestGetConfig_WrongType(t *testing.T) {
	// Create context with wrong type of value
	ctx := context.WithValue(context.Background(), configKey, "wrong-type")

	// Try to get config
	retrievedConfig := GetConfig(ctx)

	// Should return nil when wrong type is stored
	if retrievedConfig != nil {
		t.Error("Expected nil when wrong type is stored in context")
	}
}

func TestContextKey_Unique(t *testing.T) {
	// Test that our context key doesn't interfere with other keys
	testConfig := &config.Config{
		ClusterName: "test-cluster",
	}

	ctx := context.Background()

	// Add our config
	ctx = WithConfig(ctx, testConfig)

	// Add another value with a different key
	type otherKey string
	const otherContextKey otherKey = "other-config"
	ctx = context.WithValue(ctx, otherContextKey, "other-value")

	// Verify both values can be retrieved independently
	retrievedConfig := GetConfig(ctx)
	if retrievedConfig == nil {
		t.Error("Expected to retrieve our config")
	}
	if retrievedConfig.ClusterName != "test-cluster" {
		t.Errorf("Expected cluster name 'test-cluster', got %s", retrievedConfig.ClusterName)
	}

	otherValue := ctx.Value(otherContextKey)
	if otherValue != "other-value" {
		t.Errorf("Expected other value 'other-value', got %v", otherValue)
	}
}

func TestWithConfig_NilConfig(t *testing.T) {
	// Test behavior with nil config
	baseCtx := context.Background()
	ctx := WithConfig(baseCtx, nil)

	// Verify the context is not nil
	if ctx == nil {
		t.Fatal("Expected context to be non-nil even with nil config")
	}

	// Verify we get nil when retrieving
	retrievedConfig := GetConfig(ctx)
	if retrievedConfig != nil {
		t.Error("Expected nil config when nil was stored")
	}
}

func TestContextChaining(t *testing.T) {
	// Test that context values can be chained properly
	config1 := &config.Config{
		ClusterName: "cluster1",
	}
	config2 := &config.Config{
		ClusterName: "cluster2",
	}

	// Create chain of contexts
	baseCtx := context.Background()
	ctx1 := WithConfig(baseCtx, config1)
	ctx2 := WithConfig(ctx1, config2) // This should override config1

	// Verify ctx2 has config2
	retrievedConfig := GetConfig(ctx2)
	if retrievedConfig == nil {
		t.Fatal("Expected to retrieve config from ctx2")
	}
	if retrievedConfig.ClusterName != "cluster2" {
		t.Errorf("Expected cluster name 'cluster2', got %s", retrievedConfig.ClusterName)
	}

	// Verify ctx1 still has config1
	retrievedConfig1 := GetConfig(ctx1)
	if retrievedConfig1 == nil {
		t.Fatal("Expected to retrieve config from ctx1")
	}
	if retrievedConfig1.ClusterName != "cluster1" {
		t.Errorf("Expected cluster name 'cluster1', got %s", retrievedConfig1.ClusterName)
	}
}

func TestContextCancellation(t *testing.T) {
	// Test that our context works with cancellation
	testConfig := &config.Config{
		ClusterName: "test-cluster",
	}

	baseCtx := context.Background()
	ctx := WithConfig(baseCtx, testConfig)

	// Create cancellable context
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Verify config is still accessible in cancelled context
	retrievedConfig := GetConfig(cancelCtx)
	if retrievedConfig == nil {
		t.Fatal("Expected to retrieve config from cancellable context")
	}
	if retrievedConfig.ClusterName != "test-cluster" {
		t.Errorf("Expected cluster name 'test-cluster', got %s", retrievedConfig.ClusterName)
	}

	// Cancel the context
	cancel()

	// Config should still be accessible even after cancellation
	retrievedConfig = GetConfig(cancelCtx)
	if retrievedConfig == nil {
		t.Fatal("Expected to retrieve config even after cancellation")
	}
	if retrievedConfig.ClusterName != "test-cluster" {
		t.Errorf("Expected cluster name 'test-cluster', got %s", retrievedConfig.ClusterName)
	}
}

func TestConfigModification(t *testing.T) {
	// Test that modifying the original config doesn't affect the context
	testConfig := &config.Config{
		ClusterName: "original-cluster",
	}

	ctx := WithConfig(context.Background(), testConfig)

	// Modify the original config
	testConfig.ClusterName = "modified-cluster"

	// The context should still contain the reference to the same config object
	// (Go passes pointers, so this will actually see the modification)
	retrievedConfig := GetConfig(ctx)
	if retrievedConfig == nil {
		t.Fatal("Expected to retrieve config")
	}

	// Since we're storing a pointer, the modification will be visible
	if retrievedConfig.ClusterName != "modified-cluster" {
		t.Errorf("Expected cluster name 'modified-cluster', got %s", retrievedConfig.ClusterName)
	}

	// This test demonstrates that the context stores a reference, not a copy
	if retrievedConfig != testConfig {
		t.Error("Expected retrieved config to be the same reference as original")
	}
}

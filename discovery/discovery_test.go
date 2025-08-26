package discovery

import (
	"context"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	discoveryFake "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewService(t *testing.T) {
	client := fake.NewSimpleClientset()
	clusterName := "test-cluster"

	service := NewService(client, clusterName)

	if service.clusterName != clusterName {
		t.Errorf("Expected clusterName to be %s, got %s", clusterName, service.clusterName)
	}

	// Test that the service was created properly
	if service.client == nil {
		t.Error("Expected client to be set")
	}
}

func TestDiscoverNodes(t *testing.T) {
	tests := []struct {
		name     string
		nodes    []v1.Node
		expected int
	}{
		{
			name:     "no nodes",
			nodes:    []v1.Node{},
			expected: 0,
		},
		{
			name: "single ready node",
			nodes: []v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
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
							{
								Type:    v1.NodeExternalIP,
								Address: "10.0.0.10",
							},
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "multiple nodes with different statuses",
			nodes: []v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
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
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
					},
					Status: v1.NodeStatus{
						NodeInfo: v1.NodeSystemInfo{
							KubeletVersion: "v1.28.1",
						},
						Conditions: []v1.NodeCondition{
							{
								Type:   v1.NodeReady,
								Status: v1.ConditionFalse,
							},
						},
						Addresses: []v1.NodeAddress{
							{
								Type:    v1.NodeInternalIP,
								Address: "192.168.1.11",
							},
						},
					},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			for _, node := range tt.nodes {
				_, err := client.CoreV1().Nodes().Create(context.TODO(), &node, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Failed to create test node: %v", err)
				}
			}

			service := NewService(client, "test-cluster")
			result, err := service.DiscoverNodes(context.Background())

			if err != nil {
				t.Errorf("DiscoverNodes() error = %v", err)
				return
			}

			if result.NodeCount != tt.expected {
				t.Errorf("Expected NodeCount = %d, got %d", tt.expected, result.NodeCount)
			}

			if len(result.Nodes) != tt.expected {
				t.Errorf("Expected %d nodes in result, got %d", tt.expected, len(result.Nodes))
			}

			if result.ClusterInfo.Name != "test-cluster" {
				t.Errorf("Expected cluster name to be 'test-cluster', got %s", result.ClusterInfo.Name)
			}

			if result.Timestamp.IsZero() {
				t.Error("Expected timestamp to be set")
			}

			if result.Duration == "" {
				t.Error("Expected duration to be set")
			}

			// Verify node details for single node test
			if tt.name == "single ready node" && len(result.Nodes) > 0 {
				node := result.Nodes[0]
				if node.Name != "node1" {
					t.Errorf("Expected node name 'node1', got %s", node.Name)
				}
				if node.Status != "Ready" {
					t.Errorf("Expected node status 'Ready', got %s", node.Status)
				}
				if node.Version != "v1.28.2" {
					t.Errorf("Expected node version 'v1.28.2', got %s", node.Version)
				}
				if node.Addresses["InternalIP"] != "192.168.1.10" {
					t.Errorf("Expected InternalIP '192.168.1.10', got %s", node.Addresses["InternalIP"])
				}
			}
		})
	}
}

func TestGetNodeStatus(t *testing.T) {
	tests := []struct {
		name     string
		node     v1.Node
		expected string
	}{
		{
			name: "ready node",
			node: v1.Node{
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			expected: "Ready",
		},
		{
			name: "not ready node",
			node: v1.Node{
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionFalse,
						},
					},
				},
			},
			expected: "NotReady",
		},
		{
			name: "node without ready condition",
			node: v1.Node{
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeMemoryPressure,
							Status: v1.ConditionFalse,
						},
					},
				},
			},
			expected: "Unknown",
		},
		{
			name: "node with no conditions",
			node: v1.Node{
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{},
				},
			},
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getNodeStatus(&tt.node)
			if result != tt.expected {
				t.Errorf("getNodeStatus() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestGetClusterInfo(t *testing.T) {
	tests := []struct {
		name            string
		clusterName     string
		serverVersion   *version.Info
		setupError      error
		expectedName    string
		expectedVersion string
	}{
		{
			name:        "successful version retrieval",
			clusterName: "test-cluster",
			serverVersion: &version.Info{
				GitVersion: "v1.28.2",
			},
			expectedName:    "test-cluster",
			expectedVersion: "v1.28.2",
		},
		{
			name:            "version retrieval fails",
			clusterName:     "test-cluster",
			serverVersion:   nil,
			expectedName:    "test-cluster",
			expectedVersion: "v0.0.0-master+$Format:%H$", // Default fake version
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()

			// Mock the discovery client
			fakeDiscovery, ok := client.Discovery().(*discoveryFake.FakeDiscovery)
			if !ok {
				t.Fatal("Failed to get fake discovery client")
			}

			if tt.serverVersion != nil {
				fakeDiscovery.FakedServerVersion = tt.serverVersion
			}

			service := NewService(client, tt.clusterName)
			result := service.getClusterInfo()

			if result.Name != tt.expectedName {
				t.Errorf("Expected cluster name %s, got %s", tt.expectedName, result.Name)
			}

			if result.Version != tt.expectedVersion {
				t.Errorf("Expected cluster version %s, got %s", tt.expectedVersion, result.Version)
			}
		})
	}
}

func TestDiscoverNodesContext(t *testing.T) {
	client := fake.NewSimpleClientset()
	service := NewService(client, "test-cluster")

	// Test with normal context (fake client doesn't respect cancellation)
	ctx := context.Background()

	result, err := service.DiscoverNodes(ctx)
	if err != nil {
		t.Errorf("DiscoverNodes() error = %v", err)
	}

	if result == nil {
		t.Error("Expected result to be non-nil")
	}
}

func TestDiscoverNodesPerformance(t *testing.T) {
	// Create multiple nodes to test performance
	client := fake.NewSimpleClientset()

	// Create 100 test nodes
	for i := 0; i < 100; i++ {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-" + string(rune(i)),
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
						Address: "192.168.1." + string(rune(i)),
					},
				},
			},
		}
		_, err := client.CoreV1().Nodes().Create(context.TODO(), node, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create test node: %v", err)
		}
	}

	service := NewService(client, "test-cluster")

	start := time.Now()
	result, err := service.DiscoverNodes(context.Background())
	duration := time.Since(start)

	if err != nil {
		t.Errorf("DiscoverNodes() error = %v", err)
	}

	if result.NodeCount != 100 {
		t.Errorf("Expected 100 nodes, got %d", result.NodeCount)
	}

	// Should complete within reasonable time (adjust as needed)
	if duration > 5*time.Second {
		t.Errorf("Discovery took too long: %v", duration)
	}

	t.Logf("Discovery of 100 nodes completed in: %v", duration)
}

package discovery

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// KubernetesClient interface for easier testing
type KubernetesClient interface {
	kubernetes.Interface
}

type Service struct {
	client      KubernetesClient
	clusterName string
}

func NewService(client KubernetesClient, clusterName string) *Service {
	return &Service{
		client:      client,
		clusterName: clusterName,
	}
}

func (s *Service) DiscoverNodes(ctx context.Context) (*DiscoveryResult, error) {
	discoveryStart := time.Now()

	// Get cluster info
	clusterInfo := s.getClusterInfo()

	// Get nodes
	nodes, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Build discovery result
	result := &DiscoveryResult{
		Timestamp:   time.Now(),
		ClusterInfo: clusterInfo,
		NodeCount:   len(nodes.Items),
		Nodes:       make([]NodeInfo, 0, len(nodes.Items)),
		Duration:    time.Since(discoveryStart).String(),
	}

	for _, node := range nodes.Items {
		nodeInfo := NodeInfo{
			Name:      node.Name,
			Status:    getNodeStatus(&node),
			Version:   node.Status.NodeInfo.KubeletVersion,
			Addresses: make(map[string]string),
		}

		for _, address := range node.Status.Addresses {
			nodeInfo.Addresses[string(address.Type)] = address.Address
		}

		result.Nodes = append(result.Nodes, nodeInfo)
	}

	return result, nil
}

func (s *Service) getClusterInfo() ClusterInfo {
	info := ClusterInfo{
		Name:    s.clusterName, // Cluster name is required from config
		Version: "unknown",
	}

	// Get cluster version from server version
	if version, err := s.client.Discovery().ServerVersion(); err == nil && version != nil {
		info.Version = version.GitVersion
	}

	return info
}

func getNodeStatus(node *v1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady {
			if condition.Status == v1.ConditionTrue {
				return "Ready"
			}
			return "NotReady"
		}
	}
	return "Unknown"
}

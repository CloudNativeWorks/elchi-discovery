package discovery

import "time"

type ClusterInfo struct {
	Name    string `json:"cluster_name"`
	Version string `json:"cluster_version"`
}

type NodeInfo struct {
	Name      string            `json:"name"`
	Roles     []string          `json:"roles"`
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Addresses map[string]string `json:"addresses"`
}

type DiscoveryResult struct {
	Timestamp   time.Time   `json:"timestamp"`
	ClusterInfo ClusterInfo `json:"cluster_info"`
	NodeCount   int         `json:"node_count"`
	Nodes       []NodeInfo  `json:"nodes"`
	Duration    string      `json:"discovery_duration"`
}

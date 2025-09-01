package main

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ClusterConfig represents kind Cluster configuration
type ClusterConfig struct {
	Kind                    string   `yaml:"kind"`
	APIVersion              string   `yaml:"apiVersion"`
	Name                    string   `yaml:"name,omitempty"`
	Nodes                   []Node   `yaml:"nodes"`
	ContainerdConfigPatches []string `yaml:"containerdConfigPatches,omitempty"`
}

// Node represents a kind node configuration
type Node struct {
	Role                 string             `yaml:"role"`
	Image                string             `yaml:"image,omitempty"`
	ExtraPortMappings    []ExtraPortMapping `yaml:"extraPortMappings,omitempty"`
	ExtraMounts          []ExtraMount       `yaml:"extraMounts,omitempty"`
	KubeadmConfigPatches []string           `yaml:"kubeadmConfigPatches,omitempty"`
}

// ExtraPortMapping represents port mapping configuration
type ExtraPortMapping struct {
	ContainerPort int    `yaml:"containerPort"`
	HostPort      int    `yaml:"hostPort"`
	Protocol      string `yaml:"protocol,omitempty"`
}

// ExtraMount represents mount configuration
type ExtraMount struct {
	HostPath      string `yaml:"hostPath"`
	ContainerPath string `yaml:"containerPath"`
}

// buildKindYAML generates kind cluster YAML from config
func buildKindYAML(cfg *Config) (string, error) {
	if cfg.ControlPlanes < 1 {
		return "", fmt.Errorf("at least one control plane node is required")
	}

	cluster := &ClusterConfig{
		Kind:       "Cluster",
		APIVersion: cfg.APIVersion,
		Name:       cfg.Name,
		Nodes:      buildNodes(cfg),
	}

	// Add local registry configuration if requested
	if cfg.WithLocalRegistry {
		cluster.ContainerdConfigPatches = []string{
			localRegistryPatch,
		}
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cluster)
	if err != nil {
		return "", fmt.Errorf("failed to marshal yaml: %w", err)
	}

	return string(data), nil
}

// buildNodes creates node configurations based on config
func buildNodes(cfg *Config) []Node {
	nodes := make([]Node, 0, cfg.ControlPlanes+cfg.Workers)

	// Create control plane nodes
	for i := 0; i < cfg.ControlPlanes; i++ {
		node := createControlPlaneNode(cfg, i)
		nodes = append(nodes, node)
	}

	// Create worker nodes
	for i := 0; i < cfg.Workers; i++ {
		node := createWorkerNode(cfg)
		nodes = append(nodes, node)
	}

	return nodes
}

// createControlPlaneNode creates a single control plane node
func createControlPlaneNode(cfg *Config, index int) Node {
	node := Node{
		Role: "control-plane",
	}

	// Set Kubernetes version if specified
	if cfg.K8sVersion != "" {
		node.Image = fmt.Sprintf("kindest/node:%s", cfg.K8sVersion)
	}

	// Apply special configuration to first control plane only
	if index == 0 {
		// Apply ingress configuration
		if cfg.WithIngress {
			configureIngressNode(&node)
		}

		// Apply user-specified port mappings
		node.ExtraPortMappings = append(node.ExtraPortMappings, convertPortMappings(cfg.Ports)...)
	}

	// Apply mounts to all control plane nodes
	node.ExtraMounts = convertMountMappings(cfg.Mounts)

	return node
}

// createWorkerNode creates a single worker node
func createWorkerNode(cfg *Config) Node {
	node := Node{
		Role: "worker",
	}

	// Set Kubernetes version if specified
	if cfg.K8sVersion != "" {
		node.Image = fmt.Sprintf("kindest/node:%s", cfg.K8sVersion)
	}

	// Apply mounts
	node.ExtraMounts = convertMountMappings(cfg.Mounts)

	return node
}

// convertPortMappings converts config port mappings to node format
func convertPortMappings(ports PortMappings) []ExtraPortMapping {
	mappings := make([]ExtraPortMapping, 0, len(ports))
	for _, pm := range ports {
		mappings = append(mappings, ExtraPortMapping{
			ContainerPort: pm.ContainerPort,
			HostPort:      pm.HostPort,
			Protocol:      "TCP",
		})
	}
	return mappings
}

// convertMountMappings converts config mount mappings to node format
func convertMountMappings(mounts MountMappings) []ExtraMount {
	extraMounts := make([]ExtraMount, 0, len(mounts))
	for _, mm := range mounts {
		extraMounts = append(extraMounts, ExtraMount{
			HostPath:      mm.HostPath,
			ContainerPath: mm.ContainerPath,
		})
	}
	return extraMounts
}

package main

import (
	"strings"
	"testing"
)

func TestBuildNodes_MinimalCluster(t *testing.T) {
	config := &Config{
		ControlPlanes: 1,
		Workers:       0,
	}

	nodes := buildNodes(config)

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].Role != "control-plane" {
		t.Errorf("expected role control-plane, got %s", nodes[0].Role)
	}
}

func TestBuildNodes_WithWorkers(t *testing.T) {
	config := &Config{
		ControlPlanes: 1,
		Workers:       2,
	}

	nodes := buildNodes(config)

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	// First node should be control plane
	if nodes[0].Role != "control-plane" {
		t.Errorf("node 0: expected control-plane, got %s", nodes[0].Role)
	}

	// Rest should be workers
	if nodes[1].Role != "worker" {
		t.Errorf("node 1: expected worker, got %s", nodes[1].Role)
	}
	if nodes[2].Role != "worker" {
		t.Errorf("node 2: expected worker, got %s", nodes[2].Role)
	}
}

func TestBuildNodes_KubernetesVersion(t *testing.T) {
	config := &Config{
		K8sVersion:    "v1.30.0",
		ControlPlanes: 1,
		Workers:       1,
	}

	nodes := buildNodes(config)

	// All nodes should have the same image
	expectedImage := "kindest/node:v1.30.0"
	if nodes[0].Image != expectedImage {
		t.Errorf("control-plane: expected image %s, got %s", expectedImage, nodes[0].Image)
	}
	if nodes[1].Image != expectedImage {
		t.Errorf("worker: expected image %s, got %s", expectedImage, nodes[1].Image)
	}
}

func TestBuildNodes_IngressOnFirstControlPlaneOnly(t *testing.T) {
	config := &Config{
		ControlPlanes: 2,
		Workers:       1,
		WithIngress:   true,
	}

	nodes := buildNodes(config)

	// First control plane should have ingress config
	if len(nodes[0].ExtraPortMappings) != 2 {
		t.Errorf("first control plane: expected 2 port mappings, got %d", len(nodes[0].ExtraPortMappings))
	}
	if len(nodes[0].KubeadmConfigPatches) != 1 {
		t.Errorf("first control plane: expected 1 kubeadm patch, got %d", len(nodes[0].KubeadmConfigPatches))
	}

	// Verify ports 80 and 443 are mapped
	var has80, has443 bool
	for _, pm := range nodes[0].ExtraPortMappings {
		if pm.ContainerPort == 80 && pm.HostPort == 80 {
			has80 = true
		}
		if pm.ContainerPort == 443 && pm.HostPort == 443 {
			has443 = true
		}
	}
	if !has80 || !has443 {
		t.Error("expected ports 80 and 443 to be mapped")
	}

	// Second control plane should NOT have ingress config
	if len(nodes[1].ExtraPortMappings) != 0 {
		t.Error("second control plane should not have port mappings")
	}
	if len(nodes[1].KubeadmConfigPatches) != 0 {
		t.Error("second control plane should not have kubeadm patches")
	}

	// Worker should NOT have ingress config
	if len(nodes[2].ExtraPortMappings) != 0 {
		t.Error("worker should not have port mappings")
	}
}

func TestBuildNodes_CustomPortsOnFirstControlPlaneOnly(t *testing.T) {
	config := &Config{
		ControlPlanes: 1,
		Workers:       1,
		Ports: PortMappings{
			{HostPort: 8080, ContainerPort: 80},
			{HostPort: 5432, ContainerPort: 5432},
		},
	}

	nodes := buildNodes(config)

	// Control plane should have custom ports
	if len(nodes[0].ExtraPortMappings) != 2 {
		t.Fatalf("expected 2 port mappings, got %d", len(nodes[0].ExtraPortMappings))
	}

	// Worker should NOT have port mappings
	if len(nodes[1].ExtraPortMappings) != 0 {
		t.Error("worker should not have port mappings")
	}
}

func TestBuildNodes_MountsAppliedToAllNodes(t *testing.T) {
	config := &Config{
		ControlPlanes: 1,
		Workers:       2,
		Mounts: MountMappings{
			{HostPath: "/src", ContainerPath: "/app"},
			{HostPath: "/data", ContainerPath: "/data"},
		},
	}

	nodes := buildNodes(config)

	// All nodes should have mounts
	for i, node := range nodes {
		if len(node.ExtraMounts) != 2 {
			t.Errorf("node %d: expected 2 mounts, got %d", i, len(node.ExtraMounts))
		}

		// Verify mount paths
		if node.ExtraMounts[0].HostPath != "/src" || node.ExtraMounts[0].ContainerPath != "/app" {
			t.Errorf("node %d: incorrect first mount", i)
		}
		if node.ExtraMounts[1].HostPath != "/data" || node.ExtraMounts[1].ContainerPath != "/data" {
			t.Errorf("node %d: incorrect second mount", i)
		}
	}
}

func TestBuildNodes_IngressAndCustomPortsCombined(t *testing.T) {
	config := &Config{
		ControlPlanes: 1,
		WithIngress:   true,
		Ports: PortMappings{
			{HostPort: 5432, ContainerPort: 5432},
		},
	}

	nodes := buildNodes(config)

	// Should have ingress ports (80, 443) + custom port (5432)
	if len(nodes[0].ExtraPortMappings) != 3 {
		t.Fatalf("expected 3 port mappings, got %d", len(nodes[0].ExtraPortMappings))
	}

	// Count each port
	portCounts := make(map[int]int)
	for _, pm := range nodes[0].ExtraPortMappings {
		portCounts[pm.ContainerPort]++
	}

	if portCounts[80] != 1 {
		t.Error("expected port 80 to be mapped once")
	}
	if portCounts[443] != 1 {
		t.Error("expected port 443 to be mapped once")
	}
	if portCounts[5432] != 1 {
		t.Error("expected port 5432 to be mapped once")
	}
}

func TestBuildKindYAML_ErrorCases(t *testing.T) {
	// Zero control planes
	config := &Config{
		APIVersion:    "kind.x-k8s.io/v1alpha4",
		ControlPlanes: 0,
	}

	_, err := buildKindYAML(config)
	if err == nil {
		t.Error("expected error for zero control planes, got nil")
	}

	// Negative control planes
	config.ControlPlanes = -1
	_, err = buildKindYAML(config)
	if err == nil {
		t.Error("expected error for negative control planes, got nil")
	}
}

func TestBuildKindYAML_LocalRegistry(t *testing.T) {
	config := &Config{
		APIVersion:        "kind.x-k8s.io/v1alpha4",
		ControlPlanes:     1,
		WithLocalRegistry: true,
	}

	yaml, err := buildKindYAML(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the registry configuration is present
	if !strings.Contains(yaml, "containerdConfigPatches") {
		t.Error("expected containerdConfigPatches in output")
	}
	if !strings.Contains(yaml, "kind-registry:5000") {
		t.Error("expected registry endpoint in output")
	}
}

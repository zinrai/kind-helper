package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

// Config holds all configuration from command line flags
type Config struct {
	Name              string
	K8sVersion        string
	APIVersion        string
	ControlPlanes     int
	Workers           int
	WithIngress       bool
	WithLocalRegistry bool
	Ports             PortMappings
	Mounts            MountMappings
}

// PortMapping represents a single port mapping
type PortMapping struct {
	HostPort      int
	ContainerPort int
}

// PortMappings implements flag.Value for multiple port mappings
type PortMappings []PortMapping

func (p *PortMappings) String() string {
	return ""
}

func (p *PortMappings) Set(value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid port mapping format: %s (expected host:container)", value)
	}

	hostPort, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid host port: %s", parts[0])
	}

	containerPort, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid container port: %s", parts[1])
	}

	*p = append(*p, PortMapping{
		HostPort:      hostPort,
		ContainerPort: containerPort,
	})

	return nil
}

// MountMapping represents a single mount mapping
type MountMapping struct {
	HostPath      string
	ContainerPath string
}

// MountMappings implements flag.Value for multiple mount mappings
type MountMappings []MountMapping

func (m *MountMappings) String() string {
	return ""
}

func (m *MountMappings) Set(value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid mount mapping format: %s (expected host:container)", value)
	}

	if parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("mount paths cannot be empty")
	}

	*m = append(*m, MountMapping{
		HostPath:      parts[0],
		ContainerPath: parts[1],
	})

	return nil
}

// parseFlags parses command line flags and returns Config
func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Name, "name", "kind", "cluster name")
	flag.StringVar(&cfg.K8sVersion, "k8s-version", "", "Kubernetes version (e.g., v1.30.13@sha256:...)")
	flag.StringVar(&cfg.APIVersion, "api-version", "kind.x-k8s.io/v1alpha4", "kind API version")
	flag.IntVar(&cfg.ControlPlanes, "control-plane", 1, "number of control plane nodes")
	flag.IntVar(&cfg.Workers, "worker", 1, "number of worker nodes")
	flag.BoolVar(&cfg.WithIngress, "with-ingress", false, "configure cluster for ingress")
	flag.BoolVar(&cfg.WithLocalRegistry, "with-local-registry", false, "configure cluster for local registry")

	flag.Var(&cfg.Ports, "port", "port mapping HOST:CONTAINER (can be specified multiple times)")
	flag.Var(&cfg.Mounts, "mount", "mount mapping HOSTPATH:CONTAINERPATH (can be specified multiple times)")

	flag.Parse()

	// Basic validation
	if cfg.ControlPlanes < 1 {
		cfg.ControlPlanes = 1
	}

	if cfg.Workers < 0 {
		cfg.Workers = 0
	}

	return cfg
}

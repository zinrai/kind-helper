package main

// Configuration patches as constants

const (
	// ingressReadyPatch adds a label to the node for ingress controller scheduling
	ingressReadyPatch = `kind: InitConfiguration
nodeRegistration:
  kubeletExtraArgs:
    node-labels: "ingress-ready=true"`

	// localRegistryPatch configures containerd to use local registry
	localRegistryPatch = `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5001"]
  endpoint = ["http://kind-registry:5000"]`
)

// configureIngressNode applies ingress-specific configuration to a node
func configureIngressNode(node *Node) {
	// Add port mappings for HTTP and HTTPS
	node.ExtraPortMappings = append(node.ExtraPortMappings,
		ExtraPortMapping{
			ContainerPort: 80,
			HostPort:      80,
			Protocol:      "TCP",
		},
		ExtraPortMapping{
			ContainerPort: 443,
			HostPort:      443,
			Protocol:      "TCP",
		},
	)

	// Add kubeadm patch for node label
	node.KubeadmConfigPatches = append(node.KubeadmConfigPatches,
		ingressReadyPatch,
	)
}

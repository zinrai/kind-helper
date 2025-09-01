# kind-helper

A command-line tool to generate [kind](https://github.com/kubernetes-sigs/kind) cluster configuration YAML files.

## Overview

Kind cluster configurations often follow certain patterns: node counts, ingress, port mappings, and local registry. We observed that these patterns might not require manual YAML editing.

`kind-helper` is a command-line tool that generates kind configuration YAML from flags. It handles common configurations through flags like `-worker`, `-with-ingress`, `-port`, and `-with-local-registry`. Configurations outside this scope are not covered—those require manual YAML creation.

The tool outputs YAML to stdout and does nothing else. This design choice allows it to work with other command-line tools through pipes and redirections.

## Installation

```bash
$ go install github.com/example/kind-helper@latest
```

## Usage

Generate YAML to stdout:

```bash
$ kind-helper [FLAGS]
```

Create cluster directly:

```bash
$ kind-helper -worker 2 | kind create cluster --config -
```

Save configuration:

```bash
$ kind-helper -worker 2 > cluster.yaml
```

Review before creating:

```bash
$ kind-helper -worker 2 | tee cluster.yaml | kind create cluster --config -
```

## Flags

| Flag                   | Description                                                        | Default                  |
|------------------------|--------------------------------------------------------------------|--------------------------|
| `-name`                | Cluster name                                                       | `kind`                   |
| `-k8s-version`         | Kubernetes version with full image tag (e.g., v1.30.13@sha256:...) | Latest kind default      |
| `-api-version`         | kind API version                                                   | `kind.x-k8s.io/v1alpha4` |
| `-control-plane`       | Number of control plane nodes                                      | `1`                      |
| `-worker`              | Number of worker nodes                                             | `1`                      |
| `-with-ingress`        | Configure cluster for ingress                                      | `false`                  |
| `-with-local-registry` | Configure cluster for local registry                               | `false`                  |
| `-port`                | Port mapping HOST:CONTAINER (can be repeated)                      | None                     |
| `-mount`               | Mount mapping HOSTPATH:CONTAINERPATH (can be repeated)             | None                     |

## Examples

### multi-node cluster

```bash
$ kind-helper -worker 3
```

### Cluster with specific Kubernetes version

```bash
$ kind-helper -k8s-version v1.30.13@sha256:397209b3d947d154f6641f2d0ce8d473732bd91c87d9575ade99049aa33cd648 -worker 2
```

### Ingress Support

```bash
$ kind-helper -with-ingress | kind create cluster --config -
```

See [kind's ingress documentation](https://kind.sigs.k8s.io/docs/user/ingress/) for setting up an Ingress Controller.

### Local Registry Support

```bash
$ kind-helper -with-local-registry | kind create cluster --config -
```

See [kind's local registry documentation](https://kind.sigs.k8s.io/docs/user/local-registry/) for registry setup.

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) for details.

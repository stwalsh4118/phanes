# PBI-13: K3s Lightweight Kubernetes Module

[View in Backlog](../backlog.md#user-content-13)

## Overview

Create a K3s module that installs and configures K3s (lightweight Kubernetes) on servers. This enables quick setup of Kubernetes clusters for container orchestration during provisioning.

## Problem Statement

Installing K3s manually requires:
- Running the official install script
- Configuring as server or agent node
- Setting up kubectl access
- Configuring kubeconfig for users
- Managing cluster tokens for agent nodes

This is complex and time-consuming. Having K3s configured automatically during provisioning enables immediate Kubernetes cluster deployment.

## User Stories

- As a developer, I want K3s installed automatically so that I can run Kubernetes workloads
- As a developer, I want to configure server or agent nodes so that I can build clusters
- As a developer, I want kubectl configured so that I can manage the cluster
- As a developer, I want idempotent installation so that I can update configurations safely

## Technical Approach

### Module Implementation

Create `internal/modules/k3s/k3s.go` implementing the `Module` interface:

```go
type K3sModule struct{}

func (m *K3sModule) Name() string { return "k3s" }
func (m *K3sModule) Description() string { return "Installs and configures K3s Kubernetes" }
func (m *K3sModule) IsInstalled() (bool, error) { ... }
func (m *K3sModule) Install(cfg *config.Config) error { ... }
```

### Installation Process

1. Check if K3s is already installed (`k3s --version`)
2. Determine node role (server or agent)
3. For server nodes: Run `curl -sfL https://get.k3s.io | sh -`
4. For agent nodes: Run `curl -sfL https://get.k3s.io | K3S_URL=<server-url> K3S_TOKEN=<token> sh -`
5. Copy kubeconfig to user's home directory: `~/.kube/config`
6. Set proper permissions on kubeconfig
7. Verify kubectl access

### Configuration

Add to `internal/config/config.go`:

```go
type Config struct {
    // ... existing fields
    K3s K3s `yaml:"k3s"`
}

type K3s struct {
    Enabled   bool   `yaml:"enabled"`
    Role      string `yaml:"role"`       // "server" or "agent"
    ServerURL string `yaml:"server_url"` // Required for agent nodes
    Token     string `yaml:"token"`      // Required for agent nodes
}
```

### Node Roles

- **Server**: Single-node cluster or first node in multi-node cluster
- **Agent**: Worker node that joins an existing cluster

### Kubeconfig Setup

- Copy `/etc/rancher/k3s/k3s.yaml` to `~/.kube/config`
- Update server URL in kubeconfig if needed (for remote access)
- Set permissions: `chmod 600 ~/.kube/config`
- Ensure kubectl is available (installed with K3s)

### IsInstalled Check

- Check if `k3s` command exists
- Check if K3s service is running (`systemctl status k3s`)
- Verify kubectl can access cluster (`kubectl get nodes`)

## UX/UI Considerations

- Clear logging during K3s installation
- Show cluster status after installation (node count, ready state)
- Display kubectl configuration location
- Warn if agent configuration is incomplete (missing server URL or token)
- Show example kubectl commands after setup

## Acceptance Criteria

1. Module installs K3s using official install script
2. Module configures as server (single-node) or agent node
3. kubectl is available and configured for the user
4. Config struct includes k3s section with role and connection options
5. Module is idempotent (safe to run multiple times)
6. Kubeconfig is properly configured with correct permissions
7. Module validates cluster connectivity after installation

## Dependencies

- PBI-1 (Foundation) must be complete - exec helpers needed
- PBI-2 (Module Framework) must be complete - module interface needed
- For agent nodes: Server URL and token required from user

## Open Questions

- Should we support K3s options/flags? (Future enhancement - e.g., `--disable traefik`)
- Should we support multi-server HA setup? (Future enhancement)
- Should we install kubectl separately if not included? (K3s includes it)
- What default K3s version should we use? (Latest stable)

## Related Tasks

See [Tasks](./tasks.md)


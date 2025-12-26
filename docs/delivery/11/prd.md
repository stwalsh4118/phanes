# PBI-11: Tailscale VPN Module

[View in Backlog](../backlog.md#user-content-11)

## Overview

Create a Tailscale module that automatically installs Tailscale and joins servers to a tailnet during provisioning. This enables secure, zero-config VPN connectivity between provisioned servers.

## Problem Statement

Setting up Tailscale manually requires:
- Downloading and running the install script
- Authenticating with an auth key
- Configuring the service to start on boot
- Verifying connectivity

This is repetitive and error-prone when provisioning multiple servers. Having Tailscale configured automatically during provisioning enables immediate secure connectivity between servers.

## User Stories

- As a developer, I want Tailscale installed automatically so that my servers can connect to my tailnet
- As a developer, I want to provide an auth key in config so that servers authenticate automatically
- As a developer, I want idempotent installation so that I can run Phanes multiple times safely
- As a developer, I want to verify Tailscale is working so that I know connectivity is established

## Technical Approach

### Module Implementation

Create `internal/modules/tailscale/tailscale.go` implementing the `Module` interface:

```go
type TailscaleModule struct{}

func (m *TailscaleModule) Name() string { return "tailscale" }
func (m *TailscaleModule) Description() string { return "Installs and configures Tailscale VPN" }
func (m *TailscaleModule) IsInstalled() (bool, error) { ... }
func (m *TailscaleModule) Install(cfg *config.Config) error { ... }
```

### Installation Process

1. Check if Tailscale is already installed (`tailscale version`)
2. Download and run official install script: `curl -fsSL https://tailscale.com/install.sh | sh`
3. Authenticate using auth key: `tailscale up --authkey=<key>`
4. Enable and start systemd service: `systemctl enable --now tailscaled`

### IsInstalled Check

- Check if `tailscale` command exists
- Check if Tailscale is authenticated: `tailscale status` should show connected state
- Verify systemd service is enabled

### Configuration

Add to `internal/config/config.go`:

```go
type Config struct {
    // ... existing fields
    Tailscale Tailscale `yaml:"tailscale"`
}

type Tailscale struct {
    Enabled bool   `yaml:"enabled"`
    AuthKey string `yaml:"auth_key"`
}
```

### Auth Key Handling

- Auth key is required if Tailscale is enabled
- Should be a one-time use key from Tailscale admin console
- Module should validate auth key format (starts with `tskey-`)
- After authentication, auth key is no longer needed (idempotency)

## UX/UI Considerations

- Clear logging when installing Tailscale
- Show Tailscale status after installation (IP address, connected state)
- Warn if auth key is missing when Tailscale is enabled
- Display Tailscale node name/IP after successful installation

## Acceptance Criteria

1. Module installs Tailscale via official install script
2. Module authenticates using provided auth key from config
3. `IsInstalled()` correctly detects if Tailscale is already configured and authenticated
4. Module is idempotent (safe to run multiple times)
5. Config struct includes tailscale section with `enabled` and `auth_key` fields
6. Module validates auth key format
7. Tailscale service is enabled and started on boot
8. Module displays Tailscale status after installation

## Dependencies

- PBI-1 (Foundation) must be complete - exec helpers needed
- PBI-2 (Module Framework) must be complete - module interface needed
- Tailscale account and auth key required from user

## Open Questions

- Should we support exit node configuration? (Future enhancement)
- Should we support subnet routing/advertise routes? (Future enhancement)
- Should we support ACL tags? (Future enhancement)
- How to handle auth key rotation? (User responsibility - use new key)

## Related Tasks

See [Tasks](./tasks.md)


# Phanes

Phanes is a VPS provisioning system for Linux servers. It provides a modular, idempotent approach to configuring servers with predefined modules and profiles. Phanes supports dry-run mode, configuration-driven setup, and is designed to be safe to run multiple times.

## Features

- **Modular Architecture**: Compose server configurations from reusable modules
- **Profiles**: Pre-configured server setups (dev, web, database, etc.)
- **Idempotent**: Safe to run multiple times - only makes changes when needed
- **Dry-Run Mode**: Preview changes before executing
- **Configuration-Driven**: YAML-based configuration for all modules
- **Comprehensive Modules**: Baseline setup, security, Docker, databases, web servers, and more

## Quick Start

1. **Install Phanes**:
   ```bash
   go install github.com/stwalsh4118/phanes@latest
   ```

2. **Create a configuration file** (`config.yaml`):
   ```yaml
   user:
     username: "your-username"
     ssh_public_key: "ssh-ed25519 AAAA..."
   system:
     timezone: "UTC"
   ```

3. **Run a profile**:
   ```bash
   phanes --profile dev --config config.yaml
   ```

That's it! Phanes will provision your server with the selected profile's modules.

## Installation

### From Source

```bash
git clone https://github.com/stwalsh4118/phanes.git
cd phanes
go build -o phanes
sudo mv phanes /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/stwalsh4118/phanes@latest
```

### Requirements

- Go 1.21 or later (for building from source)
- Linux-based operating system (Ubuntu/Debian recommended)
- Root or sudo access for provisioning operations

## Usage

### Running Profiles

Profiles are pre-configured sets of modules for common server types:

```bash
# Development server
phanes --profile dev --config config.yaml

# Web server
phanes --profile web --config config.yaml

# Database server
phanes --profile database --config config.yaml

# Minimal server
phanes --profile minimal --config config.yaml

# Coolify hosting platform
phanes --profile coolify --config config.yaml
```

### Running Specific Modules

You can also run individual modules or a custom combination:

```bash
# Run specific modules
phanes --modules baseline,user,docker --config config.yaml

# Combine profile and additional modules
phanes --profile dev --modules nginx --config config.yaml
```

### Dry-Run Mode

Preview what changes would be made without actually executing them:

```bash
phanes --profile dev --config config.yaml --dry-run
```

### Listing Available Options

See all available modules and profiles:

```bash
phanes --list
```

## Configuration

Phanes uses a YAML configuration file to customize module behavior. See [`config.yaml.example`](config.yaml.example) for a complete example with all available options.

### Required Configuration

At minimum, you must configure:

- `user.username`: The username to create on the server
- `user.ssh_public_key`: Your SSH public key for authentication

### Configuration File Location

By default, Phanes looks for `config.yaml` in the current directory. Specify a different path with:

```bash
phanes --profile dev --config /path/to/config.yaml
```

## Available Modules

Phanes includes the following modules:

| Module | Description |
|--------|-------------|
| `baseline` | Sets timezone, locale, and runs apt update |
| `user` | Creates user and sets up SSH keys |
| `security` | Configures UFW, fail2ban, and SSH hardening |
| `swap` | Creates and configures swap file |
| `updates` | Configures automatic security updates |
| `docker` | Installs Docker CE and Docker Compose |
| `monitoring` | Installs and configures Netdata monitoring |
| `nginx` | Installs and configures Nginx web server |
| `caddy` | Installs and configures Caddy web server with automatic HTTPS |
| `coolify` | Installs and configures Coolify self-hosted PaaS |
| `postgres` | Installs and configures PostgreSQL database server |
| `redis` | Installs and configures Redis in-memory data store |
| `devtools` | Installs development tools (Git, build-essential, Node.js, Python, Go) |

## Available Profiles

Profiles combine multiple modules for common server configurations:

| Profile | Modules | Use Case |
|---------|---------|----------|
| `minimal` | baseline, user, security, swap, updates | Basic secure server setup |
| `dev` | baseline, user, security, swap, updates, docker, monitoring, devtools | Development environment |
| `web` | baseline, user, security, swap, updates, docker, monitoring, caddy | Web server with Caddy |
| `database` | baseline, user, security, swap, updates, docker, monitoring, postgres, redis | Database server |
| `coolify` | baseline, user, security, swap, updates, docker, coolify | Self-hosted PaaS platform |

## Troubleshooting

### Config File Not Found

**Error**: `Config file not found at config.yaml`

**Solution**: Create a `config.yaml` file or specify the path with `--config`:

```bash
phanes --profile dev --config /path/to/config.yaml
```

### Invalid YAML Syntax

**Error**: `Invalid YAML syntax in config file`

**Solution**: Validate your YAML syntax. Common issues:
- Missing colons after keys
- Incorrect indentation (use spaces, not tabs)
- Unquoted special characters

Use an online YAML validator or check the example file.

### Missing Required Fields

**Error**: `user.username is required` or `user.ssh_public_key is required`

**Solution**: Ensure your config file includes both required fields:

```yaml
user:
  username: "your-username"
  ssh_public_key: "ssh-ed25519 AAAA..."
```

### Module Not Found

**Error**: `Module 'xyz' not found in registry`

**Solution**: Check available modules with `phanes --list`. Ensure module names match exactly (case-sensitive).

### Permission Denied

**Error**: Permission errors during execution

**Solution**: Phanes requires root or sudo access for most operations. Run with:

```bash
sudo phanes --profile dev --config config.yaml
```

### Dry-Run Shows Changes But Nothing Happens

**Expected Behavior**: Dry-run mode only previews changes. Remove `--dry-run` to execute:

```bash
phanes --profile dev --config config.yaml
```

### Module Already Installed

**Expected Behavior**: Phanes is idempotent. If a module is already installed, it will be skipped with a message. This is normal and safe.

## Contributing

Contributions are welcome! Please ensure:

1. Code follows Go conventions
2. All tests pass
3. Documentation is updated
4. Modules are idempotent

## License

[Add your license here]

## Version

Current version: 0.1.0

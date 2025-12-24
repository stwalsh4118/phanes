# Development Guide

This guide covers development setup, hot reload, testing, and Docker workflows.

## Quick Start

### Install Development Tools

```bash
make install-tools
```

This installs:
- `air` - Hot reload for Go development
- `golangci-lint` - Linting tool

### Development with Hot Reload

Run the application with hot reload (automatically rebuilds on file changes):

```bash
make dev
# or
make air
```

This uses [air](https://github.com/cosmtrek/air) to watch for file changes and automatically rebuild/restart the application.

### Building

```bash
make build        # Build the binary to tmp/phanes
make run          # Build and run the binary
```

### Testing

```bash
make test                 # Run all tests
make test-unit            # Run unit tests only
make test-integration     # Run integration tests only
make test-race            # Run tests with race detector
make test-coverage        # Generate coverage report
```

### Code Quality

```bash
make fmt          # Format code
make vet          # Run go vet
make lint         # Run golangci-lint (if installed)
make check        # Run all checks (fmt, vet, lint, test)
```

## Docker Development

The project includes Docker support for running tests and services in containers.

### Using the Development Script

```bash
# Build Docker containers
./scripts/dev.sh build

# Start containers (PostgreSQL, Redis, etc.)
./scripts/dev.sh up

# Run tests locally
./scripts/dev.sh test

# Run tests in Docker containers
./scripts/dev.sh test docker

# View container logs
./scripts/dev.sh logs

# Stop containers
./scripts/dev.sh down

# Clean up everything
./scripts/dev.sh clean
```

### Using Makefile Targets

```bash
make docker-build    # Build Docker containers
make docker-up      # Start Docker containers
make docker-down    # Stop Docker containers
make docker-test    # Run tests in Docker containers
make docker-logs    # Show container logs
```

## Project Structure

```
phanes/
├── .air.toml              # Air configuration for hot reload
├── Makefile               # Common development tasks
├── scripts/
│   └── dev.sh            # Docker and testing script
├── docker-compose.dev.yml # Docker Compose configuration (auto-generated)
├── Dockerfile.test        # Dockerfile for test containers (auto-generated)
├── internal/              # Internal packages
│   ├── config/           # Configuration package
│   ├── exec/             # Execution helpers
│   └── log/              # Logging package
├── test/                  # Test files
│   ├── unit/             # Unit tests
│   └── integration/      # Integration tests
└── main.go               # Application entry point
```

## Air Configuration

The `.air.toml` file configures hot reload behavior:

- **Watches**: `.go` files (excludes `_test.go` files)
- **Builds**: Runs `go build -o ./tmp/phanes .`
- **Output**: Binary is built to `tmp/phanes`
- **Excludes**: `tmp/`, `vendor/`, `testdata/`, `docs/`, `.git/`

You can customize this by editing `.air.toml`.

## Docker Services

The Docker setup includes:

- **test**: Service for running tests in containers
- **postgres**: PostgreSQL 16 for integration tests (port 5432)
- **redis**: Redis 7 for integration tests (port 6379)

Services are automatically created when you run `./scripts/dev.sh build` or `./scripts/dev.sh up`.

## Common Workflows

### Daily Development

```bash
# Start hot reload
make dev

# In another terminal, run tests
make test
```

### Before Committing

```bash
# Run all checks
make check

# Or individually:
make fmt
make vet
make lint
make test
```

### Testing with Docker

```bash
# Start services
./scripts/dev.sh up

# Run tests (they'll connect to Docker services)
make test-integration

# Or run tests in containers
./scripts/dev.sh test docker
```

### Clean Up

```bash
# Clean build artifacts
make clean

# Clean Docker containers and volumes
./scripts/dev.sh clean
```

## Troubleshooting

### Air not found

```bash
make install-tools
```

### Docker not running

Make sure Docker daemon is running:
```bash
docker info
```

### Port conflicts

If ports 5432 or 6379 are already in use, edit `docker-compose.dev.yml` to change the ports.

### Permission denied on scripts

```bash
chmod +x scripts/dev.sh
```


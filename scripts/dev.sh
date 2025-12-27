#!/bin/bash

# Development script for running containers and tests
# Usage: ./scripts/dev.sh [command]
# Commands: build, up, down, test, logs, clean

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Docker compose file (can be overridden)
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.dev.yml}"

# Functions
print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Check if Docker is available
check_docker() {
    if ! command -v docker > /dev/null; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker daemon is not running"
        exit 1
    fi
}

# Build Docker images
build_containers() {
    print_header "Building Docker containers"
    check_docker
    
    if [ ! -f "$COMPOSE_FILE" ]; then
        print_info "Creating $COMPOSE_FILE..."
        create_docker_compose
    fi
    
    docker-compose -f "$COMPOSE_FILE" build
    print_success "Containers built successfully"
}

# Start containers
start_containers() {
    print_header "Starting Docker containers"
    check_docker
    
    if [ ! -f "$COMPOSE_FILE" ]; then
        print_info "Creating $COMPOSE_FILE..."
        create_docker_compose
    fi
    
    docker-compose -f "$COMPOSE_FILE" up -d
    print_success "Containers started"
    print_info "Run './scripts/dev.sh logs' to view logs"
}

# Stop containers
stop_containers() {
    print_header "Stopping Docker containers"
    check_docker
    
    if [ -f "$COMPOSE_FILE" ]; then
        docker-compose -f "$COMPOSE_FILE" down
        print_success "Containers stopped"
    else
        print_info "No docker-compose file found, nothing to stop"
    fi
}

# Show container logs
show_logs() {
    print_header "Container logs"
    check_docker
    
    if [ -f "$COMPOSE_FILE" ]; then
        docker-compose -f "$COMPOSE_FILE" logs -f
    else
        print_error "No docker-compose file found"
        exit 1
    fi
}

# Run tests
run_tests() {
    print_header "Running tests"
    
    # Run unit tests
    print_info "Running unit tests..."
    if go test -v ./internal/...; then
        print_success "Unit tests passed"
    else
        print_error "Unit tests failed"
        exit 1
    fi
    
    # Run integration tests if they exist
    if [ -d "test/integration" ]; then
        print_info "Running integration tests..."
        if go test -v ./test/integration/...; then
            print_success "Integration tests passed"
        else
            print_error "Integration tests failed"
            exit 1
        fi
    else
        print_info "No integration tests found"
    fi
    
    # Run tests with race detector
    print_info "Running tests with race detector..."
    if go test -race -v ./...; then
        print_success "Race detector tests passed"
    else
        print_error "Race detector tests failed"
        exit 1
    fi
    
    print_success "All tests passed!"
}

# Run tests in containers
run_tests_in_containers() {
    print_header "Running tests in Docker containers"
    check_docker
    
    if [ ! -f "$COMPOSE_FILE" ]; then
        print_info "Creating $COMPOSE_FILE..."
        create_docker_compose
    fi
    
    # Build test image
    docker-compose -f "$COMPOSE_FILE" build test
    
    # Run tests
    docker-compose -f "$COMPOSE_FILE" run --rm test
    print_success "Tests completed in containers"
}

# Clean up
clean() {
    print_header "Cleaning up"
    check_docker
    
    if [ -f "$COMPOSE_FILE" ]; then
        docker-compose -f "$COMPOSE_FILE" down -v --remove-orphans
        print_success "Containers and volumes removed"
    fi
    
    # Clean build artifacts
    rm -rf tmp/
    rm -f coverage.out coverage.html build-errors.log
    print_success "Build artifacts cleaned"
}

# Create docker-compose file if it doesn't exist
create_docker_compose() {
    cat > "$COMPOSE_FILE" << 'EOF'
version: '3.8'

services:
  # Test service for running tests in containers
  test:
    build:
      context: .
      dockerfile: Dockerfile.test
    volumes:
      - .:/app
      - go-modules:/go/pkg/mod
    working_dir: /app
    command: go test -v ./...
    environment:
      - CGO_ENABLED=0
    networks:
      - phanes-network

  # Example: PostgreSQL for integration tests (if needed)
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: phanes_test
      POSTGRES_PASSWORD: test_password
      POSTGRES_DB: phanes_test
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - phanes-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U phanes_test"]
      interval: 5s
      timeout: 5s
      retries: 5

  # Example: Redis for integration tests (if needed)
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - phanes-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  go-modules:
  postgres-data:
  redis-data:

networks:
  phanes-network:
    driver: bridge
EOF
    print_success "Created $COMPOSE_FILE"
}

# Create Dockerfile.test if it doesn't exist
create_dockerfile_test() {
    if [ ! -f "Dockerfile.test" ]; then
        cat > Dockerfile.test << 'EOF'
FROM golang:1.25-alpine AS test

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Default command runs tests
CMD ["go", "test", "-v", "./..."]
EOF
        print_success "Created Dockerfile.test"
    fi
}

# Main command handler
case "${1:-help}" in
    build)
        create_dockerfile_test
        build_containers
        ;;
    up|start)
        create_dockerfile_test
        start_containers
        ;;
    down|stop)
        stop_containers
        ;;
    test)
        if [ "${2:-}" = "docker" ]; then
            create_dockerfile_test
            run_tests_in_containers
        else
            run_tests
        fi
        ;;
    logs)
        show_logs
        ;;
    clean)
        clean
        ;;
    help|--help|-h)
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  build          Build Docker containers"
        echo "  up|start       Start Docker containers"
        echo "  down|stop      Stop Docker containers"
        echo "  test           Run tests locally"
        echo "  test docker    Run tests in Docker containers"
        echo "  logs           Show container logs"
        echo "  clean          Clean up containers and build artifacts"
        echo "  help           Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0 build       # Build containers"
        echo "  $0 up          # Start containers"
        echo "  $0 test        # Run tests locally"
        echo "  $0 test docker # Run tests in containers"
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Run '$0 help' for usage information"
        exit 1
        ;;
esac



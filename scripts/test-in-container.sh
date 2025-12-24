#!/bin/bash
# Script to test Phanes in a Docker container
# This allows safe testing of system provisioning without affecting the host

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "🐳 Building Phanes test container..."
docker-compose -f docker-compose.test.yml build

echo ""
echo "📋 Available test commands:"
echo "  ./scripts/test-in-container.sh list          - List modules and profiles"
echo "  ./scripts/test-in-container.sh dry-run       - Test with dry-run flag"
echo "  ./scripts/test-in-container.sh exec <args>   - Execute phanes with custom args"
echo "  ./scripts/test-in-container.sh e2e           - Run E2E tests"
echo "  ./scripts/test-in-container.sh shell          - Open interactive shell"
echo ""

if [ "$1" = "list" ]; then
    echo "🔍 Listing available modules and profiles..."
    docker-compose -f docker-compose.test.yml run --rm phanes-test phanes --list
elif [ "$1" = "dry-run" ]; then
    echo "🧪 Testing with dry-run (safe, no changes)..."
    docker-compose -f docker-compose.test.yml run --rm phanes-test phanes --modules baseline,user --config test-config.yaml --dry-run
elif [ "$1" = "exec" ]; then
    shift
    echo "🚀 Executing phanes with args: $@"
    docker-compose -f docker-compose.test.yml run --rm phanes-test phanes "$@"
elif [ "$1" = "e2e" ]; then
    echo "🧪 Running E2E tests..."
    docker-compose -f docker-compose.test.yml run --rm phanes-test go test -v ./test/integration/... -run "E2E"
elif [ "$1" = "shell" ]; then
    echo "🐚 Opening interactive shell..."
    echo "You can run phanes commands directly, e.g.:"
    echo "  phanes --list"
    echo "  phanes --modules baseline --config test-config.yaml --dry-run"
    docker-compose -f docker-compose.test.yml run --rm phanes-test /bin/bash
else
    echo "❌ Unknown command: $1"
    echo "Run without arguments to see usage"
    exit 1
fi


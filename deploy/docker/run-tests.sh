#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

DOCKER_DIR="$SCRIPT_DIR"

usage() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  up         - Start test environment"
    echo "  down       - Stop test environment"
    echo "  test       - Run E2E tests"
    echo "  seed       - Seed test data"
    echo "  logs       - View logs"
    echo "  cleanup    - Remove all test data"
    echo "  all        - Start, seed data, and run tests"
    echo ""
    echo "Examples:"
    echo "  $0 up           # Start the test environment"
    echo "  $0 test         # Run tests against running environment"
    echo "  $0 all          # Full test run (start + seed + test)"
}

wait_for_service() {
    local url=$1
    local name=$2
    echo "Waiting for $name..."
    for i in {1..30}; do
        if curl -sf "$url" > /dev/null 2>&1; then
            echo "$name is ready!"
            return 0
        fi
        sleep 2
    done
    echo "ERROR: $name failed to start"
    return 1
}

cmd_up() {
    echo "=== Starting test environment ==="
    cd "$DOCKER_DIR"
    docker-compose -f docker-compose.test.yml up -d
    
    echo ""
    echo "Waiting for services..."
    wait_for_service "http://localhost:5173" "Frontend"
    wait_for_service "http://localhost:5174/admin.html" "Admin"
    wait_for_service "http://localhost:8081/health" "Backend"
    
    echo ""
    echo "=== Services started ==="
    echo "Frontend:   http://localhost:5173"
    echo "Admin:     http://localhost:5174/admin.html"
    echo "API:       http://localhost:8081"
    echo "Swagger:   http://localhost:8081/swagger/index.html"
}

cmd_down() {
    echo "=== Stopping test environment ==="
    cd "$DOCKER_DIR"
    docker-compose -f docker-compose.test.yml down
}

cmd_seed() {
    echo "=== Seeding test data ==="
    docker exec gapi-postgres-test psql -U gapi -d gapi -f /seed.sql 2>/dev/null || {
        docker cp backend/scripts/seed.sql gapi-postgres-test:/seed.sql
        docker exec gapi-postgres-test psql -U gapi -d gapi -f /seed.sql
    }
    echo "=== Test data seeded ==="
}

cmd_logs() {
    cd "$DOCKER_DIR"
    docker-compose -f docker-compose.test.yml logs -f
}

cmd_cleanup() {
    echo "=== Cleaning up test environment ==="
    cd "$DOCKER_DIR"
    docker-compose -f docker-compose.test.yml down -v
    echo "=== Cleanup complete ==="
}

cmd_test() {
    echo "=== Running E2E tests ==="
    
    # Check if services are running
    if ! curl -sf "http://localhost:8081/health" > /dev/null 2>&1; then
        echo "ERROR: Backend is not running. Run '$0 up' first."
        exit 1
    fi
    
    # Run tests from frontend directory
    cd "$SCRIPT_DIR/frontend"
    
    # Use environment variables to point to Docker services
    export PLAYWRIGHT_BASE_URL="http://localhost:5173"
    export API_BASE_URL="http://localhost:8081"
    
    npx playwright test --reporter=list
}

cmd_all() {
    cmd_up
    sleep 5
    cmd_seed
    sleep 2
    cmd_test
}

case "${1:-}" in
    up)
        cmd_up
        ;;
    down)
        cmd_down
        ;;
    seed)
        cmd_seed
        ;;
    test)
        cmd_test
        ;;
    logs)
        cmd_logs
        ;;
    cleanup)
        cmd_cleanup
        ;;
    all)
        cmd_all
        ;;
    *)
        usage
        exit 1
        ;;
esac

#!/usr/bin/env bash
# gAPI Platform - Quick Start Script (Linux/Mac)
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  gAPI Platform - Starting...${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check Docker
if ! command -v docker &>/dev/null; then
    echo -e "${RED}Error: Docker is not installed.${NC}"
    echo "Please install Docker first: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check Docker Compose
if ! docker compose version &>/dev/null; then
    echo -e "${RED}Error: Docker Compose is not available.${NC}"
    echo "Please install Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi

# Create .env if not exists
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}[1/3] Creating .env from template...${NC}"
    cp .env.example .env
    echo -e "${YELLOW}  -> Please edit .env and set your passwords before continuing!${NC}"
    echo -e "${YELLOW}  -> After editing, run this script again.${NC}"
    echo ""
    echo "  Quick edit:  nano .env"
    echo "  Or:          code .env"
    echo ""
    exit 0
fi

# Check if defaults were changed
if grep -q "CHANGE_ME" .env 2>/dev/null; then
    echo -e "${YELLOW}Warning: You still have default passwords in .env${NC}"
    echo -e "${YELLOW}Press Ctrl+C to abort, or wait 5 seconds to continue anyway...${NC}"
    sleep 5
fi

echo -e "${GREEN}[2/3] Starting all services with Docker Compose...${NC}"
docker compose up -d

echo ""
echo -e "${GREEN}[3/3] Services starting up...${NC}"
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  All services are being started!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "  Backend API:   http://localhost:8080"
echo "  API Docs:      http://localhost:8080/swagger/index.html"
echo "  Frontend:      http://localhost:5173"
echo "  Admin Panel:   http://localhost:5174"
echo ""
echo "  Check status:  docker compose ps"
echo "  View logs:     docker compose logs -f"
echo "  Stop all:      docker compose down"
echo ""

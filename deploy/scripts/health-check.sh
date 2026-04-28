#!/bin/bash
# Health check script for mtsn2kolut-super-app services
# Usage: ./deploy/scripts/health-check.sh [service]

set -euo pipefail

SERVICE="${1:-all}"
TIMEOUT=5

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Load root directory
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${YELLOW}Performing health check for mtsn2kolut-super-app...${NC}"
echo ""

# Function to check backend health
check_backend() {
    echo -n "Checking backend API... "
    if response=$(curl -s -m "$TIMEOUT" -w "%{http_code}" "http://localhost:8080/health" -o /dev/null); then
        if [ "$response" -eq 200 ]; then
            echo -e "${GREEN}OK${NC}"
            return 0
        else
            echo -e "${RED}FAIL (HTTP $response)${NC}"
            return 1
        fi
    else
        echo -e "${RED}FAIL (timeout or connection error)${NC}"
        return 1
    fi
}

# Function to check frontend health
check_frontend() {
    echo -n "Checking frontend... "
    if response=$(curl -s -m "$TIMEOUT" -w "%{http_code}" "http://localhost:8021" -o /dev/null); then
        if [ "$response" -eq 200 ] || [ "$response" -eq 301 ] || [ "$response" -eq 302 ]; then
            echo -e "${GREEN}OK${NC}"
            return 0
        else
            echo -e "${RED}FAIL (HTTP $response)${NC}"
            return 1
        fi
    else
        echo -e "${RED}FAIL (timeout or connection error)${NC}"
        return 1
    fi
}

# Function to check worker health (via PM2)
check_worker() {
    echo -n "Checking worker... "
    if pm2 describe mtsn2kolut-pusaka-worker > /dev/null 2>&1; then
        status=$(pm2 mks)
        if echo "$status" | grep -q "mtsn2kolut-pusaka-worker"; then
            # Get the specific process status
            if pm2 show mtsn2kolut-pusaka-worker | grep -q "status.*online"; then
                echo -e "${GREEN}OK${NC}"
                return 0
            else
                echo -e "${YELLOW}NOT ONLINE${NC}"
                return 1
            fi
        else
            echo -e "${RED}NOT FOUND${NC}"
            return 1
        fi
    else
        echo -e "${RED}NOT FOUND${NC}"
        return 1
    fi
}

# Main execution
case "$SERVICE" in
    backend|all)
        check_backend || [[ "$SERVICE" == "all" ]] || exit 1
        ;;
    frontend|all)
        check_frontend || [[ "$SERVICE" == "all" ]] || exit 1
        ;;
    worker|all)
        check_worker || [[ "$SERVICE" == "all" ]] || exit 1
        ;;
    *)
        echo "Usage: $0 [backend|frontend|worker|all]"
        exit 1
        ;;
esac

echo ""
if [ "$SERVICE" = "all" ]; then
    echo -e "${YELLOW}Health check completed. Check individual service status above.${NC}"
else
    echo -e "${YELLOW}Health check for $SERVICE completed.${NC}"
fi
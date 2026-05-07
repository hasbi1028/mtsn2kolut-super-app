#!/bin/bash
# Health check script for mtsn2kolut-super-app services
# Usage: ./deploy/scripts/health-check.sh [service]
# service: backend | frontend | worker | bank-soal | all

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

is_frontend_ok_status() {
    local response="$1"
    [ "$response" -eq 200 ] || [ "$response" -eq 301 ] || [ "$response" -eq 302 ] || [ "$response" -eq 303 ] || [ "$response" -eq 307 ] || [ "$response" -eq 308 ]
}

# Function to check frontend health
check_frontend() {
    echo -n "Checking frontend... "
    if response=$(curl -s -m "$TIMEOUT" -w "%{http_code}" "http://localhost:8021" -o /dev/null); then
        if is_frontend_ok_status "$response"; then
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

# Function to smoke-check final Bank Soal page routes. Auth redirects are acceptable;
# 5xx/404 indicate a broken build, missing page, or route regression.
check_bank_soal_routes() {
    local routes=(
        "/bank-soal"
        "/bank-soal/daftar"
        "/bank-soal/tambah"
        "/bank-soal/verifikasi"
        "/bank-soal/impor"
        "/bank-soal/analisis-butir"
        "/bank-soal/mapel-kd"
        "/bank-soal/pengaturan"
    )

    echo "Checking Bank Soal routes..."
    for route in "${routes[@]}"; do
        echo -n "  ${route}... "
        if response=$(curl -s -m "$TIMEOUT" -w "%{http_code}" "http://localhost:8021${route}" -o /dev/null); then
            if is_frontend_ok_status "$response"; then
                echo -e "${GREEN}OK${NC} (HTTP $response)"
            else
                echo -e "${RED}FAIL${NC} (HTTP $response)"
                return 1
            fi
        else
            echo -e "${RED}FAIL${NC} (timeout or connection error)"
            return 1
        fi
    done
}

# Function to check worker health (via PM2)
check_worker() {
    echo -n "Checking worker... "
    if ! pm2 describe mtsn2kolut-pusaka-worker > /dev/null 2>&1; then
        echo -e "${RED}NOT FOUND${NC}"
        return 1
    fi
    if pm2 show mtsn2kolut-pusaka-worker 2>/dev/null | grep -q "status.*online"; then
        echo -e "${GREEN}OK${NC}"
        return 0
    fi
    echo -e "${YELLOW}NOT ONLINE${NC}"
    return 1
}

# Main execution
case "$SERVICE" in
    backend|all)
        check_backend || [[ "$SERVICE" == "all" ]] || exit 1
        ;;
    frontend|all)
        check_frontend || [[ "$SERVICE" == "all" ]] || exit 1
        ;;
    bank-soal)
        check_bank_soal_routes
        ;;
    worker|all)
        check_worker || [[ "$SERVICE" == "all" ]] || exit 1
        ;;
    *)
        echo "Usage: $0 [backend|frontend|worker|bank-soal|all]"
        exit 1
        ;;
esac

echo ""
if [ "$SERVICE" = "all" ]; then
    echo -e "${YELLOW}Health check completed. Check individual service status above.${NC}"
else
    echo -e "${YELLOW}Health check for $SERVICE completed.${NC}"
fi
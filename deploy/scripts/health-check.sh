#!/bin/bash
# Health check script for mtsn2kolut-super-app services
# Usage: ./deploy/scripts/health-check.sh [service]
# service: backend | frontend | worker | bank-soal | cbt-security | all

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

check_http_status() {
    local label="$1"
    local url="$2"
    local expected="$3"
    echo -n "  ${label}... "
    if response=$(curl -s -m "$TIMEOUT" -w "%{http_code}" "$url" -o /dev/null); then
        if [ "$response" -eq "$expected" ]; then
            echo -e "${GREEN}OK${NC} (HTTP $response, expected $expected)"
            return 0
        fi
        echo -e "${RED}FAIL${NC} (HTTP $response, expected $expected)"
        return 1
    fi
    echo -e "${RED}FAIL${NC} (timeout or connection error, expected $expected)"
    return 1
}


check_http_not_status() {
    local label="$1"
    local url="$2"
    local forbidden="$3"
    echo -n "  ${label}... "
    if response=$(curl -s -m "$TIMEOUT" -w "%{http_code}" "$url" -o /dev/null); then
        if [ "$response" -ne "$forbidden" ]; then
            echo -e "${GREEN}OK${NC} (HTTP $response, not public $forbidden)"
            return 0
        fi
        echo -e "${RED}FAIL${NC} (HTTP $response, public route regression)"
        return 1
    fi
    echo -e "${RED}FAIL${NC} (timeout or connection error, forbidden $forbidden)"
    return 1
}

# Focused CBT security smoke. It validates status codes only; no token/header/body
# output is printed, so the smoke remains safe for release evidence logs.
check_cbt_security_routes() {
    echo "Checking CBT security protected routes..."
    local failed=0
    local page_routes=(
        "/bank-soal"
        "/bank-soal/tambah"
        "/bank-soal/verifikasi"
        "/bank-soal/impor"
        "/bank-soal/analisis-butir"
        "/asesmen"
        "/asesmen/pengawasan"
        "/asesmen/hasil"
    )
    for route in "${page_routes[@]}"; do
        check_http_status "$route expected 302" "http://localhost:8021${route}" 302 || failed=1
    done
    check_http_status "/api/bank-soal/summary expected 401" "http://localhost:8021/api/bank-soal/summary" 401 || failed=1
    check_http_status "/api/exam/status expected 401" "http://localhost:8080/api/exam/status" 401 || failed=1
    check_http_not_status "/api/cbt/questions not public 200" "http://localhost:8021/api/cbt/questions" 200 || failed=1
    return "$failed"
}

# Function to check worker health (via PM2). A deeper heartbeat/API-key check is
# available in the web admin status UI, but this gate must at least fail if PM2
# cannot prove the supervised worker process is online.
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

run_all_checks() {
    local failed=0

    check_backend || failed=1
    check_frontend || failed=1
    check_worker || failed=1

    if [ "$failed" -ne 0 ]; then
        echo ""
        echo -e "${RED}Health check failed. See service status above.${NC}"
        return 1
    fi

    echo ""
    echo -e "${GREEN}All health checks passed.${NC}"
}

# Main execution
case "$SERVICE" in
    backend)
        check_backend
        ;;
    frontend)
        check_frontend
        ;;
    bank-soal)
        check_bank_soal_routes
        ;;
    cbt-security)
        check_cbt_security_routes
        ;;
    worker)
        check_worker
        ;;
    all)
        run_all_checks
        ;;
    *)
        echo "Usage: $0 [backend|frontend|worker|bank-soal|cbt-security|all]"
        exit 1
        ;;
esac

echo ""
if [ "$SERVICE" = "all" ]; then
    echo -e "${YELLOW}Health check completed.${NC}"
else
    echo -e "${YELLOW}Health check for $SERVICE completed.${NC}"
fi
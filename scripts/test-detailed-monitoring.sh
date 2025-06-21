#!/bin/bash

# Distributed Service - Enhanced Monitoring Test Script
# Tests all detailed monitoring endpoints and displays rich information

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Base URL
BASE_URL="http://localhost:8080"

# Print header
echo -e "${BLUE}============================================================================${NC}"
echo -e "${BLUE}            Distributed Service - Enhanced Monitoring Test${NC}"
echo -e "${BLUE}============================================================================${NC}"
echo ""

# Function to print section headers
print_section() {
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}📊 $1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Function to test endpoint and check response
test_endpoint() {
    local endpoint=$1
    local description=$2
    
    echo -e "${BLUE}Testing:${NC} $description"
    echo -e "${BLUE}URL:${NC} $BASE_URL$endpoint"
    
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$BASE_URL$endpoint")
    http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    body=$(echo $response | sed -e 's/HTTPSTATUS\:.*//g')
    
    if [ "$http_code" -eq 200 ]; then
        echo -e "${GREEN}✅ SUCCESS${NC} (HTTP $http_code)"
        return 0
    else
        echo -e "${RED}❌ FAILED${NC} (HTTP $http_code)"
        echo -e "${RED}Response:${NC} $body"
        return 1
    fi
}

# Test detailed service monitoring
print_section "DETAILED SERVICE MONITORING"

echo -e "${PURPLE}🔍 MySQL Connection Pool & Performance Details:${NC}"
curl -s "$BASE_URL/api/v1/monitor/services" | jq '.services[] | select(.name == "MySQL") | {
    name,
    status,
    latency,
    connection_pool: .details.connection_pool,
    mysql_version: .details.mysql_version,
    query_test: .details.query_test,
    dsn_info: .details.dsn_info
}' 2>/dev/null || echo "❌ Failed to get MySQL details"

echo ""
echo -e "${PURPLE}🔍 Redis Connection Pool & Stats:${NC}"
curl -s "$BASE_URL/api/v1/monitor/services" | jq '.services[] | select(.name == "Redis") | {
    name,
    status, 
    latency,
    connection_pool: .details.connection_pool,
    redis_info: .details.redis_info | {
        version: .redis_version,
        mode: .redis_mode,
        uptime_seconds: .uptime_in_seconds,
        connected_clients: .connected_clients,
        used_memory: .used_memory,
        total_commands: .total_commands_processed,
        keyspace_hits: .keyspace_hits,
        keyspace_misses: .keyspace_misses
    },
    read_write_tests: {
        write_test: .details.write_test,
        read_test: .details.read_test
    }
}' 2>/dev/null || echo "❌ Failed to get Redis details"

echo ""
echo -e "${PURPLE}🔍 RabbitMQ Connection & Queue Operations:${NC}"
curl -s "$BASE_URL/api/v1/monitor/services" | jq '.services[] | select(.name == "RabbitMQ") | {
    name,
    status,
    latency,
    connection_info: .details.connection_info,
    queue_operations: {
        queue_test: .details.queue_test,
        publish_test: .details.publish_test,
        test_queue: .details.test_queue
    }
}' 2>/dev/null || echo "❌ Failed to get RabbitMQ details"

echo ""
echo -e "${PURPLE}🔍 gRPC Connection State & Health Check:${NC}"
curl -s "$BASE_URL/api/v1/monitor/services" | jq '.services[] | select(.name == "gRPC") | {
    name,
    status,
    latency,
    connection_state: .details.connection_state,
    connection_info: .details.connection_info,
    health_check: {
        available: .details.health_check_available,
        error: .details.health_check_error
    }
}' 2>/dev/null || echo "❌ Failed to get gRPC details"

echo ""
echo -e "${PURPLE}🔍 Consul Connection Details:${NC}"
curl -s "$BASE_URL/api/v1/monitor/services" | jq '.services[] | select(.name == "Consul") | {
    name,
    status,
    latency,
    connection_string: .details.connection_string
}' 2>/dev/null || echo "❌ Failed to get Consul details"

# Test system resource monitoring
print_section "SYSTEM RESOURCE MONITORING"

echo -e "${PURPLE}🔍 CPU Usage (Per Core):${NC}"
curl -s "$BASE_URL/api/v1/monitor/system" | jq '.cpu.per_core[] | {
    core: ("Core " + (. | keys[0])),
    usage_percent: (. | values[0])
}' 2>/dev/null || echo "❌ Failed to get CPU details"

echo ""
echo -e "${PURPLE}🔍 Memory Usage Details:${NC}"
curl -s "$BASE_URL/api/v1/monitor/system" | jq '.memory | {
    total_gb: (.total / 1024 / 1024 / 1024 | floor * 100 / 100),
    used_gb: (.used / 1024 / 1024 / 1024 | floor * 100 / 100),
    available_gb: (.available / 1024 / 1024 / 1024 | floor * 100 / 100),
    usage_percent: .usage_percent,
    swap_total_gb: (.swap_total / 1024 / 1024 / 1024 | floor * 100 / 100),
    swap_used_gb: (.swap_used / 1024 / 1024 / 1024 | floor * 100 / 100)
}' 2>/dev/null || echo "❌ Failed to get memory details"

echo ""
echo -e "${PURPLE}🔍 Network Interfaces:${NC}"
curl -s "$BASE_URL/api/v1/monitor/system" | jq '.network.interfaces | to_entries[] | {
    interface: .key,
    bytes_sent: .value.bytes_sent,
    bytes_recv: .value.bytes_recv,
    packets_sent: .value.packets_sent,
    packets_recv: .value.packets_recv
}' 2>/dev/null || echo "❌ Failed to get network details"

# Test process monitoring
print_section "PROCESS MONITORING"

echo -e "${PURPLE}🔍 Current Process Details:${NC}"
curl -s "$BASE_URL/api/v1/monitor/process" | jq '{
    pid,
    cpu_percent,
    memory: {
        rss_mb: (.memory_rss / 1024 / 1024 | floor),
        vms_mb: (.memory_vms / 1024 / 1024 | floor)
    },
    threads,
    uptime_seconds,
    go_runtime: {
        goroutines: .runtime.num_goroutines,
        heap_alloc_mb: (.runtime.heap_alloc / 1024 / 1024 | floor),
        heap_sys_mb: (.runtime.heap_sys / 1024 / 1024 | floor),
        gc_runs: .runtime.num_gc
    }
}' 2>/dev/null || echo "❌ Failed to get process details"

# Test overall health
print_section "OVERALL HEALTH STATUS"

echo -e "${PURPLE}🔍 Service Health Summary:${NC}"
curl -s "$BASE_URL/api/v1/monitor/health" | jq '{
    overall_status: .status,
    timestamp: .timestamp,
    summary: .summary,
    service_details: [.services[] | {
        name,
        status,
        latency_ms: .latency,
        has_details: (.details != null)
    }]
}' 2>/dev/null || echo "❌ Failed to get health summary"

# Connection tests
print_section "CONNECTIVITY TESTS"

endpoints=(
    "/api/v1/monitor/system:System Statistics"
    "/api/v1/monitor/services:Service Health"
    "/api/v1/monitor/process:Process Statistics"
    "/api/v1/monitor/stats:Combined Statistics"
    "/api/v1/monitor/health:Health Check"
    "/monitor:Web Dashboard"
)

for endpoint_desc in "${endpoints[@]}"; do
    IFS=':' read -r endpoint description <<< "$endpoint_desc"
    test_endpoint "$endpoint" "$description"
    echo ""
done

# Performance summary
print_section "MONITORING PERFORMANCE SUMMARY"

echo -e "${PURPLE}🔍 Latest Response Times:${NC}"
curl -s "$BASE_URL/api/v1/monitor/services" | jq '.services[] | {
    service: .name,
    status: .status,
    latency_ms: .latency,
    performance: (
        if .latency < 5 then "🟢 Excellent"
        elif .latency < 20 then "🟡 Good"
        elif .latency < 50 then "🟠 Fair"
        else "🔴 Slow"
        end
    )
}' 2>/dev/null || echo "❌ Failed to get performance summary"

echo ""
echo -e "${GREEN}============================================================================${NC}"
echo -e "${GREEN}                    Enhanced Monitoring Test Complete!${NC}"
echo -e "${GREEN}============================================================================${NC}"
echo ""
echo -e "${YELLOW}📍 Access the web dashboard: ${NC}${BLUE}$BASE_URL/monitor${NC}"
echo -e "${YELLOW}📍 API documentation available at each endpoint${NC}"
echo ""
echo -e "${CYAN}Key Features Demonstrated:${NC}"
echo -e "  ${GREEN}✅${NC} MySQL connection pool statistics & query testing"
echo -e "  ${GREEN}✅${NC} Redis connection pool, server info & read/write testing"
echo -e "  ${GREEN}✅${NC} RabbitMQ connection state & queue operations testing"
echo -e "  ${GREEN}✅${NC} gRPC connection state & health check protocol"
echo -e "  ${GREEN}✅${NC} Consul connectivity verification"
echo -e "  ${GREEN}✅${NC} Detailed system resource monitoring (CPU, Memory, Network)"
echo -e "  ${GREEN}✅${NC} Process-level monitoring with Go runtime statistics"
echo -e "  ${GREEN}✅${NC} Comprehensive health status aggregation"
echo "" 
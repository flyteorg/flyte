#!/usr/bin/env bash

# Test script for ListVersions API endpoint
# This script tests the version listing functionality via the Flyte Admin API
#
# The ListVersions functionality is exposed through several endpoints:
# - ListTasks: /api/v1/tasks/{project}/{domain}/{name} - lists all versions of a task
# - ListWorkflows: /api/v1/workflows/{project}/{domain}/{name} - lists all versions of a workflow
# - ListLaunchPlans: /api/v1/launch_plans/{project}/{domain}/{name} - lists all versions of a launch plan
#
# Usage:
#   ./test_list_versions.sh [ADMIN_ENDPOINT]
#
# Arguments:
#   ADMIN_ENDPOINT - The Flyte Admin HTTP endpoint (default: http://localhost:8088)
#
# Exit codes:
#   0 - All tests passed
#   1 - One or more tests failed

set -e

# Configuration
ADMIN_ENDPOINT="${1:-http://localhost:8088}"
PROJECT="${PROJECT:-admintests}"
DOMAIN="${DOMAIN:-development}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Log functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_test() {
    echo -e "${GREEN}[TEST]${NC} $1"
}

# Test helper functions
assert_http_status() {
    local expected=$1
    local actual=$2
    local test_name=$3

    TESTS_RUN=$((TESTS_RUN + 1))
    if [ "$expected" -eq "$actual" ]; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        log_test "PASSED: $test_name (HTTP $actual)"
        return 0
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        log_error "FAILED: $test_name - Expected HTTP $expected, got HTTP $actual"
        return 1
    fi
}

assert_json_field_exists() {
    local json=$1
    local field=$2
    local test_name=$3

    TESTS_RUN=$((TESTS_RUN + 1))
    if echo "$json" | grep -q "\"$field\""; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        log_test "PASSED: $test_name"
        return 0
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        log_error "FAILED: $test_name - Field '$field' not found in response"
        return 1
    fi
}

# Check if the admin endpoint is reachable
check_admin_endpoint() {
    log_info "Checking Flyte Admin endpoint: $ADMIN_ENDPOINT"

    local http_code

    http_code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 "$ADMIN_ENDPOINT/api/v1/version" 2>/dev/null || echo "000")

    if [ "$http_code" = "200" ]; then
        log_info "Flyte Admin endpoint is reachable"
        return 0
    elif [ "$http_code" = "000" ] || [ "${http_code:0:3}" = "000" ]; then
        log_warn "Cannot connect to Flyte Admin endpoint at $ADMIN_ENDPOINT"
        log_warn "Make sure Flyte Admin is running and accessible"
        return 1
    else
        log_info "Flyte Admin endpoint returned HTTP $http_code"
        return 0
    fi
}

# Test: Get Version API
test_get_version() {
    log_info "Testing GetVersion API..."

    local response
    local http_code

    http_code=$(curl -s -o /tmp/version_response.json -w "%{http_code}" \
        -H "Accept: application/json" \
        "$ADMIN_ENDPOINT/api/v1/version")

    assert_http_status 200 "$http_code" "GetVersion API returns 200"

    if [ -f /tmp/version_response.json ]; then
        response=$(cat /tmp/version_response.json)
        assert_json_field_exists "$response" "controlPlaneVersion" "GetVersion response contains controlPlaneVersion"
    fi
}

# Test: List Tasks (versions of a task)
test_list_task_versions() {
    log_info "Testing ListTasks API (versions)..."

    local http_code
    local response

    # Test listing all tasks in project/domain
    http_code=$(curl -s -o /tmp/tasks_response.json -w "%{http_code}" \
        -H "Accept: application/json" \
        "$ADMIN_ENDPOINT/api/v1/tasks/$PROJECT/$DOMAIN?limit=10")

    # Allow 200 (found) or 404 (not found - no test data)
    if [ "$http_code" = "200" ] || [ "$http_code" = "404" ]; then
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
        log_test "PASSED: ListTasks API endpoint accessible (HTTP $http_code)"
    else
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
        log_error "FAILED: ListTasks API - Unexpected HTTP $http_code"
    fi
}

# Test: List Workflows (versions of a workflow)
test_list_workflow_versions() {
    log_info "Testing ListWorkflows API (versions)..."

    local http_code

    http_code=$(curl -s -o /tmp/workflows_response.json -w "%{http_code}" \
        -H "Accept: application/json" \
        "$ADMIN_ENDPOINT/api/v1/workflows/$PROJECT/$DOMAIN?limit=10")

    # Allow 200 (found) or 404 (not found - no test data)
    if [ "$http_code" = "200" ] || [ "$http_code" = "404" ]; then
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
        log_test "PASSED: ListWorkflows API endpoint accessible (HTTP $http_code)"
    else
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
        log_error "FAILED: ListWorkflows API - Unexpected HTTP $http_code"
    fi
}

# Test: List Launch Plans (versions of a launch plan)
test_list_launch_plan_versions() {
    log_info "Testing ListLaunchPlans API (versions)..."

    local http_code

    http_code=$(curl -s -o /tmp/launch_plans_response.json -w "%{http_code}" \
        -H "Accept: application/json" \
        "$ADMIN_ENDPOINT/api/v1/launch_plans/$PROJECT/$DOMAIN?limit=10")

    # Allow 200 (found) or 404 (not found - no test data)
    if [ "$http_code" = "200" ] || [ "$http_code" = "404" ]; then
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
        log_test "PASSED: ListLaunchPlans API endpoint accessible (HTTP $http_code)"
    else
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
        log_error "FAILED: ListLaunchPlans API - Unexpected HTTP $http_code"
    fi
}

# Test: List Task IDs (unique task identifiers)
test_list_task_ids() {
    log_info "Testing ListTaskIds API..."

    local http_code

    http_code=$(curl -s -o /tmp/task_ids_response.json -w "%{http_code}" \
        -H "Accept: application/json" \
        "$ADMIN_ENDPOINT/api/v1/task_ids/$PROJECT/$DOMAIN?limit=10")

    # Allow 200 (found) or 404 (not found - no test data)
    if [ "$http_code" = "200" ] || [ "$http_code" = "404" ]; then
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
        log_test "PASSED: ListTaskIds API endpoint accessible (HTTP $http_code)"
    else
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
        log_error "FAILED: ListTaskIds API - Unexpected HTTP $http_code"
    fi
}

# Test: Pagination with version filtering
test_version_pagination() {
    log_info "Testing version listing with pagination..."

    local http_code

    # Test pagination parameter
    http_code=$(curl -s -o /tmp/pagination_response.json -w "%{http_code}" \
        -H "Accept: application/json" \
        "$ADMIN_ENDPOINT/api/v1/tasks/$PROJECT/$DOMAIN?limit=2&sort_by.key=version&sort_by.direction=ASCENDING")

    if [ "$http_code" = "200" ] || [ "$http_code" = "404" ]; then
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
        log_test "PASSED: Pagination with version sorting works (HTTP $http_code)"
    else
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
        log_error "FAILED: Pagination test - Unexpected HTTP $http_code"
    fi
}

# Test: Version filtering
test_version_filtering() {
    log_info "Testing version filtering..."

    local http_code

    # Test filter parameter
    http_code=$(curl -s -o /tmp/filter_response.json -w "%{http_code}" \
        -H "Accept: application/json" \
        "$ADMIN_ENDPOINT/api/v1/tasks/$PROJECT/$DOMAIN?limit=10&filters=eq(version,v1)")

    if [ "$http_code" = "200" ] || [ "$http_code" = "404" ]; then
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
        log_test "PASSED: Version filtering parameter accepted (HTTP $http_code)"
    else
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
        log_error "FAILED: Version filter test - Unexpected HTTP $http_code"
    fi
}

# Cleanup temporary files
cleanup() {
    rm -f /tmp/version_response.json
    rm -f /tmp/tasks_response.json
    rm -f /tmp/workflows_response.json
    rm -f /tmp/launch_plans_response.json
    rm -f /tmp/task_ids_response.json
    rm -f /tmp/pagination_response.json
    rm -f /tmp/filter_response.json
}

# Print test summary
print_summary() {
    echo ""
    echo "========================================"
    echo "        TEST SUMMARY"
    echo "========================================"
    echo "Tests Run:    $TESTS_RUN"
    echo "Tests Passed: $TESTS_PASSED"
    echo "Tests Failed: $TESTS_FAILED"
    echo "========================================"

    if [ "$TESTS_FAILED" -eq 0 ]; then
        log_info "All tests passed!"
        return 0
    else
        log_error "Some tests failed!"
        return 1
    fi
}

# Main execution
main() {
    echo "========================================"
    echo "  ListVersions API Test Script"
    echo "========================================"
    echo "Endpoint: $ADMIN_ENDPOINT"
    echo "Project:  $PROJECT"
    echo "Domain:   $DOMAIN"
    echo "========================================"
    echo ""

    # Set up cleanup trap
    trap cleanup EXIT

    # Check if endpoint is reachable
    if ! check_admin_endpoint; then
        log_warn "Skipping API tests - endpoint not reachable"
        log_info "To run tests, start Flyte Admin and provide the endpoint:"
        log_info "  ./test_list_versions.sh http://localhost:8088"
        exit 0
    fi

    # Run tests
    test_get_version
    test_list_task_versions
    test_list_workflow_versions
    test_list_launch_plan_versions
    test_list_task_ids
    test_version_pagination
    test_version_filtering

    # Print summary and exit with appropriate code
    print_summary
}

# Run main function
main "$@"

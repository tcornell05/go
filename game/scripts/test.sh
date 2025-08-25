#!/bin/bash

# Comprehensive test runner with detailed output

echo "🧪 Running comprehensive test suite..."
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run tests for a package
run_package_tests() {
    local package=$1
    echo -e "${BLUE}Testing package: $package${NC}"
    
    # Run tests with verbose output
    if go test -v -cover $package 2>&1 | tee /tmp/test_output.txt; then
        PACKAGE_PASSED=$(grep -c "PASS:" /tmp/test_output.txt || echo 0)
        PASSED_TESTS=$((PASSED_TESTS + PACKAGE_PASSED))
        echo -e "${GREEN}✅ Package tests passed${NC}"
    else
        PACKAGE_FAILED=$(grep -c "FAIL:" /tmp/test_output.txt || echo 0)
        FAILED_TESTS=$((FAILED_TESTS + PACKAGE_FAILED))
        echo -e "${RED}❌ Package tests failed${NC}"
    fi
    
    PACKAGE_TOTAL=$(grep -c "RUN" /tmp/test_output.txt || echo 0)
    TOTAL_TESTS=$((TOTAL_TESTS + PACKAGE_TOTAL))
    
    echo ""
}

# Run tests for all packages
for package in $(go list ./...); do
    run_package_tests $package
done

# Run integration tests if they exist
if [ -f "test/integration_test.go" ]; then
    echo -e "${BLUE}Running integration tests...${NC}"
    if go test -v ./test/...; then
        echo -e "${GREEN}✅ Integration tests passed${NC}"
    else
        echo -e "${RED}❌ Integration tests failed${NC}"
    fi
    echo ""
fi

# Generate coverage report
echo -e "${YELLOW}📊 Generating coverage report...${NC}"
go test -coverprofile=coverage.out ./... >/dev/null 2>&1
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo -e "Total coverage: ${BLUE}$COVERAGE${NC}"
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${BLUE}Test Summary${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "Total tests:  $TOTAL_TESTS"
echo -e "Passed:       ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed:       ${RED}$FAILED_TESTS${NC}"
echo -e "Coverage:     ${BLUE}$COVERAGE${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}⚠️ Some tests failed. Please fix the issues.${NC}"
    exit 1
fi
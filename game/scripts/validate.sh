#!/bin/bash

# Comprehensive code validation script

echo "📝 Running comprehensive code validation..."
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Exit codes
EXIT_CODE=0

# Function to run check and update exit code
run_check() {
    local name=$1
    local command=$2
    
    echo -e "${BLUE}→ $name${NC}"
    
    if eval "$command"; then
        echo -e "${GREEN}  ✅ Passed${NC}"
    else
        echo -e "${RED}  ❌ Failed${NC}"
        EXIT_CODE=1
    fi
    echo ""
}

# 1. Go formatting check
run_check "Checking Go formatting" "gofmt -l . | (! grep .)"

# 2. Go vet
run_check "Running go vet" "go vet ./..."

# 3. Build check
run_check "Checking build compilation" "go build -o /dev/null ./..."

# 4. Module verification
run_check "Verifying module dependencies" "go mod verify"

# 5. Module tidiness
run_check "Checking module tidiness" "go mod tidy && git diff --exit-code go.mod go.sum"

# 6. Test compilation
run_check "Checking test compilation" "go test -c ./... -o /dev/null"

# 7. Security check (if gosec is available)
if command -v gosec >/dev/null 2>&1; then
    run_check "Running security analysis (gosec)" "gosec ./..."
fi

# 8. Inefficiency check (if ineffassign is available)
if command -v ineffassign >/dev/null 2>&1; then
    run_check "Checking for ineffective assignments" "ineffassign ./..."
fi

# 9. Deadcode check (if deadcode is available)
if command -v deadcode >/dev/null 2>&1; then
    run_check "Checking for dead code" "deadcode ./..."
fi

# 10. Check for TODOs and FIXMEs
echo -e "${BLUE}→ Checking for TODO/FIXME comments${NC}"
TODO_COUNT=$(grep -r "TODO\|FIXME" --include="*.go" . | wc -l)
if [ $TODO_COUNT -gt 0 ]; then
    echo -e "${YELLOW}  ⚠️ Found $TODO_COUNT TODO/FIXME comments${NC}"
    grep -r "TODO\|FIXME" --include="*.go" . | head -10
    if [ $TODO_COUNT -gt 10 ]; then
        echo -e "  ... and $((TODO_COUNT - 10)) more"
    fi
else
    echo -e "${GREEN}  ✅ No TODO/FIXME comments found${NC}"
fi
echo ""

# 11. Check for common Go anti-patterns
echo -e "${BLUE}→ Checking for common issues${NC}"
ISSUES_FOUND=0

# Check for fmt.Print in non-main packages
NON_MAIN_PRINTS=$(grep -r "fmt\.Print" --include="*.go" ./internal ./pkg 2>/dev/null | wc -l)
if [ $NON_MAIN_PRINTS -gt 0 ]; then
    echo -e "${YELLOW}  ⚠️ Found fmt.Print in non-main packages (consider using proper logging)${NC}"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

# Check for empty catch blocks
EMPTY_CATCHES=$(grep -rE "if.*err.*!=.*nil.*\{\s*\}" --include="*.go" . | wc -l)
if [ $EMPTY_CATCHES -gt 0 ]; then
    echo -e "${YELLOW}  ⚠️ Found potential empty error handling${NC}"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

if [ $ISSUES_FOUND -eq 0 ]; then
    echo -e "${GREEN}  ✅ No common issues found${NC}"
fi
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${BLUE}Validation Summary${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}🎉 All validation checks passed!${NC}"
    echo "Your code is ready for development and deployment."
else
    echo -e "${RED}⚠️ Some validation checks failed.${NC}"
    echo "Please fix the issues above before proceeding."
fi

exit $EXIT_CODE
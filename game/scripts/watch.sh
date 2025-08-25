#!/bin/bash

# Watch for changes and rebuild automatically

echo "👁️ Watching for file changes..."
echo "📁 Directory: $(pwd)"
echo "Press Ctrl+C to stop"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Function to check and report errors
check_build() {
    echo -e "${YELLOW}🔍 Checking build...${NC}"
    
    # Capture build output
    BUILD_OUTPUT=$(go build -o /dev/null ./... 2>&1)
    BUILD_EXIT=$?
    
    if [ $BUILD_EXIT -eq 0 ]; then
        echo -e "${GREEN}✅ Build successful - no errors${NC}"
    else
        echo -e "${RED}❌ Build errors found:${NC}"
        echo "$BUILD_OUTPUT" | head -30
        
        # Count errors
        ERROR_COUNT=$(echo "$BUILD_OUTPUT" | grep -c "error:")
        if [ $ERROR_COUNT -gt 0 ]; then
            echo -e "${RED}Total errors: $ERROR_COUNT${NC}"
        fi
    fi
    
    # Check for formatting issues
    UNFORMATTED=$(gofmt -l .)
    if [ ! -z "$UNFORMATTED" ]; then
        echo -e "${YELLOW}⚠️ Files need formatting:${NC}"
        echo "$UNFORMATTED"
    fi
    
    echo "---"
    echo "Waiting for changes..."
}

# Initial check
check_build

# Watch for changes
if command -v fswatch >/dev/null 2>&1; then
    fswatch -o -e ".*" -i "\\.go$" -i "\\.mod$" . | while read change; do
        check_build
    done
elif command -v inotifywait >/dev/null 2>&1; then
    while true; do
        inotifywait -r -e modify,create,delete --include '.*\.go$|.*\.mod$' . >/dev/null 2>&1
        check_build
    done
else
    # Fallback to polling
    echo "Using polling mode (every 2 seconds)..."
    LAST_MOD=""
    while true; do
        CURRENT_MOD=$(find . -name "*.go" -o -name "go.mod" -type f -exec stat -c %Y {} \; 2>/dev/null | sort -n | tail -1)
        if [ "$CURRENT_MOD" != "$LAST_MOD" ]; then
            LAST_MOD=$CURRENT_MOD
            check_build
        fi
        sleep 2
    done
fi
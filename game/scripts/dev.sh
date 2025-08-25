#!/bin/bash

# Development server with hot reload
# Watches for Go file changes and automatically rebuilds/restarts

echo "🔥 Starting development server with hot reload..."
echo "📁 Watching directory: $(pwd)"
echo "Press Ctrl+C to stop"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to build and run
build_and_run() {
    echo -e "${YELLOW}🔄 Rebuilding...${NC}"
    
    # Kill existing game process if running
    if [ ! -z "$GAME_PID" ] && kill -0 $GAME_PID 2>/dev/null; then
        echo "Stopping previous instance..."
        kill $GAME_PID 2>/dev/null
        wait $GAME_PID 2>/dev/null
    fi
    
    # Build the game
    if go build -o game .; then
        echo -e "${GREEN}✅ Build successful${NC}"
        echo "Starting game..."
        ./game &
        GAME_PID=$!
        echo -e "${GREEN}🎮 Game running (PID: $GAME_PID)${NC}"
    else
        echo -e "${RED}❌ Build failed - fix errors and save to retry${NC}"
    fi
    echo "---"
}

# Initial build
build_and_run

# Watch for changes using fswatch if available, otherwise use a loop
if command -v fswatch >/dev/null 2>&1; then
    echo "Using fswatch for file monitoring..."
    fswatch -o -e ".*" -i "\\.go$" -i "\\.mod$" . | while read change; do
        build_and_run
    done
elif command -v inotifywait >/dev/null 2>&1; then
    echo "Using inotifywait for file monitoring..."
    while true; do
        inotifywait -r -e modify,create,delete --include '.*\.go$|.*\.mod$' .
        build_and_run
    done
else
    echo "No file watcher found. Using polling mode (checking every 2 seconds)..."
    LAST_MOD=""
    while true; do
        CURRENT_MOD=$(find . -name "*.go" -o -name "go.mod" -type f -exec stat -c %Y {} \; | sort -n | tail -1)
        if [ "$CURRENT_MOD" != "$LAST_MOD" ]; then
            LAST_MOD=$CURRENT_MOD
            build_and_run
        fi
        sleep 2
    done
fi

# Cleanup on exit
trap "kill $GAME_PID 2>/dev/null; exit" INT TERM
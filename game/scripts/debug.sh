#!/bin/bash

# Quick debugging helper script

echo "🐛 Game Debugging Helper"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Show current status
echo -e "${BLUE}Current Status:${NC}"
echo "Working directory: $(pwd)"
echo "Git branch: $(git branch --show-current 2>/dev/null || echo 'Not a git repo')"
echo "Go version: $(go version)"
echo ""

# Quick build check
echo -e "${BLUE}Quick Build Check:${NC}"
if go build -o /dev/null . 2>/dev/null; then
    echo -e "${GREEN}✅ Code compiles successfully${NC}"
else
    echo -e "${RED}❌ Compilation errors:${NC}"
    go build -o /dev/null . 2>&1 | head -10
fi
echo ""

# Check if game binary exists
echo -e "${BLUE}Binary Status:${NC}"
if [ -f "game" ]; then
    SIZE=$(du -h game | cut -f1)
    echo -e "${GREEN}✅ Game binary exists ($SIZE)${NC}"
    echo "Modified: $(stat -c %y game 2>/dev/null || stat -f %Sm game)"
else
    echo -e "${YELLOW}⚠️ No game binary found (run 'make build')${NC}"
fi
echo ""

# Database status
echo -e "${BLUE}Database Status:${NC}"
if [ -f "data/game.db" ]; then
    SIZE=$(du -h data/game.db | cut -f1)
    echo -e "${GREEN}✅ Database exists ($SIZE)${NC}"
    
    # Check if sqlite3 is available to show table info
    if command -v sqlite3 >/dev/null 2>&1; then
        echo "Tables:"
        sqlite3 data/game.db ".tables" | sed 's/^/  /'
    fi
else
    echo -e "${YELLOW}⚠️ No database found (will be created on first run)${NC}"
fi
echo ""

# Asset summary
echo -e "${BLUE}Asset Summary:${NC}"
if [ -d "assets" ]; then
    PNG_COUNT=$(find assets -name "*.png" -type f | wc -l)
    TOTAL_SIZE=$(du -sh assets | cut -f1)
    echo -e "${GREEN}✅ Assets directory exists${NC}"
    echo "  PNG files: $PNG_COUNT"
    echo "  Total size: $TOTAL_SIZE"
    
    # Show recent changes
    echo "  Recent files:"
    find assets -name "*.png" -type f -printf "%TY-%Tm-%Td %TH:%TM %p\n" 2>/dev/null | sort -r | head -5 | sed 's/^/    /'
else
    echo -e "${RED}❌ Assets directory not found${NC}"
fi
echo ""

# Memory and performance info
echo -e "${BLUE}System Info:${NC}"
echo "Memory: $(free -h | awk '/Mem:/ {print $3 "/" $2}')"
echo "CPU: $(nproc) cores"
echo "Disk space: $(df -h . | awk 'NR==2 {print $4 " available"}')"
echo ""

# Common commands
echo -e "${BLUE}Quick Commands:${NC}"
echo "  make run        - Run the game"
echo "  make dev        - Start development server"
echo "  make test       - Run tests"
echo "  make validate   - Check code quality"
echo "  make clean      - Clean build files"
echo ""

# Show recent git activity if in git repo
if git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "${BLUE}Recent Git Activity:${NC}"
    git log --oneline -5 | sed 's/^/  /'
fi
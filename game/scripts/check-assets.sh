#!/bin/bash

# Asset validation and checking script

echo "🎨 Validating game assets..."
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
TOTAL_ASSETS=0
VALID_ASSETS=0
MISSING_ASSETS=0
INVALID_ASSETS=0

# Asset directories
ASSET_DIR="assets"

# Check if assets directory exists
if [ ! -d "$ASSET_DIR" ]; then
    echo -e "${RED}❌ Assets directory not found!${NC}"
    exit 1
fi

# Function to validate image file
validate_image() {
    local file=$1
    
    if [ ! -f "$file" ]; then
        echo -e "${RED}  ❌ Missing: $file${NC}"
        MISSING_ASSETS=$((MISSING_ASSETS + 1))
        return 1
    fi
    
    # Check if file is a valid image (basic check)
    if file "$file" | grep -q "image"; then
        VALID_ASSETS=$((VALID_ASSETS + 1))
        return 0
    else
        echo -e "${YELLOW}  ⚠️ Invalid image: $file${NC}"
        INVALID_ASSETS=$((INVALID_ASSETS + 1))
        return 1
    fi
}

# Check character sprites
echo -e "${BLUE}Checking character sprites...${NC}"
CHARACTER_SPRITES=(
    "Character/character_e.png"
    "Character/character_n.png"
    "Character/character_ne.png"
    "Character/character_nw.png"
    "Character/character_s.png"
    "Character/character_se.png"
    "Character/character_sw.png"
    "Character/character_w.png"
)

for sprite in "${CHARACTER_SPRITES[@]}"; do
    TOTAL_ASSETS=$((TOTAL_ASSETS + 1))
    validate_image "$ASSET_DIR/$sprite"
done

# Check floor tiles
echo -e "${BLUE}Checking floor tiles...${NC}"
for i in {1..17}; do
    TOTAL_ASSETS=$((TOTAL_ASSETS + 1))
    validate_image "$ASSET_DIR/Foor-Wall Tiles 64px/Floor_${i}_Tile(64).png"
done

# Check wall tiles
echo -e "${BLUE}Checking wall tiles...${NC}"
for file in "$ASSET_DIR"/Foor-Wall\ Tiles\ 64px/Wall_*.png; do
    if [ -f "$file" ]; then
        TOTAL_ASSETS=$((TOTAL_ASSETS + 1))
        validate_image "$file"
    fi
done

# Check furniture assets
echo -e "${BLUE}Checking furniture assets...${NC}"
FURNITURE_DIRS=("Chair" "Desk" "Sofa" "Bed" "Kitchen" "Bathroom")

for dir in "${FURNITURE_DIRS[@]}"; do
    if [ -d "$ASSET_DIR/$dir" ]; then
        for file in "$ASSET_DIR/$dir"/*.png; do
            if [ -f "$file" ]; then
                TOTAL_ASSETS=$((TOTAL_ASSETS + 1))
                validate_image "$file"
            fi
        done
    fi
done

# List all PNG files and their sizes
echo ""
echo -e "${BLUE}Asset inventory:${NC}"
find "$ASSET_DIR" -name "*.png" -type f | while read file; do
    SIZE=$(du -h "$file" | cut -f1)
    echo "  $file ($SIZE)"
done | head -20
echo "  ... and more"

# Count total assets
TOTAL_PNG_COUNT=$(find "$ASSET_DIR" -name "*.png" -type f | wc -l)

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${BLUE}Asset Validation Summary${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "Total PNG files:    $TOTAL_PNG_COUNT"
echo -e "Validated:          $TOTAL_ASSETS"
echo -e "Valid:              ${GREEN}$VALID_ASSETS${NC}"
echo -e "Missing:            ${RED}$MISSING_ASSETS${NC}"
echo -e "Invalid:            ${YELLOW}$INVALID_ASSETS${NC}"

if [ $MISSING_ASSETS -eq 0 ] && [ $INVALID_ASSETS -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ All assets validated successfully!${NC}"
    exit 0
else
    echo ""
    echo -e "${YELLOW}⚠️ Some assets have issues. Please check the output above.${NC}"
    exit 1
fi
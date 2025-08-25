package ui

import (
	"fmt"
	"image/color"
	"math"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tcornell05/go/game/internal/database"
	"github.com/tcornell05/go/game/internal/room"
	"github.com/tcornell05/go/game/pkg/assets"
)

// UIRenderer handles drawing the editor UI
type UIRenderer struct {
	inventory *TileInventory
}

// NewUIRenderer creates a new UI renderer
func NewUIRenderer(inventory *TileInventory) *UIRenderer {
	return &UIRenderer{
		inventory: inventory,
	}
}

// DrawSidebar draws the tile inventory sidebar
func (ui *UIRenderer) DrawSidebar(screen *ebiten.Image, screenWidth, screenHeight int) {
	if ui.inventory == nil {
		return
	}
	if !ui.inventory.SidebarVisible {
		return
	}

	sidebarWidth := float32(ui.inventory.SidebarWidth)
	sidebarHeight := float32(screenHeight)

	// Draw sidebar background
	vector.DrawFilledRect(screen, 0, 0, sidebarWidth, sidebarHeight, color.RGBA{40, 40, 40, 220}, false)
	
	// Draw sidebar border
	vector.StrokeRect(screen, 0, 0, sidebarWidth, sidebarHeight, 2, color.RGBA{100, 100, 100, 255}, false)

	// Draw title
	ebitenutil.DebugPrintAt(screen, "TILE INVENTORY", 10, 10)
	ebitenutil.DebugPrintAt(screen, "Keys: Tab/B/Space | R to rotate", 10, 30)

	currentY := 60
	tileSize := 48
	tilesPerRow := 4
	spacing := 10

	// Draw categories and tiles
	if len(ui.inventory.Categories) == 0 {
		ebitenutil.DebugPrintAt(screen, "No categories loaded", 10, 100)
		return
	}
	
	for _, category := range ui.inventory.Categories {
		// Draw category header
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("--- %s ---", category.Name), 10, currentY)
		currentY += 20

		// Draw tiles in grid
		for i, tile := range category.Tiles {
			x := (i % tilesPerRow) * (tileSize + spacing) + 10
			y := (i / tilesPerRow) * (tileSize + spacing) + currentY

			// Draw tile background
			bgColor := color.RGBA{60, 60, 60, 255}
			if ui.inventory.SelectedTile == &tile {
				bgColor = color.RGBA{100, 150, 100, 255} // Highlight selected
			}
			vector.DrawFilledRect(screen, float32(x), float32(y), float32(tileSize), float32(tileSize), bgColor, false)

			// Draw tile image (scaled down)
			if tile.Image != nil {
				op := &ebiten.DrawImageOptions{}
				
				// Scale to fit tile slot
				bounds := tile.Image.Bounds()
				scaleX := float64(tileSize) / float64(bounds.Dx())
				scaleY := float64(tileSize) / float64(bounds.Dy())
				scale := scaleX
				if scaleY < scaleX {
					scale = scaleY
				}
				
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(float64(x), float64(y))
				
				screen.DrawImage(tile.Image, op)
			}

			// Draw border
			vector.StrokeRect(screen, float32(x), float32(y), float32(tileSize), float32(tileSize), 1, color.RGBA{120, 120, 120, 255}, false)
		}

		// Update Y position for next category
		rows := (len(category.Tiles) + tilesPerRow - 1) / tilesPerRow
		currentY += rows*(tileSize+spacing) + 20
	}

	// Draw selected tile info at bottom
	if ui.inventory.SelectedTile != nil {
		infoY := screenHeight - 60
		tileName, rotation := ui.inventory.GetSelectedTileInfo()
		rotationText := "Normal"
		if rotation == 1 {
			rotationText = "Flipped"
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Selected: %s", tileName), 10, infoY)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rotation: %s", rotationText), 10, infoY+15)
		ebitenutil.DebugPrintAt(screen, "Left click to place | Right click to delete", 10, infoY+30)
	}
}

// HandleSidebarClick handles clicks on the sidebar for tile selection
func (ui *UIRenderer) HandleSidebarClick(mouseX, mouseY int) bool {
	if !ui.inventory.SidebarVisible {
		return false
	}

	if mouseX > ui.inventory.SidebarWidth {
		return false
	}

	currentY := 60
	tileSize := 48
	tilesPerRow := 4
	spacing := 10

	// Check clicks on tiles
	for catIdx := range ui.inventory.Categories {
		category := &ui.inventory.Categories[catIdx]
		currentY += 20 // Category header space

		for i := range category.Tiles {
			x := (i % tilesPerRow) * (tileSize + spacing) + 10
			y := (i / tilesPerRow) * (tileSize + spacing) + currentY

			if mouseX >= x && mouseX <= x+tileSize && mouseY >= y && mouseY <= y+tileSize {
				ui.inventory.SelectTile(&category.Tiles[i])
				return true
			}
		}

		// Update Y position for next category
		rows := (len(category.Tiles) + tilesPerRow - 1) / tilesPerRow
		currentY += rows*(tileSize+spacing) + 20
	}

	return true // Clicked in sidebar area, consume the click
}

// DrawPositionDialog draws the asset positioning dialog
func (ui *UIRenderer) DrawPositionDialog(screen *ebiten.Image, screenWidth, screenHeight int) {
	if !ui.inventory.PositionDialog {
		return
	}
	
	fmt.Printf("Drawing position dialog, SelectedTile: %v\n", ui.inventory.SelectedTile != nil)

	// Dialog dimensions
	dialogWidth := 400.0
	dialogHeight := 350.0
	dialogX := float32(screenWidth)/2 - float32(dialogWidth)/2
	dialogY := float32(screenHeight)/2 - float32(dialogHeight)/2

	// Draw dialog background
	vector.DrawFilledRect(screen, dialogX, dialogY, float32(dialogWidth), float32(dialogHeight), color.RGBA{30, 30, 30, 240}, false)
	vector.StrokeRect(screen, dialogX, dialogY, float32(dialogWidth), float32(dialogHeight), 2, color.RGBA{120, 120, 120, 255}, false)

	// Dialog title
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Position: %s", ui.inventory.SelectedTile.Name), int(dialogX)+10, int(dialogY)+10)
	ebitenutil.DebugPrintAt(screen, "Press P to close", int(dialogX)+10, int(dialogY)+25)

	// Draw isometric tile shape in center
	tileSize := 120.0
	tileCenterX := float64(dialogX) + dialogWidth/2
	tileCenterY := float64(dialogY) + 120

	ui.drawIsometricTile(screen, tileCenterX, tileCenterY, tileSize)

	// Draw snap points
	ui.drawSnapPoints(screen, tileCenterX, tileCenterY, tileSize)

	// Draw current asset position
	ui.drawAssetPosition(screen, tileCenterX, tileCenterY, tileSize)

	// Draw position info
	infoY := int(dialogY) + 250
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Offset: (%.2f, %.2f)", ui.inventory.SelectedPosition.OffsetX, ui.inventory.SelectedPosition.OffsetY), int(dialogX)+10, infoY)
	
	snapNames := []string{"Manual", "Top", "Top-Right", "Right", "Bottom-Right", "Bottom", "Bottom-Left", "Left", "Top-Left", "Center"}
	snapName := "Unknown"
	if ui.inventory.SelectedPosition.SnapToEdge >= 0 && ui.inventory.SelectedPosition.SnapToEdge < len(snapNames) {
		snapName = snapNames[ui.inventory.SelectedPosition.SnapToEdge]
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Snap: %s", snapName), int(dialogX)+10, infoY+15)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Depth: %.2f", ui.inventory.SelectedPosition.Depth), int(dialogX)+10, infoY+30)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rotation: %.1f°", ui.inventory.SelectedPosition.Rotation*180/math.Pi), int(dialogX)+10, infoY+45)

	// Controls help
	ebitenutil.DebugPrintAt(screen, "Drag the red asset to position | Click snap points", int(dialogX)+10, infoY+60)
	ebitenutil.DebugPrintAt(screen, "Arrow keys: depth | 0-9: Quick snap positions", int(dialogX)+10, infoY+75)
}

// drawIsometricTile draws an isometric diamond shape representing a tile
func (ui *UIRenderer) drawIsometricTile(screen *ebiten.Image, centerX, centerY, size float64) {
	// Draw isometric diamond
	halfSize := size / 2
	
	// Isometric diamond points
	topX, topY := centerX, centerY-halfSize/2
	rightX, rightY := centerX+halfSize, centerY
	bottomX, bottomY := centerX, centerY+halfSize/2
	leftX, leftY := centerX-halfSize, centerY

	// Draw diamond using lines
	diamondColor := color.RGBA{80, 80, 80, 255}
	borderColor := color.RGBA{150, 150, 150, 255}
	
	// Fill diamond (simple approach using multiple horizontal lines)
	for y := int(topY); y <= int(bottomY); y++ {
		// Calculate left and right bounds for this Y
		progress := float64(y-int(topY)) / float64(int(bottomY)-int(topY))
		if progress <= 0.5 {
			// Upper half
			halfWidth := progress * halfSize
			leftX := centerX - halfWidth
			rightX := centerX + halfWidth
			vector.StrokeLine(screen, float32(leftX), float32(y), float32(rightX), float32(y), 1, diamondColor, false)
		} else {
			// Lower half
			halfWidth := (1.0 - progress) * halfSize
			leftX := centerX - halfWidth
			rightX := centerX + halfWidth
			vector.StrokeLine(screen, float32(leftX), float32(y), float32(rightX), float32(y), 1, diamondColor, false)
		}
	}
	
	// Draw diamond border
	vector.StrokeLine(screen, float32(topX), float32(topY), float32(rightX), float32(rightY), 2, borderColor, false)
	vector.StrokeLine(screen, float32(rightX), float32(rightY), float32(bottomX), float32(bottomY), 2, borderColor, false)
	vector.StrokeLine(screen, float32(bottomX), float32(bottomY), float32(leftX), float32(leftY), 2, borderColor, false)
	vector.StrokeLine(screen, float32(leftX), float32(leftY), float32(topX), float32(topY), 2, borderColor, false)
}

// drawSnapPoints draws clickable snap points on the isometric tile
func (ui *UIRenderer) drawSnapPoints(screen *ebiten.Image, centerX, centerY, size float64) {
	halfSize := size / 2
	snapRadius := 8.0
	
	// Snap point positions - 9 total positions around the isometric tile
	snapPoints := []struct {
		x, y float64
		edge int
		name string
	}{
		{centerX, centerY - halfSize/2, 1, "T"},                    // Top edge
		{centerX + halfSize*0.5, centerY - halfSize/4, 2, "TR"},   // Top-right corner
		{centerX + halfSize*0.7, centerY, 3, "R"},                 // Right edge
		{centerX + halfSize*0.5, centerY + halfSize/4, 4, "BR"},   // Bottom-right corner
		{centerX, centerY + halfSize/2, 5, "B"},                   // Bottom edge
		{centerX - halfSize*0.5, centerY + halfSize/4, 6, "BL"},   // Bottom-left corner
		{centerX - halfSize*0.7, centerY, 7, "L"},                 // Left edge
		{centerX - halfSize*0.5, centerY - halfSize/4, 8, "TL"},   // Top-left corner
		{centerX, centerY, 9, "C"},                                // Center
	}
	
	for _, point := range snapPoints {
		// Highlight if selected
		pointColor := color.RGBA{100, 150, 200, 255}
		if ui.inventory.SelectedPosition.SnapToEdge == point.edge {
			pointColor = color.RGBA{200, 255, 100, 255}
		}
		
		// Draw snap point circle (filled rectangle as approximation)
		vector.DrawFilledRect(screen, float32(point.x-snapRadius), float32(point.y-snapRadius), float32(snapRadius*2), float32(snapRadius*2), pointColor, false)
		vector.StrokeRect(screen, float32(point.x-snapRadius), float32(point.y-snapRadius), float32(snapRadius*2), float32(snapRadius*2), 2, color.RGBA{255, 255, 255, 255}, false)
		
		// Draw label
		ebitenutil.DebugPrintAt(screen, point.name, int(point.x)-3, int(point.y)-6)
	}
}

// drawAssetPosition draws a draggable representation of where the asset will be placed
func (ui *UIRenderer) drawAssetPosition(screen *ebiten.Image, centerX, centerY, size float64) {
	// Standard isometric positioning: convert (0,0)-(1,1) offset to screen coordinates
	// (0,0) = northwest corner, (1,1) = southeast corner of tile
	halfSize := size / 2
	// Convert offset to relative position within tile: (0,0) -> (-0.5,-0.5), (1,1) -> (0.5,0.5)
	relativeX := ui.inventory.SelectedPosition.OffsetX - 0.5  // 0->-0.5, 0.5->0, 1->0.5
	relativeY := ui.inventory.SelectedPosition.OffsetY - 0.5  // 0->-0.5, 0.5->0, 1->0.5
	assetX := centerX + relativeX * halfSize
	assetY := centerY + relativeY * halfSize * 0.5 // Isometric Y scaling
	
	// Draw larger, more visible asset placeholder that can be dragged
	assetSize := 20.0
	
	// Highlight if this is being dragged or is draggable
	assetColor := color.RGBA{255, 100, 100, 255}
	borderColor := color.RGBA{255, 255, 255, 255}
	
	// Change colors if currently dragging
	if ui.inventory.DraggingAsset {
		assetColor = color.RGBA{100, 255, 100, 255} // Green when dragging
		borderColor = color.RGBA{255, 255, 100, 255} // Yellow border when dragging
	}
	
	// Draw asset as a diamond shape to match isometric style
	points := []struct{ x, y float64 }{
		{assetX, assetY - assetSize/2},     // Top
		{assetX + assetSize/2, assetY},     // Right  
		{assetX, assetY + assetSize/2},     // Bottom
		{assetX - assetSize/2, assetY},     // Left
	}
	
	// Fill the diamond
	for i := 0; i < len(points); i++ {
		next := (i + 1) % len(points)
		// Draw lines to create diamond fill (simplified)
		vector.StrokeLine(screen, float32(points[i].x), float32(points[i].y), 
			float32(points[next].x), float32(points[next].y), 2, assetColor, false)
	}
	
	// Draw center dot for precise positioning
	vector.DrawFilledRect(screen, float32(assetX-2), float32(assetY-2), 4, 4, borderColor, false)
	
	// Draw rotation indicator (small line)
	if ui.inventory.SelectedPosition.Rotation != 0 {
		lineLength := 25.0
		endX := assetX + math.Cos(ui.inventory.SelectedPosition.Rotation) * lineLength
		endY := assetY + math.Sin(ui.inventory.SelectedPosition.Rotation) * lineLength
		
		vector.StrokeLine(screen, float32(assetX), float32(assetY), float32(endX), float32(endY), 3, color.RGBA{255, 255, 100, 255}, false)
	}
	
	// Draw coordinate info next to asset
	coordText := fmt.Sprintf("(%.2f, %.2f)", ui.inventory.SelectedPosition.OffsetX, ui.inventory.SelectedPosition.OffsetY)
	ebitenutil.DebugPrintAt(screen, coordText, int(assetX)+15, int(assetY)-10)
}

// HandlePositionDialogClick handles clicks within the positioning dialog
func (ui *UIRenderer) HandlePositionDialogClick(mouseX, mouseY int, screenWidth, screenHeight int) bool {
	if !ui.inventory.PositionDialog {
		return false
	}

	dialogWidth := 400.0
	dialogHeight := 350.0
	dialogX := float64(screenWidth)/2 - dialogWidth/2
	dialogY := float64(screenHeight)/2 - dialogHeight/2

	// Check if click is within dialog
	if float64(mouseX) < dialogX || float64(mouseX) > dialogX+dialogWidth ||
	   float64(mouseY) < dialogY || float64(mouseY) > dialogY+dialogHeight {
		return false
	}

	tileCenterX := dialogX + dialogWidth/2
	tileCenterY := dialogY + 120
	tileSize := 120.0
	halfSize := tileSize / 2
	
	// Check if clicking on the asset itself for dragging
	// Standard isometric positioning: convert (0,0)-(1,1) offset to screen coordinates
	relativeX := ui.inventory.SelectedPosition.OffsetX - 0.5  // 0->-0.5, 0.5->0, 1->0.5
	relativeY := ui.inventory.SelectedPosition.OffsetY - 0.5  // 0->-0.5, 0.5->0, 1->0.5
	assetX := tileCenterX + relativeX * halfSize
	assetY := tileCenterY + relativeY * halfSize * 0.5
	assetRadius := 15.0 // Larger click area for easier dragging
	
	dx := float64(mouseX) - assetX
	dy := float64(mouseY) - assetY
	if math.Sqrt(dx*dx + dy*dy) <= assetRadius {
		// Start dragging the asset
		ui.inventory.DraggingAsset = true
		ui.inventory.DragStartX = float64(mouseX)
		ui.inventory.DragStartY = float64(mouseY)
		return true
	}
	
	// Check snap point clicks
	snapRadius := 8.0
	snapPoints := []struct {
		x, y float64
		edge int
	}{
		{tileCenterX, tileCenterY - halfSize/2, 1},                    // Top edge
		{tileCenterX + halfSize*0.5, tileCenterY - halfSize/4, 2},    // Top-right corner
		{tileCenterX + halfSize*0.7, tileCenterY, 3},                 // Right edge
		{tileCenterX + halfSize*0.5, tileCenterY + halfSize/4, 4},    // Bottom-right corner
		{tileCenterX, tileCenterY + halfSize/2, 5},                   // Bottom edge
		{tileCenterX - halfSize*0.5, tileCenterY + halfSize/4, 6},    // Bottom-left corner
		{tileCenterX - halfSize*0.7, tileCenterY, 7},                 // Left edge
		{tileCenterX - halfSize*0.5, tileCenterY - halfSize/4, 8},    // Top-left corner
		{tileCenterX, tileCenterY, 9},                                // Center
	}
	
	for _, point := range snapPoints {
		dx := float64(mouseX) - point.x
		dy := float64(mouseY) - point.y
		if math.Sqrt(dx*dx + dy*dy) <= snapRadius {
			ui.inventory.SetSnapToEdge(point.edge)
			ui.inventory.DraggingAsset = false // Stop any dragging
			return true
		}
	}

	return true // Consume click if within dialog
}

// DrawInspectionDialog draws the object inspection dialog
func (ui *UIRenderer) DrawInspectionDialog(screen *ebiten.Image, inspectedObject *room.TileData, screenWidth, screenHeight int) {
	// Dialog dimensions
	dialogWidth := 450.0
	dialogHeight := 400.0
	dialogX := float32(screenWidth)/2 - float32(dialogWidth)/2
	dialogY := float32(screenHeight)/2 - float32(dialogHeight)/2

	// Draw dialog background
	vector.DrawFilledRect(screen, dialogX, dialogY, float32(dialogWidth), float32(dialogHeight), color.RGBA{30, 30, 30, 240}, false)
	vector.StrokeRect(screen, dialogX, dialogY, float32(dialogWidth), float32(dialogHeight), 2, color.RGBA{120, 120, 120, 255}, false)

	// Dialog title
	ebitenutil.DebugPrintAt(screen, "Object Inspector", int(dialogX)+10, int(dialogY)+10)
	ebitenutil.DebugPrintAt(screen, "Press ESC to close | Click Edit to modify", int(dialogX)+10, int(dialogY)+25)

	// Asset preview section
	previewY := int(dialogY) + 50
	ebitenutil.DebugPrintAt(screen, "Asset Preview:", int(dialogX)+10, previewY)
	
	// Draw asset name
	assetName := filepath.Base(inspectedObject.AssetPath)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Asset: %s", assetName), int(dialogX)+10, previewY+20)

	// Position information section
	infoY := previewY + 80
	ebitenutil.DebugPrintAt(screen, "Position Information:", int(dialogX)+10, infoY)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Grid Position: (%d, %d)", inspectedObject.X, inspectedObject.Y), int(dialogX)+10, infoY+20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Layer: %d (%s)", inspectedObject.Layer, ui.getLayerName(inspectedObject.Layer)), int(dialogX)+10, infoY+40)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Category: %s", inspectedObject.Category), int(dialogX)+10, infoY+60)
	
	// Positioning details
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Offset: (%.2f, %.2f)", inspectedObject.OffsetX, inspectedObject.OffsetY), int(dialogX)+10, infoY+80)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Depth: %.2f", inspectedObject.Depth), int(dialogX)+10, infoY+100)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rotation: %.1f°", inspectedObject.Rotation*180/math.Pi), int(dialogX)+10, infoY+120)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Facing: %s", ui.getFacingName(inspectedObject.Facing)), int(dialogX)+10, infoY+140)

	// Edit button area
	buttonY := int(dialogY) + int(dialogHeight) - 80
	ebitenutil.DebugPrintAt(screen, "Actions:", int(dialogX)+10, buttonY)
	
	// Get mouse position for hover effects (only during draw, not for click handling)
	hoverMouseX, hoverMouseY := ebiten.CursorPosition()
	isMousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	
	// Draw Edit button
	editButtonX := int(dialogX) + 10
	editButtonY := buttonY + 15
	editButtonWidth := 200
	editButtonHeight := 25
	
	// Check if hovering edit button
	editHover := hoverMouseX >= editButtonX && hoverMouseX <= editButtonX+editButtonWidth &&
	             hoverMouseY >= editButtonY && hoverMouseY <= editButtonY+editButtonHeight
	
	// Edit button color based on state
	buttonColor := color.RGBA{60, 100, 60, 255} // Normal
	buttonBorder := color.RGBA{120, 120, 120, 255}
	if editHover {
		if isMousePressed {
			buttonColor = color.RGBA{40, 70, 40, 255} // Clicked
			buttonBorder = color.RGBA{200, 200, 200, 255}
		} else {
			buttonColor = color.RGBA{80, 130, 80, 255} // Hover
			buttonBorder = color.RGBA{180, 180, 180, 255}
		}
	}
	
	vector.DrawFilledRect(screen, float32(editButtonX), float32(editButtonY), float32(editButtonWidth), float32(editButtonHeight), buttonColor, false)
	vector.StrokeRect(screen, float32(editButtonX), float32(editButtonY), float32(editButtonWidth), float32(editButtonHeight), 1, buttonBorder, false)
	
	// Edit button text (vertically centered)
	textY := editButtonY + (editButtonHeight-7)/2
	ebitenutil.DebugPrintAt(screen, "Edit Position and Properties", editButtonX+5, textY)
	
	// Delete button
	deleteButtonX := int(dialogX) + 220
	deleteButtonY := buttonY + 15
	deleteButtonWidth := 120
	
	// Check if hovering delete button
	deleteHover := hoverMouseX >= deleteButtonX && hoverMouseX <= deleteButtonX+deleteButtonWidth &&
	              hoverMouseY >= deleteButtonY && hoverMouseY <= deleteButtonY+editButtonHeight
	
	// Delete button color based on state
	deleteColor := color.RGBA{100, 60, 60, 255} // Normal
	deleteBorder := color.RGBA{120, 120, 120, 255}
	if deleteHover {
		if isMousePressed {
			deleteColor = color.RGBA{70, 40, 40, 255} // Clicked
			deleteBorder = color.RGBA{200, 200, 200, 255}
		} else {
			deleteColor = color.RGBA{130, 80, 80, 255} // Hover
			deleteBorder = color.RGBA{180, 180, 180, 255}
		}
	}
	
	vector.DrawFilledRect(screen, float32(deleteButtonX), float32(deleteButtonY), float32(deleteButtonWidth), float32(editButtonHeight), deleteColor, false)
	vector.StrokeRect(screen, float32(deleteButtonX), float32(deleteButtonY), float32(deleteButtonWidth), float32(editButtonHeight), 1, deleteBorder, false)
	
	// Delete button text (vertically centered)
	ebitenutil.DebugPrintAt(screen, "Delete Object", deleteButtonX+25, textY)
}

// getLayerName returns a human-readable name for a layer
func (ui *UIRenderer) getLayerName(layer int) string {
	layerNames := map[int]string{
		0: "Floor",
		1: "Walls", 
		2: "Furniture",
		3: "Decorations",
	}
	if name, exists := layerNames[layer]; exists {
		return name
	}
	return "Unknown"
}

// getFacingName returns a human-readable name for facing direction
func (ui *UIRenderer) getFacingName(facing int) string {
	if facing == 0 {
		return "Normal"
	} else if facing == 1 {
		return "Flipped"
	}
	return "Unknown"
}

// HandleInspectionDialogClick handles clicks within the inspection dialog
func (ui *UIRenderer) HandleInspectionDialogClick(mouseX, mouseY int, screenWidth, screenHeight int) string {
	dialogWidth := 450.0
	dialogHeight := 400.0
	// Use same calculation as DrawInspectionDialog for consistency
	dialogX := float64(screenWidth)/2 - float64(dialogWidth)/2
	dialogY := float64(screenHeight)/2 - float64(dialogHeight)/2

	// Check if click is within dialog
	if float64(mouseX) < dialogX || float64(mouseX) > dialogX+dialogWidth ||
	   float64(mouseY) < dialogY || float64(mouseY) > dialogY+dialogHeight {
		// Click outside dialog bounds
		return ""
	}

	// Check Edit button
	buttonY := int(dialogY) + int(dialogHeight) - 80
	editButtonX := int(dialogX) + 10
	editButtonY := buttonY + 15
	editButtonWidth := 200
	editButtonHeight := 25
	
	// Checking edit button click bounds
	
	if mouseX >= editButtonX && mouseX <= editButtonX+editButtonWidth &&
	   mouseY >= editButtonY && mouseY <= editButtonY+editButtonHeight {
		// Edit button clicked
		return "edit"
	}

	// Check Delete button
	deleteButtonX := int(dialogX) + 220
	deleteButtonY := buttonY + 15
	deleteButtonWidth := 120
	
	if mouseX >= deleteButtonX && mouseX <= deleteButtonX+deleteButtonWidth &&
	   mouseY >= deleteButtonY && mouseY <= deleteButtonY+editButtonHeight {
		return "delete"
	}

	return "consume" // Consume click if within dialog but no button clicked
}

// DrawInspectionModeBorder draws a lime green border around the entire screen when in inspection mode
func (ui *UIRenderer) DrawInspectionModeBorder(screen *ebiten.Image, screenWidth, screenHeight int) {
	limeGreen := color.RGBA{50, 255, 50, 255}
	strokeWidth := float32(4.0)
	
	// Draw top border
	vector.StrokeLine(screen, 0, 0, float32(screenWidth), 0, strokeWidth, limeGreen, false)
	// Draw right border
	vector.StrokeLine(screen, float32(screenWidth), 0, float32(screenWidth), float32(screenHeight), strokeWidth, limeGreen, false)
	// Draw bottom border  
	vector.StrokeLine(screen, float32(screenWidth), float32(screenHeight), 0, float32(screenHeight), strokeWidth, limeGreen, false)
	// Draw left border
	vector.StrokeLine(screen, 0, float32(screenHeight), 0, 0, strokeWidth, limeGreen, false)
}

// DrawAssetPropertyEditor draws the asset property editor dialog with draggable preview
func (ui *UIRenderer) DrawAssetPropertyEditor(screen *ebiten.Image, asset *database.EntityProperties, isDragging bool, zoomLevel float64, screenWidth, screenHeight int) {
	dialogWidth := 900.0  // Made much wider to fit bigger preview and compass
	dialogHeight := 700.0 // Made taller
	dialogX := float64(screenWidth)/2 - dialogWidth/2
	dialogY := float64(screenHeight)/2 - dialogHeight/2

	// Draw dialog background
	vector.DrawFilledRect(screen, float32(dialogX), float32(dialogY), float32(dialogWidth), float32(dialogHeight), color.RGBA{40, 40, 40, 240}, false)
	
	// Draw border
	vector.StrokeRect(screen, float32(dialogX), float32(dialogY), float32(dialogWidth), float32(dialogHeight), 2, color.RGBA{100, 100, 100, 255}, false)
	
	// Title
	title := fmt.Sprintf("Edit Asset: %s", asset.Name)
	ebitenutil.DebugPrintAt(screen, title, int(dialogX)+10, int(dialogY)+30)
	
	// Asset Path
	pathText := fmt.Sprintf("Path: %s", asset.AssetPath)
	ebitenutil.DebugPrintAt(screen, pathText, int(dialogX)+10, int(dialogY)+60)
	
	yPos := int(dialogY) + 90
	lineHeight := 35
	
	// Basic Properties
	ebitenutil.DebugPrintAt(screen, "BASIC PROPERTIES", int(dialogX)+10, yPos)
	yPos += lineHeight
	
	nameText := fmt.Sprintf("Name: %s", asset.Name)
	ebitenutil.DebugPrintAt(screen, nameText, int(dialogX)+20, yPos)
	yPos += lineHeight
	
	categoryText := fmt.Sprintf("Category: %s", asset.Category)
	ebitenutil.DebugPrintAt(screen, categoryText, int(dialogX)+20, yPos)
	yPos += lineHeight
	
	layerText := fmt.Sprintf("Layer: %d", asset.Layer)
	ebitenutil.DebugPrintAt(screen, layerText, int(dialogX)+20, yPos)
	yPos += lineHeight + 10
	
	// Default Position Properties
	ebitenutil.DebugPrintAt(screen, "DEFAULT POSITIONING", int(dialogX)+10, yPos)
	yPos += lineHeight
	
	offsetText := fmt.Sprintf("Offset X: %.2f", asset.DefaultOffsetX)
	ebitenutil.DebugPrintAt(screen, offsetText, int(dialogX)+20, yPos)
	yPos += lineHeight
	
	offsetYText := fmt.Sprintf("Offset Y: %.2f", asset.DefaultOffsetY)
	ebitenutil.DebugPrintAt(screen, offsetYText, int(dialogX)+20, yPos)
	yPos += lineHeight
	
	depthText := fmt.Sprintf("Depth: %.2f", asset.DefaultDepth)
	ebitenutil.DebugPrintAt(screen, depthText, int(dialogX)+20, yPos)
	yPos += lineHeight
	
	flipText := "Normal"
	if asset.CurrentRotation == 1 {
		flipText = "Flipped"
	}
	rotationText := fmt.Sprintf("Position: %s", flipText)
	ebitenutil.DebugPrintAt(screen, rotationText, int(dialogX)+20, yPos)
	yPos += lineHeight + 10
	
	// Physical Properties
	ebitenutil.DebugPrintAt(screen, "PHYSICAL PROPERTIES", int(dialogX)+10, yPos)
	yPos += lineHeight
	
	sizeText := fmt.Sprintf("Size: %.1fx%.1f", asset.Width, asset.Height)
	ebitenutil.DebugPrintAt(screen, sizeText, int(dialogX)+20, yPos)
	yPos += lineHeight
	
	walkOnText := fmt.Sprintf("Can Walk On: %v", asset.CanWalkOn)
	ebitenutil.DebugPrintAt(screen, walkOnText, int(dialogX)+20, yPos)
	yPos += lineHeight
	
	sitOnText := fmt.Sprintf("Can Sit On: %v", asset.CanSitOn)
	ebitenutil.DebugPrintAt(screen, sitOnText, int(dialogX)+20, yPos)
	yPos += lineHeight + 20
	
	// Draw asset preview area on the right side
	ui.drawAssetPreview(screen, asset, isDragging, zoomLevel, dialogX, dialogY, dialogWidth, dialogHeight)
	
	// Buttons
	buttonY := int(dialogY) + int(dialogHeight) - 60
	
	// Get mouse position for hover effects (only during draw, not for click handling)
	hoverMouseX, hoverMouseY := ebiten.CursorPosition()
	isMousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	
	// Save button
	saveButtonX := int(dialogX) + 10
	saveButtonWidth := 100
	saveButtonHeight := 30
	
	// Check if hovering or clicking save button
	saveHover := hoverMouseX >= saveButtonX && hoverMouseX <= saveButtonX+saveButtonWidth &&
	             hoverMouseY >= buttonY && hoverMouseY <= buttonY+saveButtonHeight
	
	// Determine button color based on state
	saveColor := color.RGBA{60, 120, 60, 255} // Normal
	saveBorder := color.RGBA{120, 120, 120, 255}
	if saveHover {
		if isMousePressed {
			saveColor = color.RGBA{40, 80, 40, 255} // Clicked (darker)
			saveBorder = color.RGBA{200, 200, 200, 255}
		} else {
			saveColor = color.RGBA{80, 160, 80, 255} // Hover (brighter)
			saveBorder = color.RGBA{180, 180, 180, 255}
		}
	}
	
	vector.DrawFilledRect(screen, float32(saveButtonX), float32(buttonY), float32(saveButtonWidth), float32(saveButtonHeight), saveColor, false)
	vector.StrokeRect(screen, float32(saveButtonX), float32(buttonY), float32(saveButtonWidth), float32(saveButtonHeight), 1, saveBorder, false)
	
	// Center text vertically and horizontally
	// Text is approximately 7 pixels tall, button is 30 pixels tall
	// Vertical center: buttonY + (30 - 7) / 2 = buttonY + 11
	// "Save" is 4 chars * ~6px = 24px, horizontal center: (100-24)/2 = 38
	textY := buttonY + (saveButtonHeight-7)/2
	ebitenutil.DebugPrintAt(screen, "Save", saveButtonX+38, textY)
	
	// Cancel button
	cancelButtonX := int(dialogX) + 120
	cancelButtonWidth := 100
	cancelButtonHeight := 30
	
	// Check if hovering or clicking cancel button
	cancelHover := hoverMouseX >= cancelButtonX && hoverMouseX <= cancelButtonX+cancelButtonWidth &&
	               hoverMouseY >= buttonY && hoverMouseY <= buttonY+cancelButtonHeight
	
	// Determine button color based on state
	cancelColor := color.RGBA{120, 60, 60, 255} // Normal
	cancelBorder := color.RGBA{120, 120, 120, 255}
	if cancelHover {
		if isMousePressed {
			cancelColor = color.RGBA{80, 40, 40, 255} // Clicked (darker)
			cancelBorder = color.RGBA{200, 200, 200, 255}
		} else {
			cancelColor = color.RGBA{160, 80, 80, 255} // Hover (brighter)
			cancelBorder = color.RGBA{180, 180, 180, 255}
		}
	}
	
	vector.DrawFilledRect(screen, float32(cancelButtonX), float32(buttonY), float32(cancelButtonWidth), float32(cancelButtonHeight), cancelColor, false)
	vector.StrokeRect(screen, float32(cancelButtonX), float32(buttonY), float32(cancelButtonWidth), float32(cancelButtonHeight), 1, cancelBorder, false)
	
	// Center text vertically and horizontally
	// "Cancel" is 6 chars * ~6px = 36px, horizontal center: (100-36)/2 = 32
	ebitenutil.DebugPrintAt(screen, "Cancel", cancelButtonX+32, textY)
}

// HandleAssetPropertyEditorClick handles clicks on the asset property editor with draggable asset
func (ui *UIRenderer) HandleAssetPropertyEditorClick(mouseX, mouseY int, screenWidth, screenHeight int) string {
	dialogWidth := 900.0  // Updated to match new width
	dialogHeight := 700.0 // Updated to match new height
	dialogX := float64(screenWidth)/2 - dialogWidth/2
	dialogY := float64(screenHeight)/2 - dialogHeight/2
	
	// Check if click is within dialog
	if float64(mouseX) < dialogX || float64(mouseX) > dialogX+dialogWidth ||
	   float64(mouseY) < dialogY || float64(mouseY) > dialogY+dialogHeight {
		return ""
	}
	
	// Check buttons first (higher priority than dragging)
	buttonY := int(dialogY) + int(dialogHeight) - 60
	
	// Save button
	saveButtonX := int(dialogX) + 10
	saveButtonWidth := 100
	saveButtonHeight := 30
	if mouseX >= saveButtonX && mouseX <= saveButtonX+saveButtonWidth &&
	   mouseY >= buttonY && mouseY <= buttonY+saveButtonHeight {
		fmt.Printf("SAVE BUTTON CLICKED: Asset property save button clicked at (%d, %d)\n", mouseX, mouseY)
		return "save"
	}
	
	// Cancel button
	cancelButtonX := int(dialogX) + 120
	cancelButtonWidth := 100
	cancelButtonHeight := 30
	if mouseX >= cancelButtonX && mouseX <= cancelButtonX+cancelButtonWidth &&
	   mouseY >= buttonY && mouseY <= buttonY+cancelButtonHeight {
		return "cancel"
	}
	
	// Check if click is in the preview area (for dragging) - after button checks
	// These dimensions must match exactly with drawAssetPreview
	previewWidth := 350.0
	previewHeight := 450.0
	previewX := dialogX + dialogWidth - previewWidth - 20
	previewY := dialogY + 80
	
	// Allow dragging anywhere in the entire preview window
	if float64(mouseX) >= previewX && float64(mouseX) <= previewX+previewWidth &&
	   float64(mouseY) >= previewY && float64(mouseY) <= previewY+previewHeight {
		return "drag_preview"
	}
	
	// Click consumed by dialog (prevent clicking through)
	return "consume"
}

// DrawToolbar draws the main UI toolbar with pixel-style buttons
func (ui *UIRenderer) DrawToolbar(screen *ebiten.Image, editorMode, inspectionMode bool, screenWidth, screenHeight int) {
	toolbarWidth := 120
	toolbarX := screenWidth - toolbarWidth - 10
	toolbarY := 10
	buttonWidth := 100
	buttonHeight := 30
	buttonSpacing := 40
	
	// Background panel for toolbar
	panelHeight := 300
	vector.DrawFilledRect(screen, float32(toolbarX-5), float32(toolbarY-5), float32(toolbarWidth), float32(panelHeight), color.RGBA{30, 30, 30, 200}, false)
	vector.StrokeRect(screen, float32(toolbarX-5), float32(toolbarY-5), float32(toolbarWidth), float32(panelHeight), 1, color.RGBA{80, 80, 80, 255}, false)
	
	currentY := toolbarY
	
	// Editor Mode Toggle Button
	editorColor := color.RGBA{60, 60, 120, 255}
	if editorMode {
		editorColor = color.RGBA{80, 120, 80, 255} // Green when active
	}
	vector.DrawFilledRect(screen, float32(toolbarX), float32(currentY), float32(buttonWidth), float32(buttonHeight), editorColor, false)
	vector.StrokeRect(screen, float32(toolbarX), float32(currentY), float32(buttonWidth), float32(buttonHeight), 1, color.RGBA{120, 120, 120, 255}, false)
	ebitenutil.DebugPrintAt(screen, "Editor", toolbarX+25, currentY+12)
	currentY += buttonSpacing
	
	// Inventory Toggle Button (available in both editor and main game mode)
	inventoryColor := color.RGBA{60, 60, 120, 255}
	if ui.inventory.SidebarVisible {
		inventoryColor = color.RGBA{80, 80, 120, 255} // Darker blue when active
	}
	vector.DrawFilledRect(screen, float32(toolbarX), float32(currentY), float32(buttonWidth), float32(buttonHeight), inventoryColor, false)
	vector.StrokeRect(screen, float32(toolbarX), float32(currentY), float32(buttonWidth), float32(buttonHeight), 1, color.RGBA{120, 120, 120, 255}, false)
	ebitenutil.DebugPrintAt(screen, "Inventory", toolbarX+20, currentY+12)
	currentY += buttonSpacing
	
	// Inspect Mode Toggle Button (available in both editor and main game mode)
	inspectColor := color.RGBA{120, 60, 120, 255} // Purple theme for inspect
	if inspectionMode {
		inspectColor = color.RGBA{160, 100, 160, 255} // Brighter purple when active
	}
	vector.DrawFilledRect(screen, float32(toolbarX), float32(currentY), float32(buttonWidth), float32(buttonHeight), inspectColor, false)
	vector.StrokeRect(screen, float32(toolbarX), float32(currentY), float32(buttonWidth), float32(buttonHeight), 1, color.RGBA{120, 120, 120, 255}, false)
	
	// Draw magnifying glass icon
	ui.drawMagnifyingGlass(screen, toolbarX+10, currentY+8, inspectionMode)
	ebitenutil.DebugPrintAt(screen, "Inspect", toolbarX+35, currentY+12)
	currentY += buttonSpacing

	if editorMode {
		
		// Grid Toggle Button
		gridColor := color.RGBA{60, 60, 120, 255}
		vector.DrawFilledRect(screen, float32(toolbarX), float32(currentY), float32(buttonWidth), float32(buttonHeight), gridColor, false)
		vector.StrokeRect(screen, float32(toolbarX), float32(currentY), float32(buttonWidth), float32(buttonHeight), 1, color.RGBA{120, 120, 120, 255}, false)
		ebitenutil.DebugPrintAt(screen, "Grid", toolbarX+30, currentY+12)
		currentY += buttonSpacing
		
		// Save Room Button
		saveColor := color.RGBA{60, 120, 60, 255}
		vector.DrawFilledRect(screen, float32(toolbarX), float32(currentY), float32(buttonWidth), float32(buttonHeight), saveColor, false)
		vector.StrokeRect(screen, float32(toolbarX), float32(currentY), float32(buttonWidth), float32(buttonHeight), 1, color.RGBA{120, 120, 120, 255}, false)
		ebitenutil.DebugPrintAt(screen, "Save Room", toolbarX+15, currentY+12)
	}
}

// HandleToolbarClick handles clicks on the toolbar buttons
func (ui *UIRenderer) HandleToolbarClick(mouseX, mouseY int, screenWidth, screenHeight int) string {
	toolbarWidth := 120
	toolbarX := screenWidth - toolbarWidth - 10
	toolbarY := 10
	buttonWidth := 100
	buttonHeight := 30
	buttonSpacing := 40
	
	currentY := toolbarY
	
	// Check Editor Mode Toggle Button
	if mouseX >= toolbarX && mouseX <= toolbarX+buttonWidth &&
	   mouseY >= currentY && mouseY <= currentY+buttonHeight {
		return "toggle_editor"
	}
	currentY += buttonSpacing
	
	// Check Inventory Toggle Button
	if mouseX >= toolbarX && mouseX <= toolbarX+buttonWidth &&
	   mouseY >= currentY && mouseY <= currentY+buttonHeight {
		return "toggle_inventory"
	}
	currentY += buttonSpacing
	
	// Check Inspect Mode Toggle Button
	if mouseX >= toolbarX && mouseX <= toolbarX+buttonWidth &&
	   mouseY >= currentY && mouseY <= currentY+buttonHeight {
		return "toggle_inspect"
	}
	currentY += buttonSpacing
	
	// Check Grid Toggle Button
	if mouseX >= toolbarX && mouseX <= toolbarX+buttonWidth &&
	   mouseY >= currentY && mouseY <= currentY+buttonHeight {
		return "toggle_grid"
	}
	currentY += buttonSpacing
	
	// Check Save Room Button
	if mouseX >= toolbarX && mouseX <= toolbarX+buttonWidth &&
	   mouseY >= currentY && mouseY <= currentY+buttonHeight {
		return "save_room"
	}
	
	return ""
}

// drawAssetPreview draws the asset preview area with draggable asset and compass
func (ui *UIRenderer) drawAssetPreview(screen *ebiten.Image, asset *database.EntityProperties, isDragging bool, zoomLevel float64, dialogX, dialogY, dialogWidth, dialogHeight float64) {
	// Preview area dimensions (much bigger now)
	previewWidth := 350.0
	previewHeight := 450.0
	previewX := dialogX + dialogWidth - previewWidth - 20
	previewY := dialogY + 80
	
	// Draw preview area background
	vector.DrawFilledRect(screen, float32(previewX), float32(previewY), float32(previewWidth), float32(previewHeight), color.RGBA{60, 60, 60, 255}, false)
	vector.StrokeRect(screen, float32(previewX), float32(previewY), float32(previewWidth), float32(previewHeight), 1, color.RGBA{120, 120, 120, 255}, false)
	
	// Title for preview area
	ebitenutil.DebugPrintAt(screen, "Asset Preview", int(previewX)+10, int(previewY)-15)
	
	// Draw isometric tile background for reference (scaled by zoom)
	baseTileSize := 64.0
	tileSize := baseTileSize * zoomLevel
	tileCenterX := previewX + previewWidth/2
	tileCenterY := previewY + previewHeight/2 - 20
	
	// Draw isometric tile outline (diamond shape)
	tileHalfWidth := tileSize / 2
	tileHalfHeight := tileSize / 4
	
	// Diamond points (top, right, bottom, left)
	topX := tileCenterX
	topY := tileCenterY - tileHalfHeight
	rightX := tileCenterX + tileHalfWidth
	rightY := tileCenterY
	bottomX := tileCenterX
	bottomY := tileCenterY + tileHalfHeight
	leftX := tileCenterX - tileHalfWidth
	leftY := tileCenterY
	
	// Draw tile outline
	vector.StrokeLine(screen, float32(topX), float32(topY), float32(rightX), float32(rightY), 1, color.RGBA{100, 100, 100, 255}, false)
	vector.StrokeLine(screen, float32(rightX), float32(rightY), float32(bottomX), float32(bottomY), 1, color.RGBA{100, 100, 100, 255}, false)
	vector.StrokeLine(screen, float32(bottomX), float32(bottomY), float32(leftX), float32(leftY), 1, color.RGBA{100, 100, 100, 255}, false)
	vector.StrokeLine(screen, float32(leftX), float32(leftY), float32(topX), float32(topY), 1, color.RGBA{100, 100, 100, 255}, false)
	
	// FINAL FIX: Match main renderer's coordinate system exactly
	// Main renderer does: CartesianToIso(tileX + offsetX, tileY + offsetY)
	// Preview simulates tile (0,0): CartesianToIso(0 + offsetX, 0 + offsetY)
	
	previewExactX := 0.0 + asset.DefaultOffsetX
	previewExactY := 0.0 + asset.DefaultOffsetY
	
	// Use the CartesianToIso function directly to match main renderer exactly (with zoom scaling)
	previewIsoX := (previewExactX - previewExactY) * 32 * zoomLevel
	previewIsoY := ((previewExactX + previewExactY) * 16 + 16) * zoomLevel
	
	// The key difference: main renderer positions at grid intersection, 
	// but preview positions relative to visual diamond center
	// Adjust preview to match the visual positioning
	assetX := tileCenterX + previewIsoX
	assetY := tileCenterY + previewIsoY - 16 * zoomLevel  // Subtract the CartesianToIso offset to align with diamond (zoom-scaled)
	
	// Debug logging for ALL assets to compare with main renderer
	if strings.Contains(asset.AssetPath, "Chair") {
		fmt.Printf("PREVIEW: %s tile(0,0) + offset(%.2f,%.2f) = exact(%.2f,%.2f) -> iso(%.1f,%.1f) -> screen(%.1f,%.1f)\n", 
			asset.AssetPath, asset.DefaultOffsetX, asset.DefaultOffsetY, previewExactX, previewExactY, previewIsoX, previewIsoY, assetX, assetY)
	}
	
	// Load and draw the actual asset image with rotation
	ui.drawAssetImage(screen, asset.AssetPath, assetX, assetY, asset.CurrentRotation, isDragging)
	
	// Draw crosshair at center for reference
	vector.StrokeLine(screen, float32(tileCenterX-5), float32(tileCenterY), float32(tileCenterX+5), float32(tileCenterY), 1, color.RGBA{255, 0, 0, 255}, false)
	vector.StrokeLine(screen, float32(tileCenterX), float32(tileCenterY-5), float32(tileCenterX), float32(tileCenterY+5), 1, color.RGBA{255, 0, 0, 255}, false)
	
	// Display current offset values and flip status
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("X: %.2f", asset.DefaultOffsetX), int(previewX)+10, int(previewY)+int(previewHeight)-55)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Y: %.2f", asset.DefaultOffsetY), int(previewX)+10, int(previewY)+int(previewHeight)-40)
	
	// Debug info showing isometric coordinate transformation
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("IsoXY: %.1f,%.1f", previewIsoX, previewIsoY), int(previewX)+120, int(previewY)+int(previewHeight)-55)
	flipStatus := "Normal"
	if asset.CurrentRotation == 1 {
		flipStatus = "Flipped"
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Position: %s", flipStatus), int(previewX)+10, int(previewY)+int(previewHeight)-25)
	ebitenutil.DebugPrintAt(screen, "R key: Toggle flip", int(previewX)+10, int(previewY)+int(previewHeight)-10)
}

// drawAssetImage loads and draws the actual asset image at the given position with rotation
func (ui *UIRenderer) drawAssetImage(screen *ebiten.Image, assetPath string, centerX, centerY float64, rotation int, isDragging bool) {
	// Load the asset image (we need to add access to the assets package)
	img, err := ui.loadAssetImage(assetPath)
	if err != nil {
		// If we can't load the image, draw a colored rectangle as fallback
		assetSize := 16.0 // Bigger than the yellow square
		vector.DrawFilledRect(screen, float32(centerX-assetSize/2), float32(centerY-assetSize/2), float32(assetSize), float32(assetSize), color.RGBA{255, 100, 100, 255}, false)
		vector.StrokeRect(screen, float32(centerX-assetSize/2), float32(centerY-assetSize/2), float32(assetSize), float32(assetSize), 1, color.RGBA{255, 255, 255, 255}, false)
		return
	}
	
	// Get image dimensions
	imgBounds := img.Bounds()
	imgWidth := float64(imgBounds.Dx())
	imgHeight := float64(imgBounds.Dy())
	
	// Scale down for preview if needed (keep it reasonable size)
	scale := 1.0
	maxSize := 48.0 // Maximum size for preview
	if imgWidth > maxSize || imgHeight > maxSize {
		scaleX := maxSize / imgWidth
		scaleY := maxSize / imgHeight
		if scaleX < scaleY {
			scale = scaleX
		} else {
			scale = scaleY
		}
	}
	
	// Calculate draw position to match main renderer (bottom-center anchor)
	// Main renderer uses: screenX-(spriteWidth/2), screenY-spriteHeight
	drawWidth := imgWidth * scale
	drawHeight := imgHeight * scale
	drawX := centerX - drawWidth/2     // Center horizontally (same as main)
	drawY := centerY - drawHeight      // Bottom edge at centerY (same as main)
	
	// Create draw options with scaling and rotation
	op := &ebiten.DrawImageOptions{}
	
	// Apply horizontal flip if rotation is 1
	// For isometric assets, we only need normal and horizontally flipped
	if rotation == 1 {
		// Horizontal flip
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(imgWidth, 0) // Adjust position after flip
	}
	
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(drawX, drawY) // Position at final location
	
	// Add slight transparency or highlighting if being dragged
	if isDragging {
		op.ColorScale.Scale(1.2, 1.2, 1.2, 0.9) // Slightly brighter and more transparent
	}
	
	// Draw the asset image
	screen.DrawImage(img, op)
	
	// Only draw outline when actively dragging (yellow for visibility)
	if isDragging {
		outlineColor := color.RGBA{255, 255, 100, 255} // Yellow outline only when dragging
		vector.StrokeRect(screen, float32(drawX-1), float32(drawY-1), float32(drawWidth+2), float32(drawHeight+2), 1, outlineColor, false)
	}
}

// loadAssetImage loads an asset image from the given path
func (ui *UIRenderer) loadAssetImage(assetPath string) (*ebiten.Image, error) {
	// Check if we already have this image cached in the inventory
	for _, category := range ui.inventory.Categories {
		for _, tile := range category.Tiles {
			if tile.AssetPath == assetPath && tile.Image != nil {
				return tile.Image, nil
			}
		}
	}
	
	// If not in inventory, try to load it directly
	return ui.loadImageFromPath(assetPath)
}

// loadImageFromPath loads an image from the given asset path using the game's asset system
func (ui *UIRenderer) loadImageFromPath(assetPath string) (*ebiten.Image, error) {
	// Use the same asset loading system as the rest of the game
	return assets.LoadTile(assetPath)
}

// DrawMagnifyingGlass draws a magnifying glass icon (public method)
func (ui *UIRenderer) DrawMagnifyingGlass(screen *ebiten.Image, x, y int, active bool) {
	ui.drawMagnifyingGlass(screen, x, y, active)
}

// drawMagnifyingGlass draws a small magnifying glass icon
func (ui *UIRenderer) drawMagnifyingGlass(screen *ebiten.Image, x, y int, active bool) {
	// Choose colors based on active state
	glassColor := color.RGBA{200, 200, 200, 255}
	handleColor := color.RGBA{150, 150, 150, 255}
	if active {
		glassColor = color.RGBA{255, 255, 100, 255} // Yellow when active
		handleColor = color.RGBA{200, 200, 50, 255}
	}
	
	// Draw magnifying glass lens (circle)
	centerX := float32(x + 8)
	centerY := float32(y + 8)
	radius := float32(6)
	
	// Draw filled circle for lens
	for angle := float64(0); angle < 2*math.Pi; angle += 0.1 {
		px := centerX + radius*float32(math.Cos(angle))
		py := centerY + radius*float32(math.Sin(angle))
		vector.DrawFilledCircle(screen, px, py, 1, glassColor, false)
	}
	
	// Draw handle (line from bottom-right of circle)
	handleStartX := centerX + radius*0.7
	handleStartY := centerY + radius*0.7
	handleEndX := handleStartX + 4
	handleEndY := handleStartY + 4
	
	// Draw thick handle line
	vector.StrokeLine(screen, handleStartX, handleStartY, handleEndX, handleEndY, 2, handleColor, false)
}


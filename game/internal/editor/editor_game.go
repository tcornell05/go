package editor

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tcornell05/go/game/internal/database"
	"github.com/tcornell05/go/game/internal/entity"
	"github.com/tcornell05/go/game/internal/input"
	"github.com/tcornell05/go/game/internal/render"
	"github.com/tcornell05/go/game/internal/room"
	"github.com/tcornell05/go/game/internal/world"
	gamemath "github.com/tcornell05/go/game/pkg/math"
)


const (
	EditorScreenWidth  = 1024
	EditorScreenHeight = 768
)

// EditorMode represents the current editing mode
type EditorMode int

const (
	EditorModeRoom  EditorMode = iota // Room decoration mode (existing functionality)  
	EditorModeWorld                   // World tile editing mode
)

// EditorGame represents the editor game state
type EditorGame struct {
	// Current editing mode
	mode EditorMode
	
	// World data
	worldMap *world.WorldMap
	worldFile string
	currentTileID string
	
	// WorldTile editing
	player           *entity.Player
	roomController   *input.Controller     // WorldTile editor controller
	editorController *EditorController     // World editor controller
	renderer         *render.Renderer      // Isometric renderer
	worldTile        *world.WorldTile      // Current WorldTile being edited
	database         *database.Database
	camX, camY       float64
	
	// World editor state
	selectedTileID string
	showConnections bool
	showGrid       bool
	
	// Key press tracking to prevent rapid toggling
	tabKeyPressed    bool
	
	// World tile reload flag
	needsWorldTileReload bool
	escapeKeyPressed bool
}

// NewEditorGame creates a new editor game instance
func NewEditorGame(worldFile string, tileID string, newTile bool, roomWidth, roomHeight int, 
	roomController *input.Controller, renderer *render.Renderer) (*EditorGame, error) {
	// Initialize database
	db, err := database.NewDatabase("data/editor.db")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	
	// Try to load existing world, or create default
	worldMap, err := db.LoadWorldMap(worldFile)
	if err != nil {
		log.Printf("Creating new world '%s': %v", worldFile, err)
		worldMap = world.InitializeDefaultWorld()
		worldMap.Name = worldFile
		// Save the new world
		if err := db.SaveWorldMap(worldMap); err != nil {
			log.Printf("Warning: Failed to save new world: %v", err)
		}
	}
	
	
	// Create player for room editing
	player, err := entity.NewPlayer(5, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}
	
	// Create editor controller  
	editorController := NewEditorController()
	
	// Create or load WorldTile for editing
	var currentWorldTile *world.WorldTile
	if tileID != "" {
		// Try to load existing world tile from database
		if worldTile, err := db.GetWorldTile(tileID); err == nil {
			currentWorldTile = worldTile
			log.Printf("Loaded existing world tile from database: %s", tileID)
			
			// Add the loaded tile to the world map so editor mode logic can find it
			worldMap.Tiles[tileID] = worldTile
		} else {
			// Create new WorldTile
			log.Printf("Creating new WorldTile: %s", tileID)
			// Parse tile coordinates
			x, y := 0, 0
			if len(tileID) >= 2 {
				x = int(tileID[0] - 'A')
				for i := 1; i < len(tileID); i++ {
					if tileID[i] >= '0' && tileID[i] <= '9' {
						y = y*10 + int(tileID[i]-'0')
					}
				}
				y-- // Convert to 0-based indexing
			}
			currentWorldTile = world.NewWorldTile(tileID, x, y, 0)
			currentWorldTile.Type = world.TileTypeRoom
			worldMap.Tiles[tileID] = currentWorldTile
		}
	}
	
	// Determine starting mode based on tileID
	startMode := EditorModeWorld
	currentTileID := ""
	
	if tileID != "" {
		if newTile {
			// Create a new tile
			log.Printf("Creating new tile %s", tileID)
			// Parse tile coordinates
			if len(tileID) >= 2 {
				x := int(tileID[0] - 'A')
				y := 0
				for i := 1; i < len(tileID); i++ {
					if tileID[i] >= '0' && tileID[i] <= '9' {
						y = y*10 + int(tileID[i]-'0')
					}
				}
				y-- // Convert to 0-based indexing
				
				if x >= 0 && y >= 0 {
					// Create new room tile
					tile := world.NewWorldTile(tileID, x, y, 0)
					tile.Type = world.TileTypeRoom
					tile.Name = fmt.Sprintf("Room %s", tileID)
					tile.Rentable = true
					tile.RentalStatus = world.RentalAvailable
					tile.RentalPrice = 100
					
					worldMap.AddTile(tile)
					if err := db.SaveWorldMap(worldMap); err != nil {
						log.Printf("Warning: Failed to save world with new tile: %v", err)
					}
					
					startMode = EditorModeRoom
					currentTileID = tileID
					log.Printf("Created and editing new room tile: %s", tileID)
				}
			}
		} else {
			// Check if the tile exists - when editing a specific tile, always use room mode
			if tile, exists := worldMap.Tiles[tileID]; exists {
				// When editing a specific tile, always start in room mode regardless of tile type
				startMode = EditorModeRoom
				currentTileID = tileID
				log.Printf("Editing tile: %s (%s, type: %s)", tileID, tile.Name, tile.Type)
				
				// Load the room associated with this tile if it exists
				if tile.RoomID != nil {
					log.Printf("Loading room ID %d for tile %s", *tile.RoomID, tileID)
					// WorldTile already loaded from database with all its tiles
				} else {
					log.Printf("Ready to edit tile contents for %s", tileID)
				}
			} else {
				log.Printf("Tile %s not found in world %s, starting in world mode", tileID, worldFile)
			}
		}
	}

	return &EditorGame{
		mode:           startMode,
		worldMap:       worldMap,
		worldFile:      worldFile,
		currentTileID:  currentTileID,
		
		// WorldTile editing components
		player:           player,
		roomController:   roomController,
		editorController: editorController,
		renderer:         renderer,
		worldTile:        currentWorldTile,
		database:         db,
		camX:             0,
		camY:             0,
		
		// World editor defaults
		selectedTileID:   currentTileID,
		showConnections:  true,
		showGrid:        true,
	}, nil
}

// Update implements ebiten.Game interface
func (e *EditorGame) Update() error {
	// Check if world tile needs to be reloaded (after asset property changes)
	if e.needsWorldTileReload {
		if err := e.ReloadWorldTile(); err != nil {
			fmt.Printf("Failed to reload world tile: %v\n", err)
		}
		e.needsWorldTileReload = false
	}
	
	// Manage cursor based on current mode
	e.updateCursorMode()
	
	// Handle mode switching with proper key press detection
	// Tab only switches TO room mode from world mode (room mode needs Tab for inventory)
	// Escape switches back to world mode from room mode
	if e.mode == EditorModeWorld {
		tabPressed := ebiten.IsKeyPressed(ebiten.KeyTab)
		if tabPressed && !e.tabKeyPressed {
			// Tab key was just pressed (not held) - switch to room mode
			e.mode = EditorModeRoom
		}
		e.tabKeyPressed = tabPressed // Update key state for next frame
	} else {
		// In room mode: Escape switches back to world mode, Tab is free for inventory
		escapePressed := ebiten.IsKeyPressed(ebiten.KeyEscape)
		if escapePressed && !e.escapeKeyPressed {
			e.mode = EditorModeWorld
		}
		e.escapeKeyPressed = escapePressed
		e.tabKeyPressed = false // Reset so room controller can handle Tab
	}
	
	switch e.mode {
	case EditorModeRoom:
		return e.updateRoomEditor()
	case EditorModeWorld:
		return e.updateWorldEditor()
	default:
		return nil
	}
}

// updateRoomEditor handles room editing mode updates
func (e *EditorGame) updateRoomEditor() error {
	// Use full room editor controller with temporary Room adapter
	if e.worldTile != nil {
		tempRoom := e.worldTileToRoomAdapter(e.worldTile)
		direction, newCamX, newCamY := e.roomController.HandleInput(e.player, tempRoom, e.database, e.camX, e.camY, EditorScreenWidth, EditorScreenHeight)
		e.player.CurrentDirection = direction
		e.camX = newCamX
		e.camY = newCamY
		
		// Synchronize grid visibility with controller (F2 debug grid)
		e.showGrid = e.roomController.ShowDebugGrid
		
		e.player.UpdateAnimation()
		e.player.KeepInBounds(e.roomController.GridMovement, tempRoom)
		
		// Camera position is now controlled by middle-mouse dragging
		// Don't reset camera coordinates - they are managed by input handling
	}
	
	return nil
}

// handlePlayerMovement returns the direction based on input
func (e *EditorGame) handlePlayerMovement() string {
	// Simple WASD movement
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		if e.editorController.GridMovement {
			e.player.Y -= 1
		} else {
			e.player.Y -= e.player.Speed
		}
		return "N"
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		if e.editorController.GridMovement {
			e.player.Y += 1
		} else {
			e.player.Y += e.player.Speed
		}
		return "S"
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		if e.editorController.GridMovement {
			e.player.X -= 1
		} else {
			e.player.X -= e.player.Speed
		}
		return "W"
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		if e.editorController.GridMovement {
			e.player.X += 1
		} else {
			e.player.X += e.player.Speed
		}
		return "E"
	}
	return ""
}

// updateWorldEditor handles world editing mode updates
func (e *EditorGame) updateWorldEditor() error {
	// Use the editor controller for world input with camera support
	newCamX, newCamY := e.editorController.HandleWorldInput(e.worldMap, e.database, e.camX, e.camY)
	e.camX = newCamX
	e.camY = newCamY
	
	// Update selected tile from controller
	e.selectedTileID = e.editorController.SelectedTileID
	e.showConnections = e.editorController.ShowConnections
	e.showGrid = e.editorController.ShowGrid
	
	return nil
}

// getTileIDFromScreenPosition converts screen coordinates to tile ID
func (e *EditorGame) getTileIDFromScreenPosition(screenX, screenY int) string {
	// Simple grid-based calculation for now
	// This should be improved to match the actual tile rendering
	tileWidth := 80
	tileHeight := 60
	
	gridX := screenX / tileWidth
	gridY := screenY / tileHeight
	
	if gridX >= 0 && gridX < e.worldMap.Width && gridY >= 0 && gridY < e.worldMap.Height {
		return world.GenerateTileID(gridX, gridY, 0) // Floor 0 for now
	}
	
	return ""
}

// Draw implements ebiten.Game interface  
func (e *EditorGame) Draw(screen *ebiten.Image) {
	switch e.mode {
	case EditorModeRoom:
		e.drawRoomEditor(screen)
	case EditorModeWorld:
		e.drawWorldEditor(screen)
	}
	
	// Draw mode indicator
	e.drawModeIndicator(screen)
}

// drawMagnifyingGlassCursor draws a magnifying glass cursor at the specified position
func (e *EditorGame) drawMagnifyingGlassCursor(screen *ebiten.Image, x, y int) {
	e.roomController.UIRenderer.DrawMagnifyingGlass(screen, x, y, true)
}

// drawRoomEditor renders the room editing interface  
func (e *EditorGame) drawRoomEditor(screen *ebiten.Image) {
	// Temporary: create a Room adapter from WorldTile for rendering
	if e.worldTile != nil {
		tempRoom := e.worldTileToRoomAdapter(e.worldTile)
		e.renderer.Draw(screen, e.player, e.roomController, tempRoom, e.camX, e.camY, EditorScreenWidth, EditorScreenHeight)
		
		// Always draw the toolbar
		e.roomController.UIRenderer.DrawToolbar(screen, e.roomController.EditorMode, e.roomController.InspectionMode, EditorScreenWidth, EditorScreenHeight)
		
		// Always draw sidebar/inventory drawer when visible (in both editor and main game mode)
		e.roomController.UIRenderer.DrawSidebar(screen, EditorScreenWidth, EditorScreenHeight)
		
		// Draw inspection dialog if active (works in both editor and main game mode)
		if e.roomController.InspectionDialog && e.roomController.InspectedObject != nil {
			e.roomController.UIRenderer.DrawInspectionDialog(screen, e.roomController.InspectedObject, EditorScreenWidth, EditorScreenHeight)
		}
		
		// Draw asset property dialog if active (editor mode only)
		if e.roomController.AssetPropertyDialog && e.roomController.EditingAsset != nil {
			e.roomController.UIRenderer.DrawAssetPropertyEditor(screen, e.roomController.EditingAsset, e.roomController.DraggingAssetPreview, e.roomController.ZoomLevel, EditorScreenWidth, EditorScreenHeight)
		}
		
		// Draw inspection mode border when in inspection mode
		if e.roomController.InspectionMode {
			e.roomController.UIRenderer.DrawInspectionModeBorder(screen, EditorScreenWidth, EditorScreenHeight)
			// Draw magnifying glass cursor at mouse position
			mouseX, mouseY := ebiten.CursorPosition()
			e.drawMagnifyingGlassCursor(screen, mouseX, mouseY)
		}
	}
}

// worldTileToRoomAdapter creates a temporary Room object from a WorldTile for rendering compatibility
// TODO: Remove this when renderer is updated to work directly with WorldTiles
func (e *EditorGame) worldTileToRoomAdapter(worldTile *world.WorldTile) *room.Room {
	tempRoom := room.NewRoom(worldTile.ID, worldTile.Width, worldTile.Height)
	
	// Copy tiles using PlaceTile to ensure images are loaded
	for _, worldTileData := range worldTile.Tiles {
		roomTileData := room.TileData{
			AssetPath: worldTileData.AssetPath,
			X:         worldTileData.X,
			Y:         worldTileData.Y,
			Facing:    worldTileData.Facing,
			Category:  worldTileData.Category,
			Layer:     worldTileData.Layer,
			OffsetX:   worldTileData.OffsetX,
			OffsetY:   worldTileData.OffsetY,
			Depth:     worldTileData.Depth,
			Rotation:  worldTileData.Rotation,
		}
		tempRoom.PlaceTile(roomTileData)
	}
	
	// Copy walkability
	tempRoom.Walkability = make(map[string]bool)
	for key, walkable := range worldTile.Walkability {
		tempRoom.Walkability[key] = walkable
	}
	// Room adapter initialized with walkability data
	
	return tempRoom
}

// ReloadWorldTile reloads the current world tile from the database
func (e *EditorGame) ReloadWorldTile() error {
	if e.currentTileID == "" || e.database == nil {
		return fmt.Errorf("no current tile to reload")
	}
	
	// Reload world tile from database
	reloadedTile, err := e.database.GetWorldTile(e.currentTileID)
	if err != nil {
		return fmt.Errorf("failed to reload world tile %s: %w", e.currentTileID, err)
	}
	
	// Update the editor's world tile reference
	e.worldTile = reloadedTile
	
	// Update the world map reference as well
	if e.worldMap != nil {
		e.worldMap.Tiles[e.currentTileID] = reloadedTile
	}
	
	fmt.Printf("WORLD TILE RELOAD: Reloaded world tile %s from database with %d objects\n", e.currentTileID, len(reloadedTile.Tiles))
	return nil
}

// RequestWorldTileReload signals that the world tile should be reloaded on next update
func (e *EditorGame) RequestWorldTileReload() {
	e.needsWorldTileReload = true
}

// UpdateWalkability updates a tile's walkability and persists the change
func (e *EditorGame) UpdateWalkability(x, y int, walkable bool) {
	if e.worldTile == nil || e.database == nil {
		fmt.Printf("ERROR: Cannot update walkability - worldTile or database is nil\n")
		return
	}
	
	// Update the world tile's walkability
	key := fmt.Sprintf("%d:%d", x, y)
	e.worldTile.Walkability[key] = walkable
	fmt.Printf("WALKABILITY UPDATE: Set tile (%d, %d) to walkable=%v in world tile %s\n", x, y, walkable, e.worldTile.ID)
	
	// Save the world tile to persist walkability changes
	go func() {
		if err := e.database.SaveWorldTile(e.worldTile); err != nil {
			fmt.Printf("ERROR: Failed to save walkability changes: %v\n", err)
		} else {
			fmt.Printf("WALKABILITY SAVED: Persisted walkability changes for world tile %s\n", e.worldTile.ID)
		}
	}()
}

// drawSimplifiedRoom renders the room with basic isometric view
func (e *EditorGame) drawSimplifiedRoom(screen *ebiten.Image) {
	// Clear screen with dark background
	screen.Fill(color.RGBA{32, 32, 32, 255})

	// Draw isometric grid for room editing (F3)
	if e.showGrid {
		e.drawRoomGrid(screen)
	}
	
	// Draw placed room tiles
	e.drawRoomTilesSimplified(screen)
	
	// Draw player
	e.drawPlayerSimplified(screen)
	
	// Draw UI elements
	e.drawRoomEditorUI(screen)
}

// drawWorldEditor renders the world editing interface
func (e *EditorGame) drawWorldEditor(screen *ebiten.Image) {
	// Clear screen
	screen.Fill(color.RGBA{50, 50, 80, 255}) // Dark blue background
	
	// Draw world grid
	if e.showGrid {
		e.drawWorldGrid(screen)
	}
	
	// Draw world tiles
	e.drawWorldTiles(screen)
	
	// Draw connections
	if e.showConnections {
		e.drawTileConnections(screen)
	}
	
	// Draw selection highlight
	if e.selectedTileID != "" {
		e.drawTileSelection(screen, e.selectedTileID)
	}
	
	// Draw world editor UI
	e.drawWorldEditorUI(screen)
}

// drawWorldGrid draws the world tile grid
func (e *EditorGame) drawWorldGrid(screen *ebiten.Image) {
	// TODO: Implement grid rendering
	// For now, this is a placeholder
}

// drawWorldTiles draws all world tiles
func (e *EditorGame) drawWorldTiles(screen *ebiten.Image) {
	tileWidth := 80
	tileHeight := 60
	
	for _, tile := range e.worldMap.Tiles {
		x := tile.WorldX * tileWidth
		y := tile.WorldY * tileHeight
		
		// Draw tile based on type
		color := e.getTileColor(tile.Type)
		
		// TODO: Implement proper tile rendering with colors/textures
		// For now, this is a simplified representation
		_ = color
		_ = x
		_ = y
	}
}

// getTileColor returns a color for a tile type
func (e *EditorGame) getTileColor(tileType world.TileType) uint32 {
	switch tileType {
	case world.TileTypeRoom:
		return 0xFF6B6B // Red
	case world.TileTypeHallway:
		return 0x4ECDC4 // Teal
	case world.TileTypeLobby:
		return 0x45B7D1 // Blue
	case world.TileTypeShop:
		return 0x96CEB4 // Green
	case world.TileTypePublic:
		return 0xFECA57 // Yellow
	case world.TileTypeElevator:
		return 0x686DE0 // Purple
	default:
		return 0xDDD5D0 // Gray
	}
}

// drawTileConnections draws connections between tiles
func (e *EditorGame) drawTileConnections(screen *ebiten.Image) {
	// TODO: Implement connection line rendering
}

// drawTileSelection highlights the selected tile
func (e *EditorGame) drawTileSelection(screen *ebiten.Image, tileID string) {
	// TODO: Implement selection highlight
}

// drawWorldEditorUI draws the world editor user interface
func (e *EditorGame) drawWorldEditorUI(screen *ebiten.Image) {
	// TODO: Implement world editor UI (tile properties, tools, etc.)
}

// drawModeIndicator shows which mode is currently active
func (e *EditorGame) drawModeIndicator(screen *ebiten.Image) {
	modeText := "Room Editor"
	if e.mode == EditorModeWorld {
		modeText = "World Editor"
		if e.currentTileID != "" {
			modeText += " - " + e.currentTileID
		}
	} else if e.currentTileID != "" {
		modeText += " - " + e.currentTileID
	}
	
	// TODO: Implement text rendering for mode indicator
	_ = modeText
	
	// For now, show in console when tile changes
	if e.selectedTileID != e.currentTileID && e.selectedTileID != "" {
		if tile, exists := e.worldMap.Tiles[e.selectedTileID]; exists {
			log.Printf("Selected tile %s: %s (%s)", e.selectedTileID, tile.Name, tile.Type)
		}
	}
}

// drawRoomGrid draws an isometric grid for the room editor
func (e *EditorGame) drawRoomGrid(screen *ebiten.Image) {
	defaultGridColor := color.RGBA{80, 80, 80, 100}
	walkableColor := color.RGBA{50, 150, 50, 150}    // Green for walkable
	blockedColor := color.RGBA{150, 50, 50, 150}     // Red for blocked
	const roomSize = 15 // Default room size, can be overridden by room dimensions
	
	for x := 0; x < roomSize; x++ {
		for y := 0; y < roomSize; y++ {
			gridX := float64(x)
			gridY := float64(y)
			
			isoX, isoY := gamemath.CartesianToIso(gridX, gridY)
			screenX := isoX + float64(EditorScreenWidth/2) - e.camX
			screenY := isoY + float64(EditorScreenHeight/2) - e.camY
			
			// Choose color based on walkability when F3 grid is active
			gridColor := defaultGridColor
			if e.showGrid && e.worldTile != nil {
				if e.worldTile.IsWalkable(x, y) {
					gridColor = walkableColor
				} else {
					gridColor = blockedColor
				}
			}
			
			e.drawGridDiamond(screen, screenX, screenY, gridColor)
		}
	}
}

// drawRoomTilesSimplified draws room tiles without complex sorting
func (e *EditorGame) drawRoomTilesSimplified(screen *ebiten.Image) {
	if e.worldTile != nil {
		for _, tile := range e.worldTile.Tiles {
			if tileImage := e.worldTile.LoadedImages[tile.AssetPath]; tileImage != nil {
				exactX := float64(tile.X) + tile.OffsetX
				exactY := float64(tile.Y) + tile.OffsetY
				
				isoX, isoY := gamemath.CartesianToIso(exactX, exactY)
				screenX := isoX + float64(EditorScreenWidth/2) - e.camX  
				screenY := isoY + float64(EditorScreenHeight/2) - e.camY
				
				op := &ebiten.DrawImageOptions{}
				bounds := tileImage.Bounds()
				spriteWidth := float64(bounds.Dx())
				spriteHeight := float64(bounds.Dy())
				
				if tile.Facing == 1 {
					op.GeoM.Scale(-1, 1)
					op.GeoM.Translate(spriteWidth, 0)
				}
				
				op.GeoM.Translate(screenX-(spriteWidth/2), screenY-spriteHeight)
				screen.DrawImage(tileImage, op)
			}
		}
	}
}

// drawPlayerSimplified draws the player sprite
func (e *EditorGame) drawPlayerSimplified(screen *ebiten.Image) {
	playerIsoX, playerIsoY := gamemath.CartesianToIso(e.player.X, e.player.Y)
	playerScreenX := playerIsoX + float64(EditorScreenWidth/2) - e.camX
	playerScreenY := playerIsoY + float64(EditorScreenHeight/2) - e.camY
	
	// Get sprite
	var sprite *ebiten.Image
	if e.player.IsWalking && e.player.WalkingSprites[e.player.Facing] != nil {
		sprite = e.player.WalkingSprites[e.player.Facing]
	} else if e.player.StaticSprites[e.player.Facing] != nil {
		sprite = e.player.StaticSprites[e.player.Facing]
	}
	
	if sprite != nil {
		op := &ebiten.DrawImageOptions{}
		bounds := sprite.Bounds()
		spriteWidth := float64(bounds.Dx())
		spriteHeight := float64(bounds.Dy())
		
		op.GeoM.Translate(
			playerScreenX-spriteWidth/2,
			playerScreenY-spriteHeight+16,
		)
		screen.DrawImage(sprite, op)
	}
}

// drawRoomEditorUI draws the room editor user interface
func (e *EditorGame) drawRoomEditorUI(screen *ebiten.Image) {
	uiText := fmt.Sprintf("Room Editor - %s\nTab: Switch to World Editor\nWASD: Move Player\n`: Toggle Grid Movement", e.currentTileID)
	ebitenutil.DebugPrint(screen, uiText)
}

// drawGridDiamond draws a diamond shape for grid cells
func (e *EditorGame) drawGridDiamond(screen *ebiten.Image, centerX, centerY float64, gridColor color.Color) {
	points := []struct{ x, y float64 }{
		{centerX, centerY - 16}, // Top
		{centerX + 32, centerY}, // Right
		{centerX, centerY + 16}, // Bottom
		{centerX - 32, centerY}, // Left
		{centerX, centerY - 16}, // Back to top
	}
	
	for i := 0; i < len(points)-1; i++ {
		e.drawLine(screen, points[i].x, points[i].y, points[i+1].x, points[i+1].y, gridColor)
	}
}

// drawLine draws a line between two points
func (e *EditorGame) drawLine(screen *ebiten.Image, x1, y1, x2, y2 float64, c color.Color) {
	dx := x2 - x1
	dy := y2 - y1
	steps := int(math.Max(math.Abs(dx), math.Abs(dy)))
	
	if steps == 0 {
		return
	}
	
	xInc := dx / float64(steps)
	yInc := dy / float64(steps)
	
	x, y := x1, y1
	for i := 0; i <= steps; i++ {
		if int(x) >= 0 && int(x) < screen.Bounds().Dx() && int(y) >= 0 && int(y) < screen.Bounds().Dy() {
			screen.Set(int(x), int(y), c)
		}
		x += xInc
		y += yInc
	}
}

// drawInspectionCursor draws a magnifying glass cursor at the mouse position
func (e *EditorGame) drawInspectionCursor(screen *ebiten.Image) {
	mouseX, mouseY := ebiten.CursorPosition()
	
	// Draw magnifying glass cursor
	glassColor := color.RGBA{255, 255, 100, 255} // Bright yellow
	handleColor := color.RGBA{200, 200, 50, 255}
	
	// Draw magnifying glass lens (circle)
	centerX := float32(mouseX)
	centerY := float32(mouseY)
	radius := float32(10)
	
	// Draw filled circle for lens with transparency
	for angle := float64(0); angle < 2*math.Pi; angle += 0.1 {
		px := centerX + radius*float32(math.Cos(angle))
		py := centerY + radius*float32(math.Sin(angle))
		screen.Set(int(px), int(py), color.RGBA{255, 255, 100, 150})
	}
	
	// Draw lens border
	for angle := float64(0); angle < 2*math.Pi; angle += 0.1 {
		px := centerX + radius*float32(math.Cos(angle))
		py := centerY + radius*float32(math.Sin(angle))
		screen.Set(int(px), int(py), glassColor)
	}
	
	// Draw handle (line from bottom-right of circle)
	handleStartX := centerX + radius*0.7
	handleStartY := centerY + radius*0.7
	handleEndX := handleStartX + 8
	handleEndY := handleStartY + 8
	
	// Draw thick handle line
	e.drawLine(screen, float64(handleStartX), float64(handleStartY), float64(handleEndX), float64(handleEndY), handleColor)
}

// updateCursorMode manages cursor visibility based on current editor state
func (e *EditorGame) updateCursorMode() {
	// Always hide system cursor when we want to draw custom cursor
	if e.needsCustomCursor() {
		ebiten.SetCursorMode(ebiten.CursorModeHidden)
	} else {
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
	}
}

// needsCustomCursor returns true if we should draw a custom cursor
func (e *EditorGame) needsCustomCursor() bool {
	// Show custom cursor for inspection mode
	if e.roomController.InspectionMode {
		return true
	}
	
	// Show custom cursor when asset property editor is open and dragging
	if e.roomController.AssetPropertyDialog {
		return true
	}
	
	return false
}

// drawCustomCursor draws the appropriate cursor for the current mode
func (e *EditorGame) drawCustomCursor(screen *ebiten.Image) {
	if !e.needsCustomCursor() {
		return
	}
	
	// Asset property editor has highest priority
	if e.roomController.AssetPropertyDialog {
		e.drawMoveCursor(screen)
	} else if e.roomController.InspectionMode {
		e.drawInspectionCursor(screen)
	} else {
		e.drawDefaultCursor(screen)
	}
}

// drawMoveCursor draws a four-directional movement cursor
func (e *EditorGame) drawMoveCursor(screen *ebiten.Image) {
	mouseX, mouseY := ebiten.CursorPosition()
	
	centerX := float32(mouseX)
	centerY := float32(mouseY)
	arrowColor := color.RGBA{100, 200, 100, 255} // Green
	
	// Draw center dot
	screen.Set(int(centerX), int(centerY), arrowColor)
	screen.Set(int(centerX)+1, int(centerY), arrowColor)
	screen.Set(int(centerX), int(centerY)+1, arrowColor)
	screen.Set(int(centerX)+1, int(centerY)+1, arrowColor)
	
	arrowSize := float32(8)
	
	// Draw up arrow
	e.drawLine(screen, float64(centerX), float64(centerY-arrowSize), float64(centerX), float64(centerY-3), arrowColor)
	e.drawLine(screen, float64(centerX), float64(centerY-arrowSize), float64(centerX-3), float64(centerY-arrowSize+3), arrowColor)
	e.drawLine(screen, float64(centerX), float64(centerY-arrowSize), float64(centerX+3), float64(centerY-arrowSize+3), arrowColor)
	
	// Draw down arrow
	e.drawLine(screen, float64(centerX), float64(centerY+arrowSize), float64(centerX), float64(centerY+3), arrowColor)
	e.drawLine(screen, float64(centerX), float64(centerY+arrowSize), float64(centerX-3), float64(centerY+arrowSize-3), arrowColor)
	e.drawLine(screen, float64(centerX), float64(centerY+arrowSize), float64(centerX+3), float64(centerY+arrowSize-3), arrowColor)
	
	// Draw left arrow
	e.drawLine(screen, float64(centerX-arrowSize), float64(centerY), float64(centerX-3), float64(centerY), arrowColor)
	e.drawLine(screen, float64(centerX-arrowSize), float64(centerY), float64(centerX-arrowSize+3), float64(centerY-3), arrowColor)
	e.drawLine(screen, float64(centerX-arrowSize), float64(centerY), float64(centerX-arrowSize+3), float64(centerY+3), arrowColor)
	
	// Draw right arrow
	e.drawLine(screen, float64(centerX+arrowSize), float64(centerY), float64(centerX+3), float64(centerY), arrowColor)
	e.drawLine(screen, float64(centerX+arrowSize), float64(centerY), float64(centerX+arrowSize-3), float64(centerY-3), arrowColor)
	e.drawLine(screen, float64(centerX+arrowSize), float64(centerY), float64(centerX+arrowSize-3), float64(centerY+3), arrowColor)
}

// drawDefaultCursor draws a simple default cursor
func (e *EditorGame) drawDefaultCursor(screen *ebiten.Image) {
	mouseX, mouseY := ebiten.CursorPosition()
	
	cursorColor := color.RGBA{255, 255, 255, 255} // White
	outlineColor := color.RGBA{0, 0, 0, 255}       // Black outline
	
	// Draw outline first (slightly larger)
	for dy := -1; dy <= 10; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy >= 0 && dy <= 8 {
				screen.Set(mouseX+dx, mouseY+dy, outlineColor)
			}
			if dy == 0 && dx >= 0 && dx <= 8 {
				screen.Set(mouseX+dx, mouseY+dy, outlineColor)
			}
		}
	}
	
	// Draw main cursor
	// Vertical line
	for dy := 0; dy <= 8; dy++ {
		screen.Set(mouseX, mouseY+dy, cursorColor)
	}
	// Horizontal line
	for dx := 0; dx <= 8; dx++ {
		screen.Set(mouseX+dx, mouseY, cursorColor)
	}
}

// Layout implements ebiten.Game interface
func (e *EditorGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return EditorScreenWidth, EditorScreenHeight
}
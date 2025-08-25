package input

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tcornell05/go/game/internal/database"
	"github.com/tcornell05/go/game/internal/entity"
	"github.com/tcornell05/go/game/internal/ui"
	"github.com/tcornell05/go/game/internal/room"
	"github.com/tcornell05/go/game/internal/world"
	"github.com/tcornell05/go/game/pkg/math"
)

// Helper functions for world tile integration

// isValidTileID checks if a name looks like a tile ID (e.g., A1, B2, Z99)
func isValidTileID(name string) bool {
	if len(name) < 2 || len(name) > 3 {
		return false
	}
	// Check if first character is a letter and rest are digits
	if name[0] < 'A' || name[0] > 'Z' {
		return false
	}
	for i := 1; i < len(name); i++ {
		if name[i] < '0' || name[i] > '9' {
			return false
		}
	}
	return true
}

// convertRoomToWorldTile converts a Room back to WorldTile for saving
func convertRoomToWorldTile(gameRoom *room.Room) *world.WorldTile {
	worldTile := world.NewWorldTile(gameRoom.Name, 0, 0, 0)
	worldTile.Width = gameRoom.Width
	worldTile.Height = gameRoom.Height
	
	// Copy tiles - convert room.TileData to world.TileData
	worldTile.Tiles = make(map[string]world.TileData)
	for key, roomTileData := range gameRoom.Tiles {
		worldTileData := world.TileData{
			AssetPath: roomTileData.AssetPath,
			X:         roomTileData.X,
			Y:         roomTileData.Y,
			Facing:    roomTileData.Facing,
			Category:  roomTileData.Category,
			Layer:     roomTileData.Layer,
			OffsetX:   roomTileData.OffsetX,
			OffsetY:   roomTileData.OffsetY,
			Depth:     roomTileData.Depth,
			Rotation:  roomTileData.Rotation,
		}
		worldTile.Tiles[key] = worldTileData
	}
	
	// Copy walkability
	worldTile.Walkability = make(map[string]bool)
	for key, walkable := range gameRoom.Walkability {
		worldTile.Walkability[key] = walkable
	}
	
	return worldTile
}

type Controller struct {
	KeyPressed       map[ebiten.Key]bool
	MousePressed     map[ebiten.MouseButton]bool // Track mouse button states
	GridMovement     bool
	HoveredTileX     int
	HoveredTileY     int
	IsHoveringTile   bool
	EditorMode       bool
	InspectionMode   bool  // New: inspection cursor mode
	ShowDebugGrid    bool
	TileInventory    *ui.TileInventory
	UIRenderer       *ui.UIRenderer
	// InspectionMode removed - now using middle-click for direct inspection
	InspectedObject  *room.TileData
	InspectionDialog bool
	EditingInspected bool
	JustOpenedPositionDialog bool // Flag to prevent immediate closing
	
	// Camera dragging state
	CameraDragging       bool
	CameraDragStartX     int
	CameraDragStartY     int
	CameraDragCamStartX  float64
	CameraDragCamStartY  float64
	
	// Camera zoom state
	ZoomLevel            float64  // 1.0 = normal, > 1.0 = zoomed in, < 1.0 = zoomed out
	
	// Asset property editor
	AssetPropertyDialog bool
	EditingAsset        *database.EntityProperties
	
	// Asset preview dragging
	DraggingAssetPreview bool
	AssetDragStartX      int
	AssetDragStartY      int
	DragStartOffsetX     float64 // Asset's starting offset when drag began
	DragStartOffsetY     float64 // Asset's starting offset when drag began
	
	// Callback for when world tile needs to be reloaded
	OnWorldTileReloadNeeded func()
	
	// Debug frame counter
	frameCounter int
	
	// Callback for boundary editing changes
	OnBoundaryChanged func(x, y int, walkable bool)
}

// Implement render.ControllerInterface methods to break import cycle
func (c *Controller) GetShowDebugGrid() bool { return c.ShowDebugGrid }
func (c *Controller) GetIsHoveringTile() bool { return c.IsHoveringTile }
func (c *Controller) GetHoveredTile() (int, int) { return c.HoveredTileX, c.HoveredTileY }
func (c *Controller) GetIsEditorMode() bool { return c.EditorMode }
func (c *Controller) GetTileInventory() interface{} {
	if c.TileInventory == nil {
		return nil
	}
	return c.TileInventory
}
func (c *Controller) GetHasInspectionDialog() bool { return c.InspectionDialog }
func (c *Controller) GetInspectedObject() interface{} { return c.InspectedObject }
func (c *Controller) GetHasAssetPropertyDialog() bool { return c.AssetPropertyDialog }
func (c *Controller) GetEditingAsset() interface{} { return c.EditingAsset }
func (c *Controller) GetIsDraggingAssetPreview() bool { return c.DraggingAssetPreview }
func (c *Controller) GetZoomLevel() float64 { return c.ZoomLevel }

func NewController() (*Controller, error) {
	inventory, err := ui.NewTileInventory()
	if err != nil {
		return nil, err
	}
	
	uiRenderer := ui.NewUIRenderer(inventory)
	
	return &Controller{
		KeyPressed:     make(map[ebiten.Key]bool),
		MousePressed:   make(map[ebiten.MouseButton]bool),
		GridMovement:   false,
		HoveredTileX:   -1,
		HoveredTileY:   -1,
		IsHoveringTile: false,
		EditorMode:     false,
		ZoomLevel:      1.0, // Normal zoom level
		TileInventory:  inventory,
		UIRenderer:     uiRenderer,
	}, nil
}

// HandleInput processes all input for the game
func (c *Controller) HandleInput(player *entity.Player, gameRoom *room.Room, db *database.Database, camX, camY float64, screenWidth, screenHeight int) (direction string, newCamX, newCamY float64) {
	// Initialize camera coordinates (will be updated by camera dragging if active)
	newCamX = camX
	newCamY = camY
	
	// Debug frame counter (commented out to reduce console spam)
	// c.frameCounter++
	// if c.frameCounter%60 == 0 {
	//     fmt.Printf("ROOM HandleInput called, frame %d, camX=%.2f, camY=%.2f\n", c.frameCounter, camX, camY)
	// }
	
	// Toggle movement mode with backtick key
	if ebiten.IsKeyPressed(ebiten.KeyGraveAccent) && !c.KeyPressed[ebiten.KeyGraveAccent] {
		c.GridMovement = !c.GridMovement
		c.KeyPressed[ebiten.KeyGraveAccent] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyGraveAccent) {
		c.KeyPressed[ebiten.KeyGraveAccent] = false
	}
	
	// Handle rotation with R key (prioritize asset property editor, then player rotation)
	if ebiten.IsKeyPressed(ebiten.KeyR) && !c.KeyPressed[ebiten.KeyR] {
		if c.EditorMode && c.AssetPropertyDialog && c.EditingAsset != nil {
			// Toggle between 0 and 1 (normal and flipped)
			nextRotation := 1 - c.EditingAsset.CurrentRotation
			err := db.LoadRotation(c.EditingAsset, nextRotation)
			if err != nil {
				fmt.Printf("Failed to load rotation %d: %v\n", nextRotation, err)
			} else {
				fmt.Printf("Switched to position %d (flipped: %v)\n", nextRotation, nextRotation == 1)
			}
		} else {
			// Handle player rotation
			player.Rotate()
		}
		c.KeyPressed[ebiten.KeyR] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyR) {
		c.KeyPressed[ebiten.KeyR] = false
	}
	
	// Toggle editor mode with E key
	if ebiten.IsKeyPressed(ebiten.KeyE) && !c.KeyPressed[ebiten.KeyE] {
		c.EditorMode = !c.EditorMode
		c.KeyPressed[ebiten.KeyE] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyE) {
		c.KeyPressed[ebiten.KeyE] = false
	}
	
	// Toggle debug grid with F2 key
	if ebiten.IsKeyPressed(ebiten.KeyF2) && !c.KeyPressed[ebiten.KeyF2] {
		c.ShowDebugGrid = !c.ShowDebugGrid
		c.KeyPressed[ebiten.KeyF2] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyF2) {
		c.KeyPressed[ebiten.KeyF2] = false
	}
	
	// F3 key handling removed - using F2 for boundary editing instead
	
	// Handle Alt+Enter for fullscreen toggle
	altPressed := ebiten.IsKeyPressed(ebiten.KeyAlt)
	enterPressed := ebiten.IsKeyPressed(ebiten.KeyEnter) && !c.KeyPressed[ebiten.KeyEnter]
	if altPressed && enterPressed {
		if ebiten.IsFullscreen() {
			ebiten.SetFullscreen(false)
		} else {
			ebiten.SetFullscreen(true)
		}
		c.KeyPressed[ebiten.KeyEnter] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyEnter) {
		c.KeyPressed[ebiten.KeyEnter] = false
	}
	
	// Handle level changes with number keys (1-9)
	for i := 1; i <= 9; i++ {
		key := ebiten.KeyDigit1 + ebiten.Key(i-1)
		if ebiten.IsKeyPressed(key) && !c.KeyPressed[key] {
			player.Level = i
			c.KeyPressed[key] = true
		} else if !ebiten.IsKeyPressed(key) {
			c.KeyPressed[key] = false
		}
	}
	
	// Toggle sidebar with Tab key (works in both editor and main game mode)
	if ebiten.IsKeyPressed(ebiten.KeyTab) && !c.KeyPressed[ebiten.KeyTab] {
		fmt.Printf("TAB KEY DEBUG: Toggling sidebar, current visible: %v\n", c.TileInventory.SidebarVisible)
		c.TileInventory.ToggleSidebar()
		fmt.Printf("TAB KEY DEBUG: After toggle, sidebar visible: %v\n", c.TileInventory.SidebarVisible)
		c.KeyPressed[ebiten.KeyTab] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyTab) {
		c.KeyPressed[ebiten.KeyTab] = false
	}
	
	// Alternative: Toggle sidebar with B key (backup for testing)  
	if ebiten.IsKeyPressed(ebiten.KeyB) && !c.KeyPressed[ebiten.KeyB] {
		c.TileInventory.ToggleSidebar()
		c.KeyPressed[ebiten.KeyB] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyB) {
		c.KeyPressed[ebiten.KeyB] = false
	}
	
	// Toggle positioning dialog with P key (only in editor mode)
	if c.EditorMode && ebiten.IsKeyPressed(ebiten.KeyP) && !c.KeyPressed[ebiten.KeyP] {
		if c.TileInventory.PositionDialog && !c.JustOpenedPositionDialog {
			// Save changes if we were editing an inspected object
			if c.EditingInspected && c.InspectedObject != nil {
				c.saveInspectedObjectChanges(gameRoom, db)
			}
			c.TileInventory.ClosePositionDialog()
			// Reset editing state after closing
			if c.EditingInspected {
				c.EditingInspected = false
				c.InspectedObject = nil
			}
		} else if !c.TileInventory.PositionDialog {
			c.TileInventory.OpenPositionDialog()
		}
		c.KeyPressed[ebiten.KeyP] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyP) {
		c.KeyPressed[ebiten.KeyP] = false
		// Reset the flag when P is released
		if c.JustOpenedPositionDialog {
			c.JustOpenedPositionDialog = false
		}
	}
	
	// Handle inspection with I key (works in both editor and main game mode)
	if ebiten.IsKeyPressed(ebiten.KeyI) && !c.KeyPressed[ebiten.KeyI] {
		if c.IsHoveringTile {
			c.handleInspectionClick(gameRoom, db, c.HoveredTileX, c.HoveredTileY)
		}
		c.KeyPressed[ebiten.KeyI] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyI) {
		c.KeyPressed[ebiten.KeyI] = false
	}
	
	// Toggle inspection mode with M key (backup for testing)
	if ebiten.IsKeyPressed(ebiten.KeyM) && !c.KeyPressed[ebiten.KeyM] {
		c.InspectionMode = !c.InspectionMode
		c.KeyPressed[ebiten.KeyM] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyM) {
		c.KeyPressed[ebiten.KeyM] = false
	}
	
	// Test with Space key (should definitely work)
	if ebiten.IsKeyPressed(ebiten.KeySpace) && !c.KeyPressed[ebiten.KeySpace] {
		c.TileInventory.ToggleSidebar()
		c.KeyPressed[ebiten.KeySpace] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeySpace) {
		c.KeyPressed[ebiten.KeySpace] = false
	}
	
	// Close dialogs with Escape key (only in editor mode)
	if c.EditorMode && ebiten.IsKeyPressed(ebiten.KeyEscape) && !c.KeyPressed[ebiten.KeyEscape] {
		if c.AssetPropertyDialog {
			c.AssetPropertyDialog = false
			c.EditingAsset = nil
		} else if c.InspectionDialog {
			c.InspectionDialog = false
			c.EditingInspected = false
			c.InspectedObject = nil
		} else if c.TileInventory.PositionDialog && !c.JustOpenedPositionDialog {
			// Save changes if we were editing an inspected object
			if c.EditingInspected && c.InspectedObject != nil {
				c.saveInspectedObjectChanges(gameRoom, db)
			}
			c.TileInventory.ClosePositionDialog()
			// Reset editing state after closing
			if c.EditingInspected {
				c.EditingInspected = false
				c.InspectedObject = nil
			}
		}
		c.KeyPressed[ebiten.KeyEscape] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyEscape) {
		c.KeyPressed[ebiten.KeyEscape] = false
		// Reset the flag when ESC is released
		if c.JustOpenedPositionDialog {
			c.JustOpenedPositionDialog = false
		}
	}
	
	// Handle position dialog controls
	if c.EditorMode && c.TileInventory.PositionDialog {
		// Number keys for snap points (1-5 for snap points, 0 for manual)
		for i := 0; i <= 5; i++ {
			key := ebiten.KeyDigit0 + ebiten.Key(i)
			if ebiten.IsKeyPressed(key) && !c.KeyPressed[key] {
				if i == 0 {
					c.TileInventory.SetSnapToEdge(0) // Manual
				} else {
					c.TileInventory.SetSnapToEdge(i) // Snap points 1-5
				}
				c.KeyPressed[key] = true
			} else if !ebiten.IsKeyPressed(key) {
				c.KeyPressed[key] = false
			}
		}
		
		// Arrow keys for depth control
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) && !c.KeyPressed[ebiten.KeyArrowUp] {
			newDepth := c.TileInventory.SelectedPosition.Depth - 0.1
			if newDepth >= -2.0 { // Limit how high above floor
				c.TileInventory.SetDepth(newDepth)
			}
			c.KeyPressed[ebiten.KeyArrowUp] = true
		} else if !ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			c.KeyPressed[ebiten.KeyArrowUp] = false
		}
		
		if ebiten.IsKeyPressed(ebiten.KeyArrowDown) && !c.KeyPressed[ebiten.KeyArrowDown] {
			newDepth := c.TileInventory.SelectedPosition.Depth + 0.1
			if newDepth <= 2.0 { // Limit how far below floor
				c.TileInventory.SetDepth(newDepth)
			}
			c.KeyPressed[ebiten.KeyArrowDown] = true
		} else if !ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
			c.KeyPressed[ebiten.KeyArrowDown] = false
		}
	}
	
	// Rotate selected tile in editor mode
	if c.EditorMode && c.TileInventory.SelectedTile != nil && ebiten.IsKeyPressed(ebiten.KeyR) && !c.KeyPressed[ebiten.KeyR] {
		c.TileInventory.RotateSelectedTile()
		c.KeyPressed[ebiten.KeyR] = true
	}
	
	// Handle mouse input
	mouseX, mouseY := ebiten.CursorPosition()
	gridX, gridY := math.ScreenToIso(float64(mouseX), float64(mouseY), camX, camY, c.ZoomLevel, float64(screenWidth), float64(screenHeight))
	
	// Debug logging for coordinate detection
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		fmt.Printf("CLICK DETECTION: Mouse(%d,%d) -> Grid(%d,%d)\n", mouseX, mouseY, gridX, gridY)
	}
	
	// Update hover state
	if c.EditorMode && c.ShowDebugGrid {
		// F2 boundary editing mode - allow hovering on full world tile area
		if gridX >= -250 && gridX <= 250 && gridY >= -250 && gridY <= 250 {
			c.IsHoveringTile = true
			c.HoveredTileX = gridX
			c.HoveredTileY = gridY
		} else {
			c.IsHoveringTile = false
		}
	} else if c.EditorMode {
		// Regular editor mode - allow hovering on full world tile area
		if gridX >= -250 && gridX <= 250 && gridY >= -250 && gridY <= 250 {
			c.IsHoveringTile = true
			c.HoveredTileX = gridX
			c.HoveredTileY = gridY
		} else {
			c.IsHoveringTile = false
		}
	} else if gameRoom.IsWalkable(gridX, gridY) {
		// Non-editor mode - only allow hovering on walkable tiles for navigation
		c.IsHoveringTile = true
		c.HoveredTileX = gridX
		c.HoveredTileY = gridY
	} else {
		c.IsHoveringTile = false
	}
	
	// Handle drag updates for asset preview
	if c.DraggingAssetPreview && c.EditingAsset != nil {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			// Update drag - calculate new offset based on mouse position
			c.updateAssetPreviewDrag(mouseX, mouseY, screenWidth, screenHeight)
		} else {
			// End drag
			c.DraggingAssetPreview = false
		}
		// Skip other mouse handling while dragging
		mode := "Continuous"
		if c.GridMovement {
			mode = "Grid"
		}
		return "Editor Mode (" + mode + ") - Dragging", newCamX, newCamY
	}

	// Handle mouse clicks (only when not dragging) - single click detection
	leftClicked := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && !c.MousePressed[ebiten.MouseButtonLeft]
	if leftClicked {
		// Check toolbar clicks first (highest priority)
		toolbarAction := c.UIRenderer.HandleToolbarClick(mouseX, mouseY, screenWidth, screenHeight)
		if toolbarAction != "" {
			c.handleToolbarAction(toolbarAction, gameRoom, db)
		} else if c.InspectionDialog && c.InspectedObject != nil {
			// Check inspection dialog clicks before inspection mode (dialog has higher priority)
			action := c.UIRenderer.HandleInspectionDialogClick(mouseX, mouseY, screenWidth, screenHeight)
			fmt.Printf("Inspection dialog click action: %s\n", action)
			if action == "edit" {
				// Load the asset properties for editing (only available in editor mode)
				if c.EditorMode {
					entity, err := db.GetEntityByAssetPath(c.InspectedObject.AssetPath)
					if err != nil {
						fmt.Printf("Failed to get entity properties: %v\n", err)
					} else {
						c.EditingAsset = entity
						c.AssetPropertyDialog = true
						c.InspectionDialog = false // Close inspection dialog
						fmt.Printf("Opening asset property editor for: %s\n", entity.Name)
					}
				} else {
					fmt.Printf("Asset editing only available in editor mode\n")
					c.InspectionDialog = false // Close dialog in main game mode
				}
			} else if action == "delete" {
				// Delete the inspected object (only available in editor mode)
				if c.EditorMode {
					c.deleteTile(gameRoom, db, c.InspectedObject.X, c.InspectedObject.Y)
					c.InspectionDialog = false
					c.InspectedObject = nil
				} else {
					fmt.Printf("Object deletion only available in editor mode\n")
					c.InspectionDialog = false // Close dialog in main game mode
				}
			}
			// If action is "consume" or any other value, just consume the click
		} else if c.EditorMode && c.AssetPropertyDialog && c.EditingAsset != nil {
			action := c.UIRenderer.HandleAssetPropertyEditorClick(mouseX, mouseY, screenWidth, screenHeight)
			if action == "save" {
				fmt.Printf("ASSET SAVE: Starting save process for %s with offsets (%.3f, %.3f)\n", c.EditingAsset.Name, c.EditingAsset.DefaultOffsetX, c.EditingAsset.DefaultOffsetY)
				
				// Close dialog immediately for responsive UI
				c.AssetPropertyDialog = false
				editingAsset := c.EditingAsset
				c.EditingAsset = nil
				
				// Save the asset properties to database asynchronously
				go func() {
					err := db.UpdateEntityProperties(editingAsset)
					if err != nil {
						fmt.Printf("Failed to save asset properties: %v\n", err)
					} else {
						fmt.Printf("Saved asset properties and rotation %d° offsets for: %s\n", editingAsset.CurrentRotation, editingAsset.Name)
						
						// Trigger world tile reload if callback is set
						if c.OnWorldTileReloadNeeded != nil {
							c.OnWorldTileReloadNeeded()
						}
					}
				}()
			} else if action == "cancel" {
				c.AssetPropertyDialog = false
				c.EditingAsset = nil
			} else if action == "drag_preview" {
				// Start dragging the asset preview
				c.DraggingAssetPreview = true
				c.AssetDragStartX = mouseX
				c.AssetDragStartY = mouseY
				// Capture the starting offset values
				c.DragStartOffsetX = c.EditingAsset.DefaultOffsetX
				c.DragStartOffsetY = c.EditingAsset.DefaultOffsetY
			}
		} else if c.EditorMode && c.UIRenderer.HandlePositionDialogClick(mouseX, mouseY, screenWidth, screenHeight) {
			// Position dialog handled the click
		} else if c.EditorMode && c.UIRenderer.HandleSidebarClick(mouseX, mouseY) {
			// Sidebar handled the click
		} else if c.InspectionMode && c.IsHoveringTile {
			// In inspection mode, left-click should inspect objects (lower priority than dialogs)
			// Check if click is not on sidebar or other UI elements
			if !c.TileInventory.SidebarVisible || mouseX > c.TileInventory.SidebarWidth {
				fmt.Printf("INSPECTION MODE: Left-click attempting inspection at (%d, %d)\n", gridX, gridY)
				c.handleInspectionClick(gameRoom, db, gridX, gridY)
			} else {
				fmt.Printf("INSPECTION MODE: Left-click blocked by sidebar at mouseX=%d (sidebarWidth=%d)\n", mouseX, c.TileInventory.SidebarWidth)
			}
		} else if !c.EditorMode && c.IsHoveringTile {
			// Player movement (play mode only)
			fmt.Printf("Mouse click at grid (%d, %d), current player at (%.1f, %.1f)\n", gridX, gridY, player.X, player.Y)
			
			// Check if the target tile is walkable
			walkable := gameRoom.IsWalkable(gridX, gridY)
			fmt.Printf("NAVIGATION CHECK: Tile (%d, %d) walkable=%v in room adapter\n", gridX, gridY, walkable)
			
			// Debug: Show walkability map size for troubleshooting
			fmt.Printf("NAVIGATION DEBUG: Room adapter has %d walkability entries\n", len(gameRoom.Walkability))
			// Show a few key entries around the clicked tile
			for dx := -1; dx <= 1; dx++ {
				for dy := -1; dy <= 1; dy++ {
					checkX, checkY := gridX+dx, gridY+dy
					key := fmt.Sprintf("%d:%d", checkX, checkY)
					if walkableVal, exists := gameRoom.Walkability[key]; exists {
						fmt.Printf("NAVIGATION DEBUG: Walkability[%s] = %v\n", key, walkableVal)
					} else {
						fmt.Printf("NAVIGATION DEBUG: Walkability[%s] = NOT_FOUND (defaults to false)\n", key)
					}
				}
			}
			if walkable {
				// Update facing direction based on movement
				dx := float64(gridX) - player.X
				dy := float64(gridY) - player.Y
				player.SetFacingFromMovement(dx, dy)
				
				player.TargetX = float64(gridX)
				player.TargetY = float64(gridY)
				player.Animating = true
				fmt.Printf("Set target to (%.1f, %.1f), animating: %t\n", player.TargetX, player.TargetY, player.Animating)
			} else {
				fmt.Printf("Cannot move to (%d, %d) - tile is not walkable\n", gridX, gridY)
			}
		} else if c.EditorMode && c.IsHoveringTile && c.ShowDebugGrid {
			// Boundary editing mode - left click allows tiles (highest priority when F2 grid is visible)
			gameRoom.SetWalkable(gridX, gridY, true)
			fmt.Printf("BOUNDARY EDIT: Allowed tile at (%d, %d)\n", gridX, gridY)
			// Notify editor to update actual world tile and persist changes
			if c.OnBoundaryChanged != nil {
				c.OnBoundaryChanged(gridX, gridY, true)
			}
		} else if c.EditorMode && c.IsHoveringTile && c.InspectionMode {
			// Inspection mode - left click to inspect
			fmt.Printf("INSPECT MODE CLICK: Attempting inspection at (%d, %d)\n", gridX, gridY)
			c.handleInspectionClick(gameRoom, db, gridX, gridY)
		} else if c.EditorMode && c.IsHoveringTile && c.TileInventory.SelectedTile != nil {
			// Tile placement in editor mode
			fmt.Printf("Placing tile at grid (%d, %d), hovered tile is (%d, %d)\n", gridX, gridY, c.HoveredTileX, c.HoveredTileY)
			c.placeTile(gameRoom, db, gridX, gridY)
		}
	}
	
	// Handle right click for tile deletion in editor mode - single click detection
	rightClicked := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) && !c.MousePressed[ebiten.MouseButtonRight]
	if c.EditorMode && rightClicked && c.IsHoveringTile {
		// Check if F2 boundary editing mode is active
		if c.ShowDebugGrid {
			// Boundary editing mode - right click blocks tiles
			gameRoom.SetWalkable(gridX, gridY, false)
			fmt.Printf("BOUNDARY EDIT: Blocked tile at (%d, %d)\n", gridX, gridY)
			// Notify editor to update actual world tile and persist changes
			if c.OnBoundaryChanged != nil {
				c.OnBoundaryChanged(gridX, gridY, false)
			}
		} else {
			// Normal mode - right click deletes tiles
			c.deleteTile(gameRoom, db, gridX, gridY)
		}
	}
	
	// Handle middle-click camera dragging (like Dota 2 with inverted Y-axis)
	middlePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle)
	middleClicked := middlePressed && !c.MousePressed[ebiten.MouseButtonMiddle]
	middleReleased := !middlePressed && c.MousePressed[ebiten.MouseButtonMiddle]
	
	// Debug output for camera dragging (commented out to reduce console spam)
	// if middleClicked {
	//     fmt.Printf("ROOM DEBUG: Middle mouse CLICKED at (%d, %d), dragging=%v\n", mouseX, mouseY, c.CameraDragging)
	// }
	
	if middleClicked {
		// Start camera dragging
		c.CameraDragging = true
		c.CameraDragStartX = mouseX
		c.CameraDragStartY = mouseY
		c.CameraDragCamStartX = camX
		c.CameraDragCamStartY = camY
		// fmt.Printf("ROOM CAMERA DRAG: Started at mouse (%d, %d), cam (%.2f, %.2f)\n", mouseX, mouseY, camX, camY)
	} else if middleReleased {
		// Stop camera dragging
		c.CameraDragging = false
		// fmt.Printf("CAMERA DRAG: Stopped\n")
	}
	
	// Update camera position if dragging
	if c.CameraDragging && middlePressed {
		// Calculate mouse movement
		deltaX := float64(mouseX - c.CameraDragStartX)
		deltaY := float64(mouseY - c.CameraDragStartY)
		
		// Apply movement to camera with inverted axes (Dota 2 style)
		// When mouse moves right, camera moves left (inverted X)
		// When mouse moves up, camera moves down (inverted Y)
		newCamX = c.CameraDragCamStartX - deltaX  // Inverted: mouse right = camera left (negative delta)
		newCamY = c.CameraDragCamStartY - deltaY  // Inverted: mouse up = camera down (negative delta)
		
		// fmt.Printf("CAMERA DRAG: Mouse delta (%.0f, %.0f), new cam (%.2f, %.2f)\n", deltaX, deltaY, newCamX, newCamY)
	}
	
	// Handle asset dragging in position dialog
	if c.EditorMode && c.TileInventory.PositionDialog {
		if c.TileInventory.DraggingAsset && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			// Calculate mouse movement and update asset position
			c.handleAssetDrag(mouseX, mouseY, screenWidth, screenHeight)
		} else if c.TileInventory.DraggingAsset {
			// Mouse released, stop dragging
			c.TileInventory.DraggingAsset = false
		}
	}

	// Update mouse button states for next frame
	c.MousePressed[ebiten.MouseButtonLeft] = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	c.MousePressed[ebiten.MouseButtonRight] = ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	c.MousePressed[ebiten.MouseButtonMiddle] = ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle)
	
	// Handle scroll wheel zoom
	_, scrollY := ebiten.Wheel()
	if scrollY != 0 {
		const zoomSpeed = 0.1
		const minZoom = 0.5  // Can zoom out to 50%
		const maxZoom = 3.0  // Can zoom in to 300%
		
		// Zoom in/out based on scroll direction
		if scrollY > 0 {
			// Scroll up = zoom in
			c.ZoomLevel = c.ZoomLevel + zoomSpeed
			if c.ZoomLevel > maxZoom {
				c.ZoomLevel = maxZoom
			}
		} else {
			// Scroll down = zoom out  
			c.ZoomLevel = c.ZoomLevel - zoomSpeed
			if c.ZoomLevel < minZoom {
				c.ZoomLevel = minZoom
			}
		}
		// Debug output for zoom (can be removed later)
		// fmt.Printf("ZOOM: ScrollY=%.2f, ZoomLevel=%.2f\n", scrollY, c.ZoomLevel)
	}

	// Only handle movement in play mode
	if c.EditorMode {
		mode := "Continuous"
		if c.GridMovement {
			mode = "Grid"
		}
		return "Editor Mode (" + mode + ")", newCamX, newCamY
	}

	if c.GridMovement {
		direction := c.handleGridMovement(player, gameRoom)
		return direction, newCamX, newCamY
	} else {
		direction := c.handleContinuousMovement(player, gameRoom)
		return direction, newCamX, newCamY
	}
}

func (c *Controller) handleGridMovement(player *entity.Player, gameRoom *room.Room) string {
	// Track movement direction for logging
	var movingX, movingY bool
	var directionX, directionY string
	
	// If currently animating, track direction
	if player.Animating {
		dx := player.TargetX - player.X
		dy := player.TargetY - player.Y
		
		// Set movement direction for display
		if dx > 0.1 && dy > 0.1 {
			directionX, directionY = "Right", "Down"
			movingX, movingY = true, true
		} else if dx > 0.1 && dy < -0.1 {
			directionX, directionY = "Right", "Up"
			movingX, movingY = true, true
		} else if dx < -0.1 && dy > 0.1 {
			directionX, directionY = "Left", "Down"
			movingX, movingY = true, true
		} else if dx < -0.1 && dy < -0.1 {
			directionX, directionY = "Left", "Up"
			movingX, movingY = true, true
		} else if dx > 0.1 {
			directionX = "Right"
			movingX = true
		} else if dx < -0.1 {
			directionX = "Left"
			movingX = true
		} else if dy > 0.1 {
			directionY = "Down"
			movingY = true
		} else if dy < -0.1 {
			directionY = "Up"
			movingY = true
		}
	}
	
	// Only accept new movement inputs if not currently animating
	if !player.Animating {
		c.processMovementKeys(player, gameRoom, &movingX, &movingY, &directionX, &directionY, true)
	}

	// Reset key states when keys are released
	c.resetKeyStates()
	
	return c.formatDirection(movingX, movingY, directionX, directionY, "Grid")
}

func (c *Controller) handleContinuousMovement(player *entity.Player, gameRoom *room.Room) string {
	// Track movement direction for logging
	var movingX, movingY bool
	var directionX, directionY string
	
	// If currently animating, track direction
	if player.Animating {
		dx := player.TargetX - player.X
		dy := player.TargetY - player.Y
		
		// Set movement direction for display
		if dx > 0.1 && dy > 0.1 {
			directionX, directionY = "Right", "Down"
			movingX, movingY = true, true
		} else if dx > 0.1 && dy < -0.1 {
			directionX, directionY = "Right", "Up"
			movingX, movingY = true, true
		} else if dx < -0.1 && dy > 0.1 {
			directionX, directionY = "Left", "Down"
			movingX, movingY = true, true
		} else if dx < -0.1 && dy < -0.1 {
			directionX, directionY = "Left", "Up"
			movingX, movingY = true, true
		} else if dx > 0.1 {
			directionX = "Right"
			movingX = true
		} else if dx < -0.1 {
			directionX = "Left"
			movingX = true
		} else if dy > 0.1 {
			directionY = "Down"
			movingY = true
		} else if dy < -0.1 {
			directionY = "Up"
			movingY = true
		}
	}
	
	// Check for new input and calculate new target
	anyKeyPressed := c.processMovementKeys(player, gameRoom, &movingX, &movingY, &directionX, &directionY, false)
	
	// If keys are pressed, set new target and start/continue animation
	if anyKeyPressed {
		// Round target to nearest tile
		player.TargetX = float64(int(player.TargetX + 0.5))
		player.TargetY = float64(int(player.TargetY + 0.5))
		player.Animating = true
	} else if !player.Animating {
		// No keys pressed and not animating - snap to nearest tile
		player.TargetX = float64(int(player.X + 0.5))
		player.TargetY = float64(int(player.Y + 0.5))
		if player.TargetX != player.X || player.TargetY != player.Y {
			player.Animating = true
		}
	}
	
	return c.formatDirection(movingX, movingY, directionX, directionY, "Continuous")
}

func (c *Controller) processMovementKeys(player *entity.Player, gameRoom *room.Room, movingX, movingY *bool, directionX, directionY *string, gridMode bool) bool {
	anyKeyPressed := false
	
	if gridMode {
		newTargetX, newTargetY := player.TargetX, player.TargetY
		
		// Single direction keys
		leftPressed := ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
		rightPressed := ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD)
		upPressed := ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW)
		downPressed := ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS)
		
		if leftPressed && !c.KeyPressed[ebiten.KeyLeft] {
			newTargetX -= 1
			*movingX = true
			*directionX = "Left"
			c.KeyPressed[ebiten.KeyLeft] = true
		}
		if rightPressed && !c.KeyPressed[ebiten.KeyRight] {
			newTargetX += 1
			*movingX = true
			*directionX = "Right"
			c.KeyPressed[ebiten.KeyRight] = true
		}
		if upPressed && !c.KeyPressed[ebiten.KeyUp] {
			newTargetY -= 1
			*movingY = true
			*directionY = "Up"
			c.KeyPressed[ebiten.KeyUp] = true
		}
		if downPressed && !c.KeyPressed[ebiten.KeyDown] {
			newTargetY += 1
			*movingY = true
			*directionY = "Down"
			c.KeyPressed[ebiten.KeyDown] = true
		}

		// Diagonal movement keys
		if ebiten.IsKeyPressed(ebiten.KeyQ) && !c.KeyPressed[ebiten.KeyQ] {
			newTargetX -= 1
			newTargetY -= 1
			*movingX = true
			*movingY = true
			*directionX = "Left"
			*directionY = "Up"
			c.KeyPressed[ebiten.KeyQ] = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyE) && !c.KeyPressed[ebiten.KeyE] {
			newTargetX += 1
			newTargetY -= 1
			*movingX = true
			*movingY = true
			*directionX = "Right"
			*directionY = "Up"
			c.KeyPressed[ebiten.KeyE] = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyZ) && !c.KeyPressed[ebiten.KeyZ] {
			newTargetX -= 1
			newTargetY += 1
			*movingX = true
			*movingY = true
			*directionX = "Left"
			*directionY = "Down"
			c.KeyPressed[ebiten.KeyZ] = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyX) && !c.KeyPressed[ebiten.KeyX] {
			newTargetX += 1
			newTargetY += 1
			*movingX = true
			*movingY = true
			*directionX = "Right"
			*directionY = "Down"
			c.KeyPressed[ebiten.KeyX] = true
		}
		
		// If target changed, check walkability and start animation
		if newTargetX != player.TargetX || newTargetY != player.TargetY {
			// Check if the target tile is walkable (only in play mode, not editor mode)
			if !c.EditorMode && gameRoom != nil && !gameRoom.IsWalkable(int(newTargetX), int(newTargetY)) {
				fmt.Printf("Cannot move to (%.0f, %.0f) - tile is not walkable\n", newTargetX, newTargetY)
				return false
			}
			
			// Update facing direction based on movement
			dx := newTargetX - player.X
			dy := newTargetY - player.Y
			player.SetFacingFromMovement(dx, dy)
			
			player.TargetX = newTargetX
			player.TargetY = newTargetY
			player.Animating = true
		}
	} else {
		// Continuous movement - recalculate direction every frame
		dx, dy := 0.0, 0.0
		
		// Check currently pressed keys and calculate movement direction
		if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
			dx -= 1.0
			*movingX = true
			*directionX = "Left"
			anyKeyPressed = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
			dx += 1.0
			*movingX = true
			*directionX = "Right"
			anyKeyPressed = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
			dy -= 1.0
			*movingY = true
			*directionY = "Up"
			anyKeyPressed = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
			dy += 1.0
			*movingY = true
			*directionY = "Down"
			anyKeyPressed = true
		}

		// Diagonal movement keys (override if pressed)
		if ebiten.IsKeyPressed(ebiten.KeyQ) {
			dx = -1.0
			dy = -1.0
			*movingX = true
			*movingY = true
			*directionX = "Left"
			*directionY = "Up"
			anyKeyPressed = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyE) {
			dx = 1.0
			dy = -1.0
			*movingX = true
			*movingY = true
			*directionX = "Right"
			*directionY = "Up"
			anyKeyPressed = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyZ) {
			dx = -1.0
			dy = 1.0
			*movingX = true
			*movingY = true
			*directionX = "Left"
			*directionY = "Down"
			anyKeyPressed = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyX) {
			dx = 1.0
			dy = 1.0
			*movingX = true
			*movingY = true
			*directionX = "Right"
			*directionY = "Down"
			anyKeyPressed = true
		}
		
		// Update target and facing based on combined input
		if anyKeyPressed {
			newTargetX := player.X + dx
			newTargetY := player.Y + dy
			
			// Check walkability in continuous mode (only in play mode, not editor mode)
			if !c.EditorMode && gameRoom != nil && !gameRoom.IsWalkable(int(newTargetX+0.5), int(newTargetY+0.5)) {
				// Block movement but still update facing
				player.SetFacingFromMovement(dx, dy)
				return false
			}
			
			player.TargetX = newTargetX
			player.TargetY = newTargetY
			// Always update facing when keys are pressed
			player.SetFacingFromMovement(dx, dy)
		}
	}
	
	return anyKeyPressed
}

func (c *Controller) resetKeyStates() {
	// Reset directional keys
	leftPressed := ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
	rightPressed := ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD)
	upPressed := ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW)
	downPressed := ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS)
	
	if !leftPressed { c.KeyPressed[ebiten.KeyLeft] = false }
	if !rightPressed { c.KeyPressed[ebiten.KeyRight] = false }
	if !upPressed { c.KeyPressed[ebiten.KeyUp] = false }
	if !downPressed { c.KeyPressed[ebiten.KeyDown] = false }
	
	// Reset diagonal keys
	if !ebiten.IsKeyPressed(ebiten.KeyQ) { c.KeyPressed[ebiten.KeyQ] = false }
	if !ebiten.IsKeyPressed(ebiten.KeyE) { c.KeyPressed[ebiten.KeyE] = false }
	if !ebiten.IsKeyPressed(ebiten.KeyZ) { c.KeyPressed[ebiten.KeyZ] = false }
	if !ebiten.IsKeyPressed(ebiten.KeyX) { c.KeyPressed[ebiten.KeyX] = false }
}

func (c *Controller) formatDirection(movingX, movingY bool, directionX, directionY, mode string) string {
	if movingX && movingY {
		return directionY + "-" + directionX + " (" + mode + ")"
	} else if movingX {
		return directionX + " (" + mode + ")"
	} else if movingY {
		return directionY + " (" + mode + ")"
	} else {
		return "Idle (" + mode + ")"
	}
}

// placeTile places the selected tile at the specified grid position
func (c *Controller) placeTile(gameRoom *room.Room, db *database.Database, gridX, gridY int) {
	if c.TileInventory.SelectedTile == nil {
		return
	}

	selectedTile := c.TileInventory.SelectedTile
	
	// Get values from asset properties if we're editing one, otherwise use defaults
	var offsetX, offsetY, depth float64
	var rotation int
	
	// Try to load asset properties from database first
	entity, err := db.GetEntityByAssetPath(selectedTile.AssetPath)
	if err == nil && entity != nil {
		// Use the saved asset properties
		offsetX = entity.DefaultOffsetX
		offsetY = entity.DefaultOffsetY
		depth = entity.DefaultDepth
		rotation = entity.CurrentRotation
	} else {
		// Fall back to inventory defaults
		offsetX = c.TileInventory.SelectedPosition.OffsetX
		offsetY = c.TileInventory.SelectedPosition.OffsetY
		depth = c.TileInventory.SelectedPosition.Depth
		rotation = c.TileInventory.SelectedRotation
	}
	
	// Create tile data with positioning information
	tileData := room.TileData{
		AssetPath: selectedTile.AssetPath,
		X:         gridX,
		Y:         gridY,
		Facing:    rotation, // Use the rotation/flip position
		Category:  selectedTile.Category,
		Layer:     selectedTile.Layer,
		OffsetX:   offsetX,
		OffsetY:   offsetY,
		Depth:     depth,
		Rotation:  0, // Legacy field, no longer used
	}
	
	// Debug logging for tile placement
	fmt.Printf("TILE PLACEMENT: Placing at Grid(%d,%d) with Offset(%.2f,%.2f)\n", 
		gridX, gridY, offsetX, offsetY)

	// Add tile to room
	addErr := gameRoom.AddTile(tileData)
	if addErr != nil {
		fmt.Printf("Error placing tile: %v\n", addErr)
	} else {
		// Save to database asynchronously to avoid blocking the UI
		go func() {
			if saveErr := db.SaveWorldTileObject(gameRoom.Name, &tileData); saveErr != nil {
				fmt.Printf("Error saving tile to database: %v\n", saveErr)
			} else {
				// Trigger world tile reload to show the placed item immediately
				if c.OnWorldTileReloadNeeded != nil {
					c.OnWorldTileReloadNeeded()
				}
			}
		}()
		fmt.Printf("Placed %s at GRID (%d, %d) with offset (%.2f, %.2f), depth %.2f, rotation %.1f°\n", 
			selectedTile.Name, gridX, gridY, 
			c.TileInventory.SelectedPosition.OffsetX, 
			c.TileInventory.SelectedPosition.OffsetY,
			c.TileInventory.SelectedPosition.Depth,
			c.TileInventory.SelectedPosition.Rotation * 180 / 3.14159)
	}
}

// deleteTile removes tiles at the specified grid position
func (c *Controller) deleteTile(gameRoom *room.Room, db *database.Database, gridX, gridY int) {
	// Try to remove tiles from all layers at this position
	removed := false
	for layer := 3; layer >= 0; layer-- { // Remove from top layer first
		if _, exists := gameRoom.GetTile(layer, gridX, gridY); exists {
			gameRoom.RemoveTile(layer, gridX, gridY)
			// Also remove from database
			if deleteErr := db.DeleteWorldTileObject(gameRoom.Name, gridX, gridY, layer); deleteErr != nil {
				fmt.Printf("Error removing tile from database: %v\n", deleteErr)
			} else {
				// Trigger world tile reload to show the removal immediately
				if c.OnWorldTileReloadNeeded != nil {
					c.OnWorldTileReloadNeeded()
				}
			}
			fmt.Printf("Removed tile from layer %d at (%d, %d)\n", layer, gridX, gridY)
			removed = true
			break // Remove only one tile per click
		}
	}
	
	if !removed {
		fmt.Printf("No tiles to remove at (%d, %d)\n", gridX, gridY)
	}
}

// handleInspectionClick handles clicking on objects in inspection mode
func (c *Controller) handleInspectionClick(gameRoom *room.Room, db *database.Database, gridX, gridY int) {
	fmt.Printf("INSPECTION: Looking for object at tile (%d, %d) in room '%s'\n", gridX, gridY, gameRoom.Name)
	
	// Use world tile lookup exclusively (room system has been deprecated)
	obj, err := db.GetWorldTileObjectAt(gameRoom.Name, gridX, gridY)
	if err != nil {
		fmt.Printf("INSPECTION ERROR: Getting object for inspection: %v\n", err)
		return
	}
	
	if obj == nil {
		fmt.Printf("INSPECTION: No object found at tile (%d, %d) in world tile '%s'\n", gridX, gridY, gameRoom.Name)
		return
	}
	
	// Set the inspected object and open dialog
	c.InspectedObject = obj
	c.InspectionDialog = true
	fmt.Printf("INSPECTION SUCCESS: Inspecting %s assigned to tile (%d, %d)\n", obj.AssetPath, gridX, gridY)
}

// saveInspectedObjectChanges saves changes made to an inspected object back to room and database
func (c *Controller) saveInspectedObjectChanges(gameRoom *room.Room, db *database.Database) {
	if c.InspectedObject == nil {
		return
	}

	// Update the inspected object with new position data
	c.InspectedObject.OffsetX = c.TileInventory.SelectedPosition.OffsetX
	c.InspectedObject.OffsetY = c.TileInventory.SelectedPosition.OffsetY
	c.InspectedObject.Depth = c.TileInventory.SelectedPosition.Depth
	c.InspectedObject.Rotation = c.TileInventory.SelectedPosition.Rotation

	// Update the object in the room
	key := room.GetTileKey(c.InspectedObject.Layer, c.InspectedObject.X, c.InspectedObject.Y)
	gameRoom.Tiles[key] = *c.InspectedObject

	// Save changes to database asynchronously
	inspectedObj := *c.InspectedObject // Copy for async operation
	roomName := gameRoom.Name
	go func() {
		if err := db.SaveObject(roomName, &inspectedObj); err != nil {
			fmt.Printf("Error saving object changes to database: %v\n", err)
		} else {
			fmt.Printf("Saved changes to %s at (%d, %d)\n", inspectedObj.AssetPath, inspectedObj.X, inspectedObj.Y)
		}
	}()

	// Reset editing state
	c.EditingInspected = false
}

// handleAssetDrag updates asset position based on mouse drag in position dialog
func (c *Controller) handleAssetDrag(mouseX, mouseY, screenWidth, screenHeight int) {
	dialogWidth := 400.0
	dialogHeight := 350.0
	dialogX := float64(screenWidth)/2 - dialogWidth/2
	dialogY := float64(screenHeight)/2 - dialogHeight/2
	
	tileCenterX := dialogX + dialogWidth/2
	tileCenterY := dialogY + 120
	tileSize := 120.0
	halfSize := tileSize / 2
	
	// Convert mouse position relative to tile center
	relativeX := float64(mouseX) - tileCenterX
	relativeY := float64(mouseY) - tileCenterY
	
	// Convert screen coordinates back to standard (0,0)-(1,1) offset system
	// Account for isometric Y scaling  
	normalizedX := relativeX / halfSize        // Screen offset to relative (-0.5 to 0.5)
	normalizedY := (relativeY / halfSize) * 2.0 // Reverse the 0.5 scaling used in drawing
	
	// Convert relative (-0.5 to 0.5) back to standard (0 to 1) offset system
	offsetX := normalizedX + 0.5  // -0.5->0, 0->0.5, 0.5->1
	offsetY := normalizedY + 0.5  // -0.5->0, 0->0.5, 0.5->1
	
	// Clamp to standard bounds (0 to 1, with slight extension for flexibility)
	if offsetX < -0.1 { offsetX = -0.1 }
	if offsetX > 1.1 { offsetX = 1.1 }
	if offsetY < -0.1 { offsetY = -0.1 }
	if offsetY > 1.1 { offsetY = 1.1 }
	
	// Update position and set to manual mode
	c.TileInventory.SelectedPosition.OffsetX = offsetX
	c.TileInventory.SelectedPosition.OffsetY = offsetY
	c.TileInventory.SelectedPosition.SnapToEdge = 0 // Manual positioning
}

// handleToolbarAction processes toolbar button clicks
func (c *Controller) handleToolbarAction(action string, gameRoom *room.Room, db *database.Database) {
	switch action {
	case "toggle_editor":
		c.EditorMode = !c.EditorMode
		if !c.EditorMode {
			// Clear editor state when exiting editor mode
			c.InspectionDialog = false
			c.AssetPropertyDialog = false
			c.InspectedObject = nil
			c.EditingAsset = nil
			c.EditingInspected = false
			c.TileInventory.SidebarVisible = false
			c.TileInventory.PositionDialog = false
		}
		fmt.Printf("Editor mode: %v\n", c.EditorMode)
		
	case "toggle_inspect":
		c.InspectionMode = !c.InspectionMode
		fmt.Printf("INSPECT BUTTON: Inspection mode toggled to: %v\n", c.InspectionMode)
		
	case "toggle_inventory":
		c.TileInventory.ToggleSidebar()
		fmt.Printf("Inventory visible: %v\n", c.TileInventory.SidebarVisible)
		
	case "toggle_grid":
		if c.EditorMode {
			c.ShowDebugGrid = !c.ShowDebugGrid
			fmt.Printf("Debug grid: %v\n", c.ShowDebugGrid)
		}
		
	case "save_room":
		if c.EditorMode && gameRoom != nil {
			// Convert Room to WorldTile and save to database (unified persistence)
			worldTile := convertRoomToWorldTile(gameRoom)
			err := db.SaveWorldTile(worldTile)
			if err != nil {
				fmt.Printf("Failed to save world tile to database: %v\n", err)
			} else {
				fmt.Printf("World tile %s saved to database\n", worldTile.ID)
			}
		}
	}
}

// updateAssetPreviewDrag updates asset offset values using smooth delta movement
func (c *Controller) updateAssetPreviewDrag(mouseX, mouseY, screenWidth, screenHeight int) {
	if c.EditingAsset == nil {
		return
	}

	// Only update if there's been actual movement (prevent snap-to-cursor)
	dragThreshold := 3 // pixels
	if abs(mouseX-c.AssetDragStartX) < dragThreshold && abs(mouseY-c.AssetDragStartY) < dragThreshold {
		return // Not enough movement to constitute a drag
	}

	// Calculate mouse movement delta from drag start
	deltaScreenX := float64(mouseX - c.AssetDragStartX)
	deltaScreenY := float64(mouseY - c.AssetDragStartY)
	
	// Convert screen pixel delta to isometric grid delta
	// Reverse the isometric transformation: ix = (x - y) * 32, iy = (x + y) * 16
	// Solving for grid deltas: x = (ix/32 + iy/16) / 2, y = (iy/16 - ix/32) / 2
	deltaGridX := (deltaScreenX/32 + deltaScreenY/16) / 2
	deltaGridY := (deltaScreenY/16 - deltaScreenX/32) / 2
	
	// Apply delta to starting position (smooth movement)
	// Convert grid deltas back to 0-1 offset range
	newOffsetX := c.DragStartOffsetX + deltaGridX
	newOffsetY := c.DragStartOffsetY + deltaGridY
	
	// Allow dragging anywhere in the preview window (expanded bounds)
	// Extend bounds to allow positioning outside the tile
	if newOffsetX < -1.0 { newOffsetX = -1.0 }
	if newOffsetX > 2.0 { newOffsetX = 2.0 }
	if newOffsetY < -1.0 { newOffsetY = -1.0 }
	if newOffsetY > 2.0 { newOffsetY = 2.0 }
	
	// Update the asset's default offset values
	c.EditingAsset.DefaultOffsetX = newOffsetX
	c.EditingAsset.DefaultOffsetY = newOffsetY
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// updatePlacedAssets updates all placed instances of an asset with new properties
func (c *Controller) updatePlacedAssets(gameRoom *room.Room, db *database.Database, assetPath string) {
	fmt.Printf("ASSET UPDATE: Reloading objects for room: '%s'\n", gameRoom.Name)
	// Reload all objects from database to get updated properties
	objects, err := db.LoadObjects(gameRoom.Name)
	if err != nil {
		fmt.Printf("Failed to reload objects: %v\n", err)
		return
	}
	fmt.Printf("ASSET UPDATE: Found %d objects to reload\n", len(objects))
	
	// Clear current room tiles
	gameRoom.Tiles = make(map[string]room.TileData)
	
	// Reload all tiles with updated properties
	for _, obj := range objects {
		gameRoom.PlaceTile(obj)
	}
	
	fmt.Printf("Updated all placed instances of %s in room\n", assetPath)
}


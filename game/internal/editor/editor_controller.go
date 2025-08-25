package editor

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tcornell05/go/game/internal/database"
	"github.com/tcornell05/go/game/internal/room"
	"github.com/tcornell05/go/game/internal/world"
)

// EditorController handles input for both world and room editing
type EditorController struct {
	KeyPressed   map[ebiten.Key]bool
	MousePressed map[ebiten.MouseButton]bool
	
	// World editor state
	SelectedTileID   string
	ShowConnections  bool
	ShowGrid        bool
	
	// Room editor state (simplified from input.Controller)
	EditorMode      bool
	GridMovement    bool
	HoveredTileX    int
	HoveredTileY    int
	IsHoveringTile  bool
	
	// Camera dragging state (same as input.Controller)
	CameraDragging       bool
	CameraDragStartX     int
	CameraDragStartY     int
	CameraDragCamStartX  float64
	CameraDragCamStartY  float64
	
	// Camera zoom state
	ZoomLevel            float64  // 1.0 = normal, > 1.0 = zoomed in, < 1.0 = zoomed out
}

// NewEditorController creates a new editor controller
func NewEditorController() *EditorController {
	return &EditorController{
		KeyPressed:      make(map[ebiten.Key]bool),
		MousePressed:    make(map[ebiten.MouseButton]bool),
		ShowConnections: true,
		ShowGrid:       true,
		EditorMode:     true, // Always in editor mode
		ZoomLevel:      1.0,  // Normal zoom level
	}
}

// HandleWorldInput processes input for world editing mode
func (ec *EditorController) HandleWorldInput(worldMap *world.WorldMap, db *database.Database, camX, camY float64) (newCamX, newCamY float64) {
	// Initialize camera coordinates (will be updated by camera dragging if active)
	newCamX = camX
	newCamY = camY
	
	// Get mouse position for camera dragging
	mouseX, mouseY := ebiten.CursorPosition()
	
	// Handle middle-click camera dragging (same as input.Controller)
	middlePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle)
	middleClicked := middlePressed && !ec.MousePressed[ebiten.MouseButtonMiddle]
	middleReleased := !middlePressed && ec.MousePressed[ebiten.MouseButtonMiddle]
	
	if middleClicked {
		// Start camera dragging
		ec.CameraDragging = true
		ec.CameraDragStartX = mouseX
		ec.CameraDragStartY = mouseY
		ec.CameraDragCamStartX = camX
		ec.CameraDragCamStartY = camY
		// Debug output to confirm it works
		// fmt.Printf("WORLD CAMERA DRAG: Started at mouse (%d, %d), cam (%.2f, %.2f)\n", mouseX, mouseY, camX, camY)
	} else if middleReleased {
		// Stop camera dragging
		ec.CameraDragging = false
		// fmt.Printf("WORLD CAMERA DRAG: Stopped\n")
	}
	
	// Update camera position if dragging
	if ec.CameraDragging && middlePressed {
		// Calculate mouse movement
		deltaX := float64(mouseX - ec.CameraDragStartX)
		deltaY := float64(mouseY - ec.CameraDragStartY)
		
		// Apply movement to camera with inverted axes (Dota 2 style)
		newCamX = ec.CameraDragCamStartX - deltaX  // Inverted: mouse right = camera left (negative delta)
		newCamY = ec.CameraDragCamStartY - deltaY  // Inverted: mouse up = camera down (negative delta)
		
		// Debug output to see if it's working
		// fmt.Printf("WORLD CAMERA DRAG: Mouse delta (%.0f, %.0f), new cam (%.2f, %.2f)\n", deltaX, deltaY, newCamX, newCamY)
	}
	
	// Update middle mouse button state for next frame
	ec.MousePressed[ebiten.MouseButtonMiddle] = middlePressed
	
	// Handle tile selection
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && !ec.MousePressed[ebiten.MouseButtonLeft] {
		x, y := ebiten.CursorPosition()
		tileID := ec.getTileIDFromPosition(x, y, worldMap)
		if tileID != "" {
			ec.SelectedTileID = tileID
		}
		ec.MousePressed[ebiten.MouseButtonLeft] = true
	} else if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		ec.MousePressed[ebiten.MouseButtonLeft] = false
	}
	
	// Toggle connections display
	if ebiten.IsKeyPressed(ebiten.KeyC) && !ec.KeyPressed[ebiten.KeyC] {
		ec.ShowConnections = !ec.ShowConnections
		ec.KeyPressed[ebiten.KeyC] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyC) {
		ec.KeyPressed[ebiten.KeyC] = false
	}
	
	// Toggle grid display
	if ebiten.IsKeyPressed(ebiten.KeyG) && !ec.KeyPressed[ebiten.KeyG] {
		ec.ShowGrid = !ec.ShowGrid
		ec.KeyPressed[ebiten.KeyG] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyG) {
		ec.KeyPressed[ebiten.KeyG] = false
	}
	
	// Save world
	if ebiten.IsKeyPressed(ebiten.KeyS) && ebiten.IsKeyPressed(ebiten.KeyControl) {
		if !ec.KeyPressed[ebiten.KeyS] {
			if err := db.SaveWorldMap(worldMap); err != nil {
				// TODO: Show error message in UI
			}
			ec.KeyPressed[ebiten.KeyS] = true
		}
	} else if !ebiten.IsKeyPressed(ebiten.KeyS) {
		ec.KeyPressed[ebiten.KeyS] = false
	}
	
	// Handle scroll wheel zoom (same as input.Controller)
	_, scrollY := ebiten.Wheel()
	if scrollY != 0 {
		const zoomSpeed = 0.1
		const minZoom = 0.5  // Can zoom out to 50%
		const maxZoom = 3.0  // Can zoom in to 300%
		
		// Zoom in/out based on scroll direction
		if scrollY > 0 {
			// Scroll up = zoom in
			ec.ZoomLevel = ec.ZoomLevel + zoomSpeed
			if ec.ZoomLevel > maxZoom {
				ec.ZoomLevel = maxZoom
			}
		} else {
			// Scroll down = zoom out  
			ec.ZoomLevel = ec.ZoomLevel - zoomSpeed
			if ec.ZoomLevel < minZoom {
				ec.ZoomLevel = minZoom
			}
		}
	}
	
	// Return the camera coordinates (potentially modified by dragging)
	return newCamX, newCamY
}

// HandleRoomInput processes input for room editing mode (simplified)
func (ec *EditorController) HandleRoomInput(gameRoom *room.Room, db *database.Database) {
	// Simplified room input handling
	// This would be expanded based on the existing input.Controller logic
	
	// Toggle movement mode
	if ebiten.IsKeyPressed(ebiten.KeyGraveAccent) && !ec.KeyPressed[ebiten.KeyGraveAccent] {
		ec.GridMovement = !ec.GridMovement
		ec.KeyPressed[ebiten.KeyGraveAccent] = true
	} else if !ebiten.IsKeyPressed(ebiten.KeyGraveAccent) {
		ec.KeyPressed[ebiten.KeyGraveAccent] = false
	}
}

// getTileIDFromPosition converts screen position to tile ID
func (ec *EditorController) getTileIDFromPosition(screenX, screenY int, worldMap *world.WorldMap) string {
	// Simple grid-based calculation
	tileWidth := 80
	tileHeight := 60
	
	gridX := screenX / tileWidth
	gridY := screenY / tileHeight
	
	if gridX >= 0 && gridX < worldMap.Width && gridY >= 0 && gridY < worldMap.Height {
		return world.GenerateTileID(gridX, gridY, 0) // Floor 0 for now
	}
	
	return ""
}
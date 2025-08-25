package game

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tcornell05/go/game/internal/database"
	"github.com/tcornell05/go/game/internal/entity"
	"github.com/tcornell05/go/game/internal/input"
	"github.com/tcornell05/go/game/internal/render"
	"github.com/tcornell05/go/game/internal/room"
	"github.com/tcornell05/go/game/pkg/assets"
)

const (
	ScreenWidth  = 1024
	ScreenHeight = 768
)

type Game struct {
	player     *entity.Player
	controller *input.Controller
	renderer   *render.Renderer
	room       *room.Room
	database   *database.Database
	camX, camY float64
}

func New() *Game {
	// Load tiles
	floorTile, err := assets.LoadTile("assets/Foor-Wall Tiles 64px/Floor_1_Tile(64).png")
	if err != nil {
		log.Fatal("Failed to load floor tile:", err)
	}

	wallTile1, err := assets.LoadTile("assets/Foor-Wall Tiles 64px/Wall_1_Tile(64).png")
	if err != nil {
		log.Fatal("Failed to load wall tile 1:", err)
	}

	wallTile2, err := assets.LoadTile("assets/Foor-Wall Tiles 64px/Wall_2_Tile(64).png")
	if err != nil {
		log.Fatal("Failed to load wall tile 2:", err)
	}

	// Create player
	player, err := entity.NewPlayer(5, 5)
	if err != nil {
		log.Fatal("Failed to create player:", err)
	}

	// Create systems
	controller, err := input.NewController()
	if err != nil {
		log.Fatal("Failed to create controller:", err)
	}
	renderer := render.NewRenderer(floorTile, wallTile1, wallTile2)
	
	// Initialize database
	db, err := database.NewDatabase("data/game.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Create default room
	gameRoom := room.NewRoom("Default Room", 15, 15)
	
	// Load existing objects from database
	if objects, err := db.LoadObjects(gameRoom.Name); err == nil {
		for _, obj := range objects {
			fmt.Printf("LOADING FROM DB: Tile at (%d,%d) with offset (%.2f,%.2f) - %s\n", 
				obj.X, obj.Y, obj.OffsetX, obj.OffsetY, obj.AssetPath)
			gameRoom.PlaceTile(obj)
		}
	}

	return &Game{
		player:     player,
		controller: controller,
		renderer:   renderer,
		room:       gameRoom,
		database:   db,
		camX:       0,
		camY:       0,
	}
}

func (g *Game) Update() error {
	// Update cursor mode based on inspection state
	if g.controller.InspectionMode {
		ebiten.SetCursorMode(ebiten.CursorModeHidden)
	} else {
		ebiten.SetCursorMode(ebiten.CursorModeVisible) 
	}

	// Handle input and get current direction and camera updates
	direction, newCamX, newCamY := g.controller.HandleInput(g.player, g.room, g.database, g.camX, g.camY, ScreenWidth, ScreenHeight)
	g.player.CurrentDirection = direction
	g.camX = newCamX
	g.camY = newCamY

	// Update player animation
	g.player.UpdateAnimation()

	// Keep player within bounds using walkability data
	g.player.KeepInBounds(g.controller.GridMovement, g.room)

	// Camera position is now controlled by middle-mouse dragging
	// Don't reset camera coordinates here - they are managed by input handling

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Draw(screen, g.player, g.controller, g.room, g.camX, g.camY, ScreenWidth, ScreenHeight)
	
	// Always draw the toolbar (even outside editor mode)
	g.controller.UIRenderer.DrawToolbar(screen, g.controller.EditorMode, g.controller.InspectionMode, ScreenWidth, ScreenHeight)
	
	// Always draw sidebar/inventory drawer when visible (in both editor and main game mode)
	g.controller.UIRenderer.DrawSidebar(screen, ScreenWidth, ScreenHeight)
	
	// Draw inspection dialog if active (works in both editor and main game mode)
	if g.controller.InspectionDialog && g.controller.InspectedObject != nil {
		g.controller.UIRenderer.DrawInspectionDialog(screen, g.controller.InspectedObject, ScreenWidth, ScreenHeight)
	}
	
	// Draw inspection mode border when in inspection mode
	if g.controller.InspectionMode {
		g.controller.UIRenderer.DrawInspectionModeBorder(screen, ScreenWidth, ScreenHeight)
		// Draw magnifying glass cursor at mouse position
		mouseX, mouseY := ebiten.CursorPosition()
		g.drawMagnifyingGlassCursor(screen, mouseX, mouseY)
	}
	
	// Draw editor-specific UI on top (only in editor mode)
	if g.controller.EditorMode {
		g.controller.UIRenderer.DrawPositionDialog(screen, ScreenWidth, ScreenHeight)
		
		// Draw asset property editor if active
		if g.controller.AssetPropertyDialog && g.controller.EditingAsset != nil {
			g.controller.UIRenderer.DrawAssetPropertyEditor(screen, g.controller.EditingAsset, g.controller.DraggingAssetPreview, g.controller.ZoomLevel, ScreenWidth, ScreenHeight)
		}
		
		// Inspection mode border removed - now using middle-click for direct inspection
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// drawMagnifyingGlassCursor draws a magnifying glass cursor at the specified position
func (g *Game) drawMagnifyingGlassCursor(screen *ebiten.Image, x, y int) {
	g.controller.UIRenderer.DrawMagnifyingGlass(screen, x, y, true)
}
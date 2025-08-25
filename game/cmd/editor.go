package cmd

import (
	"embed"
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
	"github.com/tcornell05/go/game/internal/editor"
	"github.com/tcornell05/go/game/internal/input"
	"github.com/tcornell05/go/game/internal/render"
	"github.com/tcornell05/go/game/pkg/assets"
)

var (
	worldFile string
	tileID    string
	newTile   bool
	roomWidth int
	roomHeight int
)

var editorCmd = &cobra.Command{
	Use:   "editor",
	Short: "Launch the world and room editor",
	Long: `Launch the editor mode to:
- Design and configure world tiles (no tile specified)
- Edit a specific room tile (with --tile A1)
- Create new tiles (with --new-tile --tile A1)
- Set custom room dimensions (--width 20 --height 20)
- Full isometric editor with asset placement and room decoration`,
	Run: runEditor,
}

func init() {
	rootCmd.AddCommand(editorCmd)
	
	// Editor-specific flags
	editorCmd.Flags().StringVarP(&worldFile, "world", "w", "default", "World configuration to edit")
	editorCmd.Flags().StringVarP(&tileID, "tile", "t", "", "Specific tile to edit (e.g., A1, B2)")
	editorCmd.Flags().BoolVar(&newTile, "new-tile", false, "Create a new tile (requires --tile)")
	editorCmd.Flags().IntVar(&roomWidth, "width", 15, "Room width for new tiles")
	editorCmd.Flags().IntVar(&roomHeight, "height", 15, "Room height for new tiles")
}

func runEditor(cmd *cobra.Command, args []string) {
	// Validate new tile flag
	if newTile && tileID == "" {
		log.Fatal("--new-tile requires --tile to be specified")
	}
	
	if newTile {
		log.Printf("Creating new tile %s in world: %s (%dx%d)\n", tileID, worldFile, roomWidth, roomHeight)
	} else if tileID != "" {
		log.Printf("Starting editor mode for tile %s in world: %s\n", tileID, worldFile)
	} else {
		log.Printf("Starting world editor mode for world: %s\n", worldFile)
	}
	
	// Initialize editor components
	roomController, err := input.NewController()
	if err != nil {
		log.Fatal("Failed to create room controller:", err)
	}
	
	// Set editor mode to enable editor-specific features like asset property editing
	roomController.EditorMode = true
	
	// Load tiles for renderer
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
	
	renderer := render.NewRenderer(floorTile, wallTile1, wallTile2)
	
	// Initialize editor game
	ebiten.SetWindowSize(1024, 768)
	
	if tileID != "" {
		ebiten.SetWindowTitle(fmt.Sprintf("Room Editor - %s:%s", worldFile, tileID))
	} else {
		ebiten.SetWindowTitle("World Editor - " + worldFile)
	}
	
	editorGame, err := editor.NewEditorGame(worldFile, tileID, newTile, roomWidth, roomHeight, roomController, renderer)
	if err != nil {
		log.Fatal("Failed to initialize editor:", err)
	}
	
	// Set up callback so controller can trigger world tile reload when asset properties change
	roomController.OnWorldTileReloadNeeded = func() {
		editorGame.RequestWorldTileReload()
	}
	
	// Set up callback for boundary editing changes
	roomController.OnBoundaryChanged = func(x, y int, walkable bool) {
		editorGame.UpdateWalkability(x, y, walkable)
	}
	
	if err := ebiten.RunGame(editorGame); err != nil {
		log.Fatal(err)
	}
}

// SetEmbeddedAssets sets the embedded assets for the editor
func SetEmbeddedAssets(fs embed.FS) {
	assets.GlobalAssets = fs
}
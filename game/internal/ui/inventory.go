package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tcornell05/go/game/pkg/assets"
)

// TileCategory represents a category of tiles
type TileCategory struct {
	Name  string
	Tiles []TileInfo
}

// TileInfo represents an available tile for placement
type TileInfo struct {
	Name      string
	AssetPath string
	Category  string
	Layer     int // 0=floor, 1=walls, 2=furniture, 3=decorations
	Image     *ebiten.Image
}

// PositionData represents the positioning of an asset within a tile
type PositionData struct {
	OffsetX    float64 // Offset within the tile (-0.5 to 0.5)
	OffsetY    float64 // Offset within the tile (-0.5 to 0.5)
	Depth      float64 // Vertical depth offset (negative = above floor, positive = below)
	Rotation   float64 // Rotation in radians
	SnapToEdge int     // 0=none, 1=top, 2=right, 3=bottom, 4=left, 5=center
}

// TileInventory manages the available tiles for building
type TileInventory struct {
	Categories         []TileCategory
	SelectedTile       *TileInfo
	SelectedRotation   int // 0=normal, 1=flipped horizontally
	PositionDialog     bool // Whether the positioning dialog is open
	SelectedPosition   PositionData
	SidebarVisible     bool
	SidebarWidth       int
	DraggingAsset      bool // Whether the user is dragging the asset in the position dialog
	DragStartX         float64 // Starting position for drag
	DragStartY         float64
}

// NewTileInventory creates a new tile inventory with predefined categories
func NewTileInventory() (*TileInventory, error) {
	inv := &TileInventory{
		Categories:       make([]TileCategory, 0),
		SelectedTile:     nil,
		SelectedRotation: 0,
		PositionDialog:   false,
		SelectedPosition: PositionData{
			OffsetX:    0.5, // Default to center position
			OffsetY:    0.5, // Default to center position
			Depth:      0.0, // Default on floor level
			Rotation:   0.0,
			SnapToEdge: 9, // Default to center
		},
		SidebarVisible:   false,
		SidebarWidth:     300,
	}

	// Load floor tiles
	floorTiles := []TileInfo{
		{"Gray Floor", "assets/Foor-Wall Tiles 64px/Floor_1_Tile(64).png", "floors", 0, nil},
		{"Wood Floor", "assets/Foor-Wall Tiles 64px/Floor_2_Tile(64).png", "floors", 0, nil},
		{"Marble Floor", "assets/Foor-Wall Tiles 64px/Floor_3_Tile(64).png", "floors", 0, nil},
		{"Dark Floor", "assets/Foor-Wall Tiles 64px/Floor_4_Tile(64).png", "floors", 0, nil},
	}

	// Load wall tiles
	wallTiles := []TileInfo{
		{"Orange Wall", "assets/Foor-Wall Tiles 64px/Wall_1_Tile(64).png", "walls", 1, nil},
		{"Blue Wall", "assets/Foor-Wall Tiles 64px/Wall_2_Tile(64).png", "walls", 1, nil},
		{"Green Wall", "assets/Foor-Wall Tiles 64px/Wall_3_Tile(64).png", "walls", 1, nil},
	}

	// Load furniture
	furnitureTiles := []TileInfo{
		{"Lamp A", "assets/Lamp/Lamp_8_A_Tile.png", "furniture", 2, nil},
		{"Lamp B", "assets/Lamp/Lamp_8_B_Tile.png", "furniture", 2, nil},
		{"Chair A", "assets/Chair/Chair_2_A_Tile.png", "furniture", 2, nil},
		{"Chair B", "assets/Chair/Chair_2_B_Tile.png", "furniture", 2, nil},
		{"Chair C", "assets/Chair/Chair_2_C_Tile.png", "furniture", 2, nil},
		{"Chair D", "assets/Chair/Chair_2_D_Tile.png", "furniture", 2, nil},
		{"Desk", "assets/Desk/Desk_1_Tile.png", "furniture", 2, nil},
		{"Desk B", "assets/Desk/Desk_1_B_Tile.png", "furniture", 2, nil},
	}

	// Load plants and decorations
	decorationTiles := []TileInfo{
		{"Plant 1", "assets/Plants/Plant_1_Tile.png", "decorations", 3, nil},
		{"Plant 5", "assets/Plants/Plant_5_Tile.png", "decorations", 3, nil},
		{"Cactus", "assets/Plants/Cactus_4_Tile.png", "decorations", 3, nil},
		{"Small Plant", "assets/Plants/Plant_10.png", "decorations", 3, nil},
	}

	// Create categories
	inv.Categories = []TileCategory{
		{"Floors", floorTiles},
		{"Walls", wallTiles},
		{"Furniture", furnitureTiles},
		{"Decorations", decorationTiles},
	}

	// Load all images
	for catIdx := range inv.Categories {
		for tileIdx := range inv.Categories[catIdx].Tiles {
			tile := &inv.Categories[catIdx].Tiles[tileIdx]
			img, err := assets.LoadTile(tile.AssetPath)
			if err != nil {
				return nil, err
			}
			tile.Image = img
		}
	}

	return inv, nil
}

// ToggleSidebar toggles the visibility of the tile inventory sidebar
func (inv *TileInventory) ToggleSidebar() {
	inv.SidebarVisible = !inv.SidebarVisible
}

// SelectTile selects a tile for placement
func (inv *TileInventory) SelectTile(tile *TileInfo) {
	inv.SelectedTile = tile
	
	// Initialize default position to northwest corner (standard isometric positioning)
	inv.SelectedPosition = PositionData{
		OffsetX:    0.0, // Northwest corner (standard starting position)
		OffsetY:    0.0, // Northwest corner (standard starting position)
		Depth:      0.0, // Floor level
		Rotation:   0.0, // No rotation
		SnapToEdge: 0,   // Manual positioning (no snap)
	}
}

// RotateSelectedTile rotates the currently selected tile
func (inv *TileInventory) RotateSelectedTile() {
	inv.SelectedRotation = (inv.SelectedRotation + 1) % 2 // Only flip for now
}

// GetSelectedTileInfo returns info about the currently selected tile
func (inv *TileInventory) GetSelectedTileInfo() (string, int) {
	if inv.SelectedTile == nil {
		return "None", 0
	}
	return inv.SelectedTile.Name, inv.SelectedRotation
}

// OpenPositionDialog opens the positioning dialog for the selected tile
func (inv *TileInventory) OpenPositionDialog() {
	if inv.SelectedTile != nil {
		inv.PositionDialog = true
	}
}

// ClosePositionDialog closes the positioning dialog
func (inv *TileInventory) ClosePositionDialog() {
	inv.PositionDialog = false
}

// SetPositionOffset sets the position offset within a tile
func (inv *TileInventory) SetPositionOffset(x, y float64) {
	inv.SelectedPosition.OffsetX = x
	inv.SelectedPosition.OffsetY = y
}

// SetSnapToEdge sets the edge snapping mode
func (inv *TileInventory) SetSnapToEdge(edge int) {
	inv.SelectedPosition.SnapToEdge = edge
	
	// Standard isometric tile positioning: (0,0) = northwest, (1,1) = southeast
	// Snap points within the tile bounds
	switch edge {
	case 0: // Manual positioning - keep current offsets
		// Keep current offsets
	case 1: // Northwest corner
		inv.SelectedPosition.OffsetX = 0.0
		inv.SelectedPosition.OffsetY = 0.0
	case 2: // Northeast corner  
		inv.SelectedPosition.OffsetX = 1.0
		inv.SelectedPosition.OffsetY = 0.0
	case 3: // Southeast corner
		inv.SelectedPosition.OffsetX = 1.0
		inv.SelectedPosition.OffsetY = 1.0
	case 4: // Southwest corner
		inv.SelectedPosition.OffsetX = 0.0
		inv.SelectedPosition.OffsetY = 1.0
	case 5: // Center
		inv.SelectedPosition.OffsetX = 0.5
		inv.SelectedPosition.OffsetY = 0.5
	case 6: // North edge (top)
		inv.SelectedPosition.OffsetX = 0.5
		inv.SelectedPosition.OffsetY = 0.0
	case 7: // East edge (right)
		inv.SelectedPosition.OffsetX = 1.0
		inv.SelectedPosition.OffsetY = 0.5
	case 8: // South edge (bottom)
		inv.SelectedPosition.OffsetX = 0.5
		inv.SelectedPosition.OffsetY = 1.0
	case 9: // West edge (left)
		inv.SelectedPosition.OffsetX = 0.0
		inv.SelectedPosition.OffsetY = 0.5
	}
}

// SetRotation sets the rotation of the asset
func (inv *TileInventory) SetRotation(rotation float64) {
	inv.SelectedPosition.Rotation = rotation
}

// SetDepth sets the depth/vertical position of the asset
func (inv *TileInventory) SetDepth(depth float64) {
	inv.SelectedPosition.Depth = depth
}
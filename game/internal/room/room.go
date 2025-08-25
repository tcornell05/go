package room

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tcornell05/go/game/pkg/assets"
)

// TileData represents a placed tile in the room
type TileData struct {
	AssetPath string  `json:"asset_path"`
	X         int     `json:"x"`
	Y         int     `json:"y"`
	Facing    int     `json:"facing"`
	Category  string  `json:"category"`
	Layer     int     `json:"layer"` // 0=floor, 1=walls, 2=furniture, 3=decorations
	OffsetX   float64 `json:"offset_x"` // Position offset within the tile
	OffsetY   float64 `json:"offset_y"` // Position offset within the tile
	Depth     float64 `json:"depth"`    // Vertical depth offset
	Rotation  float64 `json:"rotation"`  // Rotation in radians
}

// Room represents a complete room layout
type Room struct {
	Name      string              `json:"name"`
	Width     int                 `json:"width"`
	Height    int                 `json:"height"`
	Tiles     map[string]TileData `json:"tiles"` // key: "layer:x:y"
	Walkability map[string]bool   `json:"walkability"` // key: "x:y", true=walkable, false/missing=blocked
	
	// Runtime data (not serialized)
	LoadedImages map[string]*ebiten.Image `json:"-"`
}

// NewRoom creates a new empty room
func NewRoom(name string, width, height int) *Room {
	room := &Room{
		Name:         name,
		Width:        width,
		Height:       height,
		Tiles:        make(map[string]TileData),
		Walkability:  make(map[string]bool),
		LoadedImages: make(map[string]*ebiten.Image),
	}
	
	// Initialize default walkability (assume floor area is walkable)
	for x := 0; x <= 10; x++ {
		for y := 0; y <= 10; y++ {
			key := fmt.Sprintf("%d:%d", x, y)
			room.Walkability[key] = true // Default to walkable
		}
	}
	
	return room
}

// GetTileKey generates a key for tile storage
func GetTileKey(layer, x, y int) string {
	return fmt.Sprintf("%d:%d:%d", layer, x, y)
}

// AddTile adds a tile to the room
func (r *Room) AddTile(tile TileData) error {
	// For wall tiles (layer 1), allow up to 2 tiles per position
	if tile.Layer == 1 {
		// Check if there's already a wall tile at this position
		key1 := fmt.Sprintf("1.0:%d:%d", tile.X, tile.Y)
		if _, exists := r.Tiles[key1]; !exists {
			// Use first wall slot
			r.Tiles[key1] = tile
		} else {
			// Check second wall slot
			key2 := fmt.Sprintf("1.1:%d:%d", tile.X, tile.Y)
			if _, exists := r.Tiles[key2]; !exists {
				// Use second wall slot
				r.Tiles[key2] = tile
			} else {
				// Both wall slots occupied, replace the first one
				r.Tiles[key1] = tile
			}
		}
	} else {
		// For non-wall tiles, use normal single-tile placement
		key := GetTileKey(tile.Layer, tile.X, tile.Y)
		r.Tiles[key] = tile
	}
	
	// Load the image if not already loaded
	if _, exists := r.LoadedImages[tile.AssetPath]; !exists {
		img, err := assets.LoadTile(tile.AssetPath)
		if err != nil {
			return err
		}
		r.LoadedImages[tile.AssetPath] = img
	}
	
	return nil
}

// PlaceTile places a tile in the room (alias for AddTile for consistency with game code)
func (r *Room) PlaceTile(tile TileData) error {
	return r.AddTile(tile)
}

// RemoveTile removes a tile from the room (for walls, removes the top-most wall)
func (r *Room) RemoveTile(layer, x, y int) {
	if layer == 1 {
		// For walls, remove the second wall first (top-most), then the first wall
		key2 := fmt.Sprintf("1.1:%d:%d", x, y)
		if _, exists := r.Tiles[key2]; exists {
			delete(r.Tiles, key2)
			return
		}
		key1 := fmt.Sprintf("1.0:%d:%d", x, y)
		delete(r.Tiles, key1)
		return
	}
	
	// For non-wall tiles, use normal deletion
	key := GetTileKey(layer, x, y)
	delete(r.Tiles, key)
}

// GetTile gets a tile at the specified position and layer (returns first tile for walls)
func (r *Room) GetTile(layer, x, y int) (TileData, bool) {
	if layer == 1 {
		// For walls, return the first wall tile if it exists
		key1 := fmt.Sprintf("1.0:%d:%d", x, y)
		if tile, exists := r.Tiles[key1]; exists {
			return tile, true
		}
	}
	key := GetTileKey(layer, x, y)
	tile, exists := r.Tiles[key]
	return tile, exists
}

// GetWallTilesAtPosition gets all wall tiles (up to 2) at a specific position
func (r *Room) GetWallTilesAtPosition(x, y int) []TileData {
	var tiles []TileData
	key1 := fmt.Sprintf("1.0:%d:%d", x, y)
	key2 := fmt.Sprintf("1.1:%d:%d", x, y)
	
	if tile, exists := r.Tiles[key1]; exists {
		tiles = append(tiles, tile)
	}
	if tile, exists := r.Tiles[key2]; exists {
		tiles = append(tiles, tile)
	}
	return tiles
}

// GetTilesAtPosition gets all tiles at a specific position across all layers
func (r *Room) GetTilesAtPosition(x, y int) []TileData {
	var tiles []TileData
	
	// Handle non-wall layers normally
	for layer := 0; layer <= 3; layer++ {
		if layer == 1 {
			// For walls, get both possible wall tiles
			wallTiles := r.GetWallTilesAtPosition(x, y)
			tiles = append(tiles, wallTiles...)
		} else {
			if tile, exists := r.GetTile(layer, x, y); exists {
				tiles = append(tiles, tile)
			}
		}
	}
	return tiles
}

// SaveToFile saves the room to a JSON file
func (r *Room) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, data, 0644)
}

// LoadFromFile loads a room from a JSON file
func LoadFromFile(filename string) (*Room, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	var room Room
	err = json.Unmarshal(data, &room)
	if err != nil {
		return nil, err
	}
	
	// Initialize runtime data
	room.LoadedImages = make(map[string]*ebiten.Image)
	
	// Load all tile images
	for _, tile := range room.Tiles {
		if _, exists := room.LoadedImages[tile.AssetPath]; !exists {
			img, err := assets.LoadTile(tile.AssetPath)
			if err != nil {
				return nil, err
			}
			room.LoadedImages[tile.AssetPath] = img
		}
	}
	
	// Initialize walkability if not present
	if room.Walkability == nil {
		room.Walkability = make(map[string]bool)
		// Initialize default walkability for backward compatibility
		for x := 0; x <= 10; x++ {
			for y := 0; y <= 10; y++ {
				key := fmt.Sprintf("%d:%d", x, y)
				room.Walkability[key] = true // Default to walkable
			}
		}
	}
	
	return &room, nil
}

// GetWalkabilityKey generates a key for walkability storage
func GetWalkabilityKey(x, y int) string {
	return fmt.Sprintf("%d:%d", x, y)
}

// IsWalkable returns true if the tile at (x,y) is walkable
func (r *Room) IsWalkable(x, y int) bool {
	key := GetWalkabilityKey(x, y)
	walkable, exists := r.Walkability[key]
	if !exists {
		return false // Unknown tiles are blocked by default
	}
	return walkable
}

// SetWalkable sets the walkability of a tile at (x,y)
func (r *Room) SetWalkable(x, y int, walkable bool) {
	key := GetWalkabilityKey(x, y)
	r.Walkability[key] = walkable
}
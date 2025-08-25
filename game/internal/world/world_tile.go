package world

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tcornell05/go/game/pkg/assets"
)

// TileData represents a placed tile in the world tile
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

// WorldTile represents a single tile in the world grid
type WorldTile struct {
	ID          string               `json:"id"`           // e.g., "A1", "B2"
	Name        string               `json:"name"`         // Display name
	Type        TileType             `json:"type"`         // Type of tile
	Description string               `json:"description"`  // Tile description
	Owner       *string              `json:"owner"`        // Owner username (nil for public)
	AccessLevel AccessLevel          `json:"access_level"` // Who can access
	Rentable    bool                 `json:"rentable"`     // Can be rented
	RentalStatus RentalStatus        `json:"rental_status"`// Current rental status
	RentalPrice  int                 `json:"rental_price"` // Daily/weekly price
	
	// Position in world grid
	WorldX      int                  `json:"world_x"`      // X position in world
	WorldY      int                  `json:"world_y"`      // Y position in world
	Floor       int                  `json:"floor"`        // Floor level (0 = ground)
	
	// Connections to other tiles
	Connections map[Direction]string `json:"connections"`  // Direction -> TileID
	
	// Tile layout data (500x500 world tile with placed objects)
	Width       int                 `json:"width"`        // Tile dimensions (500)
	Height      int                 `json:"height"`       // Tile dimensions (500)
	Tiles       map[string]TileData `json:"tiles"`        // key: "layer:x:y", placed objects
	Walkability map[string]bool     `json:"walkability"`  // key: "x:y", walkable areas
	
	// Room data (if this tile contains a room)
	RoomID      *int                 `json:"room_id"`      // Reference to room in DB (deprecated)
	MaxCapacity int                  `json:"max_capacity"` // Max players allowed
	
	// Runtime data (not serialized)
	LoadedImages map[string]*ebiten.Image `json:"-"`
	
	// Metadata
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	LastVisit   *time.Time           `json:"last_visit"`   // Last player visit
}

// NewWorldTile creates a new world tile with default values (500x500 with 11x11 walkable area)
func NewWorldTile(id string, x, y, floor int) *WorldTile {
	tile := &WorldTile{
		ID:           id,
		Name:         fmt.Sprintf("Tile %s", id),
		Type:         TileTypeEmpty,
		AccessLevel:  AccessPublic,
		Rentable:     false,
		RentalStatus: RentalNotRentable,
		WorldX:       x,
		WorldY:       y,
		Floor:        floor,
		Width:        500,
		Height:       500,
		Connections:  make(map[Direction]string),
		Tiles:        make(map[string]TileData),
		Walkability:  make(map[string]bool),
		LoadedImages: make(map[string]*ebiten.Image),
		MaxCapacity:  25,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	// Initialize default 11x11 walkable area in the center
	for x := 0; x <= 10; x++ {
		for y := 0; y <= 10; y++ {
			key := fmt.Sprintf("%d:%d", x, y)
			tile.Walkability[key] = true // Default 11x11 grid is walkable
		}
	}
	
	return tile
}

// IsAccessibleBy checks if a player can access this tile
func (t *WorldTile) IsAccessibleBy(playerName string, isStaff bool) bool {
	switch t.AccessLevel {
	case AccessPublic:
		return true
	case AccessPrivate:
		return t.Owner != nil && *t.Owner == playerName
	case AccessStaff:
		return isStaff
	case AccessRestricted:
		// Would check against access list or special conditions
		return t.Owner != nil && *t.Owner == playerName || isStaff
	default:
		return false
	}
}

// CanBeRentedBy checks if this tile can be rented by a player
func (t *WorldTile) CanBeRentedBy(playerName string) bool {
	if !t.Rentable {
		return false
	}
	
	if t.RentalStatus != RentalAvailable {
		return false
	}
	
	// Additional checks could go here (e.g., player balance, level requirements)
	return true
}

// SetOwner sets the owner of this tile
func (t *WorldTile) SetOwner(playerName string) {
	t.Owner = &playerName
	t.RentalStatus = RentalOccupied
	t.AccessLevel = AccessPrivate
	t.UpdatedAt = time.Now()
}

// ReleaseOwnership releases the ownership of this tile
func (t *WorldTile) ReleaseOwnership() {
	t.Owner = nil
	if t.Rentable {
		t.RentalStatus = RentalAvailable
		t.AccessLevel = AccessPublic
	}
	t.UpdatedAt = time.Now()
}

// AddConnection adds a connection to another tile
func (t *WorldTile) AddConnection(direction Direction, tileID string) {
	if t.Connections == nil {
		t.Connections = make(map[Direction]string)
	}
	t.Connections[direction] = tileID
	t.UpdatedAt = time.Now()
}

// RemoveConnection removes a connection in the specified direction
func (t *WorldTile) RemoveConnection(direction Direction) {
	delete(t.Connections, direction)
	t.UpdatedAt = time.Now()
}

// GetConnectedTileID returns the ID of the tile connected in the given direction
func (t *WorldTile) GetConnectedTileID(direction Direction) (string, bool) {
	tileID, exists := t.Connections[direction]
	return tileID, exists
}

// UpdateLastVisit updates the last visit timestamp
func (t *WorldTile) UpdateLastVisit() {
	now := time.Now()
	t.LastVisit = &now
}

// GetDisplayInfo returns formatted information about the tile
func (t *WorldTile) GetDisplayInfo() string {
	ownerStr := "Public"
	if t.Owner != nil {
		ownerStr = *t.Owner
	}
	
	return fmt.Sprintf("[%s] %s\nType: %s\nOwner: %s\nAccess: %s\nRentable: %v",
		t.ID, t.Name, t.Type, ownerStr, t.AccessLevel, t.Rentable)
}

// Room-like methods for WorldTile

// GetTileKey generates a key for tile storage
func GetTileKey(layer, x, y int) string {
	return fmt.Sprintf("%d:%d:%d", layer, x, y)
}

// AddTile adds a tile to the world tile
func (t *WorldTile) AddTile(tile TileData) error {
	// For wall tiles (layer 1), allow up to 2 tiles per position
	if tile.Layer == 1 {
		// Check if there's already a wall tile at this position
		key1 := fmt.Sprintf("1.0:%d:%d", tile.X, tile.Y)
		if _, exists := t.Tiles[key1]; !exists {
			// Use first wall slot
			t.Tiles[key1] = tile
		} else {
			// Check second wall slot
			key2 := fmt.Sprintf("1.1:%d:%d", tile.X, tile.Y)
			if _, exists := t.Tiles[key2]; !exists {
				// Use second wall slot
				t.Tiles[key2] = tile
			} else {
				// Both wall slots occupied, replace the first one
				t.Tiles[key1] = tile
			}
		}
	} else {
		// For non-wall tiles, use normal single-tile placement
		key := GetTileKey(tile.Layer, tile.X, tile.Y)
		t.Tiles[key] = tile
	}
	
	// Load the image if not already loaded
	if _, exists := t.LoadedImages[tile.AssetPath]; !exists {
		img, err := assets.LoadTile(tile.AssetPath)
		if err != nil {
			return err
		}
		t.LoadedImages[tile.AssetPath] = img
	}
	
	t.UpdatedAt = time.Now()
	return nil
}

// PlaceTile places a tile in the world tile (alias for AddTile for consistency)
func (t *WorldTile) PlaceTile(tile TileData) error {
	return t.AddTile(tile)
}

// RemoveTile removes a tile from the world tile (for walls, removes the top-most wall)
func (t *WorldTile) RemoveTile(layer, x, y int) {
	if layer == 1 {
		// For walls, remove the second wall first (top-most), then the first wall
		key2 := fmt.Sprintf("1.1:%d:%d", x, y)
		if _, exists := t.Tiles[key2]; exists {
			delete(t.Tiles, key2)
			t.UpdatedAt = time.Now()
			return
		}
		key1 := fmt.Sprintf("1.0:%d:%d", x, y)
		delete(t.Tiles, key1)
		t.UpdatedAt = time.Now()
		return
	}
	
	// For non-wall tiles, use normal deletion
	key := GetTileKey(layer, x, y)
	delete(t.Tiles, key)
	t.UpdatedAt = time.Now()
}

// GetTile gets a tile at the specified position and layer (returns first tile for walls)
func (t *WorldTile) GetTile(layer, x, y int) (TileData, bool) {
	if layer == 1 {
		// For walls, return the first wall tile if it exists
		key1 := fmt.Sprintf("1.0:%d:%d", x, y)
		if tile, exists := t.Tiles[key1]; exists {
			return tile, true
		}
	}
	key := GetTileKey(layer, x, y)
	tile, exists := t.Tiles[key]
	return tile, exists
}

// GetWallTilesAtPosition gets all wall tiles (up to 2) at a specific position
func (t *WorldTile) GetWallTilesAtPosition(x, y int) []TileData {
	var tiles []TileData
	key1 := fmt.Sprintf("1.0:%d:%d", x, y)
	key2 := fmt.Sprintf("1.1:%d:%d", x, y)
	
	if tile, exists := t.Tiles[key1]; exists {
		tiles = append(tiles, tile)
	}
	if tile, exists := t.Tiles[key2]; exists {
		tiles = append(tiles, tile)
	}
	return tiles
}

// GetTilesAtPosition gets all tiles at a specific position across all layers
func (t *WorldTile) GetTilesAtPosition(x, y int) []TileData {
	var tiles []TileData
	
	// Handle non-wall layers normally
	for layer := 0; layer <= 3; layer++ {
		if layer == 1 {
			// For walls, get both possible wall tiles
			wallTiles := t.GetWallTilesAtPosition(x, y)
			tiles = append(tiles, wallTiles...)
		} else {
			if tile, exists := t.GetTile(layer, x, y); exists {
				tiles = append(tiles, tile)
			}
		}
	}
	return tiles
}

// GetWalkabilityKey generates a key for walkability storage
func GetWalkabilityKey(x, y int) string {
	return fmt.Sprintf("%d:%d", x, y)
}

// IsWalkable returns true if the tile at (x,y) is walkable
func (t *WorldTile) IsWalkable(x, y int) bool {
	key := GetWalkabilityKey(x, y)
	walkable, exists := t.Walkability[key]
	if !exists {
		return false // Unknown tiles are blocked by default
	}
	return walkable
}

// SetWalkable sets the walkability of a tile at (x,y)
func (t *WorldTile) SetWalkable(x, y int, walkable bool) {
	key := GetWalkabilityKey(x, y)
	t.Walkability[key] = walkable
	t.UpdatedAt = time.Now()
}

// SaveToFile saves the world tile to a JSON file
func (t *WorldTile) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, data, 0644)
}

// LoadFromFile loads a world tile from a JSON file
func LoadWorldTileFromFile(filename string) (*WorldTile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	var tile WorldTile
	err = json.Unmarshal(data, &tile)
	if err != nil {
		return nil, err
	}
	
	// Initialize runtime data
	tile.LoadedImages = make(map[string]*ebiten.Image)
	
	// Load all tile images
	for _, tileData := range tile.Tiles {
		if _, exists := tile.LoadedImages[tileData.AssetPath]; !exists {
			img, err := assets.LoadTile(tileData.AssetPath)
			if err != nil {
				return nil, err
			}
			tile.LoadedImages[tileData.AssetPath] = img
		}
	}
	
	// Initialize walkability if not present (for backward compatibility)
	if tile.Walkability == nil {
		tile.Walkability = make(map[string]bool)
		// Initialize default walkability for 11x11 grid
		for x := 0; x <= 10; x++ {
			for y := 0; y <= 10; y++ {
				key := fmt.Sprintf("%d:%d", x, y)
				tile.Walkability[key] = true // Default to walkable
			}
		}
	}
	
	return &tile, nil
}
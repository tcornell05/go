package world

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorldMap manages the entire world grid of tiles
type WorldMap struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Width       int                    `json:"width"`  // Grid width
	Height      int                    `json:"height"` // Grid height
	Floors      int                    `json:"floors"` // Number of floors
	Tiles       map[string]*WorldTile  `json:"tiles"`  // TileID -> Tile
	Grid        [][][]string           `json:"grid"`   // [floor][y][x] = TileID
}

// NewWorldMap creates a new world map
func NewWorldMap(name string, width, height, floors int) *WorldMap {
	wm := &WorldMap{
		Name:        name,
		Description: "A new world",
		Width:       width,
		Height:      height,
		Floors:      floors,
		Tiles:       make(map[string]*WorldTile),
		Grid:        make([][][]string, floors),
	}
	
	// Initialize grid
	for f := 0; f < floors; f++ {
		wm.Grid[f] = make([][]string, height)
		for y := 0; y < height; y++ {
			wm.Grid[f][y] = make([]string, width)
		}
	}
	
	return wm
}

// GenerateTileID generates a tile ID based on position
func GenerateTileID(x, y, floor int) string {
	// Convert x to letter (A, B, C, ...)
	col := string(rune('A' + x))
	// Row is y + 1 (1-indexed)
	row := y + 1
	
	if floor > 0 {
		return fmt.Sprintf("%s%d-F%d", col, row, floor+1)
	}
	return fmt.Sprintf("%s%d", col, row)
}

// AddTile adds a tile to the world map
func (wm *WorldMap) AddTile(tile *WorldTile) error {
	if tile.WorldX < 0 || tile.WorldX >= wm.Width ||
	   tile.WorldY < 0 || tile.WorldY >= wm.Height ||
	   tile.Floor < 0 || tile.Floor >= wm.Floors {
		return fmt.Errorf("tile position out of bounds: (%d, %d, floor:%d)", 
			tile.WorldX, tile.WorldY, tile.Floor)
	}
	
	// Store in map
	wm.Tiles[tile.ID] = tile
	
	// Store in grid
	wm.Grid[tile.Floor][tile.WorldY][tile.WorldX] = tile.ID
	
	// Auto-connect to adjacent tiles
	wm.autoConnectTile(tile)
	
	return nil
}

// autoConnectTile automatically creates connections to adjacent tiles
func (wm *WorldMap) autoConnectTile(tile *WorldTile) {
	// Check all cardinal directions
	directions := []struct {
		dir Direction
		dx, dy int
	}{
		{DirectionNorth, 0, -1},
		{DirectionEast, 1, 0},
		{DirectionSouth, 0, 1},
		{DirectionWest, -1, 0},
	}
	
	for _, d := range directions {
		adjX := tile.WorldX + d.dx
		adjY := tile.WorldY + d.dy
		
		// Check bounds
		if adjX >= 0 && adjX < wm.Width && adjY >= 0 && adjY < wm.Height {
			adjID := wm.Grid[tile.Floor][adjY][adjX]
			if adjID != "" {
				adjTile := wm.Tiles[adjID]
				if adjTile != nil && wm.canConnect(tile, adjTile) {
					// Create bidirectional connection
					tile.AddConnection(d.dir, adjID)
					adjTile.AddConnection(oppositeDirection(d.dir), tile.ID)
				}
			}
		}
	}
}

// canConnect determines if two tiles can be connected
func (wm *WorldMap) canConnect(tile1, tile2 *WorldTile) bool {
	// Empty tiles cannot be connected
	if tile1.Type == TileTypeEmpty || tile2.Type == TileTypeEmpty {
		return false
	}
	
	// Rooms can connect to hallways
	if (tile1.Type == TileTypeRoom && tile2.Type == TileTypeHallway) ||
	   (tile1.Type == TileTypeHallway && tile2.Type == TileTypeRoom) {
		return true
	}
	
	// Hallways can connect to other hallways, lobbies, etc.
	publicTypes := []TileType{TileTypeHallway, TileTypeLobby, TileTypePublic, TileTypeElevator, TileTypeStairs}
	
	tile1IsPublic := false
	tile2IsPublic := false
	
	for _, t := range publicTypes {
		if tile1.Type == t {
			tile1IsPublic = true
		}
		if tile2.Type == t {
			tile2IsPublic = true
		}
	}
	
	return tile1IsPublic && tile2IsPublic
}

// oppositeDirection returns the opposite direction
func oppositeDirection(dir Direction) Direction {
	switch dir {
	case DirectionNorth:
		return DirectionSouth
	case DirectionSouth:
		return DirectionNorth
	case DirectionEast:
		return DirectionWest
	case DirectionWest:
		return DirectionEast
	case DirectionUp:
		return DirectionDown
	case DirectionDown:
		return DirectionUp
	default:
		return dir
	}
}

// GetTile retrieves a tile by ID
func (wm *WorldMap) GetTile(tileID string) (*WorldTile, bool) {
	tile, exists := wm.Tiles[tileID]
	return tile, exists
}

// GetTileAt retrieves a tile at specific coordinates
func (wm *WorldMap) GetTileAt(x, y, floor int) (*WorldTile, bool) {
	if x < 0 || x >= wm.Width || y < 0 || y >= wm.Height || floor < 0 || floor >= wm.Floors {
		return nil, false
	}
	
	tileID := wm.Grid[floor][y][x]
	if tileID == "" {
		return nil, false
	}
	
	return wm.GetTile(tileID)
}

// GetConnectedTiles returns all tiles connected to the given tile
func (wm *WorldMap) GetConnectedTiles(tileID string) []*WorldTile {
	tile, exists := wm.Tiles[tileID]
	if !exists {
		return nil
	}
	
	var connected []*WorldTile
	for _, connectedID := range tile.Connections {
		if connectedTile, exists := wm.Tiles[connectedID]; exists {
			connected = append(connected, connectedTile)
		}
	}
	
	return connected
}

// SaveToFile saves the world map to a JSON file
func (wm *WorldMap) SaveToFile(filename string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	data, err := json.MarshalIndent(wm, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal world map: %w", err)
	}
	
	return os.WriteFile(filename, data, 0644)
}

// LoadFromFile loads a world map from a JSON file
func LoadFromFile(filename string) (*WorldMap, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	var wm WorldMap
	if err := json.Unmarshal(data, &wm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal world map: %w", err)
	}
	
	return &wm, nil
}

// GetMapDisplay returns a text representation of a floor
func (wm *WorldMap) GetMapDisplay(floor int) string {
	if floor < 0 || floor >= wm.Floors {
		return "Invalid floor"
	}
	
	var sb strings.Builder
	
	// Header
	sb.WriteString("   ")
	for x := 0; x < wm.Width; x++ {
		sb.WriteString(fmt.Sprintf(" %c  ", 'A'+x))
	}
	sb.WriteString("\n")
	
	// Grid
	for y := 0; y < wm.Height; y++ {
		sb.WriteString(fmt.Sprintf("%2d ", y+1))
		
		for x := 0; x < wm.Width; x++ {
			tileID := wm.Grid[floor][y][x]
			if tileID == "" {
				sb.WriteString("[  ]")
			} else {
				tile := wm.Tiles[tileID]
				icon := wm.getTileIcon(tile)
				sb.WriteString(fmt.Sprintf("[%s]", icon))
			}
		}
		sb.WriteString("\n")
	}
	
	// Legend
	sb.WriteString("\nLegend: [RM]=Room [HW]=Hallway [LB]=Lobby [PS]=Public [  ]=Empty\n")
	
	return sb.String()
}

// getTileIcon returns a 2-character icon for the tile type
func (wm *WorldMap) getTileIcon(tile *WorldTile) string {
	if tile == nil {
		return "  "
	}
	
	switch tile.Type {
	case TileTypeRoom:
		if tile.Owner != nil {
			return "RM"
		}
		return "rm"
	case TileTypeHallway:
		return "HW"
	case TileTypeLobby:
		return "LB"
	case TileTypeShop:
		return "SH"
	case TileTypePublic:
		return "PS"
	case TileTypeElevator:
		return "EL"
	case TileTypeStairs:
		return "ST"
	case TileTypeOutdoor:
		return "OD"
	default:
		return "  "
	}
}

// InitializeDefaultWorld creates a simple default world layout
func InitializeDefaultWorld() *WorldMap {
	wm := NewWorldMap("Default Hotel", 5, 5, 1)
	wm.Description = "A small hotel with rooms and public spaces"
	
	// Create tiles for ground floor
	// Row 1: Rooms
	for x := 0; x < 5; x++ {
		tile := NewWorldTile(GenerateTileID(x, 0, 0), x, 0, 0)
		tile.Type = TileTypeRoom
		tile.Name = fmt.Sprintf("Room %s", tile.ID)
		tile.Rentable = true
		tile.RentalStatus = RentalAvailable
		tile.RentalPrice = 100
		wm.AddTile(tile)
	}
	
	// Row 2: Hallway
	for x := 0; x < 5; x++ {
		tile := NewWorldTile(GenerateTileID(x, 1, 0), x, 1, 0)
		tile.Type = TileTypeHallway
		tile.Name = "Main Hallway"
		tile.AccessLevel = AccessPublic
		wm.AddTile(tile)
	}
	
	// Row 3: Mix of rooms and public
	tiles := []TileType{TileTypeRoom, TileTypePublic, TileTypeLobby, TileTypePublic, TileTypeRoom}
	for x := 0; x < 5; x++ {
		tile := NewWorldTile(GenerateTileID(x, 2, 0), x, 2, 0)
		tile.Type = tiles[x]
		
		switch tile.Type {
		case TileTypeRoom:
			tile.Name = fmt.Sprintf("Room %s", tile.ID)
			tile.Rentable = true
			tile.RentalStatus = RentalAvailable
			tile.RentalPrice = 150
		case TileTypeLobby:
			tile.Name = "Main Lobby"
			tile.MaxCapacity = 50
		case TileTypePublic:
			tile.Name = "Lounge"
		}
		
		wm.AddTile(tile)
	}
	
	// Row 4: Hallway
	for x := 0; x < 5; x++ {
		tile := NewWorldTile(GenerateTileID(x, 3, 0), x, 3, 0)
		tile.Type = TileTypeHallway
		tile.Name = "South Hallway"
		tile.AccessLevel = AccessPublic
		wm.AddTile(tile)
	}
	
	// Row 5: Shops and services
	types := []TileType{TileTypeShop, TileTypeShop, TileTypePublic, TileTypeShop, TileTypeElevator}
	names := []string{"Gift Shop", "Cafe", "Reception", "Boutique", "Elevator"}
	for x := 0; x < 5; x++ {
		tile := NewWorldTile(GenerateTileID(x, 4, 0), x, 4, 0)
		tile.Type = types[x]
		tile.Name = names[x]
		tile.AccessLevel = AccessPublic
		wm.AddTile(tile)
	}
	
	return wm
}
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tcornell05/go/game/internal/room"
	"github.com/tcornell05/go/game/internal/world"
	"github.com/tcornell05/go/game/pkg/assets"
)

type Database struct {
	db *sql.DB
}

// NewDatabase creates or opens the SQLite database
func NewDatabase(dbPath string) (*Database, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{db: db}
	
	// Initialize tables
	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return database, nil
}

// createTables creates the necessary tables for rooms, entities, and room_entities
func (d *Database) createTables() error {
	query := `
	-- Rooms table
	CREATE TABLE IF NOT EXISTS rooms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		width INTEGER NOT NULL,
		height INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Entities table (asset definitions with properties)
	CREATE TABLE IF NOT EXISTS entities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		asset_path TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		category TEXT NOT NULL,
		layer INTEGER NOT NULL,
		width REAL DEFAULT 1.0,
		height REAL DEFAULT 1.0,
		can_walk_on BOOLEAN DEFAULT 0,
		can_sit_on BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Entity rotations table (stores offset data per rotation)
	CREATE TABLE IF NOT EXISTS entity_rotations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		entity_id INTEGER NOT NULL,
		rotation INTEGER NOT NULL, -- 0, 90, 180, 270
		offset_x REAL DEFAULT 0.5,
		offset_y REAL DEFAULT 0.5,
		depth REAL DEFAULT 0.0,
		snap_to_edge INTEGER DEFAULT 9,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE,
		UNIQUE(entity_id, rotation)
	);

	-- Room entities table (instances of entities placed in rooms)
	CREATE TABLE IF NOT EXISTS room_entities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room_id INTEGER NOT NULL,
		entity_id INTEGER NOT NULL,
		grid_x INTEGER NOT NULL,
		grid_y INTEGER NOT NULL,
		facing INTEGER DEFAULT 0,
		placed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
		FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE
	);

	-- World tiles table (new)
	CREATE TABLE IF NOT EXISTS world_tiles (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		description TEXT DEFAULT '',
		owner TEXT NULL,
		access_level TEXT DEFAULT 'public',
		rentable BOOLEAN DEFAULT 0,
		rental_status TEXT DEFAULT 'not_rentable',
		rental_price INTEGER DEFAULT 0,
		world_x INTEGER NOT NULL,
		world_y INTEGER NOT NULL,
		floor INTEGER DEFAULT 0,
		room_id INTEGER NULL,
		max_capacity INTEGER DEFAULT 25,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_visit DATETIME NULL,
		FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE SET NULL
	);

	-- World tile connections table
	CREATE TABLE IF NOT EXISTS world_tile_connections (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_tile_id TEXT NOT NULL,
		to_tile_id TEXT NOT NULL,
		direction TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (from_tile_id) REFERENCES world_tiles(id) ON DELETE CASCADE,
		FOREIGN KEY (to_tile_id) REFERENCES world_tiles(id) ON DELETE CASCADE,
		UNIQUE(from_tile_id, direction)
	);

	-- World tile data table (stores actual placed tiles/furniture)
	CREATE TABLE IF NOT EXISTS world_tile_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		world_tile_id TEXT NOT NULL,
		asset_path TEXT NOT NULL,
		x INTEGER NOT NULL,
		y INTEGER NOT NULL,
		facing INTEGER DEFAULT 0,
		category TEXT NOT NULL,
		layer INTEGER NOT NULL,
		offset_x REAL DEFAULT 0.0,
		offset_y REAL DEFAULT 0.0,
		depth REAL DEFAULT 0.0,
		rotation REAL DEFAULT 0.0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (world_tile_id) REFERENCES world_tiles(id) ON DELETE CASCADE
	);

	-- World tile walkability table (stores walkable/blocked areas)
	CREATE TABLE IF NOT EXISTS world_tile_walkability (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		world_tile_id TEXT NOT NULL,
		x INTEGER NOT NULL,
		y INTEGER NOT NULL,
		walkable BOOLEAN NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (world_tile_id) REFERENCES world_tiles(id) ON DELETE CASCADE,
		UNIQUE(world_tile_id, x, y)
	);

	-- World maps table  
	CREATE TABLE IF NOT EXISTS world_maps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		description TEXT DEFAULT '',
		width INTEGER NOT NULL,
		height INTEGER NOT NULL,
		floors INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_room_entities_room_position ON room_entities(room_id, grid_x, grid_y);
	CREATE INDEX IF NOT EXISTS idx_room_entities_entity ON room_entities(entity_id);
	CREATE INDEX IF NOT EXISTS idx_entities_asset_path ON entities(asset_path);
	CREATE INDEX IF NOT EXISTS idx_world_tiles_position ON world_tiles(world_x, world_y, floor);
	CREATE INDEX IF NOT EXISTS idx_world_tiles_owner ON world_tiles(owner);
	CREATE INDEX IF NOT EXISTS idx_world_tile_connections_from ON world_tile_connections(from_tile_id);
	CREATE INDEX IF NOT EXISTS idx_entity_rotations_entity ON entity_rotations(entity_id, rotation);
	CREATE INDEX IF NOT EXISTS idx_world_tile_data_tile ON world_tile_data(world_tile_id);
	CREATE INDEX IF NOT EXISTS idx_world_tile_data_position ON world_tile_data(world_tile_id, x, y, layer);
	CREATE INDEX IF NOT EXISTS idx_world_tile_walkability_tile ON world_tile_walkability(world_tile_id);
	CREATE INDEX IF NOT EXISTS idx_world_tile_walkability_position ON world_tile_walkability(world_tile_id, x, y);
	`

	_, err := d.db.Exec(query)
	return err
}

// SaveObject - DEPRECATED: Use SaveWorldTileObject instead
func (d *Database) SaveObject(roomName string, obj *room.TileData) error {
	return fmt.Errorf("SaveObject deprecated - room system has been migrated to world tiles, use SaveWorldTileObject instead")
}

// SaveWorldTileObject saves or updates an object in a world tile
func (d *Database) SaveWorldTileObject(worldTileID string, obj *room.TileData) error {
	// Get or create entity in the entities table (this ensures the asset exists in entities table)
	_, err := d.getOrCreateEntity(obj)
	if err != nil {
		return fmt.Errorf("failed to get/create entity: %w", err)
	}
	
	// Check if object already exists at this position in the world tile
	query := `SELECT id FROM world_tile_data WHERE world_tile_id = ? AND x = ? AND y = ? AND layer = ?`
	var existingID int64
	err = d.db.QueryRow(query, worldTileID, obj.X, obj.Y, obj.Layer).Scan(&existingID)
	
	if err == nil {
		// Update existing world tile object
		updateQuery := `UPDATE world_tile_data SET 
			asset_path = ?, facing = ?, category = ?, 
			offset_x = ?, offset_y = ?, depth = ?, rotation = ?
			WHERE id = ?`
		_, err = d.db.Exec(updateQuery, obj.AssetPath, obj.Facing, obj.Category,
			obj.OffsetX, obj.OffsetY, obj.Depth, obj.Rotation, existingID)
		fmt.Printf("WORLD TILE UPDATE: Updated object at (%d, %d, layer %d) in world tile %s\n", obj.X, obj.Y, obj.Layer, worldTileID)
		return err
	} else if err == sql.ErrNoRows {
		// Insert new world tile object
		insertQuery := `INSERT INTO world_tile_data 
			(world_tile_id, asset_path, x, y, facing, category, layer, offset_x, offset_y, depth, rotation)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err = d.db.Exec(insertQuery, worldTileID, obj.AssetPath, obj.X, obj.Y, 
			obj.Facing, obj.Category, obj.Layer, obj.OffsetX, obj.OffsetY, obj.Depth, obj.Rotation)
		// Object inserted into world tile
		return err
	} else {
		return fmt.Errorf("failed to check existing world tile object: %w", err)
	}
}

// LoadObjects - DEPRECATED: Use LoadWorldTileObjects instead
// Room system has been migrated to world tiles
func (d *Database) LoadObjects(roomName string) ([]room.TileData, error) {
	return nil, fmt.Errorf("LoadObjects deprecated - room system has been migrated to world tiles, use LoadWorldTileObjects instead")
}

// DeleteObject - DEPRECATED: Use DeleteWorldTileObject instead  
func (d *Database) DeleteObject(roomName string, x, y, layer int) error {
	return fmt.Errorf("DeleteObject deprecated - room system has been migrated to world tiles, use DeleteWorldTileObject instead")
}

// DeleteWorldTileObject removes an object at a specific position from world tile data
func (d *Database) DeleteWorldTileObject(worldTileID string, x, y, layer int) error {
	query := `DELETE FROM world_tile_data WHERE world_tile_id = ? AND x = ? AND y = ? AND layer = ?`
	result, err := d.db.Exec(query, worldTileID, x, y, layer)
	if err != nil {
		return fmt.Errorf("failed to delete world tile object: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("WORLD TILE DELETE: Removed %d object(s) at (%d, %d, layer %d) from world tile %s\n", rowsAffected, x, y, layer, worldTileID)
	return nil
}

// GetObjectAt retrieves an object at a specific position using new schema
// GetObjectAt - DEPRECATED: Use GetWorldTileObjectAt instead
// Room system has been migrated to world tiles
func (d *Database) GetObjectAt(roomName string, x, y int) (*room.TileData, error) {
	return nil, fmt.Errorf("GetObjectAt deprecated - room system has been migrated to world tiles, use GetWorldTileObjectAt instead")
}

// GetWorldTileObjectAt retrieves an object at a specific position from world tile data
func (d *Database) GetWorldTileObjectAt(worldTileID string, x, y int) (*room.TileData, error) {
	query := `
	SELECT asset_path, x, y, facing, category, layer, offset_x, offset_y, depth, rotation
	FROM world_tile_data 
	WHERE world_tile_id = ? AND x = ? AND y = ?
	ORDER BY layer DESC
	LIMIT 1`
	
	row := d.db.QueryRow(query, worldTileID, x, y)
	
	var obj room.TileData
	var offsetX, offsetY, depth sql.NullFloat64
	err := row.Scan(&obj.AssetPath, &obj.X, &obj.Y, &obj.Facing,
		&obj.Category, &obj.Layer, &offsetX, &offsetY, &depth, &obj.Rotation)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No object found
		}
		return nil, fmt.Errorf("failed to scan world tile object: %w", err)
	}
	
	// Set offset values
	if offsetX.Valid {
		obj.OffsetX = offsetX.Float64
	}
	if offsetY.Valid {
		obj.OffsetY = offsetY.Float64
	}
	if depth.Valid {
		obj.Depth = depth.Float64
	}
	
	return &obj, nil
}


// MigrateRoomDataToWorldTiles migrates any existing room data to world tile format
func (d *Database) MigrateRoomDataToWorldTiles() error {
	fmt.Println("Starting migration from room system to world tile system...")
	
	// Check if room tables exist
	hasRoomTables := true
	if err := d.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='rooms'").Err(); err != nil {
		hasRoomTables = false
	}
	
	if !hasRoomTables {
		fmt.Println("No room tables found, migration not needed")
		return nil
	}
	
	// Get all rooms and their entities
	rows, err := d.db.Query(`
		SELECT r.name, r.width, r.height, 
		       e.asset_path, re.grid_x, re.grid_y, re.facing,
		       e.category, e.layer
		FROM rooms r
		LEFT JOIN room_entities re ON r.id = re.room_id  
		LEFT JOIN entities e ON re.entity_id = e.id
		ORDER BY r.name, e.layer, re.grid_y, re.grid_x
	`)
	if err != nil {
		return fmt.Errorf("failed to query room data: %w", err)
	}
	defer rows.Close()
	
	roomsToMigrate := make(map[string][]world.TileData)
	
	for rows.Next() {
		var roomName string
		var width, height int
		var assetPath sql.NullString
		var gridX, gridY, facing sql.NullInt64
		var category sql.NullString
		var layer sql.NullInt64
		
		err := rows.Scan(&roomName, &width, &height, &assetPath, &gridX, &gridY, &facing, &category, &layer)
		if err != nil {
			return fmt.Errorf("failed to scan room data: %w", err)
		}
		
		// If this room has objects, add them to migration data
		if assetPath.Valid {
			tileData := world.TileData{
				AssetPath: assetPath.String,
				X:         int(gridX.Int64),
				Y:         int(gridY.Int64),
				Facing:    int(facing.Int64),
				Category:  category.String,
				Layer:     int(layer.Int64),
			}
			roomsToMigrate[roomName] = append(roomsToMigrate[roomName], tileData)
		}
	}
	
	// Create world tiles for each room
	for roomName, objects := range roomsToMigrate {
		fmt.Printf("Migrating room '%s' with %d objects to world tile\n", roomName, len(objects))
		
		// Convert slice to map format expected by WorldTile
		tilesMap := make(map[string]world.TileData)
		for _, obj := range objects {
			key := fmt.Sprintf("%d:%d:%d", obj.Layer, obj.X, obj.Y)
			tilesMap[key] = obj
		}
		
		// Create world tile
		worldTile := &world.WorldTile{
			ID:       roomName,
			Name:     fmt.Sprintf("Migrated Room %s", roomName),
			Type:     "room",
			Width:    500, // Default world tile size
			Height:   500,
			Tiles:    tilesMap,
		}
		
		// Save as world tile
		if err := d.SaveWorldTile(worldTile); err != nil {
			return fmt.Errorf("failed to save migrated world tile %s: %w", roomName, err)
		}
	}
	
	fmt.Printf("Successfully migrated %d rooms to world tiles\n", len(roomsToMigrate))
	return nil
}

// DropRoomTables removes the deprecated room system tables
func (d *Database) DropRoomTables() error {
	fmt.Println("Dropping deprecated room system tables...")
	
	queries := []string{
		"DROP INDEX IF EXISTS idx_room_entities_room_position",
		"DROP INDEX IF EXISTS idx_room_entities_entity", 
		"DROP TABLE IF EXISTS room_entities",
		"DROP TABLE IF EXISTS rooms",
	}
	
	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute %s: %w", query, err)
		}
	}
	
	fmt.Println("Successfully dropped room system tables")
	return nil
}

// Helper methods for new schema

// getOrCreateRoom gets an existing room or creates a new one
func (d *Database) getOrCreateRoom(name string, width, height int) (int, error) {
	// Try to get existing room
	var id int
	err := d.db.QueryRow("SELECT id FROM rooms WHERE name = ?", name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	// Create new room
	result, err := d.db.Exec("INSERT INTO rooms (name, width, height) VALUES (?, ?, ?)", name, width, height)
	if err != nil {
		return 0, err
	}
	
	roomID, err := result.LastInsertId()
	return int(roomID), err
}

// getOrCreateEntity gets an existing entity or creates a new one
func (d *Database) getOrCreateEntity(obj *room.TileData) (int, error) {
	// Try to get existing entity
	var id int
	err := d.db.QueryRow("SELECT id FROM entities WHERE asset_path = ?", obj.AssetPath).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	// Create new entity with basic properties
	query := `INSERT INTO entities (asset_path, name, category, layer) VALUES (?, ?, ?, ?)`
	
	name := getAssetName(obj.AssetPath)
	result, err := d.db.Exec(query, obj.AssetPath, name, obj.Category, obj.Layer)
	if err != nil {
		return 0, err
	}
	
	entityID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	
	// Create default rotation data (rotation 0)
	rotQuery := `INSERT INTO entity_rotations (entity_id, rotation, offset_x, offset_y, depth, snap_to_edge)
				 VALUES (?, 0, ?, ?, ?, ?)`
	
	_, err = d.db.Exec(rotQuery, int(entityID), obj.OffsetX, obj.OffsetY, obj.Depth, getSnapToEdgeFromPosition(obj))
	if err != nil {
		return 0, err
	}
	
	return int(entityID), nil
}

// getRoomID gets the ID of a room by name
func (d *Database) getRoomID(name string) (int, error) {
	var id int
	err := d.db.QueryRow("SELECT id FROM rooms WHERE name = ?", name).Scan(&id)
	return id, err
}

// getRoomEntityIDAt gets the database ID of a room entity at a position
func (d *Database) getRoomEntityIDAt(roomID, x, y int) (int, error) {
	var id int
	err := d.db.QueryRow("SELECT id FROM room_entities WHERE room_id = ? AND grid_x = ? AND grid_y = ?", roomID, x, y).Scan(&id)
	return id, err
}

// getAssetName extracts a display name from an asset path
func getAssetName(assetPath string) string {
	// Extract filename without extension
	parts := strings.Split(assetPath, "/")
	filename := parts[len(parts)-1]
	name := strings.Split(filename, ".")[0]
	return strings.ReplaceAll(name, "_", " ")
}

// getSnapToEdgeFromPosition determines snap-to-edge value from position offsets
func getSnapToEdgeFromPosition(obj *room.TileData) int {
	// This is a simplified mapping - in a real implementation you might
	// want to store this directly or calculate it more precisely
	if obj.OffsetX == 0.5 && obj.OffsetY == 0.5 {
		return 9 // Center
	}
	// Add more mappings as needed
	return 0 // Manual/unknown
}

// GetEntityByAssetPath gets entity properties by asset path with rotation data
func (d *Database) GetEntityByAssetPath(assetPath string) (*EntityProperties, error) {
	// First get the basic entity info
	query := `SELECT id, name, category, layer, width, height, can_walk_on, can_sit_on 
			  FROM entities WHERE asset_path = ?`
	
	var entity EntityProperties
	err := d.db.QueryRow(query, assetPath).Scan(
		&entity.ID, &entity.Name, &entity.Category, &entity.Layer,
		&entity.Width, &entity.Height, &entity.CanWalkOn, &entity.CanSitOn)
		
	if err != nil {
		return nil, err
	}
	
	entity.AssetPath = assetPath
	entity.CurrentRotation = 0 // Default to rotation 0
	
	// Load rotation data for rotation 0 (default)
	err = d.loadRotationData(&entity, 0)
	if err != nil {
		// If no rotation data exists, create default
		err = d.createDefaultRotationData(&entity, 0)
		if err != nil {
			return nil, err
		}
		// Try loading again
		err = d.loadRotationData(&entity, 0)
		if err != nil {
			return nil, err
		}
	}
	
	return &entity, nil
}

// UpdateEntityProperties updates the basic properties of an entity
func (d *Database) UpdateEntityProperties(entity *EntityProperties) error {
	// Update basic entity properties
	query := `UPDATE entities SET 
		name = ?, category = ?, layer = ?, width = ?, height = ?, 
		can_walk_on = ?, can_sit_on = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`
		
	_, err := d.db.Exec(query, entity.Name, entity.Category, entity.Layer,
		entity.Width, entity.Height, entity.CanWalkOn, entity.CanSitOn, entity.ID)
		
	if err != nil {
		return err
	}
	
	// Update rotation-specific data
	err = d.UpdateRotationData(entity)
	if err != nil {
		return err
	}
	
	// Update all instances of this asset in world tiles with the new default offsets
	return d.UpdateWorldTileInstanceOffsets(entity.AssetPath, entity.DefaultOffsetX, entity.DefaultOffsetY, entity.DefaultDepth)
}

// UpdateWorldTileInstanceOffsets updates all instances of an asset in world tiles with new offset values
func (d *Database) UpdateWorldTileInstanceOffsets(assetPath string, offsetX, offsetY, depth float64) error {
	query := `UPDATE world_tile_data SET offset_x = ?, offset_y = ?, depth = ? WHERE asset_path = ?`
	result, err := d.db.Exec(query, offsetX, offsetY, depth, assetPath)
	if err != nil {
		return fmt.Errorf("failed to update world tile instance offsets: %v", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Updated %d instances of %s with new offsets (%.2f, %.2f) and depth %.2f\n", rowsAffected, assetPath, offsetX, offsetY, depth)
	return nil
}

// EntityProperties represents entity asset properties
type EntityProperties struct {
	ID                 int
	AssetPath          string
	Name               string
	Category           string
	Layer              int
	Width              float64
	Height             float64
	CanWalkOn          bool
	CanSitOn           bool
	
	// Current rotation data (loaded dynamically)
	CurrentRotation    int     // 0, 90, 180, 270
	DefaultOffsetX     float64 // Offset for current rotation
	DefaultOffsetY     float64 // Offset for current rotation
	DefaultDepth       float64 // Depth for current rotation
	DefaultSnapToEdge  int     // Snap setting for current rotation
}

// loadRotationData loads offset data for a specific rotation
func (d *Database) loadRotationData(entity *EntityProperties, rotation int) error {
	query := `SELECT offset_x, offset_y, depth, snap_to_edge 
			  FROM entity_rotations WHERE entity_id = ? AND rotation = ?`
	
	return d.db.QueryRow(query, entity.ID, rotation).Scan(
		&entity.DefaultOffsetX, &entity.DefaultOffsetY, 
		&entity.DefaultDepth, &entity.DefaultSnapToEdge)
}

// createDefaultRotationData creates default rotation data for an entity
func (d *Database) createDefaultRotationData(entity *EntityProperties, rotation int) error {
	query := `INSERT INTO entity_rotations (entity_id, rotation, offset_x, offset_y, depth, snap_to_edge)
			  VALUES (?, ?, 0.5, 0.5, 0.0, 9)`
	
	_, err := d.db.Exec(query, entity.ID, rotation)
	return err
}

// UpdateRotationData updates the offset data for a specific rotation
func (d *Database) UpdateRotationData(entity *EntityProperties) error {
	query := `UPDATE entity_rotations SET 
			  offset_x = ?, offset_y = ?, depth = ?, snap_to_edge = ?, updated_at = CURRENT_TIMESTAMP
			  WHERE entity_id = ? AND rotation = ?`
			  
	_, err := d.db.Exec(query, entity.DefaultOffsetX, entity.DefaultOffsetY, 
		entity.DefaultDepth, entity.DefaultSnapToEdge, entity.ID, entity.CurrentRotation)
		
	return err
}

// LoadRotation loads offset data for a different rotation
func (d *Database) LoadRotation(entity *EntityProperties, rotation int) error {
	// First ensure rotation data exists
	err := d.loadRotationData(entity, rotation)
	if err != nil {
		// Create default if doesn't exist
		err = d.createDefaultRotationData(entity, rotation)
		if err != nil {
			return err
		}
		// Load the default values
		err = d.loadRotationData(entity, rotation)
		if err != nil {
			return err
		}
	}
	
	entity.CurrentRotation = rotation
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// World Tile Database Methods

// SaveWorldMap saves a complete world map to the database
func (d *Database) SaveWorldMap(worldMap *world.WorldMap) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert or update world map
	_, err = tx.Exec(`INSERT OR REPLACE INTO world_maps (name, description, width, height, floors, updated_at) 
					  VALUES (?, ?, ?, ?, ?, ?)`,
		worldMap.Name, worldMap.Description, worldMap.Width, worldMap.Height, worldMap.Floors, time.Now())
	if err != nil {
		return err
	}

	// Clear existing tiles for this world (assuming we're replacing the whole world)
	_, err = tx.Exec(`DELETE FROM world_tiles WHERE id IN (
		SELECT wt.id FROM world_tiles wt 
		WHERE wt.world_x < ? AND wt.world_y < ? AND wt.floor < ?)`,
		worldMap.Width, worldMap.Height, worldMap.Floors)
	if err != nil {
		return err
	}

	// Insert all tiles
	for _, tile := range worldMap.Tiles {
		err = d.saveWorldTileInTx(tx, tile)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// saveWorldTileInTx saves a world tile within a transaction
func (d *Database) saveWorldTileInTx(tx *sql.Tx, tile *world.WorldTile) error {
	// Insert or update world tile
	_, err := tx.Exec(`INSERT OR REPLACE INTO world_tiles 
		(id, name, type, description, owner, access_level, rentable, rental_status, rental_price, 
		 world_x, world_y, floor, room_id, max_capacity, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tile.ID, tile.Name, tile.Type, tile.Description, tile.Owner, tile.AccessLevel,
		tile.Rentable, tile.RentalStatus, tile.RentalPrice, tile.WorldX, tile.WorldY,
		tile.Floor, tile.RoomID, tile.MaxCapacity, tile.UpdatedAt)
	if err != nil {
		return err
	}

	// Clear existing connections
	_, err = tx.Exec(`DELETE FROM world_tile_connections WHERE from_tile_id = ?`, tile.ID)
	if err != nil {
		return err
	}

	// Insert connections
	for direction, toTileID := range tile.Connections {
		_, err = tx.Exec(`INSERT INTO world_tile_connections (from_tile_id, to_tile_id, direction) 
						  VALUES (?, ?, ?)`, tile.ID, toTileID, direction)
		if err != nil {
			return err
		}
	}

	// Clear existing tile data
	_, err = tx.Exec(`DELETE FROM world_tile_data WHERE world_tile_id = ?`, tile.ID)
	if err != nil {
		return err
	}

	// Insert tile data (placed furniture, decorations, etc.)
	for _, tileData := range tile.Tiles {
		_, err = tx.Exec(`INSERT INTO world_tile_data 
			(world_tile_id, asset_path, x, y, facing, category, layer, offset_x, offset_y, depth, rotation)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			tile.ID, tileData.AssetPath, tileData.X, tileData.Y, tileData.Facing,
			tileData.Category, tileData.Layer, tileData.OffsetX, tileData.OffsetY,
			tileData.Depth, tileData.Rotation)
		if err != nil {
			return err
		}
	}

	// Clear existing walkability data
	_, err = tx.Exec(`DELETE FROM world_tile_walkability WHERE world_tile_id = ?`, tile.ID)
	if err != nil {
		return err
	}

	// Insert walkability data
	for walkKey, walkable := range tile.Walkability {
		// Parse the walkability key to get x, y coordinates
		var x, y int
		_, err = fmt.Sscanf(walkKey, "%d:%d", &x, &y)
		if err != nil {
			continue // Skip invalid keys
		}

		_, err = tx.Exec(`INSERT INTO world_tile_walkability 
			(world_tile_id, x, y, walkable) VALUES (?, ?, ?, ?)`,
			tile.ID, x, y, walkable)
		if err != nil {
			return err
		}
	}

	return nil
}

// LoadWorldMap loads a world map from the database
func (d *Database) LoadWorldMap(name string) (*world.WorldMap, error) {
	// Load world map metadata
	var worldMap world.WorldMap
	var createdAt, updatedAt time.Time
	err := d.db.QueryRow(`SELECT name, description, width, height, floors, created_at, updated_at 
						  FROM world_maps WHERE name = ?`, name).Scan(
		&worldMap.Name, &worldMap.Description, &worldMap.Width, &worldMap.Height,
		&worldMap.Floors, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	// Initialize structures
	worldMap.Tiles = make(map[string]*world.WorldTile)
	worldMap.Grid = make([][][]string, worldMap.Floors)
	for f := 0; f < worldMap.Floors; f++ {
		worldMap.Grid[f] = make([][]string, worldMap.Height)
		for y := 0; y < worldMap.Height; y++ {
			worldMap.Grid[f][y] = make([]string, worldMap.Width)
		}
	}

	// Load all tiles for this world
	tiles, err := d.loadWorldTiles(worldMap.Width, worldMap.Height, worldMap.Floors)
	if err != nil {
		return nil, err
	}

	// Add tiles to world map
	for _, tile := range tiles {
		worldMap.Tiles[tile.ID] = tile
		if tile.WorldX < worldMap.Width && tile.WorldY < worldMap.Height && tile.Floor < worldMap.Floors {
			worldMap.Grid[tile.Floor][tile.WorldY][tile.WorldX] = tile.ID
		}
	}

	return &worldMap, nil
}

// loadWorldTiles loads all world tiles within the specified bounds
func (d *Database) loadWorldTiles(maxWidth, maxHeight, maxFloors int) ([]*world.WorldTile, error) {
	query := `SELECT id, name, type, description, owner, access_level, rentable, rental_status,
			  rental_price, world_x, world_y, floor, room_id, max_capacity, created_at, updated_at, last_visit
			  FROM world_tiles 
			  WHERE world_x < ? AND world_y < ? AND floor < ?
			  ORDER BY floor, world_y, world_x`

	rows, err := d.db.Query(query, maxWidth, maxHeight, maxFloors)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiles []*world.WorldTile
	for rows.Next() {
		tile := &world.WorldTile{
			Connections: make(map[world.Direction]string),
		}

		var owner sql.NullString
		var lastVisit sql.NullTime
		var roomIDInt sql.NullInt64

		err := rows.Scan(&tile.ID, &tile.Name, &tile.Type, &tile.Description,
			&owner, &tile.AccessLevel, &tile.Rentable, &tile.RentalStatus,
			&tile.RentalPrice, &tile.WorldX, &tile.WorldY, &tile.Floor,
			&roomIDInt, &tile.MaxCapacity, &tile.CreatedAt, &tile.UpdatedAt, &lastVisit)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if owner.Valid {
			tile.Owner = &owner.String
		}
		if roomIDInt.Valid {
			roomIDValue := int(roomIDInt.Int64)
			tile.RoomID = &roomIDValue
		}
		if lastVisit.Valid {
			tile.LastVisit = &lastVisit.Time
		}

		// Load connections
		connections, err := d.loadTileConnections(tile.ID)
		if err != nil {
			return nil, err
		}
		tile.Connections = connections

		tiles = append(tiles, tile)
	}

	return tiles, rows.Err()
}

// loadTileConnections loads all connections for a tile
func (d *Database) loadTileConnections(tileID string) (map[world.Direction]string, error) {
	query := `SELECT direction, to_tile_id FROM world_tile_connections WHERE from_tile_id = ?`
	rows, err := d.db.Query(query, tileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	connections := make(map[world.Direction]string)
	for rows.Next() {
		var direction, toTileID string
		err := rows.Scan(&direction, &toTileID)
		if err != nil {
			return nil, err
		}
		connections[world.Direction(direction)] = toTileID
	}

	return connections, rows.Err()
}

// loadWorldTileData loads all placed tile data for a world tile
func (d *Database) loadWorldTileData(tile *world.WorldTile) error {
	query := `SELECT asset_path, x, y, facing, category, layer, offset_x, offset_y, depth, rotation
			  FROM world_tile_data WHERE world_tile_id = ? ORDER BY layer, y, x`
	
	rows, err := d.db.Query(query, tile.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tileData world.TileData
		err := rows.Scan(&tileData.AssetPath, &tileData.X, &tileData.Y, &tileData.Facing,
			&tileData.Category, &tileData.Layer, &tileData.OffsetX, &tileData.OffsetY,
			&tileData.Depth, &tileData.Rotation)
		if err != nil {
			return err
		}

		// Generate the appropriate tile key
		var key string
		if tileData.Layer == 1 {
			// For walls, we need to handle dual placement
			key1 := fmt.Sprintf("1.0:%d:%d", tileData.X, tileData.Y)
			if _, exists := tile.Tiles[key1]; !exists {
				key = key1
			} else {
				key = fmt.Sprintf("1.1:%d:%d", tileData.X, tileData.Y)
			}
		} else {
			key = world.GetTileKey(tileData.Layer, tileData.X, tileData.Y)
		}

		tile.Tiles[key] = tileData

		// Load the image if not already loaded
		if _, exists := tile.LoadedImages[tileData.AssetPath]; !exists {
			img, err := assets.LoadTile(tileData.AssetPath)
			if err == nil {
				tile.LoadedImages[tileData.AssetPath] = img
			}
		}
	}

	return rows.Err()
}

// loadWorldTileWalkability loads all walkability data for a world tile
func (d *Database) loadWorldTileWalkability(tile *world.WorldTile) error {
	query := `SELECT x, y, walkable FROM world_tile_walkability WHERE world_tile_id = ?`
	
	rows, err := d.db.Query(query, tile.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	hasWalkabilityData := false
	for rows.Next() {
		hasWalkabilityData = true
		var x, y int
		var walkable bool
		err := rows.Scan(&x, &y, &walkable)
		if err != nil {
			return err
		}

		key := world.GetWalkabilityKey(x, y)
		tile.Walkability[key] = walkable
	}
	
	// If no walkability data exists in database, initialize default 11x11 walkable area
	if !hasWalkabilityData {
		// Initializing default walkability for new tile
		for x := 0; x <= 10; x++ {
			for y := 0; y <= 10; y++ {
				key := world.GetWalkabilityKey(x, y)
				tile.Walkability[key] = true // Default 11x11 grid is walkable
			}
		}
	}

	return rows.Err()
}

// SaveWorldTile saves a single world tile
func (d *Database) SaveWorldTile(tile *world.WorldTile) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = d.saveWorldTileInTx(tx, tile)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetWorldTile retrieves a world tile by ID
func (d *Database) GetWorldTile(tileID string) (*world.WorldTile, error) {
	tile := &world.WorldTile{
		Connections: make(map[world.Direction]string),
		Tiles:       make(map[string]world.TileData),
		Walkability: make(map[string]bool),
		LoadedImages: make(map[string]*ebiten.Image),
		Width:       500,
		Height:      500,
	}

	var owner sql.NullString
	var lastVisit sql.NullTime
	var roomIDInt sql.NullInt64

	query := `SELECT id, name, type, description, owner, access_level, rentable, rental_status,
			  rental_price, world_x, world_y, floor, room_id, max_capacity, created_at, updated_at, last_visit
			  FROM world_tiles WHERE id = ?`

	err := d.db.QueryRow(query, tileID).Scan(&tile.ID, &tile.Name, &tile.Type, &tile.Description,
		&owner, &tile.AccessLevel, &tile.Rentable, &tile.RentalStatus,
		&tile.RentalPrice, &tile.WorldX, &tile.WorldY, &tile.Floor,
		&roomIDInt, &tile.MaxCapacity, &tile.CreatedAt, &tile.UpdatedAt, &lastVisit)

	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if owner.Valid {
		tile.Owner = &owner.String
	}
	if roomIDInt.Valid {
		roomIDValue := int(roomIDInt.Int64)
		tile.RoomID = &roomIDValue
	}
	if lastVisit.Valid {
		tile.LastVisit = &lastVisit.Time
	}

	// Load connections
	connections, err := d.loadTileConnections(tile.ID)
	if err != nil {
		return nil, err
	}
	tile.Connections = connections

	// Load tile data (placed furniture, decorations, etc.)
	err = d.loadWorldTileData(tile)
	if err != nil {
		return nil, err
	}

	// Load walkability data
	err = d.loadWorldTileWalkability(tile)
	if err != nil {
		return nil, err
	}

	return tile, nil
}

// UpdateTileOwnership updates the ownership of a world tile
func (d *Database) UpdateTileOwnership(tileID string, owner *string) error {
	query := `UPDATE world_tiles SET owner = ?, updated_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, owner, time.Now(), tileID)
	return err
}

// GetTilesByOwner gets all tiles owned by a specific player
func (d *Database) GetTilesByOwner(owner string) ([]*world.WorldTile, error) {
	query := `SELECT id FROM world_tiles WHERE owner = ?`
	rows, err := d.db.Query(query, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiles []*world.WorldTile
	for rows.Next() {
		var tileID string
		err := rows.Scan(&tileID)
		if err != nil {
			return nil, err
		}

		tile, err := d.GetWorldTile(tileID)
		if err != nil {
			return nil, err
		}
		tiles = append(tiles, tile)
	}

	return tiles, rows.Err()
}

// GetAvailableRentableTiles gets all tiles available for rent
func (d *Database) GetAvailableRentableTiles() ([]*world.WorldTile, error) {
	query := `SELECT id FROM world_tiles WHERE rentable = 1 AND rental_status = 'available'`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiles []*world.WorldTile
	for rows.Next() {
		var tileID string
		err := rows.Scan(&tileID)
		if err != nil {
			return nil, err
		}

		tile, err := d.GetWorldTile(tileID)
		if err != nil {
			return nil, err
		}
		tiles = append(tiles, tile)
	}

	return tiles, rows.Err()
}
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tcornell05/go/game/internal/database"
	"github.com/tcornell05/go/game/internal/world"
)

// TestDatabaseManager manages a persistent test database for integration test scenarios
type TestDatabaseManager struct {
	db   *database.Database
	path string
}

// NewTestDatabaseManager creates or connects to a persistent test database
func NewTestDatabaseManager() (*TestDatabaseManager, error) {
	// Use a persistent test database in the project
	testDBPath := filepath.Join("test_data", "integration_test.db")
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(testDBPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create test data directory: %w", err)
	}
	
	db, err := database.NewDatabase(testDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create test database: %w", err)
	}
	
	return &TestDatabaseManager{
		db:   db,
		path: testDBPath,
	}, nil
}

// SeedTestData populates the test database with comprehensive test scenarios
func (tm *TestDatabaseManager) SeedTestData() error {
	// Test Scenario 1: Basic Room (A1) - Simple furniture setup
	roomA1 := world.NewWorldTile("A1", 0, 0, 0)
	roomA1.Name = "Basic Test Room"
	roomA1.Type = world.TileTypeRoom
	roomA1.Description = "Simple room for basic functionality tests"
	
	// Add basic furniture
	chair := world.TileData{
		AssetPath: "assets/Chair/Chair_2_A_Tile.png",
		X:         3,
		Y:         3,
		Layer:     2,
		Category:  "furniture",
		OffsetX:   0.5,
		OffsetY:   0.5,
	}
	if err := roomA1.AddTile(chair); err != nil {
		return err
	}
	
	desk := world.TileData{
		AssetPath: "assets/Desk/Desk_1_Tile.png",
		X:         5,
		Y:         5,
		Layer:     2,
		Category:  "furniture",
		OffsetX:   0.5,
		OffsetY:   0.5,
	}
	if err := roomA1.AddTile(desk); err != nil {
		return err
	}
	
	if err := tm.db.SaveWorldTile(roomA1); err != nil {
		return fmt.Errorf("failed to save A1: %w", err)
	}
	
	// Test Scenario 2: Complex Room (B2) - Multiple furniture types and custom walkability
	roomB2 := world.NewWorldTile("B2", 1, 1, 0)
	roomB2.Name = "Complex Test Room"
	roomB2.Type = world.TileTypeRoom
	roomB2.Description = "Complex room for advanced functionality tests"
	
	// Add multiple furniture pieces
	furnitureItems := []world.TileData{
		{AssetPath: "assets/Chair/Chair_2_A_Tile.png", X: 2, Y: 2, Layer: 2, Category: "furniture", OffsetX: 0.3, OffsetY: 0.4},
		{AssetPath: "assets/Chair/Chair_2_B_Tile.png", X: 2, Y: 3, Layer: 2, Category: "furniture", OffsetX: 0.7, OffsetY: 0.6},
		{AssetPath: "assets/Desk/Desk_1_Tile.png", X: 6, Y: 6, Layer: 2, Category: "furniture", OffsetX: 0.8, OffsetY: 0.2},
		{AssetPath: "assets/Plants/Plant_1_Tile.png", X: 8, Y: 4, Layer: 2, Category: "decorations", OffsetX: 0.1, OffsetY: 0.9},
		{AssetPath: "assets/Lamp/Lamp_8_A_Tile.png", X: 9, Y: 8, Layer: 3, Category: "decorations", OffsetX: 0.4, OffsetY: 0.7},
	}
	
	for _, item := range furnitureItems {
		if err := roomB2.AddTile(item); err != nil {
			return err
		}
	}
	
	// Custom walkability - block some areas, expand others
	roomB2.SetWalkable(1, 1, false) // Block northwest corner
	roomB2.SetWalkable(9, 9, false) // Block southeast corner
	roomB2.SetWalkable(12, 12, true) // Expand beyond default
	roomB2.SetWalkable(13, 13, true) // Expand further
	
	if err := tm.db.SaveWorldTile(roomB2); err != nil {
		return fmt.Errorf("failed to save B2: %w", err)
	}
	
	// Test Scenario 3: Public Room (C3) - Different access levels and rental properties
	roomC3 := world.NewWorldTile("C3", 2, 2, 0)
	roomC3.Name = "Public Test Room"
	roomC3.Type = world.TileTypeLobby
	roomC3.Description = "Public room for access control tests"
	roomC3.AccessLevel = world.AccessPublic
	roomC3.Rentable = true
	roomC3.RentalStatus = world.RentalAvailable
	roomC3.RentalPrice = 100
	roomC3.MaxCapacity = 50
	
	// Add some public seating
	publicChairs := []world.TileData{
		{AssetPath: "assets/Chair/Chair_2_A_Tile.png", X: 4, Y: 4, Layer: 2, Category: "furniture"},
		{AssetPath: "assets/Chair/Chair_2_A_Tile.png", X: 4, Y: 5, Layer: 2, Category: "furniture"},
		{AssetPath: "assets/Chair/Chair_2_A_Tile.png", X: 6, Y: 4, Layer: 2, Category: "furniture"},
		{AssetPath: "assets/Chair/Chair_2_A_Tile.png", X: 6, Y: 5, Layer: 2, Category: "furniture"},
	}
	
	for _, chair := range publicChairs {
		if err := roomC3.AddTile(chair); err != nil {
			return err
		}
	}
	
	if err := tm.db.SaveWorldTile(roomC3); err != nil {
		return fmt.Errorf("failed to save C3: %w", err)
	}
	
	// Test Scenario 4: Private Room (D4) - Owner-specific room for access testing
	roomD4 := world.NewWorldTile("D4", 3, 3, 0)
	roomD4.Name = "Private Test Room"
	roomD4.Type = world.TileTypeRoom
	roomD4.Description = "Private room for ownership tests"
	roomD4.AccessLevel = world.AccessPrivate
	ownerName := "test_user"
	roomD4.Owner = &ownerName
	roomD4.Rentable = false
	roomD4.RentalStatus = world.RentalNotRentable
	
	// Add private furniture
	privateFurniture := []world.TileData{
		{AssetPath: "assets/Desk/Desk_1_Tile.png", X: 5, Y: 5, Layer: 2, Category: "furniture"},
		{AssetPath: "assets/Lamp/Lamp_8_A_Tile.png", X: 7, Y: 7, Layer: 3, Category: "decorations"},
		{AssetPath: "assets/Plants/Plant_1_Tile.png", X: 3, Y: 7, Layer: 2, Category: "decorations"},
	}
	
	for _, item := range privateFurniture {
		if err := roomD4.AddTile(item); err != nil {
			return err
		}
	}
	
	if err := tm.db.SaveWorldTile(roomD4); err != nil {
		return fmt.Errorf("failed to save D4: %w", err)
	}
	
	// Test Scenario 5: Empty Room (E5) - For new content creation tests
	roomE5 := world.NewWorldTile("E5", 4, 4, 0)
	roomE5.Name = "Empty Test Room"
	roomE5.Type = world.TileTypeEmpty
	roomE5.Description = "Empty room for content creation tests"
	
	if err := tm.db.SaveWorldTile(roomE5); err != nil {
		return fmt.Errorf("failed to save E5: %w", err)
	}
	
	return nil
}

// GetTestRoom returns a specific test room for integration tests
func (tm *TestDatabaseManager) GetTestRoom(tileID string) (*world.WorldTile, error) {
	return tm.db.GetWorldTile(tileID)
}

// ListTestRooms returns all available test rooms
func (tm *TestDatabaseManager) ListTestRooms() ([]string, error) {
	// For now, return the known test room IDs
	// In a full implementation, we'd query the database
	return []string{"A1", "B2", "C3", "D4", "E5"}, nil
}

// ResetTestData clears and re-seeds the test database
func (tm *TestDatabaseManager) ResetTestData() error {
	// Close and delete the existing database
	if err := tm.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	
	if err := os.Remove(tm.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete test database: %w", err)
	}
	
	// Recreate the database
	db, err := database.NewDatabase(tm.path)
	if err != nil {
		return fmt.Errorf("failed to recreate test database: %w", err)
	}
	tm.db = db
	
	// Re-seed with test data
	return tm.SeedTestData()
}

// Close closes the test database connection
func (tm *TestDatabaseManager) Close() error {
	return tm.db.Close()
}

// GetDatabase returns the underlying database for direct access in tests
func (tm *TestDatabaseManager) GetDatabase() *database.Database {
	return tm.db
}
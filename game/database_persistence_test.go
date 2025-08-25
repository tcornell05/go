package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/tcornell05/go/game/internal/database"
	"github.com/tcornell05/go/game/internal/world"
	"github.com/tcornell05/go/game/pkg/assets"
)

//go:embed assets/*
var dbTestAssets embed.FS

func init() {
	assets.GlobalAssets = dbTestAssets
}

// TestDatabaseWorldTilePersistence tests that world tiles should persist in database, not JSON files
func TestDatabaseWorldTilePersistence(t *testing.T) {
	// Create test database
	testDir, err := os.MkdirTemp("", "db_persistence_test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	db, err := database.NewDatabase(filepath.Join(testDir, "persistence_test.db"))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test the exact user workflow: go run . editor --tile A2, add furniture, save, reload
	t.Log("=== PHASE 1: Create WorldTile A2 with furniture ===")
	
	// Create world tile
	worldTile := world.NewWorldTile("A2", 2, 0, 0)
	
	// Add furniture
	chairs := []world.TileData{
		{AssetPath: "assets/Chair/Chair_2_A_Tile.png", X: 2, Y: 3, Layer: 2, Category: "furniture", OffsetX: 0.3, OffsetY: 0.4},
		{AssetPath: "assets/Chair/Chair_2_B_Tile.png", X: 4, Y: 5, Layer: 2, Category: "furniture", OffsetX: 0.7, OffsetY: 0.6},
	}
	
	for _, chair := range chairs {
		err = worldTile.AddTile(chair)
		if err != nil {
			t.Fatalf("Failed to add chair: %v", err)
		}
	}
	
	// Modify walkability (F2 boundary editing)
	worldTile.SetWalkable(15, 15, true)  // Expand
	worldTile.SetWalkable(5, 5, false)   // Block
	
	t.Logf("Created A2 with %d tiles", len(worldTile.Tiles))
	
	// Save to database instead of JSON file
	err = db.SaveWorldTile(worldTile)
	if err != nil {
		t.Fatalf("Failed to save world tile to database: %v", err)
	}
	
	t.Log("✅ Saved A2 to database")
	
	// Clear memory (simulate app restart)
	worldTile = nil
	
	t.Log("=== PHASE 2: Reload WorldTile A2 from database ===")
	
	// Load from database
	reloadedTile, err := db.GetWorldTile("A2")
	if err != nil {
		t.Fatalf("Failed to load world tile from database: %v", err)
	}
	
	// Verify basic properties
	if reloadedTile.ID != "A2" {
		t.Errorf("Expected tile ID A2, got %s", reloadedTile.ID)
	}
	
	if reloadedTile.WorldX != 2 {
		t.Errorf("Expected WorldX 2, got %d", reloadedTile.WorldX)
	}
	
	// Verify walkability persisted
	if !reloadedTile.IsWalkable(15, 15) {
		t.Error("Expanded walkable area (15,15) should persist")
	}
	if reloadedTile.IsWalkable(5, 5) {
		t.Error("Blocked area (5,5) should persist as non-walkable")
	}
	
	// Verify furniture tiles persist in database
	if len(reloadedTile.Tiles) != 2 {
		t.Errorf("Expected 2 tiles to persist, got %d", len(reloadedTile.Tiles))
	}
	
	// Check specific chairs
	chair1Found := false
	chair2Found := false
	for _, tileData := range reloadedTile.Tiles {
		if tileData.AssetPath == "assets/Chair/Chair_2_A_Tile.png" && tileData.X == 2 && tileData.Y == 3 {
			chair1Found = true
			if tileData.OffsetX != 0.3 || tileData.OffsetY != 0.4 {
				t.Errorf("Chair 1 offset mismatch: expected (0.3, 0.4), got (%f, %f)", tileData.OffsetX, tileData.OffsetY)
			}
		}
		if tileData.AssetPath == "assets/Chair/Chair_2_B_Tile.png" && tileData.X == 4 && tileData.Y == 5 {
			chair2Found = true
			if tileData.OffsetX != 0.7 || tileData.OffsetY != 0.6 {
				t.Errorf("Chair 2 offset mismatch: expected (0.7, 0.6), got (%f, %f)", tileData.OffsetX, tileData.OffsetY)
			}
		}
	}
	
	if !chair1Found {
		t.Error("Chair 1 (Chair_2_A_Tile.png at 2,3) not found after database reload")
	}
	if !chair2Found {
		t.Error("Chair 2 (Chair_2_B_Tile.png at 4,5) not found after database reload")
	}
	
	// Verify walkability map persists in database
	totalWalkable := 0
	totalNonWalkable := 0
	for x := 0; x <= 15; x++ { // Check expanded range to include (15,15)
		for y := 0; y <= 15; y++ {
			if reloadedTile.IsWalkable(x, y) {
				totalWalkable++
			} else {
				totalNonWalkable++
			}
		}
	}
	
	// Should be: 
	// - 121 default walkable tiles (11x11 grid: 0-10, 0-10)
	// - 1 additional expanded tile (15,15) = +1 walkable
	// - 1 blocked tile (5,5) = -1 walkable  
	// Total: 121 walkable tiles in core grid + (15,15) - (5,5) = 121 walkable total
	// But we need to check if (15,15) is actually being stored and retrieved
	expectedWalkableCore := 120 // 11x11 - 1 blocked (5,5) = 121 - 1 = 120
	if totalWalkable < expectedWalkableCore {
		t.Errorf("Expected at least %d walkable tiles in core area, got %d", expectedWalkableCore, totalWalkable)
	}
	
	// Specifically test the expanded area
	if !reloadedTile.IsWalkable(15, 15) {
		t.Error("Expanded walkable area (15,15) should persist as walkable")
	} else {
		t.Log("✅ Expanded area (15,15) persisted correctly")
	}
	
	t.Log("✅ Furniture tiles and walkability map persist correctly in database")
}

// TestDatabaseWorldTileSchema tests the database schema for world tiles
func TestDatabaseWorldTileSchema(t *testing.T) {
	testDir, err := os.MkdirTemp("", "db_schema_test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	db, err := database.NewDatabase(filepath.Join(testDir, "schema_test.db"))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test that world_tiles table exists and has correct schema
	// This should work as the table is already created
	
	worldTile := world.NewWorldTile("SCHEMA_TEST", 0, 0, 0)
	err = db.SaveWorldTile(worldTile)
	if err != nil {
		t.Fatalf("Failed to save world tile to test schema: %v", err)
	}
	
	loaded, err := db.GetWorldTile("SCHEMA_TEST")
	if err != nil {
		t.Fatalf("Failed to load world tile to test schema: %v", err)
	}
	
	if loaded.ID != "SCHEMA_TEST" {
		t.Errorf("Schema test failed: expected SCHEMA_TEST, got %s", loaded.ID)
	}
	
	t.Log("✅ Database world_tiles table schema works")
}

// TestDatabaseMigrationFromJSON tests migrating from JSON files to database
func TestDatabaseMigrationFromJSON(t *testing.T) {
	testDir, err := os.MkdirTemp("", "db_migration_test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Step 1: Create JSON file (old system)
	worldTile := world.NewWorldTile("MIGRATE_TEST", 5, 3, 0)
	
	// Add furniture to JSON version
	furniture := world.TileData{
		AssetPath: "assets/Desk/Desk_1_Tile.png",
		X:         3,
		Y:         4,
		Layer:     2,
		Category:  "furniture",
		OffsetX:   0.5,
		OffsetY:   0.5,
	}
	
	err = worldTile.AddTile(furniture)
	if err != nil {
		t.Fatalf("Failed to add furniture to migration test: %v", err)
	}
	
	// Save to JSON file
	jsonFile := filepath.Join(testDir, "MIGRATE_TEST.json")
	err = worldTile.SaveToFile(jsonFile)
	if err != nil {
		t.Fatalf("Failed to save JSON file for migration test: %v", err)
	}
	
	// Step 2: Load from JSON and migrate to database
	db, err := database.NewDatabase(filepath.Join(testDir, "migration_test.db"))
	if err != nil {
		t.Fatalf("Failed to create migration database: %v", err)
	}
	defer db.Close()
	
	// Load from JSON
	jsonTile, err := world.LoadWorldTileFromFile(jsonFile)
	if err != nil {
		t.Fatalf("Failed to load JSON tile for migration: %v", err)
	}
	
	// Save to database
	err = db.SaveWorldTile(jsonTile)
	if err != nil {
		t.Fatalf("Failed to migrate JSON tile to database: %v", err)
	}
	
	// Step 3: Verify migration worked
	dbTile, err := db.GetWorldTile("MIGRATE_TEST")
	if err != nil {
		t.Fatalf("Failed to load migrated tile from database: %v", err)
	}
	
	if dbTile.ID != "MIGRATE_TEST" {
		t.Errorf("Migration failed: expected MIGRATE_TEST, got %s", dbTile.ID)
	}
	
	if dbTile.WorldX != 5 || dbTile.WorldY != 3 {
		t.Errorf("Migration failed: expected position (5,3), got (%d,%d)", dbTile.WorldX, dbTile.WorldY)
	}
	
	// TODO: Verify furniture migrated (requires database schema extension)
	
	t.Log("✅ JSON to database migration test PASSED")
}

// TestDatabasePerformanceVsJSON compares database vs JSON performance
func TestDatabasePerformanceVsJSON(t *testing.T) {
	testDir, err := os.MkdirTemp("", "db_perf_test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	db, err := database.NewDatabase(filepath.Join(testDir, "perf_test.db"))
	if err != nil {
		t.Fatalf("Failed to create perf database: %v", err)
	}
	defer db.Close()

	// Create test world tile
	worldTile := world.NewWorldTile("PERF_TEST", 0, 0, 0)
	
	// Add multiple furniture items
	for i := 0; i < 10; i++ {
		furniture := world.TileData{
			AssetPath: "assets/Chair/Chair_2_A_Tile.png",
			X:         i,
			Y:         i,
			Layer:     2,
			Category:  "furniture",
		}
		worldTile.AddTile(furniture)
	}
	
	// Test JSON save/load performance
	
	jsonFile := filepath.Join(testDir, "PERF_TEST.json")
	
	start := time.Now()
	err = worldTile.SaveToFile(jsonFile)
	if err != nil {
		t.Fatalf("Failed to save JSON for perf test: %v", err)
	}
	jsonSaveTime := time.Since(start)
	
	start = time.Now()
	_, err = world.LoadWorldTileFromFile(jsonFile)
	if err != nil {
		t.Fatalf("Failed to load JSON for perf test: %v", err)
	}
	jsonLoadTime := time.Since(start)
	
	// Test database save/load performance
	start = time.Now()
	err = db.SaveWorldTile(worldTile)
	if err != nil {
		t.Fatalf("Failed to save to database for perf test: %v", err)
	}
	dbSaveTime := time.Since(start)
	
	start = time.Now()
	_, err = db.GetWorldTile("PERF_TEST")
	if err != nil {
		t.Fatalf("Failed to load from database for perf test: %v", err)
	}
	dbLoadTime := time.Since(start)
	
	t.Logf("Performance comparison:")
	t.Logf("  JSON Save: %v, Load: %v (Total: %v)", jsonSaveTime, jsonLoadTime, jsonSaveTime+jsonLoadTime)
	t.Logf("  DB Save: %v, Load: %v (Total: %v)", dbSaveTime, dbLoadTime, dbSaveTime+dbLoadTime)
	
	// Both should be under 100ms for non-blocking UI
	totalJSON := jsonSaveTime + jsonLoadTime
	totalDB := dbSaveTime + dbLoadTime
	
	if totalJSON > 100*time.Millisecond {
		t.Errorf("JSON persistence too slow: %v", totalJSON)
	}
	
	if totalDB > 100*time.Millisecond {
		t.Errorf("Database persistence too slow: %v", totalDB)
	}
	
	t.Log("✅ Both JSON and database persistence are fast enough")
}

// TestConcurrentDatabaseAccess tests multiple world tiles accessing database concurrently
func TestConcurrentDatabaseAccess(t *testing.T) {
	testDir, err := os.MkdirTemp("", "db_concurrent_test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	db, err := database.NewDatabase(filepath.Join(testDir, "concurrent_test.db"))
	if err != nil {
		t.Fatalf("Failed to create concurrent database: %v", err)
	}
	defer db.Close()

	// Test concurrent saves and loads (simulating multiple editor instances)
	
	var wg sync.WaitGroup
	numTiles := 5
	
	// Concurrent saves
	for i := 0; i < numTiles; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			tileID := fmt.Sprintf("CONCURRENT_%d", index)
			tile := world.NewWorldTile(tileID, index, 0, 0)
			
			err := db.SaveWorldTile(tile)
			if err != nil {
				t.Errorf("Failed to save tile %s concurrently: %v", tileID, err)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Concurrent loads
	for i := 0; i < numTiles; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			tileID := fmt.Sprintf("CONCURRENT_%d", index)
			_, err := db.GetWorldTile(tileID)
			if err != nil {
				t.Errorf("Failed to load tile %s concurrently: %v", tileID, err)
			}
		}(i)
	}
	
	wg.Wait()
	
	t.Log("✅ Concurrent database access test PASSED")
}
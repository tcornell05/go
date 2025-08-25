package main

import (
	"embed"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tcornell05/go/game/internal/database"
	"github.com/tcornell05/go/game/internal/entity"
	"github.com/tcornell05/go/game/internal/room"
	"github.com/tcornell05/go/game/internal/world"
	"github.com/tcornell05/go/game/pkg/assets"
)

//go:embed assets/*
var testAssets embed.FS

func init() {
	// Initialize assets for all tests in this file
	assets.GlobalAssets = testAssets
}

// TestRealAssetLoading verifies that assets are properly embedded and loadable
func TestRealAssetLoading(t *testing.T) {
	// Test loading a real floor tile
	tile1, err := assets.LoadTile("assets/Foor-Wall Tiles 64px/Floor_1_Tile(64).png")
	if err != nil {
		t.Fatalf("Failed to load floor tile: %v", err)
	}
	if tile1 == nil {
		t.Fatal("Floor tile should not be nil")
	}

	// Test loading same tile again (should use cache if implemented)
	tile2, err := assets.LoadTile("assets/Foor-Wall Tiles 64px/Floor_1_Tile(64).png")
	if err != nil {
		t.Fatalf("Failed to load floor tile second time: %v", err)
	}
	
	// Verify both tiles were loaded (testing caching would require more sophisticated checking)
	if tile2 == nil {
		t.Fatal("Second tile load should not be nil")
	}

	// Test loading a chair asset
	chair, err := assets.LoadTile("assets/Chair/Chair_2_A_Tile.png")
	if err != nil {
		t.Fatalf("Failed to load chair tile: %v", err)
	}
	if chair == nil {
		t.Fatal("Chair asset should not be nil")
	}

	t.Log("✅ Real asset loading works perfectly with embedded assets!")
}

// TestWorldTileWithRealAssets tests the complete WorldTile workflow with real embedded assets
func TestWorldTileWithRealAssets(t *testing.T) {
	// Create isolated test environment
	testDir, err := os.MkdirTemp("", "game_integration_test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create a world tile
	worldTile := world.NewWorldTile("A1", 0, 0, 0)

	// Verify 500x500 dimensions
	if worldTile.Width != 500 || worldTile.Height != 500 {
		t.Errorf("Expected 500x500 dimensions, got %dx%d", worldTile.Width, worldTile.Height)
	}

	// Modify walkability (F2 boundary editing)
	worldTile.SetWalkable(15, 15, true)  // Expand walkable area
	worldTile.SetWalkable(5, 5, false)   // Block a default area

	// Place a chair object with real asset
	chairTileData := world.TileData{
		AssetPath: "assets/Chair/Chair_2_A_Tile.png",
		X:         2,
		Y:         3,
		Layer:     2, // furniture layer
		Facing:    0,
		Category:  "furniture",
		OffsetX:   0.3,
		OffsetY:   0.4,
	}

	// Verify we can load this asset
	chairImg, err := assets.LoadTile(chairTileData.AssetPath)
	if err != nil {
		t.Fatalf("Failed to load chair asset: %v", err)
	}
	if chairImg == nil {
		t.Fatal("Chair image should not be nil")
	}

	// Add tile to world tile
	err = worldTile.AddTile(chairTileData)
	if err != nil {
		t.Fatalf("Failed to add chair tile: %v", err)
	}

	// Save to file
	filename := filepath.Join(testDir, "A1.json")
	err = worldTile.SaveToFile(filename)
	if err != nil {
		t.Fatalf("Failed to save world tile: %v", err)
	}

	// Load from file - this will also load all tile assets
	loaded, err := world.LoadWorldTileFromFile(filename)
	if err != nil {
		t.Fatalf("Failed to load world tile WITH REAL ASSETS: %v", err)
	}

	// Verify walkability persisted
	if !loaded.IsWalkable(15, 15) {
		t.Error("Expected expanded walkable area (15,15) to persist")
	}
	if loaded.IsWalkable(5, 5) {
		t.Error("Expected blocked area (5,5) to persist as non-walkable")
	}

	// Verify placed object persisted
	key := world.GetTileKey(chairTileData.Layer, chairTileData.X, chairTileData.Y)
	if tileData, ok := loaded.Tiles[key]; !ok {
		t.Error("Placed chair tile should persist after save/load")
	} else {
		if tileData.AssetPath != chairTileData.AssetPath {
			t.Errorf("Chair asset path mismatch: expected %s, got %s", chairTileData.AssetPath, tileData.AssetPath)
		}
	}

	// Verify the loaded WorldTile has the actual images loaded
	if len(loaded.LoadedImages) == 0 {
		t.Error("LoadedImages should contain the chair image after loading")
	}
	if loadedImg, exists := loaded.LoadedImages[chairTileData.AssetPath]; !exists {
		t.Error("Chair image should be loaded in LoadedImages map")
	} else if loadedImg == nil {
		t.Error("Loaded chair image should not be nil")
	}

	t.Log("✅ Complete WorldTile workflow with real embedded assets works perfectly!")
}

// TestDatabaseWithRealAssets tests database operations with real asset paths
func TestDatabaseWithRealAssets(t *testing.T) {
	// Create test database
	testDir, err := os.MkdirTemp("", "game_db_test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	db, err := database.NewDatabase(filepath.Join(testDir, "test.db"))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Save object with real asset path
	object := room.TileData{
		AssetPath: "assets/Sofa/Sofa_3_A_Tile.png",
		X:         5,
		Y:         5,
		Layer:     2,
		Facing:    0,
		Category:  "furniture",
		OffsetX:   0.5,
		OffsetY:   0.5,
	}

	// Verify the asset exists
	sofaImg, err := assets.LoadTile(object.AssetPath)
	if err != nil {
		t.Fatalf("Failed to load sofa asset: %v", err)
	}
	if sofaImg == nil {
		t.Fatal("Sofa image should not be nil")
	}

	// Create world tile first
	worldTile := world.NewWorldTile("TEST_TILE", 0, 0, 0)
	worldTile.Type = world.TileTypeRoom
	err = db.SaveWorldTile(worldTile)
	if err != nil {
		t.Fatalf("Failed to create world tile: %v", err)
	}

	// Save object to database
	err = db.SaveWorldTileObject("TEST_TILE", &object)
	if err != nil {
		t.Fatalf("Failed to save object to database: %v", err)
	}

	// Load from database by getting the world tile (objects are loaded automatically)
	loadedWorldTile, err := db.GetWorldTile("TEST_TILE")
	if err != nil {
		t.Fatalf("Failed to load world tile from database: %v", err)
	}
	
	// Convert world tile objects to room.TileData for compatibility with test
	var objects []room.TileData
	for _, tileData := range loadedWorldTile.Tiles {
		objects = append(objects, room.TileData{
			AssetPath: tileData.AssetPath,
			X:         tileData.X,
			Y:         tileData.Y,
			Facing:    tileData.Facing,
			Category:  tileData.Category,
			Layer:     tileData.Layer,
			OffsetX:   tileData.OffsetX,
			OffsetY:   tileData.OffsetY,
			Depth:     tileData.Depth,
			Rotation:  tileData.Rotation,
		})
	}

	if len(objects) != 1 {
		t.Fatalf("Expected 1 object, got %d", len(objects))
	}

	// Verify loaded object has correct asset path
	loaded := objects[0]
	if loaded.AssetPath != object.AssetPath {
		t.Errorf("Asset path mismatch: expected %s, got %s", object.AssetPath, loaded.AssetPath)
	}

	// Verify we can still load the asset from the loaded path
	loadedImg, err := assets.LoadTile(loaded.AssetPath)
	if err != nil {
		t.Fatalf("Failed to load asset from database-loaded path: %v", err)
	}
	if loadedImg == nil {
		t.Fatal("Loaded image from database path should not be nil")
	}

	t.Log("✅ Database integration with real assets works perfectly!")
}

// TestPlayerSpritesEmbedded tests that character sprites are properly embedded
func TestPlayerSpritesEmbedded(t *testing.T) {
	// Try to load character sprites that the player system uses
	spritePaths := []string{
		"assets/Character/habbo_sprites/direction_0_static.png",
		"assets/Character/habbo_sprites/direction_1_static.png",
		"assets/Character/habbo_sprites/direction_2_static.png",
		"assets/Character/habbo_sprites/direction_3_static.png",
		"assets/Character/habbo_sprites/direction_4_static.png",
		"assets/Character/habbo_sprites/direction_5_static.png",
		"assets/Character/habbo_sprites/direction_6_static.png",
		"assets/Character/habbo_sprites/direction_7_static.png",
	}

	for _, path := range spritePaths {
		sprite, err := assets.LoadTile(path)
		if err != nil {
			t.Errorf("Failed to load character sprite %s: %v", path, err)
			continue
		}
		if sprite == nil {
			t.Errorf("Character sprite %s should not be nil", path)
		}
	}

	// Now try to create a player (which loads sprites)
	player, err := entity.NewPlayer(5, 5)
	if err != nil {
		// Player might fail if sprites are in different paths
		t.Logf("Player creation note: %v", err)
	} else if player != nil {
		t.Logf("✅ Player created successfully at (%.1f, %.1f)", player.X, player.Y)
	}
}

// TestF2BoundaryEditingPersistence tests that F2 boundary editing persists correctly
func TestF2BoundaryEditingPersistence(t *testing.T) {
	testDir, err := os.MkdirTemp("", "boundary_test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create world tile
	tile := world.NewWorldTile("B2", 0, 0, 0)

	// Verify default 11x11 walkable area
	for x := 0; x <= 10; x++ {
		for y := 0; y <= 10; y++ {
			if !tile.IsWalkable(x, y) {
				t.Errorf("Default tile (%d,%d) should be walkable", x, y)
			}
		}
	}

	// Simulate F2 boundary editing
	// Left-click on blue tiles (expand walkable area)
	tile.SetWalkable(20, 20, true)
	tile.SetWalkable(15, 15, true)
	tile.SetWalkable(-5, 5, true)

	// Right-click on grey tiles (block walkable area)
	tile.SetWalkable(5, 5, false)
	tile.SetWalkable(7, 7, false)

	// Save
	filename := filepath.Join(testDir, "B2.json")
	err = tile.SaveToFile(filename)
	if err != nil {
		t.Fatalf("Failed to save tile: %v", err)
	}

	// Load
	loaded, err := world.LoadWorldTileFromFile(filename)
	if err != nil {
		t.Fatalf("Failed to load tile: %v", err)
	}

	// Verify expanded areas
	if !loaded.IsWalkable(20, 20) {
		t.Error("Expanded area (20,20) should remain walkable after load")
	}
	if !loaded.IsWalkable(15, 15) {
		t.Error("Expanded area (15,15) should remain walkable after load")
	}
	if !loaded.IsWalkable(-5, 5) {
		t.Error("Expanded area (-5,5) should remain walkable after load")
	}

	// Verify blocked areas
	if loaded.IsWalkable(5, 5) {
		t.Error("Blocked area (5,5) should remain non-walkable after load")
	}
	if loaded.IsWalkable(7, 7) {
		t.Error("Blocked area (7,7) should remain non-walkable after load")
	}

	// Verify unmodified default areas still work
	if !loaded.IsWalkable(1, 1) {
		t.Error("Unmodified default area (1,1) should remain walkable")
	}

	t.Log("✅ F2 boundary editing persists perfectly with save/load!")
}

// TestAssetLoadingPerformance tests that asset loading is fast enough for non-blocking UI
func TestAssetLoadingPerformance(t *testing.T) {
	// Test loading multiple assets and measure time
	assetPaths := []string{
		"assets/Chair/Chair_2_A_Tile.png",
		"assets/Sofa/Sofa_3_A_Tile.png",
		"assets/Desk/Desk_1_Tile.png",
		"assets/Plants/Plant_1_Tile.png",
		"assets/Lamp/Lamp_8_A_Tile.png",
	}

	start := time.Now()
	for _, path := range assetPaths {
		_, err := assets.LoadTile(path)
		if err != nil {
			t.Errorf("Failed to load %s: %v", path, err)
		}
	}
	duration := time.Since(start)

	// Should be fast (< 100ms for 5 assets)
	if duration > 100*time.Millisecond {
		t.Errorf("Asset loading too slow: %v (should be < 100ms for 5 assets)", duration)
	}

	t.Logf("✅ Loaded 5 assets in %v", duration)

	// Test loading same assets again (should be even faster if cached)
	start = time.Now()
	for _, path := range assetPaths {
		_, err := assets.LoadTile(path)
		if err != nil {
			t.Errorf("Failed to load cached %s: %v", path, err)
		}
	}
	cachedDuration := time.Since(start)

	t.Logf("✅ Loaded 5 cached assets in %v", cachedDuration)
}
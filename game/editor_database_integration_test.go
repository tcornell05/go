package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/tcornell05/go/game/internal/database"
	"github.com/tcornell05/go/game/internal/world"
	"github.com/tcornell05/go/game/pkg/assets"
)

//go:embed assets/*
var editorTestAssets embed.FS

func init() {
	assets.GlobalAssets = editorTestAssets
}

// TestEditorDatabaseIntegration tests the complete editor workflow with database persistence
func TestEditorDatabaseIntegration(t *testing.T) {
	testDir, err := os.MkdirTemp("", "editor_db_integration_test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create test database
	dbPath := filepath.Join(testDir, "editor_integration.db")
	db, err := database.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	t.Log("=== PHASE 1: Create and save WorldTile C3 through database ===")

	// Create a WorldTile directly and save to database (simulating editor creation)
	worldTile := world.NewWorldTile("C3", 2, 2, 0)
	worldTile.Name = "Editor Test Room"
	worldTile.Type = world.TileTypeRoom

	// Add some furniture
	chair := world.TileData{
		AssetPath: "assets/Chair/Chair_2_A_Tile.png",
		X:         3,
		Y:         4,
		Layer:     2,
		Category:  "furniture",
		OffsetX:   0.2,
		OffsetY:   0.3,
	}
	err = worldTile.AddTile(chair)
	if err != nil {
		t.Fatalf("Failed to add chair to world tile: %v", err)
	}

	// Add a desk
	desk := world.TileData{
		AssetPath: "assets/Desk/Desk_1_Tile.png",
		X:         6,
		Y:         7,
		Layer:     2,
		Category:  "furniture",
		OffsetX:   0.8,
		OffsetY:   0.9,
	}
	err = worldTile.AddTile(desk)
	if err != nil {
		t.Fatalf("Failed to add desk to world tile: %v", err)
	}

	// Set custom walkability
	worldTile.SetWalkable(8, 8, true)  // Expand walkable area
	worldTile.SetWalkable(2, 2, false) // Block an area

	// Save to database
	err = db.SaveWorldTile(worldTile)
	if err != nil {
		t.Fatalf("Failed to save world tile to database: %v", err)
	}

	t.Logf("Created and saved WorldTile C3 with %d tiles", len(worldTile.Tiles))

	t.Log("=== PHASE 2: Simulate editor loading WorldTile C3 from database ===")

	// Clear the tile from memory (simulate app restart)
	worldTile = nil

	// Load from database (simulating editor loading)
	loadedTile, err := db.GetWorldTile("C3")
	if err != nil {
		t.Fatalf("Failed to load world tile from database: %v", err)
	}

	// Verify basic properties
	if loadedTile.ID != "C3" {
		t.Errorf("Expected tile ID C3, got %s", loadedTile.ID)
	}
	if loadedTile.Name != "Editor Test Room" {
		t.Errorf("Expected name 'Editor Test Room', got %s", loadedTile.Name)
	}
	if loadedTile.WorldX != 2 || loadedTile.WorldY != 2 {
		t.Errorf("Expected position (2,2), got (%d,%d)", loadedTile.WorldX, loadedTile.WorldY)
	}

	// Verify furniture persisted
	if len(loadedTile.Tiles) != 2 {
		t.Errorf("Expected 2 furniture pieces, got %d", len(loadedTile.Tiles))
	}

	// Find the chair and desk
	chairFound, deskFound := false, false
	for _, tileData := range loadedTile.Tiles {
		if tileData.AssetPath == "assets/Chair/Chair_2_A_Tile.png" {
			chairFound = true
			if tileData.X != 3 || tileData.Y != 4 {
				t.Errorf("Chair position mismatch: expected (3,4), got (%d,%d)", tileData.X, tileData.Y)
			}
			if tileData.OffsetX != 0.2 || tileData.OffsetY != 0.3 {
				t.Errorf("Chair offset mismatch: expected (0.2,0.3), got (%f,%f)", tileData.OffsetX, tileData.OffsetY)
			}
		}
		if tileData.AssetPath == "assets/Desk/Desk_1_Tile.png" {
			deskFound = true
			if tileData.X != 6 || tileData.Y != 7 {
				t.Errorf("Desk position mismatch: expected (6,7), got (%d,%d)", tileData.X, tileData.Y)
			}
			if tileData.OffsetX != 0.8 || tileData.OffsetY != 0.9 {
				t.Errorf("Desk offset mismatch: expected (0.8,0.9), got (%f,%f)", tileData.OffsetX, tileData.OffsetY)
			}
		}
	}

	if !chairFound {
		t.Error("Chair not found after database reload")
	}
	if !deskFound {
		t.Error("Desk not found after database reload")
	}

	// Verify walkability changes persisted
	if !loadedTile.IsWalkable(8, 8) {
		t.Error("Expanded walkable area (8,8) should persist")
	}
	if loadedTile.IsWalkable(2, 2) {
		t.Error("Blocked area (2,2) should persist as non-walkable")
	}

	t.Log("✅ Editor database integration test PASSED")

	t.Log("=== PHASE 3: Test editor creation workflow simulation ===")

	// Simulate editor creating a new WorldTile through normal workflow
	// This would normally happen through NewEditorGame with tileID parameter
	newTileID := "D4"
	
	// The editor would call GetWorldTile for a non-existent tile, which should fail
	_, err = db.GetWorldTile(newTileID)
	if err == nil {
		t.Errorf("Expected error when loading non-existent tile %s", newTileID)
	}

	// Editor would then create a new WorldTile
	newTile := world.NewWorldTile(newTileID, 3, 3, 0)
	newTile.Name = fmt.Sprintf("Tile %s", newTileID)

	// Add some content and save
	lamp := world.TileData{
		AssetPath: "assets/Lamp/Lamp_8_A_Tile.png",
		X:         5,
		Y:         6,
		Layer:     2,
		Category:  "decorations",
	}
	err = newTile.AddTile(lamp)
	if err != nil {
		t.Fatalf("Failed to add lamp to new tile: %v", err)
	}

	err = db.SaveWorldTile(newTile)
	if err != nil {
		t.Fatalf("Failed to save new tile to database: %v", err)
	}

	// Verify the new tile can be loaded
	verifyTile, err := db.GetWorldTile(newTileID)
	if err != nil {
		t.Fatalf("Failed to load newly created tile: %v", err)
	}

	if len(verifyTile.Tiles) != 1 {
		t.Errorf("Expected 1 tile in new WorldTile, got %d", len(verifyTile.Tiles))
	}

	t.Log("✅ Editor new tile creation workflow test PASSED")

	t.Log("=== ALL EDITOR DATABASE INTEGRATION TESTS PASSED ===")
}
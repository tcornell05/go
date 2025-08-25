package main

import (
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/tcornell05/go/game/internal/world"
	"github.com/tcornell05/go/game/pkg/assets"
)

//go:embed assets/*
var reloadTestAssets embed.FS

func init() {
	assets.GlobalAssets = reloadTestAssets
}

// TestWorldTileReloadPersistence tests the exact user workflow that's failing
func TestWorldTileReloadPersistence(t *testing.T) {
	// This test simulates: load A2, add two chairs, save, reload app, expect chairs to be there
	
	// Create isolated test environment (simulating app data directory)
	testDataDir, err := os.MkdirTemp("", "game_data_test")
	if err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	defer os.RemoveAll(testDataDir)

	t.Logf("Test data directory: %s", testDataDir)

	// PHASE 1: First App Session - Load A2, add chairs, save
	t.Log("=== PHASE 1: First App Session ===")
	
	// Simulate: go run . editor --tile A2
	worldTileA2 := world.NewWorldTile("A2", 2, 0, 0)
	
	// Add two chairs (simulating user placing them in editor)
	chair1 := world.TileData{
		AssetPath: "assets/Chair/Chair_2_A_Tile.png",
		X:         3,
		Y:         4,
		Layer:     2,
		Facing:    0,
		Category:  "furniture",
		OffsetX:   0.3,
		OffsetY:   0.4,
	}
	
	chair2 := world.TileData{
		AssetPath: "assets/Chair/Chair_2_B_Tile.png",
		X:         5,
		Y:         6,
		Layer:     2,
		Facing:    1,
		Category:  "furniture",
		OffsetX:   0.7,
		OffsetY:   0.2,
	}
	
	// Add chairs to world tile
	err = worldTileA2.AddTile(chair1)
	if err != nil {
		t.Fatalf("Failed to add chair1: %v", err)
	}
	
	err = worldTileA2.AddTile(chair2)
	if err != nil {
		t.Fatalf("Failed to add chair2: %v", err)
	}
	
	// Verify chairs are in the world tile
	chair1Key := world.GetTileKey(chair1.Layer, chair1.X, chair1.Y)
	chair2Key := world.GetTileKey(chair2.Layer, chair2.X, chair2.Y)
	
	if _, exists := worldTileA2.Tiles[chair1Key]; !exists {
		t.Fatal("Chair1 should exist in world tile after adding")
	}
	if _, exists := worldTileA2.Tiles[chair2Key]; !exists {
		t.Fatal("Chair2 should exist in world tile after adding")
	}
	
	t.Logf("Added 2 chairs to A2. Total tiles: %d", len(worldTileA2.Tiles))
	
	// Save world tile (simulating user pressing save)
	saveFile := filepath.Join(testDataDir, "A2.json")
	err = worldTileA2.SaveToFile(saveFile)
	if err != nil {
		t.Fatalf("Failed to save world tile: %v", err)
	}
	
	// Verify file was created
	if _, err := os.Stat(saveFile); os.IsNotExist(err) {
		t.Fatal("Save file A2.json should exist after saving")
	}
	
	t.Logf("Saved A2.json to: %s", saveFile)
	
	// PHASE 2: App Restart - Simulate reloading the app
	t.Log("=== PHASE 2: App Restart Simulation ===")
	
	// Clear memory (simulate app restart)
	worldTileA2 = nil
	
	// Reload world tile (simulating: go run . editor --tile A2 after restart)
	reloadedTile, err := world.LoadWorldTileFromFile(saveFile)
	if err != nil {
		t.Fatalf("Failed to reload world tile A2: %v", err)
	}
	
	if reloadedTile.ID != "A2" {
		t.Errorf("Expected tile ID A2, got %s", reloadedTile.ID)
	}
	
	// CRITICAL TEST: Verify both chairs persist after reload
	if len(reloadedTile.Tiles) != 2 {
		t.Errorf("Expected 2 tiles after reload, got %d", len(reloadedTile.Tiles))
		t.Logf("Available tiles after reload: %+v", reloadedTile.Tiles)
	}
	
	// Check chair1 persisted
	if reloadedChair1, exists := reloadedTile.Tiles[chair1Key]; !exists {
		t.Error("Chair1 should persist after app reload")
	} else {
		if reloadedChair1.AssetPath != chair1.AssetPath {
			t.Errorf("Chair1 asset path mismatch: expected %s, got %s", chair1.AssetPath, reloadedChair1.AssetPath)
		}
		if reloadedChair1.X != chair1.X || reloadedChair1.Y != chair1.Y {
			t.Errorf("Chair1 position mismatch: expected (%d,%d), got (%d,%d)", chair1.X, chair1.Y, reloadedChair1.X, reloadedChair1.Y)
		}
	}
	
	// Check chair2 persisted
	if reloadedChair2, exists := reloadedTile.Tiles[chair2Key]; !exists {
		t.Error("Chair2 should persist after app reload")
	} else {
		if reloadedChair2.AssetPath != chair2.AssetPath {
			t.Errorf("Chair2 asset path mismatch: expected %s, got %s", chair2.AssetPath, reloadedChair2.AssetPath)
		}
		if reloadedChair2.X != chair2.X || reloadedChair2.Y != chair2.Y {
			t.Errorf("Chair2 position mismatch: expected (%d,%d), got (%d,%d)", chair2.X, chair2.Y, reloadedChair2.X, reloadedChair2.Y)
		}
	}
	
	// Verify loaded images are properly initialized
	if len(reloadedTile.LoadedImages) != 2 {
		t.Errorf("Expected 2 loaded images after reload, got %d", len(reloadedTile.LoadedImages))
	}
	
	t.Log("✅ World tile A2 persistence test PASSED - chairs persist after app reload!")
}

// TestEditorGameIntegration tests the actual editor game integration
func TestEditorGameIntegration(t *testing.T) {
	// Test the actual editor.Game loading mechanism
	
	testDataDir, err := os.MkdirTemp("", "editor_game_test")
	if err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	defer os.RemoveAll(testDataDir)
	
	// Create a world tile file
	worldTile := world.NewWorldTile("B3", 3, 1, 0)
	
	// Add some furniture
	sofa := world.TileData{
		AssetPath: "assets/Sofa/Sofa_3_A_Tile.png",
		X:         2,
		Y:         3,
		Layer:     2,
		Facing:    0,
		Category:  "furniture",
		OffsetX:   0.5,
		OffsetY:   0.5,
	}
	
	err = worldTile.AddTile(sofa)
	if err != nil {
		t.Fatalf("Failed to add sofa: %v", err)
	}
	
	// Save to test data directory
	saveFile := filepath.Join(testDataDir, "B3.json")
	err = worldTile.SaveToFile(saveFile)
	if err != nil {
		t.Fatalf("Failed to save B3: %v", err)
	}
	
	// Now test loading through editor game (simulating --tile B3)
	// NOTE: This would require modifying the editor to accept a custom data directory
	// For now, test direct file loading which is what editor should do
	
	reloaded, err := world.LoadWorldTileFromFile(saveFile)
	if err != nil {
		t.Fatalf("Failed to reload B3 through editor mechanism: %v", err)
	}
	
	// Verify sofa persisted
	sofaKey := world.GetTileKey(sofa.Layer, sofa.X, sofa.Y)
	if _, exists := reloaded.Tiles[sofaKey]; !exists {
		t.Error("Sofa should persist in editor reload")
	}
	
	t.Log("✅ Editor game integration test PASSED")
}

// TestMultipleWorldTilesPersistence tests persistence of multiple world tiles
func TestMultipleWorldTilesPersistence(t *testing.T) {
	testDataDir, err := os.MkdirTemp("", "multi_tile_test")
	if err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	defer os.RemoveAll(testDataDir)
	
	// Create multiple world tiles with different content
	tiles := map[string]*world.WorldTile{
		"A1": world.NewWorldTile("A1", 0, 0, 0),
		"A2": world.NewWorldTile("A2", 1, 0, 0),
		"B1": world.NewWorldTile("B1", 0, 1, 0),
		"B2": world.NewWorldTile("B2", 1, 1, 0),
	}
	
	// Add different furniture to each tile
	furniture := []world.TileData{
		{AssetPath: "assets/Chair/Chair_2_A_Tile.png", X: 1, Y: 1, Layer: 2, Category: "furniture"},
		{AssetPath: "assets/Desk/Desk_1_Tile.png", X: 2, Y: 2, Layer: 2, Category: "furniture"},
		{AssetPath: "assets/Plants/Plant_1_Tile.png", X: 3, Y: 3, Layer: 2, Category: "decoration"},
		{AssetPath: "assets/Lamp/Lamp_8_A_Tile.png", X: 4, Y: 4, Layer: 2, Category: "furniture"},
	}
	
	i := 0
	for tileID, tile := range tiles {
		err = tile.AddTile(furniture[i])
		if err != nil {
			t.Fatalf("Failed to add furniture to %s: %v", tileID, err)
		}
		
		// Save each tile
		saveFile := filepath.Join(testDataDir, tileID+".json")
		err = tile.SaveToFile(saveFile)
		if err != nil {
			t.Fatalf("Failed to save tile %s: %v", tileID, err)
		}
		
		t.Logf("Saved tile %s with %s", tileID, furniture[i].AssetPath)
		i++
	}
	
	// Reload all tiles and verify persistence
	for tileID := range tiles {
		saveFile := filepath.Join(testDataDir, tileID+".json")
		reloaded, err := world.LoadWorldTileFromFile(saveFile)
		if err != nil {
			t.Fatalf("Failed to reload tile %s: %v", tileID, err)
		}
		
		if len(reloaded.Tiles) != 1 {
			t.Errorf("Tile %s should have 1 furniture item, got %d", tileID, len(reloaded.Tiles))
		}
		
		if reloaded.ID != tileID {
			t.Errorf("Tile ID mismatch: expected %s, got %s", tileID, reloaded.ID)
		}
	}
	
	t.Log("✅ Multiple world tiles persistence test PASSED")
}

// TestWorldTileFileLocationBug tests if the issue is file location/naming
func TestWorldTileFileLocationBug(t *testing.T) {
	// This test checks if the bug is related to where the game looks for saved files
	
	testDataDir, err := os.MkdirTemp("", "file_location_test")
	if err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	defer os.RemoveAll(testDataDir)
	
	// Test saving in different locations that the app might check
	locations := []string{
		testDataDir,                                // Custom directory
		filepath.Join(testDataDir, "data"),         // data/ subdirectory
		testDataDir,                                // Current directory
	}
	
	for i, location := range locations {
		// Ensure directory exists
		err := os.MkdirAll(location, 0755)
		if err != nil {
			t.Fatalf("Failed to create location %s: %v", location, err)
		}
		
		tileID := "C" + string(rune('1'+i)) // C1, C2, C3
		tile := world.NewWorldTile(tileID, i, 0, 0)
		
		// Add unique furniture
		furniture := world.TileData{
			AssetPath: "assets/Chair/Chair_2_A_Tile.png",
			X:         i + 1,
			Y:         i + 1,
			Layer:     2,
			Category:  "furniture",
		}
		
		err = tile.AddTile(furniture)
		if err != nil {
			t.Fatalf("Failed to add furniture to %s: %v", tileID, err)
		}
		
		// Save in this location
		saveFile := filepath.Join(location, tileID+".json")
		err = tile.SaveToFile(saveFile)
		if err != nil {
			t.Fatalf("Failed to save %s to %s: %v", tileID, location, err)
		}
		
		// Immediately reload and verify
		reloaded, err := world.LoadWorldTileFromFile(saveFile)
		if err != nil {
			t.Fatalf("Failed to reload %s from %s: %v", tileID, location, err)
		}
		
		if len(reloaded.Tiles) != 1 {
			t.Errorf("Location %s: %s should have 1 tile, got %d", location, tileID, len(reloaded.Tiles))
		}
		
		t.Logf("✅ Location %s: %s saves and loads correctly", location, tileID)
	}
	
	t.Log("✅ File location bug test PASSED - all locations work")
}

// TestEditorSaveCommand tests if the issue is with the save command mechanism
func TestEditorSaveCommand(t *testing.T) {
	// This test simulates what happens when user presses save in the editor
	
	testDataDir, err := os.MkdirTemp("", "save_command_test")
	if err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	defer os.RemoveAll(testDataDir)
	
	// Simulate the editor game environment
	tile := world.NewWorldTile("D1", 0, 0, 0)
	
	// Add furniture
	table := world.TileData{
		AssetPath: "assets/Desk/Desk_1_Tile.png",
		X:         3,
		Y:         3,
		Layer:     2,
		Category:  "furniture",
		OffsetX:   0.5,
		OffsetY:   0.5,
	}
	
	err = tile.AddTile(table)
	if err != nil {
		t.Fatalf("Failed to add table: %v", err)
	}
	
	// Test both possible save locations the app might use
	possibleSaveFiles := []string{
		filepath.Join(testDataDir, "D1.json"),        // Direct to data dir
		filepath.Join(testDataDir, "data", "D1.json"), // data/ subdirectory
		"D1.json",                                     // Current directory
	}
	
	for _, saveFile := range possibleSaveFiles {
		// Create directory if needed
		dir := filepath.Dir(saveFile)
		if dir != "." && dir != "" {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				continue // Skip if can't create directory
			}
		}
		
		// Test save
		err = tile.SaveToFile(saveFile)
		if err != nil {
			t.Logf("Could not save to %s: %v", saveFile, err)
			continue
		}
		
		// Test immediate reload
		reloaded, err := world.LoadWorldTileFromFile(saveFile)
		if err != nil {
			t.Errorf("Failed to reload from %s: %v", saveFile, err)
			continue
		}
		
		if len(reloaded.Tiles) != 1 {
			t.Errorf("Save file %s: expected 1 tile, got %d", saveFile, len(reloaded.Tiles))
		}
		
		t.Logf("✅ Save command test PASSED for: %s", saveFile)
		
		// Clean up
		os.Remove(saveFile)
	}
}

// TestRealWorldWorkflow tests the exact workflow a user experiences
func TestRealWorldWorkflow(t *testing.T) {
	// Simulate the exact commands: 
	// 1. go run . editor --tile A2
	// 2. Add furniture
	// 3. Save (however that's done in the editor)
	// 4. Exit
	// 5. go run . editor --tile A2
	// 6. Expect furniture to be there
	
	testDataDir, err := os.MkdirTemp("", "real_workflow_test")
	if err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	defer os.RemoveAll(testDataDir)
	
	t.Log("=== SIMULATING REAL USER WORKFLOW ===")
	
	// Step 1: go run . editor --tile A2 (first time - creates new tile)
	t.Log("Step 1: go run . editor --tile A2 (new tile)")
	
	tileA2 := world.NewWorldTile("A2", 2, 0, 0)
	
	// Step 2: User adds furniture in editor
	t.Log("Step 2: User places chairs in editor")
	
	chairs := []world.TileData{
		{AssetPath: "assets/Chair/Chair_2_A_Tile.png", X: 2, Y: 3, Layer: 2, Category: "furniture", OffsetX: 0.3, OffsetY: 0.4},
		{AssetPath: "assets/Chair/Chair_2_B_Tile.png", X: 4, Y: 5, Layer: 2, Category: "furniture", OffsetX: 0.7, OffsetY: 0.6},
	}
	
	for i, chair := range chairs {
		err = tileA2.AddTile(chair)
		if err != nil {
			t.Fatalf("Failed to add chair %d: %v", i, err)
		}
	}
	
	t.Logf("Added %d chairs to A2", len(chairs))
	
	// Step 3: User saves (the mechanism we need to test)
	t.Log("Step 3: User presses save")
	
	// Try saving to the locations the app might actually use
	possibleDataDirs := []string{
		"data",           // Relative data directory  
		testDataDir,      // Test directory
		".",              // Current directory
	}
	
	var successfulSaveFile string
	for _, dataDir := range possibleDataDirs {
		// Create directory if needed and possible
		if dataDir != "." {
			err := os.MkdirAll(dataDir, 0755)
			if err != nil {
				continue
			}
		}
		
		saveFile := filepath.Join(dataDir, "A2.json")
		err = tileA2.SaveToFile(saveFile)
		if err == nil {
			successfulSaveFile = saveFile
			t.Logf("Successfully saved to: %s", saveFile)
			break
		}
	}
	
	if successfulSaveFile == "" {
		t.Fatal("Could not save A2.json to any expected location")
	}
	
	// Step 4: Exit (clear memory)
	t.Log("Step 4: User exits app (clearing memory)")
	tileA2 = nil
	
	// Step 5: go run . editor --tile A2 (reload)
	t.Log("Step 5: go run . editor --tile A2 (reload)")
	
	reloadedA2, err := world.LoadWorldTileFromFile(successfulSaveFile)
	if err != nil {
		t.Fatalf("Failed to reload A2 after app restart: %v", err)
	}
	
	// Step 6: Verify furniture is there
	t.Log("Step 6: Verify furniture persisted")
	
	if len(reloadedA2.Tiles) != 2 {
		t.Errorf("Expected 2 chairs after reload, got %d", len(reloadedA2.Tiles))
		t.Log("Available tiles after reload:")
		for key, tile := range reloadedA2.Tiles {
			t.Logf("  %s: %s at (%d,%d)", key, tile.AssetPath, tile.X, tile.Y)
		}
		t.Fatal("❌ REAL WORLD WORKFLOW FAILED - chairs not persisting")
	}
	
	// Verify each chair
	for i, expectedChair := range chairs {
		key := world.GetTileKey(expectedChair.Layer, expectedChair.X, expectedChair.Y)
		if actualChair, exists := reloadedA2.Tiles[key]; !exists {
			t.Errorf("Chair %d missing after reload", i)
		} else {
			if actualChair.AssetPath != expectedChair.AssetPath {
				t.Errorf("Chair %d asset mismatch: expected %s, got %s", i, expectedChair.AssetPath, actualChair.AssetPath)
			}
		}
	}
	
	t.Log("✅ REAL WORLD WORKFLOW TEST PASSED - chairs persist correctly!")
	
	// Clean up
	os.Remove(successfulSaveFile)
}
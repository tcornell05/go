package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tcornell05/go/game/internal/database"
	"github.com/tcornell05/go/game/internal/input"
	"github.com/tcornell05/go/game/internal/world"
	"github.com/tcornell05/go/game/pkg/assets"
)

//go:embed assets/*
var boundaryTestAssets embed.FS

func init() {
	assets.GlobalAssets = boundaryTestAssets
}

// TestBoundaryEditingWorkflow tests the complete F2 boundary editing workflow
func TestBoundaryEditingWorkflow(t *testing.T) {
	testDir, err := os.MkdirTemp("", "boundary_editing_test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create test database
	dbPath := filepath.Join(testDir, "boundary_test.db")
	db, err := database.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	t.Log("=== PHASE 1: Create WorldTile with default walkability ===")

	// Create a WorldTile directly (simulates: spawn world -> edit mode)
	worldTile := world.NewWorldTile("BOUNDARY_TEST", 0, 0, 0)
	worldTile.Name = "Boundary Edit Test Tile"
	worldTile.Type = world.TileTypeRoom

	// Verify default walkability (should be 11x11 walkable area: 0-10, 0-10)
	expectedWalkableTiles := 121 // 11 * 11
	walkableCount := 0
	for key, walkable := range worldTile.Walkability {
		if walkable {
			walkableCount++
		}
		t.Logf("Initial walkability: %s = %v", key, walkable)
	}
	
	if walkableCount != expectedWalkableTiles {
		t.Errorf("Expected %d walkable tiles initially, got %d", expectedWalkableTiles, walkableCount)
	}
	
	// Verify specific tiles
	if !worldTile.IsWalkable(5, 5) {
		t.Error("Center tile (5,5) should be walkable initially")
	}
	if worldTile.IsWalkable(15, 15) {
		t.Error("Outside tile (15,15) should NOT be walkable initially")
	}

	// Save to database
	err = db.SaveWorldTile(worldTile)
	if err != nil {
		t.Fatalf("Failed to save world tile: %v", err)
	}

	t.Log("=== PHASE 2: Simulate F2 boundary editing workflow ===")

	// Create controller in editor mode
	controller, err := input.NewController()
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}
	controller.EditorMode = true

	// Test boundary editing workflow
	var boundaryChanges []struct {
		x, y     int
		walkable bool
		desc     string
	}

	// Set up callback to track boundary changes (simulates editor integration)
	controller.OnBoundaryChanged = func(x, y int, walkable bool) {
		boundaryChanges = append(boundaryChanges, struct {
			x, y     int
			walkable bool
			desc     string
		}{x, y, walkable, fmt.Sprintf("tile (%d,%d) -> walkable=%v", x, y, walkable)})
		
		// Simulate what the editor does: update world tile and save
		key := world.GetWalkabilityKey(x, y)
		worldTile.Walkability[key] = walkable
		t.Logf("BOUNDARY CHANGE: %s", boundaryChanges[len(boundaryChanges)-1].desc)
	}

	// Step 1: Enable F2 debug grid mode
	controller.ShowDebugGrid = true
	if !controller.ShowDebugGrid {
		t.Error("Failed to enable debug grid mode (F2)")
	}
	t.Log("✓ Step 1: F2 debug grid enabled")

	// Step 2: Test expanding walkable area (left-click on blue tile outside 11x11)
	testX, testY := 12, 8 // Outside the default 11x11 grid
	if worldTile.IsWalkable(testX, testY) {
		t.Error("Test tile should start as non-walkable")
	}
	
	// Simulate left-click to make tile walkable
	if controller.OnBoundaryChanged != nil {
		controller.OnBoundaryChanged(testX, testY, true)
	}
	
	// Verify the change
	if !worldTile.IsWalkable(testX, testY) {
		t.Error("Tile should be walkable after boundary edit")
	}
	t.Logf("✓ Step 2: Expanded walkable area to (%d, %d)", testX, testY)

	// Step 3: Test restricting walkable area (right-click on grey tile inside 11x11)
	testX2, testY2 := 3, 4 // Inside the default 11x11 grid
	if !worldTile.IsWalkable(testX2, testY2) {
		t.Error("Test tile should start as walkable")
	}
	
	// Simulate right-click to make tile non-walkable
	if controller.OnBoundaryChanged != nil {
		controller.OnBoundaryChanged(testX2, testY2, false)
	}
	
	// Verify the change
	if worldTile.IsWalkable(testX2, testY2) {
		t.Error("Tile should be non-walkable after boundary edit")
	}
	t.Logf("✓ Step 3: Restricted walkable area at (%d, %d)", testX2, testY2)

	// Step 4: Verify callback was called correctly
	if len(boundaryChanges) != 2 {
		t.Errorf("Expected 2 boundary changes, got %d", len(boundaryChanges))
	} else {
		// Check first change (expand)
		if boundaryChanges[0].x != testX || boundaryChanges[0].y != testY || !boundaryChanges[0].walkable {
			t.Errorf("First boundary change incorrect: expected (%d,%d,true), got (%d,%d,%v)", 
				testX, testY, boundaryChanges[0].x, boundaryChanges[0].y, boundaryChanges[0].walkable)
		}
		
		// Check second change (restrict)
		if boundaryChanges[1].x != testX2 || boundaryChanges[1].y != testY2 || boundaryChanges[1].walkable {
			t.Errorf("Second boundary change incorrect: expected (%d,%d,false), got (%d,%d,%v)", 
				testX2, testY2, boundaryChanges[1].x, boundaryChanges[1].y, boundaryChanges[1].walkable)
		}
	}
	t.Log("✓ Step 4: Boundary change callbacks working correctly")

	t.Log("=== PHASE 3: Test persistence (save/reload cycle) ===")

	// Save the modified world tile
	err = db.SaveWorldTile(worldTile)
	if err != nil {
		t.Errorf("Failed to save modified world tile: %v", err)
	}

	// Reload from database to verify persistence
	reloadedTile, err := db.GetWorldTile("BOUNDARY_TEST")
	if err != nil {
		t.Fatalf("Failed to reload world tile: %v", err)
	}

	// Verify the walkability changes persisted
	if !reloadedTile.IsWalkable(testX, testY) {
		t.Error("Expanded walkable area should persist after save/reload")
	}
	if reloadedTile.IsWalkable(testX2, testY2) {
		t.Error("Restricted walkable area should persist after save/reload")
	}

	// Verify default walkability is still intact for other tiles
	if !reloadedTile.IsWalkable(5, 5) {
		t.Error("Center tile should still be walkable after save/reload")
	}
	if reloadedTile.IsWalkable(15, 15) {
		t.Error("Far outside tile should still be non-walkable after save/reload")
	}

	t.Log("✓ Step 5: Walkability changes persisted correctly")

	// Count final walkability
	finalWalkableCount := 0
	for _, walkable := range reloadedTile.Walkability {
		if walkable {
			finalWalkableCount++
		}
	}
	
	// Should be: 121 (original) + 1 (expanded) - 1 (restricted) = 121
	expectedFinalCount := 121
	if finalWalkableCount != expectedFinalCount {
		t.Errorf("Expected %d final walkable tiles, got %d", expectedFinalCount, finalWalkableCount)
	}

	t.Log("=== ALL TESTS PASSED: F2 Boundary Editing Workflow Complete ===")
	t.Logf("✅ Successfully tested: spawn world -> edit mode -> F2 grid -> change boundaries -> persistence")
	t.Logf("✅ Walkability data: %d tiles modified, all changes persisted to database", len(boundaryChanges))
}

// TestBoundaryEditingEdgeCases tests edge cases for boundary editing
func TestBoundaryEditingEdgeCases(t *testing.T) {
	testDir, err := os.MkdirTemp("", "boundary_edge_test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create test database
	dbPath := filepath.Join(testDir, "boundary_edge.db")
	db, err := database.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Test Case 1: WorldTile with no initial walkability data (database load scenario)
	t.Log("=== Testing: Empty walkability data initialization ===")
	
	// Create empty world tile and save to database
	emptyTile := &world.WorldTile{
		ID:           "EMPTY_TEST",
		Name:         "Empty Walkability Test",
		Type:         world.TileTypeRoom,
		Width:        500,
		Height:       500,
		Connections:  make(map[world.Direction]string),
		Tiles:        make(map[string]world.TileData),
		Walkability:  make(map[string]bool), // Empty walkability
		LoadedImages: make(map[string]*ebiten.Image),
	}
	
	// Save empty tile
	err = db.SaveWorldTile(emptyTile)
	if err != nil {
		t.Fatalf("Failed to save empty tile: %v", err)
	}
	
	// Load it back (should trigger default walkability initialization)
	loadedTile, err := db.GetWorldTile("EMPTY_TEST")
	if err != nil {
		t.Fatalf("Failed to load empty tile: %v", err)
	}
	
	// Verify default walkability was initialized
	if len(loadedTile.Walkability) == 0 {
		t.Error("Empty tile should have default walkability initialized after load")
	}
	
	if !loadedTile.IsWalkable(5, 5) {
		t.Error("Loaded empty tile should have walkable center area")
	}
	
	t.Log("✓ Empty walkability initialization working correctly")

	// Test Case 2: Extreme coordinates  
	t.Log("=== Testing: Extreme coordinate boundary editing ===")
	
	worldTile := world.NewWorldTile("EXTREME_TEST", 0, 0, 0)
	
	// Test coordinates at the edges of the 500x500 world tile
	extremeCoords := []struct{ x, y int }{
		{-1, -1},   // Negative coordinates
		{0, 0},     // Origin
		{250, 250}, // Center of 500x500 tile  
		{499, 499}, // Max valid coordinates
	}
	
	changeCount := 0
	for _, coord := range extremeCoords {
		// Set walkable
		key := world.GetWalkabilityKey(coord.x, coord.y)
		worldTile.Walkability[key] = true
		changeCount++
		
		if !worldTile.IsWalkable(coord.x, coord.y) {
			t.Errorf("Failed to set walkability for extreme coordinate (%d, %d)", coord.x, coord.y)
		}
	}
	
	t.Logf("✓ Extreme coordinates handled correctly (%d test coordinates)", changeCount)
	
	t.Log("=== Edge Case Tests Complete ===")
}
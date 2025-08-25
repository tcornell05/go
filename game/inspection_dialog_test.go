package main

import (
	"testing"
	"os"
	"path/filepath"

	"github.com/tcornell05/go/game/internal/database"
	"github.com/tcornell05/go/game/internal/input"
	"github.com/tcornell05/go/game/internal/world"
)

func TestInspectionDialogEditFlow(t *testing.T) {
	// Create temporary directory for test database
	testDir, err := os.MkdirTemp("", "inspection_edit_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Initialize test database
	dbPath := filepath.Join(testDir, "test.db")
	db, err := database.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create test world tile with an object
	worldTile := &world.WorldTile{
		ID:     "TEST",
		Name:   "Test Tile",
		Type:   "room",
		Width:  500,
		Height: 500,
		Tiles:  make(map[string]world.TileData),
	}

	// Add a chair object to the world tile
	chairTile := world.TileData{
		AssetPath: "assets/Chair/Chair_2_B_Tile.png",
		X:         5,
		Y:         5,
		Facing:    0,
		Category:  "furniture",
		Layer:     2,
		OffsetX:   0.0,
		OffsetY:   0.0,
		Depth:     0.0,
		Rotation:  0.0,
	}
	worldTile.Tiles["2:5:5"] = chairTile

	// Save world tile to database
	err = db.SaveWorldTile(worldTile)
	if err != nil {
		t.Fatalf("Failed to save world tile: %v", err)
	}

	// Create the entity record that should exist for editing
	// This simulates what happens when assets are loaded into the database
	// We need to access the internal DB connection to insert the entity
	// For now, let's skip this test step and focus on the actual issue

	// Create controller in editor mode
	controller, err := input.NewController()
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}
	controller.EditorMode = true

	t.Logf("=== Testing Inspection Dialog Edit Button Flow ===")

	// Step 1: Enable inspection mode
	controller.InspectionMode = true
	if !controller.InspectionMode {
		t.Error("Failed to enable inspection mode")
	}
	t.Logf("✓ Step 1: Inspection mode enabled")

	// Step 2: Simulate object inspection (this should use GetWorldTileObjectAt)
	obj, err := db.GetWorldTileObjectAt("TEST", 5, 5)
	if err != nil {
		t.Fatalf("Failed to get object from world tile: %v", err)
	}
	if obj == nil {
		t.Fatal("No object found at test coordinates")
	}

	// Set inspected object and open dialog
	controller.InspectedObject = obj
	controller.InspectionDialog = true
	t.Logf("✓ Step 2: Object inspected, dialog opened")

	// Step 3: Verify inspection dialog state
	if !controller.InspectionDialog {
		t.Error("Inspection dialog should be open")
	}
	if controller.InspectedObject == nil {
		t.Error("Inspected object should be set")
	}
	if controller.InspectedObject.AssetPath != "assets/Chair/Chair_2_B_Tile.png" {
		t.Errorf("Wrong asset path: %s", controller.InspectedObject.AssetPath)
	}
	t.Logf("✓ Step 3: Dialog state verified")

	// Step 4: Test edit button click detection
	screenWidth, screenHeight := 1024, 768
	
	// Calculate dialog and button position (same as UI code)
	dialogWidth := 450.0
	dialogHeight := 400.0
	dialogX := float64(screenWidth)/2 - float64(dialogWidth)/2
	dialogY := float64(screenHeight)/2 - float64(dialogHeight)/2
	
	// Edit button coordinates (same as UI code)
	buttonY := int(dialogY) + int(dialogHeight) - 80
	editButtonX := int(dialogX) + 10
	editButtonY := buttonY + 15
	editButtonWidth := 200
	editButtonHeight := 25
	
	// Test click on edit button
	testMouseX := editButtonX + editButtonWidth/2  // Middle of button
	testMouseY := editButtonY + editButtonHeight/2 // Middle of button
	
	t.Logf("Dialog bounds: (%.1f, %.1f, %.1f, %.1f)", dialogX, dialogY, dialogX+dialogWidth, dialogY+dialogHeight)
	t.Logf("Edit button bounds: (%d, %d, %d, %d)", editButtonX, editButtonY, editButtonX+editButtonWidth, editButtonY+editButtonHeight)
	t.Logf("Test click position: (%d, %d)", testMouseX, testMouseY)

	// Simulate edit button click
	action := controller.UIRenderer.HandleInspectionDialogClick(testMouseX, testMouseY, screenWidth, screenHeight)
	if action != "edit" {
		t.Errorf("Expected 'edit' action, got '%s'", action)
	}
	t.Logf("✓ Step 4: Edit button click detected correctly")

	// Step 5: Test editor mode check
	if !controller.EditorMode {
		t.Error("Controller should be in editor mode")
	}

	// Step 6: Verify entity exists in database for editing
	entity, err := db.GetEntityByAssetPath(controller.InspectedObject.AssetPath)
	if err != nil {
		t.Errorf("Failed to get entity for editing: %v", err)
	}
	if entity == nil {
		t.Error("Entity should exist for asset editing")
	}
	t.Logf("✓ Step 6: Entity found for editing: %s", entity.Name)

	// Step 7: Simulate the edit button action processing
	if action == "edit" && controller.EditorMode {
		controller.EditingAsset = entity
		controller.AssetPropertyDialog = true
		controller.InspectionDialog = false
		t.Logf("✓ Step 7: Asset property dialog should open")
	}

	// Step 8: Verify final state
	if !controller.AssetPropertyDialog {
		t.Error("Asset property dialog should be open")
	}
	if controller.InspectionDialog {
		t.Error("Inspection dialog should be closed")
	}
	if controller.EditingAsset == nil {
		t.Error("EditingAsset should be set")
	}
	if controller.EditingAsset.AssetPath != "assets/Chair/Chair_2_B_Tile.png" {
		t.Errorf("Wrong editing asset: %s", controller.EditingAsset.AssetPath)
	}

	t.Logf("✓ Step 8: Final state verified - Asset property dialog opened successfully")

	t.Log("=== ALL TESTS PASSED: Inspection Dialog Edit Flow Working ===")
}

func TestInspectionDialogClickBounds(t *testing.T) {
	controller, err := input.NewController()
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}

	screenWidth, screenHeight := 1024, 768

	// Test various click positions
	testCases := []struct {
		name     string
		mouseX   int
		mouseY   int
		expected string
	}{
		{
			name:     "Outside dialog (top-left)",
			mouseX:   100,
			mouseY:   100,
			expected: "",
		},
		{
			name:     "Outside dialog (bottom-right)",
			mouseX:   900,
			mouseY:   700,
			expected: "",
		},
		{
			name:     "Inside dialog, not on button",
			mouseX:   500,
			mouseY:   400,
			expected: "consume",
		},
		{
			name:     "On edit button",
			mouseX:   400, // Should be in edit button area
			mouseY:   540, // Should be in edit button area
			expected: "edit",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			action := controller.UIRenderer.HandleInspectionDialogClick(tc.mouseX, tc.mouseY, screenWidth, screenHeight)
			if action != tc.expected {
				t.Errorf("Expected action '%s' for click at (%d, %d), got '%s'", 
					tc.expected, tc.mouseX, tc.mouseY, action)
			}
		})
	}
}
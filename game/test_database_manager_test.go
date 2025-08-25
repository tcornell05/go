package main

import (
	"embed"
	"os"
	"testing"

	"github.com/tcornell05/go/game/internal/world"
	"github.com/tcornell05/go/game/pkg/assets"
)

//go:embed assets/*
var testDBAssets embed.FS

func init() {
	assets.GlobalAssets = testDBAssets
}

// TestTestDatabaseManager tests the test database manager functionality
func TestTestDatabaseManager(t *testing.T) {
	// Clean up any existing test database
	os.RemoveAll("test_data")
	defer os.RemoveAll("test_data")

	t.Log("=== PHASE 1: Create and seed test database ===")

	// Create test database manager
	tm, err := NewTestDatabaseManager()
	if err != nil {
		t.Fatalf("Failed to create test database manager: %v", err)
	}
	defer tm.Close()

	// Seed with test data
	err = tm.SeedTestData()
	if err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}

	t.Log("✅ Test database created and seeded successfully")

	t.Log("=== PHASE 2: Verify test scenarios ===")

	// Test Scenario 1: Basic Room (A1)
	roomA1, err := tm.GetTestRoom("A1")
	if err != nil {
		t.Errorf("Failed to load test room A1: %v", err)
	} else {
		if roomA1.Name != "Basic Test Room" {
			t.Errorf("A1 name mismatch: expected 'Basic Test Room', got '%s'", roomA1.Name)
		}
		if len(roomA1.Tiles) != 2 {
			t.Errorf("A1 should have 2 tiles, got %d", len(roomA1.Tiles))
		}
		t.Log("✅ Test scenario A1 (Basic Room) verified")
	}

	// Test Scenario 2: Complex Room (B2)
	roomB2, err := tm.GetTestRoom("B2")
	if err != nil {
		t.Errorf("Failed to load test room B2: %v", err)
	} else {
		if roomB2.Name != "Complex Test Room" {
			t.Errorf("B2 name mismatch: expected 'Complex Test Room', got '%s'", roomB2.Name)
		}
		if len(roomB2.Tiles) != 5 {
			t.Errorf("B2 should have 5 tiles, got %d", len(roomB2.Tiles))
		}
		// Verify custom walkability
		if roomB2.IsWalkable(1, 1) {
			t.Error("B2 (1,1) should be blocked")
		}
		if roomB2.IsWalkable(9, 9) {
			t.Error("B2 (9,9) should be blocked")
		}
		if !roomB2.IsWalkable(12, 12) {
			t.Error("B2 (12,12) should be walkable (expanded)")
		}
		t.Log("✅ Test scenario B2 (Complex Room) verified")
	}

	// Test Scenario 3: Public Room (C3)
	roomC3, err := tm.GetTestRoom("C3")
	if err != nil {
		t.Errorf("Failed to load test room C3: %v", err)
	} else {
		if roomC3.AccessLevel != world.AccessPublic {
			t.Errorf("C3 should be public access, got %v", roomC3.AccessLevel)
		}
		if !roomC3.Rentable {
			t.Error("C3 should be rentable")
		}
		if roomC3.RentalStatus != world.RentalAvailable {
			t.Errorf("C3 should be available for rent, got %v", roomC3.RentalStatus)
		}
		if roomC3.RentalPrice != 100 {
			t.Errorf("C3 rental price should be 100, got %d", roomC3.RentalPrice)
		}
		if len(roomC3.Tiles) != 4 {
			t.Errorf("C3 should have 4 chairs, got %d tiles", len(roomC3.Tiles))
		}
		t.Log("✅ Test scenario C3 (Public Room) verified")
	}

	// Test Scenario 4: Private Room (D4)
	roomD4, err := tm.GetTestRoom("D4")
	if err != nil {
		t.Errorf("Failed to load test room D4: %v", err)
	} else {
		if roomD4.AccessLevel != world.AccessPrivate {
			t.Errorf("D4 should be private access, got %v", roomD4.AccessLevel)
		}
		if roomD4.Owner == nil || *roomD4.Owner != "test_user" {
			t.Error("D4 should be owned by test_user")
		}
		if roomD4.Rentable {
			t.Error("D4 should not be rentable")
		}
		if len(roomD4.Tiles) != 3 {
			t.Errorf("D4 should have 3 furniture items, got %d tiles", len(roomD4.Tiles))
		}
		t.Log("✅ Test scenario D4 (Private Room) verified")
	}

	// Test Scenario 5: Empty Room (E5)
	roomE5, err := tm.GetTestRoom("E5")
	if err != nil {
		t.Errorf("Failed to load test room E5: %v", err)
	} else {
		if roomE5.Type != world.TileTypeEmpty {
			t.Errorf("E5 should be empty type, got %v", roomE5.Type)
		}
		if len(roomE5.Tiles) != 0 {
			t.Errorf("E5 should be empty, got %d tiles", len(roomE5.Tiles))
		}
		t.Log("✅ Test scenario E5 (Empty Room) verified")
	}

	t.Log("=== PHASE 3: Test database persistence ===")

	// Close and reopen to test persistence
	tm.Close()

	tm2, err := NewTestDatabaseManager()
	if err != nil {
		t.Fatalf("Failed to reconnect to test database: %v", err)
	}
	defer tm2.Close()

	// Verify data persisted
	roomA1Again, err := tm2.GetTestRoom("A1")
	if err != nil {
		t.Errorf("Failed to load A1 after reconnection: %v", err)
	} else {
		if roomA1Again.Name != "Basic Test Room" {
			t.Error("A1 data not persistent after reconnection")
		}
		t.Log("✅ Test database persistence verified")
	}

	t.Log("=== PHASE 4: Test database reset ===")

	// Test reset functionality
	err = tm2.ResetTestData()
	if err != nil {
		t.Errorf("Failed to reset test data: %v", err)
	}

	// Verify data is fresh after reset
	roomA1Fresh, err := tm2.GetTestRoom("A1")
	if err != nil {
		t.Errorf("Failed to load A1 after reset: %v", err)
	} else {
		if roomA1Fresh.Name != "Basic Test Room" {
			t.Error("A1 not properly restored after reset")
		}
		t.Log("✅ Test database reset verified")
	}

	t.Log("=== ALL TEST DATABASE MANAGER TESTS PASSED ===")
}

// TestTestDatabaseIntegrationUsage demonstrates how to use the test database in integration tests
func TestTestDatabaseIntegrationUsage(t *testing.T) {
	// Clean up any existing test database
	os.RemoveAll("test_data")
	defer os.RemoveAll("test_data")

	tm, err := NewTestDatabaseManager()
	if err != nil {
		t.Fatalf("Failed to create test database manager: %v", err)
	}
	defer tm.Close()

	// Seed test data
	err = tm.SeedTestData()
	if err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}

	t.Log("=== Integration Test Example: Room Access Control ===")

	// Test public room access
	publicRoom, err := tm.GetTestRoom("C3")
	if err != nil {
		t.Fatalf("Failed to load public room: %v", err)
	}

	// Anyone should be able to access public room
	if !publicRoom.IsAccessibleBy("any_user", false) {
		t.Error("Public room should be accessible by anyone")
	}

	// Test private room access
	privateRoom, err := tm.GetTestRoom("D4")
	if err != nil {
		t.Fatalf("Failed to load private room: %v", err)
	}

	// Only owner should access private room
	if !privateRoom.IsAccessibleBy("test_user", false) {
		t.Error("Private room should be accessible by owner")
	}
	if privateRoom.IsAccessibleBy("other_user", false) {
		t.Error("Private room should not be accessible by non-owner")
	}

	t.Log("✅ Integration test example completed successfully")

	t.Log("=== Integration Test Example: Content Modification ===")

	// Get empty room for content addition test
	emptyRoom, err := tm.GetTestRoom("E5")
	if err != nil {
		t.Fatalf("Failed to load empty room: %v", err)
	}

	// Add furniture to empty room
	newChair := world.TileData{
		AssetPath: "assets/Chair/Chair_2_A_Tile.png",
		X:         5,
		Y:         5,
		Layer:     2,
		Category:  "furniture",
		OffsetX:   0.5,
		OffsetY:   0.5,
	}

	err = emptyRoom.AddTile(newChair)
	if err != nil {
		t.Errorf("Failed to add tile to empty room: %v", err)
	}

	// Save modified room back to database
	db := tm.GetDatabase()
	err = db.SaveWorldTile(emptyRoom)
	if err != nil {
		t.Errorf("Failed to save modified room: %v", err)
	}

	// Reload and verify modification persisted
	modifiedRoom, err := tm.GetTestRoom("E5")
	if err != nil {
		t.Errorf("Failed to reload modified room: %v", err)
	}

	if len(modifiedRoom.Tiles) != 1 {
		t.Errorf("Modified room should have 1 tile, got %d", len(modifiedRoom.Tiles))
	}

	t.Log("✅ Content modification test completed successfully")
}
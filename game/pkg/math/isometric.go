package math

// CartesianToIso converts grid coordinates to isometric screen coordinates
// Returns the position where sprites should be anchored (bottom-center of tile)
func CartesianToIso(x, y float64) (float64, float64) {
	// Standard isometric projection: 2:1 ratio
	// For 64x64 tiles in isometric view:
	// - Diamond width spans 64 pixels horizontally
	// - Diamond height spans 32 pixels vertically  
	ix := (x - y) * 32 // Half of tile width (64/2)
	iy := (x + y) * 16 // Half of diamond height (32/2)
	
	// Adjust to bottom-center of tile for proper sprite anchoring
	// This is where floor tiles should align and furniture should sit
	iy += 16 // Move down to bottom of diamond
	
	return ix, iy
}

// CartesianToIsoZoomed converts cartesian coordinates to isometric display coordinates with zoom scaling
func CartesianToIsoZoomed(x, y, zoomLevel float64) (float64, float64) {
	// Scale the base isometric projection by zoom level
	ix := (x - y) * 32 * zoomLevel // Scaled by zoom
	iy := (x + y) * 16 * zoomLevel // Scaled by zoom
	
	// The positioning offset should NOT be scaled - it's a fixed anchor point
	iy += 16 // Fixed offset for proper sprite anchoring
	
	return ix, iy
}

// ScreenToIso converts screen coordinates to isometric grid coordinates
func ScreenToIso(screenX, screenY, camX, camY, zoomLevel, screenWidth, screenHeight float64) (int, int) {
	// Convert screen coordinates to isometric space relative to camera
	isoX := screenX - float64(screenWidth/2) + camX
	isoY := screenY - float64(screenHeight/2) + camY
	
	// Apply zoom level correction - when zoomed, isometric coordinates are scaled
	// so we need to divide by zoom to get back to original coordinate space
	isoX /= zoomLevel
	isoY /= zoomLevel
	
	// For center-based click detection, we need to account for the diamond center
	// The CartesianToIso function adds +16 to position sprites at the bottom
	// For click detection, we want the center of the diamond, so we subtract 8
	// This centers the detection on the diamond instead of putting it at SE corner
	adjustedIsoY := isoY - 8
	
	// Convert isometric coordinates back to grid coordinates
	// Inverse of CartesianToIso transformation
	gridX := (isoX/32 + adjustedIsoY/16) / 2
	gridY := (adjustedIsoY/16 - isoX/32) / 2
	
	// Add 0.5 for proper rounding to nearest integer
	return int(gridX + 0.5), int(gridY + 0.5)
}

// IsPointInFloorArea checks if grid coordinates are within the floor area
func IsPointInFloorArea(gridX, gridY int) bool {
	return gridX >= 0 && gridX <= 10 && gridY >= 0 && gridY <= 10
}
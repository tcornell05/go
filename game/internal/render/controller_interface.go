package render

// ControllerInterface defines the minimal interface that the renderer needs from a controller
// This breaks the circular dependency between render -> input -> editor
type ControllerInterface interface {
	// Debug and display properties
	GetShowDebugGrid() bool
	GetIsHoveringTile() bool
	GetHoveredTile() (int, int)
	GetZoomLevel() float64
	
	// Editor state
	GetIsEditorMode() bool
	GetTileInventory() interface{}
	
	// Inspection system
	GetHasInspectionDialog() bool
	GetInspectedObject() interface{}
	GetHasAssetPropertyDialog() bool
	GetEditingAsset() interface{}
	GetIsDraggingAssetPreview() bool
}

// TileInventoryInterface defines the minimal tile inventory interface
type TileInventoryInterface interface {
	GetSelectedTileInfo() (string, int)
	GetSelectedTile() interface{}
}

// UIRendererInterface defines the minimal UI renderer interface
type UIRendererInterface interface {
	DrawToolbar(screen interface{}, editorMode bool, showGrid bool, screenWidth, screenHeight int)
	DrawSidebar(screen interface{}, screenWidth, screenHeight int)
	DrawPositionDialog(screen interface{}, screenWidth, screenHeight int)
	DrawInspectionDialog(screen interface{}, obj interface{}, screenWidth, screenHeight int)
	DrawAssetPropertyEditor(screen interface{}, asset interface{}, dragging bool, zoomLevel float64, screenWidth, screenHeight int)
}
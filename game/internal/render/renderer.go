package render

import (
	"fmt"
	"image/color"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tcornell05/go/game/internal/entity"
	"github.com/tcornell05/go/game/internal/room"
	"github.com/tcornell05/go/game/internal/world"
	gamemath "github.com/tcornell05/go/game/pkg/math"
)

// WalkabilityChecker interface for objects that can check if tiles are walkable
type WalkabilityChecker interface {
	IsWalkable(x, y int) bool
}

type Renderer struct {
	FloorTile  *ebiten.Image
	WallTile1  *ebiten.Image
	WallTile2  *ebiten.Image
	frameCount int // For animations
}

func NewRenderer(floorTile, wallTile1, wallTile2 *ebiten.Image) *Renderer {
	return &Renderer{
		FloorTile:  floorTile,
		WallTile1:  wallTile1,
		WallTile2:  wallTile2,
		frameCount: 0,
	}
}

func (r *Renderer) Draw(screen *ebiten.Image, player *entity.Player, controller ControllerInterface, gameRoom *room.Room, camX, camY float64, screenWidth, screenHeight int) {
	screen.Fill(color.RGBA{32, 32, 32, 255}) // Dark background

	// Increment frame counter for animations
	r.frameCount++
	
	// Get zoom level for scaling
	zoomLevel := controller.GetZoomLevel()

	// Define room dimensions
	const roomSize = 11 // 11x11 tiles

	// Draw debug grid if enabled (F2)
	if controller.GetShowDebugGrid() {
		r.drawDebugGrid(screen, gameRoom, controller.GetIsEditorMode(), camX, camY, zoomLevel, screenWidth, screenHeight)
	}

	// Draw grid outline for building reference - now dynamic based on walkability
	r.drawGrid(screen, gameRoom, controller.GetIsEditorMode(), camX, camY, zoomLevel, screenWidth, screenHeight)

	// Draw floor tiles (only placed ones)
	r.drawFloor(screen, roomSize, gameRoom, camX, camY, zoomLevel, screenWidth, screenHeight)

	// Draw placed room tiles (layer by layer)
	r.drawRoomTiles(screen, gameRoom, camX, camY, zoomLevel, screenWidth, screenHeight)

	// Draw hover highlight
	if controller.GetIsHoveringTile() {
		x, y := controller.GetHoveredTile()
		r.drawTileHighlight(screen, x, y, camX, camY, zoomLevel, screenWidth, screenHeight)
	}

	// Draw player
	r.drawPlayer(screen, player, camX, camY, zoomLevel, screenWidth, screenHeight)

	// Draw debug info
	r.drawDebugInfo(screen, player, controller, camX, camY)
}

func (r *Renderer) drawFloor(screen *ebiten.Image, roomSize int, gameRoom *room.Room, camX, camY, zoomLevel float64, screenWidth, screenHeight int) {
	// Draw only placed floor tiles (no default floor)
	for x := 0; x < roomSize; x++ {
		for y := 0; y < roomSize; y++ {
			// Only draw if there's a placed floor tile at this position
			if placedTile, exists := gameRoom.GetTile(0, x, y); exists {
				if tileImage := gameRoom.LoadedImages[placedTile.AssetPath]; tileImage != nil {
					r.drawPlacedTile(screen, &placedTile, tileImage, camX, camY, zoomLevel, screenWidth, screenHeight)
				}
			}
		}
	}
}

func (r *Renderer) drawGrid(screen *ebiten.Image, gameRoom interface{}, isEditorMode bool, camX, camY, zoomLevel float64, screenWidth, screenHeight int) {
	// Try to cast to WorldTile first, then fallback to Room for compatibility
	var walkabilityChecker WalkabilityChecker
	if worldTile, ok := gameRoom.(*world.WorldTile); ok {
		walkabilityChecker = worldTile
	} else if roomObj, ok := gameRoom.(*room.Room); ok {
		walkabilityChecker = roomObj
	} else {
		// Unknown type, fall back to default 11x11 grid
		r.drawDefaultGrid(screen, 11, camX, camY, zoomLevel, screenWidth, screenHeight)
		return
	}
	
	gridColor := color.RGBA{80, 80, 80, 100} // Subtle gray grid lines
	
	// Calculate the range of grid coordinates that could be visible on screen
	const gridRange = 30 // Check from -30 to +30 in both directions

	for x := -gridRange; x <= gridRange; x++ {
		for y := -gridRange; y <= gridRange; y++ {
			gridX := float64(x)
			gridY := float64(y)

			isoX, isoY := gamemath.CartesianToIsoZoomed(gridX, gridY, zoomLevel)
			screenX := isoX + float64(screenWidth/2) - camX
			screenY := isoY + float64(screenHeight/2) - camY

			// Skip if off-screen
			if screenX < -100 || screenX > float64(screenWidth)+100 ||
				screenY < -100 || screenY > float64(screenHeight)+100 {
				continue
			}

			// In editor mode, show ALL tiles within reasonable bounds with different colors
			// In non-editor mode, only show walkable tiles
			if isEditorMode {
				// In editor mode, show a larger grid area for object placement
				if x >= -20 && x <= 20 && y >= -20 && y <= 20 {
					// Use different colors based on walkability in edit mode
					var tileColor color.Color
					if walkabilityChecker.IsWalkable(x, y) {
						tileColor = color.RGBA{80, 80, 80, 100}   // Grey for walkable tiles
					} else {
						tileColor = color.RGBA{0, 100, 255, 100} // Blue for non-walkable tiles
					}
					r.drawGridDiamond(screen, screenX, screenY, zoomLevel, tileColor)
				}
			} else {
				// In non-editor mode, only show walkable tiles
				if walkabilityChecker.IsWalkable(x, y) {
					r.drawGridDiamond(screen, screenX, screenY, zoomLevel, gridColor)
				}
			}
		}
	}
}

// drawDefaultGrid draws a fixed-size grid (fallback for unknown game room types)
func (r *Renderer) drawDefaultGrid(screen *ebiten.Image, roomSize int, camX, camY, zoomLevel float64, screenWidth, screenHeight int) {
	gridColor := color.RGBA{80, 80, 80, 100} // Subtle gray grid lines

	for x := 0; x < roomSize; x++ {
		for y := 0; y < roomSize; y++ {
			gridX := float64(x)
			gridY := float64(y)

			isoX, isoY := gamemath.CartesianToIsoZoomed(gridX, gridY, zoomLevel)
			screenX := isoX + float64(screenWidth/2) - camX
			screenY := isoY + float64(screenHeight/2) - camY

			// Draw diamond outline for each grid cell
			r.drawGridDiamond(screen, screenX, screenY, zoomLevel, gridColor)
		}
	}
}

func (r *Renderer) drawGridDiamond(screen *ebiten.Image, centerX, centerY, zoomLevel float64, gridColor color.Color) {
	// Draw diamond outline for grid cell scaled by zoom level
	tileHalfWidth := 32 * zoomLevel
	tileHalfHeight := 16 * zoomLevel
	points := []struct{ x, y float64 }{
		{centerX, centerY - tileHalfHeight}, // Top
		{centerX + tileHalfWidth, centerY}, // Right
		{centerX, centerY + tileHalfHeight}, // Bottom
		{centerX - tileHalfWidth, centerY}, // Left
		{centerX, centerY - tileHalfHeight}, // Back to top
	}

	// Draw lines between consecutive points
	for i := 0; i < len(points)-1; i++ {
		r.drawLine(screen, points[i].x, points[i].y, points[i+1].x, points[i+1].y, gridColor)
	}
}

func (r *Renderer) drawRoomTiles(screen *ebiten.Image, gameRoom *room.Room, camX, camY, zoomLevel float64, screenWidth, screenHeight int) {
	// Draw tiles layer by layer (1=walls, 2=furniture, 3=decorations)
	// Floor tiles (layer 0) are handled by drawFloor separately
	for layer := 1; layer <= 3; layer++ {
		// Collect all tiles for this layer
		var layerTiles []room.TileData
		for _, tile := range gameRoom.Tiles {
			if tile.Layer == layer {
				layerTiles = append(layerTiles, tile)
			}
		}
		
		// Sort tiles by isometric depth within each layer using standard isometric sorting
		// Standard isometric back-to-front: Y first (back to front), then X (left to right)
		sort.Slice(layerTiles, func(i, j int) bool {
			tile1 := layerTiles[i]
			tile2 := layerTiles[j]
			
			// Apply position offsets for more precise sorting
			y1 := float64(tile1.Y) + tile1.OffsetY
			y2 := float64(tile2.Y) + tile2.OffsetY
			x1 := float64(tile1.X) + tile1.OffsetX
			x2 := float64(tile2.X) + tile2.OffsetX
			
			// Primary sort: Y coordinate (objects with lower Y drawn first - back to front)
			if y1 != y2 {
				return y1 < y2
			}
			
			// Secondary sort: X coordinate (objects with lower X drawn first - left to right) 
			if x1 != x2 {
				return x1 < x2
			}
			
			// Tertiary sort: Depth offset for fine-tuning
			return tile1.Depth < tile2.Depth
		})
		
		// Draw sorted tiles
		for _, tile := range layerTiles {
			r.drawPlacedTile(screen, &tile, gameRoom.LoadedImages[tile.AssetPath], camX, camY, zoomLevel, screenWidth, screenHeight)
		}
	}
}

func (r *Renderer) drawPlacedTile(screen *ebiten.Image, tile *room.TileData, tileImage *ebiten.Image, camX, camY, zoomLevel float64, screenWidth, screenHeight int) {
	if tileImage == nil {
		return
	}

	// Standard isometric positioning: (0,0) offset = northwest corner, (1,1) = southeast corner
	// Tile coordinates are integers, offsets are 0-1 within the tile
	// Calculate exact position by adding tile coordinate + offset within tile
	exactX := float64(tile.X) + tile.OffsetX
	exactY := float64(tile.Y) + tile.OffsetY
	
	isoX, isoY := gamemath.CartesianToIsoZoomed(exactX, exactY, zoomLevel)

	// Apply depth offset (vertical positioning)
	depthPixelY := tile.Depth * 20 // Fixed depth offset - do not scale with zoom

	screenX := isoX + float64(screenWidth/2) - camX
	screenY := isoY + float64(screenHeight/2) - camY + depthPixelY
	

	op := &ebiten.DrawImageOptions{}
	
	// Get image dimensions for transformations
	bounds := tileImage.Bounds()
	spriteWidth := float64(bounds.Dx())
	spriteHeight := float64(bounds.Dy())

	// Apply horizontal flip if Facing is 1
	// For isometric assets, we only need normal (0) and flipped (1)
	if tile.Facing == 1 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(spriteWidth, 0)
	}

	// Note: Do NOT apply zoom scaling here - CartesianToIsoZoomed already handles zoom
	// Additional sprite scaling would cause double-scaling and sync issues
	
	// Clean positioning: CartesianToIso now returns bottom-center anchor point
	// Center sprite horizontally, bottom of sprite sits on tile surface
	// Use unscaled sprite dimensions for centering since zoom is applied after positioning
	op.GeoM.Translate(screenX-(spriteWidth/2), screenY-spriteHeight)

	screen.DrawImage(tileImage, op)
}

func (r *Renderer) drawWalls(screen *ebiten.Image, roomSize int, camX, camY float64, screenWidth, screenHeight int) {
	// Left wall - runs along the left edge of the floor area, aligned with grid
	for y := 0; y < roomSize; y++ {
		gridX := -0.65             // Shift slightly more northwest (west)
		gridY := float64(y) - 0.65 // Shift slightly more northwest (north)

		isoX, isoY := gamemath.CartesianToIso(gridX, gridY)
		screenX := isoX + float64(screenWidth/2) - camX
		screenY := isoY + float64(screenHeight/2) - camY

		wallOp := &ebiten.DrawImageOptions{}
		// Use same positioning as tiles for consistency
		bounds := r.WallTile1.Bounds()
		spriteWidth := float64(bounds.Dx())
		spriteHeight := float64(bounds.Dy())
		wallOp.GeoM.Translate(screenX-(spriteWidth/2), screenY-spriteHeight)
		screen.DrawImage(r.WallTile1, wallOp)
	}

	// Back/Top wall - runs along the top edge of the floor area, aligned with grid
	for x := 0; x < roomSize; x++ {
		gridX := float64(x) - 0.5  // Position exactly on the grid line
		gridY := -0.5              // Position exactly on the top edge of grid

		isoX, isoY := gamemath.CartesianToIso(gridX, gridY)
		screenX := isoX + float64(screenWidth/2) - camX
		screenY := isoY + float64(screenHeight/2) - camY

		wallOp := &ebiten.DrawImageOptions{}
		// Use same positioning as tiles for consistency
		bounds := r.WallTile2.Bounds()
		spriteWidth := float64(bounds.Dx())
		spriteHeight := float64(bounds.Dy())
		wallOp.GeoM.Translate(screenX-(spriteWidth/2), screenY-spriteHeight)
		screen.DrawImage(r.WallTile2, wallOp)
	}
}

func (r *Renderer) drawPlayer(screen *ebiten.Image, player *entity.Player, camX, camY, zoomLevel float64, screenWidth, screenHeight int) {
	// Use exact player coordinates without offset - debug should match visual
	playerIsoX, playerIsoY := gamemath.CartesianToIsoZoomed(player.X, player.Y, zoomLevel)
	
	playerScreenX := playerIsoX + float64(screenWidth/2) - camX
	playerScreenY := playerIsoY + float64(screenHeight/2) - camY

	// Get the appropriate sprite based on movement state and facing direction
	var sprite *ebiten.Image
	if player.IsWalking {
		// 4-frame animation cycle: static -> walking -> static -> walking
		if (player.WalkAnimFrame == 1 || player.WalkAnimFrame == 3) && player.WalkingSprites[player.Facing] != nil {
			sprite = player.WalkingSprites[player.Facing] // Frames 1 & 3: walking pose
		} else if player.StaticSprites[player.Facing] != nil {
			sprite = player.StaticSprites[player.Facing] // Frames 0 & 2: static pose
		}
	} else if player.StaticSprites[player.Facing] != nil {
		sprite = player.StaticSprites[player.Facing]
	}

	if sprite == nil {
		return // No sprite available
	}

	playerOp := &ebiten.DrawImageOptions{}
	
	// Note: Do NOT apply zoom scaling - CartesianToIsoZoomed already handles zoom
	
	// Get sprite bounds to center properly
	bounds := sprite.Bounds()
	spriteWidth := float64(bounds.Dx())
	spriteHeight := float64(bounds.Dy())

	// Center the sprite horizontally and position base on tile
	playerOp.GeoM.Translate(
		playerScreenX-spriteWidth/2,
		playerScreenY-spriteHeight+16, // Fixed offset - base sits on tile
	)
	screen.DrawImage(sprite, playerOp)
}

func (r *Renderer) drawTileHighlight(screen *ebiten.Image, tileX, tileY int, camX, camY, zoomLevel float64, screenWidth, screenHeight int) {
	isoX, isoY := gamemath.CartesianToIsoZoomed(float64(tileX), float64(tileY), zoomLevel)
	screenX := isoX + float64(screenWidth/2) - camX
	screenY := isoY + float64(screenHeight/2) - camY

	// Draw highlight border around the tile
	r.drawTileHighlightBorder(screen, screenX, screenY, zoomLevel)
}

func (r *Renderer) drawTileHighlightBorder(screen *ebiten.Image, x, y, zoomLevel float64) {
	// Define the diamond shape of an isometric tile (64x64) scaled by zoom
	// The tile is centered at the provided coordinates
	centerX := x
	centerY := y

	// Isometric tile corners scaled by zoom level
	tileHalfWidth := 32 * zoomLevel
	tileHalfHeight := 16 * zoomLevel
	topX, topY := centerX, centerY-tileHalfHeight       // Top point
	rightX, rightY := centerX+tileHalfWidth, centerY   // Right point
	bottomX, bottomY := centerX, centerY+tileHalfHeight // Bottom point
	leftX, leftY := centerX-tileHalfWidth, centerY     // Left point

	// Draw outline using vector graphics
	points := []float64{
		topX, topY,
		rightX, rightY,
		bottomX, bottomY,
		leftX, leftY,
		topX, topY, // Close the shape
	}

	// Draw the outline by drawing lines between consecutive points
	highlightColor := color.RGBA{255, 255, 0, 180} // Yellow with transparency
	for i := 0; i < len(points)-2; i += 2 {
		r.drawLine(screen, points[i], points[i+1], points[i+2], points[i+3], highlightColor)
	}
}

func (r *Renderer) drawLine(screen *ebiten.Image, x1, y1, x2, y2 float64, c color.Color) {
	// Simple line drawing using Bresenham-like algorithm
	dx := x2 - x1
	dy := y2 - y1
	steps := int(math.Max(math.Abs(dx), math.Abs(dy)))

	if steps == 0 {
		return
	}

	xInc := dx / float64(steps)
	yInc := dy / float64(steps)

	x, y := x1, y1
	for i := 0; i <= steps; i++ {
		if int(x) >= 0 && int(x) < screen.Bounds().Dx() && int(y) >= 0 && int(y) < screen.Bounds().Dy() {
			screen.Set(int(x), int(y), c)
		}
		x += xInc
		y += yInc
	}
}

func (r *Renderer) drawDebugGrid(screen *ebiten.Image, gameRoom interface{}, isEditorMode bool, camX, camY, zoomLevel float64, screenWidth, screenHeight int) {
	// Try to cast to WorldTile first, then fallback to Room for compatibility
	var walkabilityChecker WalkabilityChecker
	if worldTile, ok := gameRoom.(*world.WorldTile); ok {
		walkabilityChecker = worldTile
	} else if roomObj, ok := gameRoom.(*room.Room); ok {
		walkabilityChecker = roomObj
	} else {
		// Unknown type, can't check walkability
		return
	}
	
	// Define colors
	blueColor := color.RGBA{0, 100, 255, 150}       // Blue for blocked tiles
	greyColor := color.RGBA{100, 100, 100, 150}    // Grey for walkable tiles

	// Calculate the range of grid coordinates that could be visible on screen
	const gridRange = 250 // Draw from -250 to +250 to cover full world tile area

	for x := -gridRange; x <= gridRange; x++ {
		for y := -gridRange; y <= gridRange; y++ {
			gridX := float64(x)
			gridY := float64(y)

			isoX, isoY := gamemath.CartesianToIsoZoomed(gridX, gridY, zoomLevel)
			screenX := isoX + float64(screenWidth/2) - camX
			screenY := isoY + float64(screenHeight/2) - camY

			// Only draw if the diamond is potentially visible on screen
			if screenX > -100 && screenX < float64(screenWidth)+100 &&
				screenY > -100 && screenY < float64(screenHeight)+100 {
				
				// Use different colors in editor mode based on walkability
				if isEditorMode {
					if walkabilityChecker.IsWalkable(x, y) {
						r.drawDebugDiamondWithWalkability(screen, screenX, screenY, gridX, gridY, zoomLevel, greyColor, true)
					} else {
						r.drawDebugDiamondWithWalkability(screen, screenX, screenY, gridX, gridY, zoomLevel, blueColor, false)
					}
				} else {
					// Non-editor mode: just show blue grid
					r.drawDebugDiamond(screen, screenX, screenY, gridX, gridY, zoomLevel, blueColor)
				}
			}
		}
	}
}

func (r *Renderer) drawDebugDiamond(screen *ebiten.Image, centerX, centerY, gridX, gridY, zoomLevel float64, c color.Color) {
	// Draw diamond outline scaled by zoom level
	tileHalfWidth := 32 * zoomLevel
	tileHalfHeight := 16 * zoomLevel
	points := []struct{ x, y float64 }{
		{centerX, centerY - tileHalfHeight}, // Top
		{centerX + tileHalfWidth, centerY}, // Right
		{centerX, centerY + tileHalfHeight}, // Bottom
		{centerX - tileHalfWidth, centerY}, // Left
		{centerX, centerY - tileHalfHeight}, // Back to top
	}

	// Draw lines between consecutive points
	for i := 0; i < len(points)-1; i++ {
		r.drawLine(screen, points[i].x, points[i].y, points[i+1].x, points[i+1].y, c)
	}

	// Draw coordinate labels for key grid positions
	if int(gridX) >= 0 && int(gridY) >= 0 && int(gridX) <= 10 && int(gridY) <= 10 {
		// Draw a small cross at the center scaled by zoom
		crossSize := 2 * zoomLevel
		r.drawLine(screen, centerX-crossSize, centerY, centerX+crossSize, centerY, color.RGBA{255, 255, 0, 255})
		r.drawLine(screen, centerX, centerY-crossSize, centerX, centerY+crossSize, color.RGBA{255, 255, 0, 255})
	}
}

func (r *Renderer) drawDebugDiamondWithWalkability(screen *ebiten.Image, centerX, centerY, gridX, gridY, zoomLevel float64, c color.Color, isWalkable bool) {
	// Draw diamond outline scaled by zoom level
	tileHalfWidth := 32 * zoomLevel
	tileHalfHeight := 16 * zoomLevel
	points := []struct{ x, y float64 }{
		{centerX, centerY - tileHalfHeight}, // Top
		{centerX + tileHalfWidth, centerY}, // Right
		{centerX, centerY + tileHalfHeight}, // Bottom
		{centerX - tileHalfWidth, centerY}, // Left
		{centerX, centerY - tileHalfHeight}, // Back to top
	}

	// Draw lines between consecutive points
	for i := 0; i < len(points)-1; i++ {
		r.drawLine(screen, points[i].x, points[i].y, points[i+1].x, points[i+1].y, c)
	}

	// Draw the small + for ALL walkable tiles
	if isWalkable {
		// Draw a small cross at the center scaled by zoom
		crossSize := 2 * zoomLevel
		r.drawLine(screen, centerX-crossSize, centerY, centerX+crossSize, centerY, color.RGBA{255, 255, 0, 255})
		r.drawLine(screen, centerX, centerY-crossSize, centerX, centerY+crossSize, color.RGBA{255, 255, 0, 255})
	}
}

func (r *Renderer) drawDebugInfo(screen *ebiten.Image, player *entity.Player, controller ControllerInterface, camX, camY float64) {
	modeText := "Play Mode"
	controls := "WASD to move | QEZX for diagonals | Click to move | ` toggle movement | R rotate | E toggle editor | F2 debug grid | 1-9 set level"

	if controller.GetIsEditorMode() {
		modeText = "EDITOR MODE"
		controls = "E toggle editor | Tab toggle sidebar | P position dialog | R rotate tile | I or middle-click inspect | Inspect button | Click tiles to select"

		// TODO: Fix tile inventory interface access
		// inventory := controller.GetTileInventory()
		// if inventory != nil && inventory.GetSelectedTile() != nil {
		// 	tileName, rotation := inventory.GetSelectedTileInfo()
		// 	rotText := "Normal"
		// 	if rotation == 1 {
		// 		rotText = "Flipped"
		// 	}
		// 	controls += fmt.Sprintf(" | Selected: %s (%s)", tileName, rotText)
		// }
	}

	walkingText := "Static"
	if player.IsWalking {
		walkingText = fmt.Sprintf("Walking(F%d)", player.WalkAnimFrame)
	}
	// Calculate FPS (using ebiten's built-in TPS which runs at 60)
	fps := ebiten.ActualFPS()
	
	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %.1f | TPS: %.1f | Zoom: %.1fx\n%s\n%s - Player: (%.1f, %.1f) Facing: %s(%d) %s Level: %d\nDirection: %s | Camera: (%.1f, %.1f)",
		fps, ebiten.ActualTPS(), controller.GetZoomLevel(), controls, modeText, player.X, player.Y, player.GetFacingName(), player.Facing, walkingText, player.Level, player.CurrentDirection, camX, camY))

	// Draw the compass
	r.drawCompass(screen, player)
}

func (r *Renderer) drawCompass(screen *ebiten.Image, player *entity.Player) {
	// Position the compass on the isometric plane near the player
	// Place it at grid position (13, 5) to the right of the play area
	compassGridX := 13.0
	compassGridY := 5.0

	// Convert to isometric screen coordinates
	centerIsoX, centerIsoY := gamemath.CartesianToIso(compassGridX, compassGridY)
	compassX := centerIsoX + float64(screen.Bounds().Dx()/2) - 0 // Using camera at 0,0 for fixed position
	compassY := centerIsoY + float64(screen.Bounds().Dy()/2) - 0

	// Direction data - matching isometric grid directions
	directions := []struct {
		label  string
		gridDX float64 // Grid direction X
		gridDY float64 // Grid direction Y
		index  int     // Sprite index matching player facing
	}{
		{"S", 0, 1, 0},    // South - down in grid
		{"SW", -1, 1, 1},  // Southwest
		{"W", -1, 0, 2},   // West - left in grid
		{"NW", -1, -1, 3}, // Northwest
		{"N", 0, -1, 4},   // North - up in grid
		{"NE", 1, -1, 5},  // Northeast
		{"E", 1, 0, 6},    // East - right in grid
		{"SE", 1, 1, 7},   // Southeast
	}

	// Draw all direction lines
	for _, dir := range directions {
		// Calculate end point in grid space
		endGridX := compassGridX + dir.gridDX*1.5
		endGridY := compassGridY + dir.gridDY*1.5

		// Convert to isometric coordinates
		endIsoX, endIsoY := gamemath.CartesianToIso(endGridX, endGridY)
		endScreenX := endIsoX + float64(screen.Bounds().Dx()/2)
		endScreenY := endIsoY + float64(screen.Bounds().Dy()/2)

		// Choose color based on if this is the active direction
		lineColor := color.RGBA{100, 100, 100, 255} // Gray for inactive
		if dir.index == player.Facing {
			lineColor = color.RGBA{255, 255, 0, 255} // Yellow for active
		}

		// Draw the direction line
		r.drawLine(screen, compassX, compassY, endScreenX, endScreenY, lineColor)

		// Draw direction label at the end
		labelOffset := 10.0
		labelX := endScreenX + dir.gridDX*labelOffset
		labelY := endScreenY + dir.gridDY*labelOffset*0.5 // Half for isometric
		ebitenutil.DebugPrintAt(screen, dir.label, int(labelX)-5, int(labelY)-5)
	}

	// Draw UP/DN indicators above and below center
	upX, upY := gamemath.CartesianToIso(compassGridX, compassGridY-0.5)
	dnX, dnY := gamemath.CartesianToIso(compassGridX, compassGridY+0.5)
	ebitenutil.DebugPrintAt(screen, "UP", int(upX+float64(screen.Bounds().Dx()/2))-10, int(upY+float64(screen.Bounds().Dy()/2))-20)
	ebitenutil.DebugPrintAt(screen, "DN", int(dnX+float64(screen.Bounds().Dx()/2))-10, int(dnY+float64(screen.Bounds().Dy()/2))+10)

	// Draw active direction arrow that moves based on level
	for _, dir := range directions {
		if dir.index == player.Facing {
			// Calculate arrow base position with level offset
			levelOffset := (float64(player.Level) - 5) * 2 // Vertical offset based on level

			// Add oscillation animation
			time := float64(r.frameCount) * 0.05
			oscillation := math.Sin(time) * 1.5

			// Calculate arrow start with vertical offset
			arrowStartY := compassY + levelOffset + oscillation

			// Calculate arrow end point
			endGridX := compassGridX + dir.gridDX*1.2
			endGridY := compassGridY + dir.gridDY*1.2
			endIsoX, endIsoY := gamemath.CartesianToIso(endGridX, endGridY)
			endScreenX := endIsoX + float64(screen.Bounds().Dx()/2)
			endScreenY := endIsoY + float64(screen.Bounds().Dy()/2) + levelOffset + oscillation

			// Draw the active arrow (thicker and red)
			r.drawArrow(screen, compassX, arrowStartY, endScreenX, endScreenY, color.RGBA{255, 50, 50, 255}, 2)
			break
		}
	}

	// Draw center point
	r.drawCircle(screen, compassX, compassY, 2, color.RGBA{255, 255, 255, 255})
}

func (r *Renderer) drawCircle(screen *ebiten.Image, centerX, centerY, radius float64, c color.Color) {
	// Draw a filled circle using simple algorithm
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if x*x+y*y <= radius*radius {
				px := int(centerX + x)
				py := int(centerY + y)
				if px >= 0 && px < screen.Bounds().Dx() && py >= 0 && py < screen.Bounds().Dy() {
					screen.Set(px, py, c)
				}
			}
		}
	}
}

func (r *Renderer) drawText(screen *ebiten.Image, text string, x, y float64, c color.Color) {
	// Simple text drawing (using debug print for now)
	// In a real implementation, you'd use a proper font
	ebitenutil.DebugPrintAt(screen, text, int(x), int(y))
}

func (r *Renderer) drawArrow(screen *ebiten.Image, x1, y1, x2, y2 float64, c color.Color, thickness float64) {
	// Draw main line with thickness
	for t := -thickness / 2; t <= thickness/2; t++ {
		r.drawLine(screen, x1+t, y1, x2+t, y2, c)
		r.drawLine(screen, x1, y1+t, x2, y2+t, c)
	}

	// Calculate arrowhead
	angle := math.Atan2(y2-y1, x2-x1)
	arrowLength := 10.0
	arrowAngle := 0.5

	// Draw arrowhead lines
	leftX := x2 - arrowLength*math.Cos(angle-arrowAngle)
	leftY := y2 - arrowLength*math.Sin(angle-arrowAngle)
	rightX := x2 - arrowLength*math.Cos(angle+arrowAngle)
	rightY := y2 - arrowLength*math.Sin(angle+arrowAngle)

	for t := 0.0; t <= thickness; t++ {
		r.drawLine(screen, x2, y2, leftX, leftY, c)
		r.drawLine(screen, x2, y2, rightX, rightY, c)
	}
}



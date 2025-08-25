package entity

import (
	"fmt"
	"image"
	"image/color"
	"math"
	
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tcornell05/go/game/pkg/assets"
)

type Player struct {
	X, Y             float64
	Speed            float64
	StaticSprites    [8]*ebiten.Image // 8 directional static sprites
	WalkingSprites   [8]*ebiten.Image // 8 directional walking sprites (first frame of APNG)
	CurrentDirection string
	Facing           int // 0=S, 1=SW, 2=W, 3=NW, 4=N, 5=NE, 6=E, 7=SE
	IsWalking        bool // Whether character is currently moving
	Level            int  // Player level for debugging display
	
	// Animation variables for grid movement
	TargetX, TargetY float64
	Animating        bool
	AnimationSpeed   float64
	
	// Walking animation frame tracking
	WalkAnimTimer    float64 // Timer for frame switching
	WalkAnimFrame    int     // Current animation frame (0-3 for 4-frame cycle)
}

func NewPlayer(startX, startY float64) (*Player, error) {
	var staticSprites [8]*ebiten.Image
	var walkingSprites [8]*ebiten.Image
	
	// Map direction indices to file names (matching actual asset filenames)
	// Our internal order: 0=S, 1=SW, 2=W, 3=NW, 4=N, 5=NE, 6=E, 7=SE
	staticFiles := []string{
		"character_south",  // 0: South
		"character_sw",     // 1: Southwest
		"character_w",      // 2: West
		"character_nw",     // 3: Northwest
		"character_n",      // 4: North
		"character_ne",     // 5: Northeast
		"character_east",   // 6: East
		"character_se",     // 7: Southeast
	}
	
	hasRealSprites := true
	
	// Load static sprites from assets/Character/static/
	for i := 0; i < 8; i++ {
		sprite, err := assets.LoadTile(fmt.Sprintf("assets/Character/static/%s.png", staticFiles[i]))
		if err != nil {
			fmt.Printf("Failed to load sprite: assets/Character/static/%s.png - %v\n", staticFiles[i], err)
			hasRealSprites = false
			break
		}
		staticSprites[i] = sprite
		// For now, use static sprites for walking too
		walkingSprites[i] = sprite
	}
	
	// If we couldn't load real sprites, create colored placeholders
	if !hasRealSprites {
		// Define colors for each direction (simple placeholder characters)
		directionColors := []color.Color{
			color.RGBA{255, 100, 100, 255}, // S - Red
			color.RGBA{255, 150, 100, 255}, // SW - Orange
			color.RGBA{255, 255, 100, 255}, // W - Yellow
			color.RGBA{100, 255, 100, 255}, // NW - Green
			color.RGBA{100, 255, 255, 255}, // N - Cyan
			color.RGBA{100, 100, 255, 255}, // NE - Blue
			color.RGBA{150, 100, 255, 255}, // E - Purple
			color.RGBA{255, 100, 200, 255}, // SE - Pink
		}
		
		// Create simple character sprites (32x48 rectangles with direction indicators)
		for i := 0; i < 8; i++ {
			img := image.NewRGBA(image.Rect(0, 0, 32, 48))
			
			// Fill with direction color
			for y := 0; y < 48; y++ {
				for x := 0; x < 32; x++ {
					// Body
					if y > 8 {
						img.Set(x, y, directionColors[i])
					}
					// Head (circle-ish)
					if y <= 16 && x >= 8 && x < 24 {
						dx := float64(x - 16)
						dy := float64(y - 8)
						if dx*dx+dy*dy < 64 {
							img.Set(x, y, directionColors[i])
						}
					}
				}
			}
			
			// Add direction indicator (small triangle or line)
			indicatorColor := color.RGBA{0, 0, 0, 255}
			switch i {
			case 0: // S - down arrow
				for x := 14; x <= 18; x++ {
					img.Set(x, 40, indicatorColor)
				}
				img.Set(15, 41, indicatorColor)
				img.Set(16, 42, indicatorColor)
				img.Set(17, 41, indicatorColor)
			case 4: // N - up arrow
				for x := 14; x <= 18; x++ {
					img.Set(x, 20, indicatorColor)
				}
				img.Set(15, 19, indicatorColor)
				img.Set(16, 18, indicatorColor)
				img.Set(17, 19, indicatorColor)
			case 2: // W - left arrow
				for y := 28; y <= 32; y++ {
					img.Set(8, y, indicatorColor)
				}
				img.Set(7, 29, indicatorColor)
				img.Set(6, 30, indicatorColor)
				img.Set(7, 31, indicatorColor)
			case 6: // E - right arrow
				for y := 28; y <= 32; y++ {
					img.Set(24, y, indicatorColor)
				}
				img.Set(25, 29, indicatorColor)
				img.Set(26, 30, indicatorColor)
				img.Set(25, 31, indicatorColor)
			}
			
			staticSprites[i] = ebiten.NewImageFromImage(img)
			walkingSprites[i] = ebiten.NewImageFromImage(img) // Same for walking
		}
	}

	return &Player{
		X:                startX,
		Y:                startY,
		Speed:            0.1,
		StaticSprites:    staticSprites,
		WalkingSprites:   walkingSprites,
		CurrentDirection: "Idle",
		Facing:           0, // Default facing South
		IsWalking:        false,
		Level:            1, // Start at level 1
		TargetX:          startX,
		TargetY:          startY,
		Animating:        false,
		AnimationSpeed:   0.05,
		WalkAnimTimer:    0.0,
		WalkAnimFrame:    0,
	}, nil
}

// WalkabilityChecker interface for checking if tiles are walkable
type WalkabilityChecker interface {
	IsWalkable(x, y int) bool
}

// KeepInBounds constrains the player within the walkable area
func (p *Player) KeepInBounds(gridMovement bool, walkabilityChecker WalkabilityChecker) {
	// If no walkability checker provided, fall back to original hardcoded bounds
	if walkabilityChecker == nil {
		const (
			minX = 0.0   // Left edge of world tile (0-499 range)
			maxX = 499.0 // Right edge of world tile 
			minY = 0.0   // Top edge of world tile
			maxY = 499.0 // Bottom edge of world tile
		)

		if gridMovement {
			// For grid movement, constrain current position and round to grid
			if p.X < minX { p.X = minX }
			if p.X > maxX { p.X = maxX }
			if p.Y < minY { p.Y = minY }
			if p.Y > maxY { p.Y = maxY }
			// Round to grid positions
			p.X = float64(int(p.X + 0.5))
			p.Y = float64(int(p.Y + 0.5))
		} else {
			// For continuous movement, constrain current position
			if p.X < minX { p.X = minX }
			if p.X > maxX { p.X = maxX }
			if p.Y < minY { p.Y = minY }
			if p.Y > maxY { p.Y = maxY }
		}
		return
	}

	// Dynamic bounds checking using walkability - constrain current position
	currentX := int(p.X)
	currentY := int(p.Y)
	
	// If the current position is not walkable, find the nearest walkable position
	if !walkabilityChecker.IsWalkable(currentX, currentY) {
		// Search for the nearest walkable tile in expanding rings
		found := false
		for radius := 1; radius <= 5 && !found; radius++ {
			for dx := -radius; dx <= radius && !found; dx++ {
				for dy := -radius; dy <= radius && !found; dy++ {
					checkX := currentX + dx
					checkY := currentY + dy
					if walkabilityChecker.IsWalkable(checkX, checkY) {
						p.X = float64(checkX)
						p.Y = float64(checkY)
						found = true
					}
				}
			}
		}
		
		// If no walkable position found, don't change position
		// (this maintains the old behavior where invalid positions stay as-is)
	}
	// If current position is walkable, no need to constrain it
}

// UpdateAnimation updates the player's position during animation
func (p *Player) UpdateAnimation() {
	if !p.Animating {
		p.IsWalking = false
		p.WalkAnimTimer = 0.0
		p.WalkAnimFrame = 0
		return
	}

	// Set walking state when animating
	p.IsWalking = true

	// Update walking animation frame timing (change frame every 0.15 seconds for 4-frame cycle)
	p.WalkAnimTimer += 1.0/60.0 // Assuming 60 FPS
	if p.WalkAnimTimer >= 0.15 {
		p.WalkAnimFrame = (p.WalkAnimFrame + 1) % 4
		p.WalkAnimTimer = 0.0
	}

	// Calculate distance to target
	dx := p.TargetX - p.X
	dy := p.TargetY - p.Y
	distance := math.Sqrt(dx*dx + dy*dy)
	
	// Check if we're close enough to target (animation complete)
	if distance < 0.01 {
		p.X = p.TargetX
		p.Y = p.TargetY
		p.Animating = false
		p.IsWalking = false
		p.WalkAnimTimer = 0.0
		p.WalkAnimFrame = 0
	} else {
		// Move at constant speed (same as keyboard movement)
		// Normalize direction and apply speed
		moveSpeed := p.AnimationSpeed * 2.0 // Doubled to match keyboard movement feel
		if distance < moveSpeed {
			// If we're very close, just snap to target
			p.X = p.TargetX
			p.Y = p.TargetY
		} else {
			// Move towards target at constant speed
			p.X += (dx / distance) * moveSpeed
			p.Y += (dy / distance) * moveSpeed
		}
	}
}

// Rotate rotates the player to face the next direction
func (p *Player) Rotate() {
	p.Facing = (p.Facing + 1) % 8
}

// GetFacingName returns the direction name for the current facing
func (p *Player) GetFacingName() string {
	directions := []string{"S", "SW", "W", "NW", "N", "NE", "E", "SE"}
	return directions[p.Facing]
}

// SetFacingFromMovement updates the player's facing direction based on movement
func (p *Player) SetFacingFromMovement(dx, dy float64) {
	// Our sprite order: 0=S, 1=SW, 2=W, 3=NW, 4=N, 5=NE, 6=E, 7=SE
	if dx == 0 && dy > 0 {
		p.Facing = 0 // Moving down (south) -> face south
	} else if dx == 0 && dy < 0 {
		p.Facing = 4 // Moving up (north) -> face north
	} else if dx < 0 && dy == 0 {
		p.Facing = 2 // Moving left (west) -> face west
	} else if dx > 0 && dy == 0 {
		p.Facing = 6 // Moving right (east) -> face east
	} else if dx < 0 && dy > 0 {
		p.Facing = 1 // Moving down-left (southwest) -> face southwest
	} else if dx > 0 && dy > 0 {
		p.Facing = 7 // Moving down-right (southeast) -> face southeast
	} else if dx < 0 && dy < 0 {
		p.Facing = 3 // Moving up-left (northwest) -> face northwest
	} else if dx > 0 && dy < 0 {
		p.Facing = 5 // Moving up-right (northeast) -> face northeast
	}
}
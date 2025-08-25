package assets

import (
	"embed"
	"image"
	_ "image/png"  // Register PNG decoder
	_ "image/jpeg" // Register JPEG decoder

	"github.com/hajimehoshi/ebiten/v2"
)

// GlobalAssets holds the embedded assets
var GlobalAssets embed.FS

// LoadTile loads a single tile image from the global assets
func LoadTile(path string) (*ebiten.Image, error) {
	file, err := GlobalAssets.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}
package main

import (
	"embed"

	"github.com/tcornell05/go/game/cmd"
	"github.com/tcornell05/go/game/pkg/assets"
)

//go:embed assets/*
var embeddedAssets embed.FS

func main() {
	// Initialize global assets
	assets.GlobalAssets = embeddedAssets
	
	// Set embedded assets for commands
	cmd.SetEmbeddedAssets(embeddedAssets)
	
	// Execute cobra command
	cmd.Execute()
}
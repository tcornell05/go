package cmd

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
	"github.com/tcornell05/go/game/internal/game"
)

var (
	playerName   string
	startTile    string
	playWorldFile string
)

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Start the game",
	Long: `Start the game to:
- Navigate between world tiles
- Enter your rented rooms
- Visit public spaces
- Interact with furniture and other players`,
	Run: runGame,
}

func init() {
	rootCmd.AddCommand(playCmd)
	
	// Game-specific flags
	playCmd.Flags().StringVarP(&playerName, "name", "n", "Player", "Your player name")
	playCmd.Flags().StringVarP(&startTile, "start", "s", "A1", "Starting world tile")
	playCmd.Flags().StringVarP(&playWorldFile, "world", "w", "default", "World to play in")
}

func runGame(cmd *cobra.Command, args []string) {
	log.Printf("Starting game as %s in world %s at tile %s\n", playerName, playWorldFile, startTile)
	
	// Initialize game
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("Isometric Room Game - " + playerName)
	// Let the game handle cursor mode dynamically
	
	// For now, use the existing game implementation
	// Later this will be replaced with a world-aware game
	g := game.New()
	
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
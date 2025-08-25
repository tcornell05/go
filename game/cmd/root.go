package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/tcornell05/go/game/internal/database"
)

var rootCmd = &cobra.Command{
	Use:   "game",
	Short: "Isometric room game inspired by Habbo Hotel",
	Long: `An isometric room decoration and social game where players can:
- Explore and navigate between world tiles
- Rent and decorate private rooms
- Visit public spaces and interact with others`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Add migration command for cleaning up room system
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrate from room system to world tile system",
		Long: `Migrate any existing room data to world tile format and remove deprecated room tables.
This command will:
- Convert any existing rooms to world tiles
- Drop the deprecated room_entities and rooms tables
- Clean up the database to use only world tiles`,
		Run: func(cmd *cobra.Command, args []string) {
			runMigration()
		},
	}
	rootCmd.AddCommand(migrateCmd)
}

func runMigration() {
	log.Println("Starting room system migration...")
	
	// Migrate both game and editor databases
	databases := []string{"data/game.db", "data/editor.db"}
	
	for _, dbPath := range databases {
		log.Printf("Migrating database: %s", dbPath)
		
		db, err := database.NewDatabase(dbPath)
		if err != nil {
			log.Printf("Failed to open database %s: %v", dbPath, err)
			continue
		}
		
		// Run migration
		if err := db.MigrateRoomDataToWorldTiles(); err != nil {
			log.Printf("Migration failed for %s: %v", dbPath, err)
			continue
		}
		
		// Drop room tables
		if err := db.DropRoomTables(); err != nil {
			log.Printf("Failed to drop room tables in %s: %v", dbPath, err)
			continue
		}
		
		log.Printf("Successfully migrated %s", dbPath)
		db.Close()
	}
	
	log.Println("Room system migration complete!")
}
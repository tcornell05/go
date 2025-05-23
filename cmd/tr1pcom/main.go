package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tcornell05/go/tr1pcom/internal/client"
)

func main() {
	clientName := flag.String("name", "", "Name of this client (e.g., barn, office)")
	clientsConfigFile := flag.String("clients", "config/clients.yaml", "Path to the clients configuration file")
	flag.Parse()

	if *clientName == "" {
		fmt.Println("Error: --name flag is required.")
		flag.Usage() // Print usage information
		os.Exit(1)   // Exit with an error code
	}

	fmt.Printf("Starting tr1pcom client: %s\n", *clientName)
	fmt.Printf("Using clients configuration: %s\n", *clientsConfigFile)

	// Initialize the application client
	appClient, err := client.NewClient(*clientName, *clientsConfigFile)
	if err != nil {
		log.Fatalf("Error initializing client: %v", err)
	}

	// Start the client (this will eventually run the Ebiten game loop)
	if err := appClient.Start(); err != nil {
		log.Fatalf("Error running client: %v", err)
	}

	log.Println("tr1pcom client finished.")
}

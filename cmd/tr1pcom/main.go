package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	clientName := flag.String("name", "", "Name of this client (e.g., barn, office)")
	clientsConfigFile := flag.String("clients", "config/clients.yaml", "Path to the clients configuration file")
	flag.Parse()

	if *clientName == "" {
		log.Fatal("Error: --name flag is required")
	}

	fmt.Printf("Starting tr1pcom client: %s\n", *clientName)
	fmt.Printf("Using clients configuration: %s\n", *clientsConfigFile)

	// TODO: Load client configuration
	// TODO: Initialize Ebiten UI
	// TODO: Implement client connection logic
	// TODO: Implement host election/management
	// TODO: Implement audio streaming and session management
}

package client

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// ClientConfig represents a single client's configuration in the system.
type ClientConfig struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

// SystemConfig represents the overall system configuration, primarily a list of clients.
type SystemConfig struct {
	Clients []ClientConfig `yaml:"clients"`
}

// Client represents the main application logic for an intercom client.
type Client struct {
	Name         string
	SystemConfig SystemConfig
	// TODO: Add fields for network state, UI, audio streams, host status, etc.
}

// LoadSystemConfig loads the client configurations from a YAML file.
func LoadSystemConfig(filePath string) (*SystemConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	var config SystemConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data from %s: %w", filePath, err)
	}

	if len(config.Clients) == 0 {
		log.Printf("Warning: No clients defined in the configuration file: %s", filePath)
	}

	return &config, nil
}

// NewClient creates and initializes a new Client.
func NewClient(clientName string, clientsConfigFile string) (*Client, error) {
	cfg, err := LoadSystemConfig(clientsConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load system config: %w", err)
	}

	// Validate if the current clientName exists in the loaded configuration
	clientExists := false
	var currentClientConfig ClientConfig
	for _, c := range cfg.Clients {
		if c.Name == clientName {
			clientExists = true
			currentClientConfig = c
			break
		}
	}
	if !clientExists {
		return nil, fmt.Errorf("client name '%s' not found in configuration file '%s'. Ensure it's listed under 'clients:'", clientName, clientsConfigFile)
	}

	log.Printf("Initializing client: %s with address %s", currentClientConfig.Name, currentClientConfig.Address)

	return &Client{
		Name:         clientName,
		SystemConfig: *cfg,
	}, nil
}

// Start begins the client's operations.
// This is where Ebiten UI will be initialized and the main application loop will run.
func (c *Client) Start() error {
	log.Printf("Client '%s' starting...", c.Name)
	log.Printf("System Configuration Loaded. Known clients:")
	for _, clientCfg := range c.SystemConfig.Clients {
		log.Printf("  - Name: %s, Address: %s", clientCfg.Name, clientCfg.Address)
		if clientCfg.Name == c.Name {
			log.Printf("    (This is me: %s)", c.Name)
		}
	}

	// TODO: Initialize Ebiten UI (e.g., ebiten.RunGame(newGame(c)))
	// TODO: Implement client connection logic (listening for incoming, connecting to host)
	// TODO: Implement host election/management
	// TODO: Implement audio streaming and session management

	log.Println("Client setup complete. Further implementation (UI, networking) will go here.")
	// For now, the application will exit after this.
	// Once Ebiten is integrated, ebiten.RunGame() will block and run the game loop.
	return nil
}

package client

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// ClientConfig represents a single client's configuration in the system.
type ClientConfig struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"` // Expected format: "host:port" or ":port"
}

// SystemConfig represents the overall system configuration, primarily a list of clients.
type SystemConfig struct {
	Clients []ClientConfig `yaml:"clients"`
}

// Client represents the main application logic for an intercom client.
type Client struct {
	config       ClientConfig   // This client's own configuration
	systemConfig SystemConfig   // All clients in the system
	listener     net.Listener
	peers        map[string]net.Conn // Key: peer name
	mu           sync.Mutex
	// TODO: Add Ebiten game state, UI elements etc.
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

	var currentClientConfig ClientConfig
	clientFound := false
	for _, c := range cfg.Clients {
		if c.Name == clientName {
			currentClientConfig = c
			clientFound = true
			break
		}
	}
	if !clientFound {
		return nil, fmt.Errorf("client name '%s' not found in configuration file '%s'. Ensure it's listed under 'clients:'", clientName, clientsConfigFile)
	}

	log.Printf("Initializing client: %s with address %s", currentClientConfig.Name, currentClientConfig.Address)

	return &Client{
		config:       currentClientConfig,
		systemConfig: *cfg,
		peers:        make(map[string]net.Conn),
	}, nil
}

// Start begins the client's operations.
func (c *Client) Start() error {
	log.Printf("Client '%s' starting at address %s...", c.config.Name, c.config.Address)

	// Start listening for incoming connections
	go c.runListener()

	// Attempt to connect to other known peers
	// Add a small delay to allow listeners to start on other clients, especially in local testing.
	// The connection logic itself has retries, so this is mostly for smoother local startups.
	time.Sleep(1 * time.Second)
	go c.connectToPeers()

	// TODO: Initialize Ebiten UI (e.g., ebiten.RunGame(newGame(c)))
	// For now, block indefinitely. Ebiten's RunGame will do this in the future.
	log.Println("Client setup complete. Listening for connections and connecting to peers.")
	log.Println("Press Ctrl+C to exit (this may take a moment due to active connections).")
	select {} // Block forever
}

// runListener starts and manages the TCP listener for incoming connections.
func (c *Client) runListener() {
	var err error
	c.listener, err = net.Listen("tcp", c.config.Address)
	if err != nil {
		// If listener fails to start, it's a critical error for this client instance.
		log.Fatalf("FATAL: Client '%s' failed to start listener on %s: %v", c.config.Name, c.config.Address, err)
		return // log.Fatalf will exit, but return for completeness.
	}
	defer c.listener.Close()
	log.Printf("Client '%s' is now listening on %s", c.config.Name, c.config.Address)

	for {
		conn, err := c.listener.Accept()
		if err != nil {
			// Check if the error is due to the listener being closed.
			if strings.Contains(err.Error(), "use of closed network connection") {
				log.Printf("Client '%s': Listener on %s closed.", c.config.Name, c.config.Address)
				return // Exit goroutine as listener is closed.
			}
			log.Printf("Client '%s': Error accepting connection on %s: %v", c.config.Name, c.config.Address, err)
			continue // Continue trying to accept connections
		}
		// Handle each incoming connection in a new goroutine to not block the listener.
		go c.handleIncomingConnection(conn)
	}
}

// handleIncomingConnection manages a newly accepted connection.
func (c *Client) handleIncomingConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	log.Printf("Client '%s': Received incoming connection from %s", c.config.Name, remoteAddr)

	// Defer closing the connection to ensure it's cleaned up.
	// This defer will be the last thing to run for this function's scope.
	// If managePeerLifecycle is called, its own defer will handle peer map cleanup.
	// This one ensures conn.Close() if handshake fails early.
	var peerName string // Declare here to be accessible in defer if needed, though not strictly used by this defer.
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Client '%s': Recovered in handleIncomingConnection from %s: %v", c.config.Name, remoteAddr, r)
		}
		// If peerName is not set (handshake failed), we don't remove from map, just close.
		// If peerName is set, managePeerLifecycle's defer will handle map removal.
		// log.Printf("Client '%s': Closing incoming connection from %s (peer: %s)", c.config.Name, remoteAddr, peerName)
		conn.Close()
	}()

	// Handshake: Read the connecting client's name
	reader := bufio.NewReader(conn)
	peerNameMsg, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Client '%s': Failed to read peer name from %s: %v", c.config.Name, remoteAddr, err)
		return // Error in handshake, connection will be closed by defer.
	}
	peerName = strings.TrimSpace(peerNameMsg)
	if peerName == "" {
		log.Printf("Client '%s': Received empty peer name from %s. Closing connection.", c.config.Name, remoteAddr)
		return
	}

	log.Printf("Client '%s': Identified incoming peer from %s as '%s'", c.config.Name, remoteAddr, peerName)

	// Handshake: Send own name back as acknowledgment
	_, err = fmt.Fprintf(conn, "%s\n", c.config.Name)
	if err != nil {
		log.Printf("Client '%s': Failed to send own name to peer '%s' (%s): %v", c.config.Name, peerName, remoteAddr, err)
		return // Error in handshake
	}

	c.mu.Lock()
	if existingConn, ok := c.peers[peerName]; ok {
		log.Printf("Client '%s': Existing connection found for peer '%s'. Closing old one before accepting new incoming.", c.config.Name, peerName)
		existingConn.Close() // Close the old connection
	}
	c.peers[peerName] = conn
	c.mu.Unlock()

	log.Printf("Client '%s': Successfully established connection with peer '%s' (incoming from %s). Total peers: %d", c.config.Name, peerName, remoteAddr, len(c.peers))
	c.managePeerLifecycle(peerName, conn) // This function will handle the ongoing communication and cleanup
}

// connectToPeers attempts to establish outgoing connections to all other clients in the system config.
func (c *Client) connectToPeers() {
	for _, peerCfg := range c.systemConfig.Clients {
		if peerCfg.Name == c.config.Name {
			continue // Don't connect to self
		}

		// Launch a separate goroutine for each peer to connect to, allowing parallel attempts and retries.
		go func(pcfg ClientConfig) {
			for { // Infinite loop for retrying connections to this specific peer
				c.mu.Lock()
				_, isConnected := c.peers[pcfg.Name]
				c.mu.Unlock()

				if isConnected {
					// If already connected (either we dialed them or they dialed us),
					// just wait and periodically re-check. The managePeerLifecycle
					// for the existing connection will handle its lifetime.
					time.Sleep(15 * time.Second) // Check periodically
					continue
				}

				log.Printf("Client '%s': Attempting to connect to peer '%s' at %s", c.config.Name, pcfg.Name, pcfg.Address)
				conn, err := net.DialTimeout("tcp", pcfg.Address, 5*time.Second)
				if err != nil {
					log.Printf("Client '%s': Failed to connect to peer '%s' (%s): %v. Retrying in 10s.", c.config.Name, pcfg.Name, pcfg.Address, err)
					time.Sleep(10 * time.Second)
					continue // Retry connection in the next iteration of the for loop
				}
				
				remoteAddr := conn.RemoteAddr().String() // Get it early for logging
				log.Printf("Client '%s': Established outgoing TCP connection to peer '%s' at %s (remote: %s)", c.config.Name, pcfg.Name, pcfg.Address, remoteAddr)

				// Defer closing this new connection if handshake fails or managePeerLifecycle exits.
				// This defer is specific to this attempt within the loop.
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Client '%s': Recovered in connectToPeers goroutine for %s: %v", c.config.Name, pcfg.Name, r)
					}
					// This defer ensures conn.Close() if this specific Dial attempt fails post-connection (e.g. handshake)
					// or if managePeerLifecycle returns. The map cleanup is handled by managePeerLifecycle.
					// log.Printf("Client '%s': Closing outgoing connection attempt to %s (%s)", c.config.Name, pcfg.Name, remoteAddr)
					conn.Close() 
				}()

				// Handshake: Send own name
				_, err = fmt.Fprintf(conn, "%s\n", c.config.Name)
				if err != nil {
					log.Printf("Client '%s': Failed to send own name to peer '%s' (%s): %v. Retrying connection.", c.config.Name, pcfg.Name, remoteAddr, err)
					// conn.Close() is handled by defer. Wait before retrying.
					time.Sleep(10 * time.Second)
					continue
				}

				// Handshake: Read peer's name (as acknowledgment)
				reader := bufio.NewReader(conn)
				ackPeerNameMsg, err := reader.ReadString('\n')
				if err != nil {
					log.Printf("Client '%s': Failed to read ack name from peer '%s' (%s): %v. Retrying connection.", c.config.Name, pcfg.Name, remoteAddr, err)
					time.Sleep(10 * time.Second)
					continue
				}
				ackPeerName := strings.TrimSpace(ackPeerNameMsg)
				if ackPeerName != pcfg.Name {
					log.Printf("Client '%s': Received unexpected ack name '%s' from peer '%s' (%s). Expected '%s'. Retrying connection.", c.config.Name, ackPeerName, pcfg.Name, remoteAddr, pcfg.Name)
					time.Sleep(10 * time.Second)
					continue
				}

				c.mu.Lock()
				if existingConn, ok := c.peers[pcfg.Name]; ok {
					log.Printf("Client '%s': Existing connection found for peer '%s' while establishing outgoing. Closing old one.", c.config.Name, pcfg.Name)
					existingConn.Close()
				}
				c.peers[pcfg.Name] = conn
				c.mu.Unlock()

				log.Printf("Client '%s': Successfully established connection with peer '%s' (outgoing to %s). Total peers: %d", c.config.Name, pcfg.Name, remoteAddr, len(c.peers))
				c.managePeerLifecycle(pcfg.Name, conn) // This blocks until connection is lost

				// If managePeerLifecycle returns, it means the connection was lost.
				// The loop will then re-evaluate `isConnected` and attempt to reconnect.
				log.Printf("Client '%s': Connection to peer '%s' (%s) ended. Will re-evaluate and attempt to reconnect if necessary.", c.config.Name, pcfg.Name, remoteAddr)
				// No explicit sleep here, loop will check `isConnected` and potentially sleep if Dial fails.
			}
		}(peerCfg))
	}
}

// managePeerLifecycle handles the lifetime of an established peer connection (for both incoming and outgoing).
// It reads messages from the peer and handles disconnection.
func (c *Client) managePeerLifecycle(peerName string, conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	log.Printf("Client '%s': Managing lifecycle for peer '%s' (%s)", c.config.Name, peerName, remoteAddr)
	reader := bufio.NewReader(conn)

	// Defer is crucial for cleanup when this function exits (connection lost/error).
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Client '%s': Recovered in managePeerLifecycle for %s (%s): %v", c.config.Name, peerName, remoteAddr, r)
		}
		log.Printf("Client '%s': Connection with peer '%s' (%s) is closing/ending.", c.config.Name, peerName, remoteAddr)
		conn.Close() // Ensure the connection is closed.

		c.mu.Lock()
		// Only delete from map if this specific connection instance is still the one in the map.
		// This prevents accidentally deleting a newer connection to the same peer.
		if currentConn, ok := c.peers[peerName]; ok && currentConn == conn {
			delete(c.peers, peerName)
			log.Printf("Client '%s': Removed peer '%s' from active connections. Total peers now: %d", c.config.Name, peerName, len(c.peers))
		} else if ok {
			log.Printf("Client '%s': Peer '%s' (%s) was likely replaced by a new connection. Not removing this instance from map.", c.config.Name, peerName, remoteAddr)
		} else {
			log.Printf("Client '%s': Peer '%s' (%s) not found in active connections map during cleanup (already removed).", c.config.Name, peerName, remoteAddr)
		}
		c.mu.Unlock()
	}()

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Client '%s': Peer '%s' (%s) disconnected (EOF).", c.config.Name, peerName, remoteAddr)
			} else if strings.Contains(err.Error(), "use of closed network connection") {
				log.Printf("Client '%s': Connection to peer '%s' (%s) closed locally.", c.config.Name, peerName, remoteAddr)
			} else {
				log.Printf("Client '%s': Error reading from peer '%s' (%s): %v", c.config.Name, peerName, remoteAddr, err)
			}
			return // Exit goroutine, defer will clean up.
		}

		trimmedMessage := strings.TrimSpace(message)
		if trimmedMessage == "" && err == nil { // Just an empty line, might be keep-alive or accidental
			continue
		}
		log.Printf("Client '%s': Received from '%s' (%s): \"%s\"", c.config.Name, peerName, remoteAddr, trimmedMessage)
		// TODO: Process the message (e.g., call requests, audio data, text chat, etc.)
		// Example: c.handlePeerMessage(peerName, trimmedMessage)
	}
}

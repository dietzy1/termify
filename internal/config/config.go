package config

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Server struct {
		// Port for the server to listen on
		Port string `yaml:"port"`
	} `yaml:"server"`

	// Spotify configuration
	Spotify struct {
		// ClientID for Spotify API
		ClientID string `yaml:"client_id"`
		// Whether to use daemon mode (run without Spotify being open)
		DaemonMode bool `yaml:"daemon_mode"`
		// Which Spotify connect client to use
		ConnectClient string `yaml:"connect_client"`
	} `yaml:"spotify"`

	// Internal configuration
	ConfigPath string `yaml:"-"` // Not stored in config file
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	cfg := &Config{}

	// Set default values
	cfg.Server.Port = "8080"
	cfg.Spotify.ConnectClient = "default"
	cfg.Spotify.DaemonMode = false

	return cfg
}

// LoadConfig loads configuration from file, environment variables, and command-line flags
func LoadConfig() (*Config, error) {
	// Start with default configuration
	cfg := DefaultConfig()

	// Define and parse command-line flags
	var (
		configPath    = flag.String("config", "", "Path to config file")  //Verify that the config file exists and can be read
		port          = flag.String("port", "", "Server port")            // Verify that the port is a valid port number and that it is not already in use
		clientID      = flag.String("client-id", "", "Spotify client ID") // Verify that the client ID is a valid Spotify client ID aka 32 characters long
		daemonMode    = flag.Bool("daemon", false, "Run in daemon mode")  // Verify that the daemon mode flag is a boolean
		connectClient = flag.String("connect-client", "", "Spotify connect client to use")
	)
	flag.Parse()

	// Determine config file path
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config directory: %w", err)
	}

	// Default config path if not specified
	defaultConfigPath := filepath.Join(userConfigDir, "termify", "config.yaml")

	// Use specified config path or default
	configFilePath := defaultConfigPath
	if *configPath != "" {
		configFilePath = *configPath
	}
	cfg.ConfigPath = configFilePath

	// Try to load from config file
	if err := cfg.loadFromFile(configFilePath); err != nil {
		// Only return error if file exists but couldn't be loaded
		if !errors.Is(err, fs.ErrNotExist) {
			log.Printf("Config file not found at %s, using default configuration", configFilePath)
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables
	if envPort := os.Getenv("TERMIFY_PORT"); envPort != "" {
		cfg.Server.Port = envPort
	}
	if envClientID := os.Getenv("TERMIFY_CLIENT_ID"); envClientID != "" {
		cfg.Spotify.ClientID = envClientID
	}
	if envDaemonMode := os.Getenv("TERMIFY_DAEMON_MODE"); envDaemonMode == "true" {
		cfg.Spotify.DaemonMode = true
	}
	if envConnectClient := os.Getenv("TERMIFY_CONNECT_CLIENT"); envConnectClient != "" {
		cfg.Spotify.ConnectClient = envConnectClient
	}

	// Override with command-line flags (highest priority)
	if *port != "" {
		cfg.Server.Port = *port
	}
	if *clientID != "" {
		cfg.Spotify.ClientID = *clientID
	}
	if *daemonMode {
		cfg.Spotify.DaemonMode = true
	}
	if *connectClient != "" {
		cfg.Spotify.ConnectClient = *connectClient
	}

	// Ensure port has colon prefix
	if cfg.Server.Port != "" && cfg.Server.Port[0] != ':' {
		cfg.Server.Port = ":" + cfg.Server.Port
	}

	//Debug log all config values
	log.Println("=== CONFIGURATION VALUES ===")
	log.Println("Server:")
	log.Printf("  Port: %s", cfg.Server.Port)
	log.Println("Spotify:")
	log.Printf("  Client ID: %s", cfg.Spotify.ClientID)
	log.Printf("  Daemon mode: %t", cfg.Spotify.DaemonMode)
	log.Printf("  Connect client: %s", cfg.Spotify.ConnectClient)
	log.Println("============================")

	return cfg, nil
}

// loadFromFile loads configuration from a YAML file
func (c *Config) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}

// SaveConfig saves the current configuration to file
func (c *Config) SaveConfig() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(c.ConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(c.ConfigPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetClientID returns the Spotify client ID
func (c *Config) GetClientID() string {
	return c.Spotify.ClientID
}

// SetClientID sets the Spotify client ID and saves the config
func (c *Config) SetClientID(clientID string) error {
	c.Spotify.ClientID = clientID
	return c.SaveConfig()
}

// GetPort returns the server port with colon prefix
func (c *Config) GetPort() string {
	return c.Server.Port
}

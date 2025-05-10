package config

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

func (c *Config) validate() error {
	// Validate Port
	if c.Server.Port != "" { // Assuming port can be empty to disable server
		portStr := strings.TrimPrefix(c.Server.Port, ":")
		if portNum, err := strconv.Atoi(portStr); err != nil || portNum < 1 || portNum > 65535 {
			return fmt.Errorf("invalid server port '%s': must be a number between 1 and 65535", c.Server.Port)
		}
	} else {
		return errors.New("server port cannot be empty if server is enabled")
	}

	// Validate Spotify Client ID
	if c.Spotify.ClientID == "" {
		log.Println("Warning: Spotify Client ID is not set in configuration.")
	} else if len(c.Spotify.ClientID) != 32 {
		return fmt.Errorf("invalid Spotify Client ID: must be 32 characters long, got %d", len(c.Spotify.ClientID))
	}

	return nil
}

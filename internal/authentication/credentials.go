package authentication

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// Credentials represents the stored authentication credentials
type credentials struct {
	ClientID     string    `json:"client_id"`
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

// CredentialManager handles the storage and retrieval of authentication credentials
type credentialManager struct {
	filePath string
	mu       sync.RWMutex
}

// NewCredentialManager creates a new credential manager
func NewCredentialManager(configDir string) (*credentialManager, error) {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &credentialManager{
		filePath: filepath.Join(configDir, "spotify_credentials.json"),
	}, nil
}

// SaveToken stores the OAuth2 token to disk
func (cm *credentialManager) SaveToken(token *oauth2.Token, clientId string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	creds := credentials{
		ClientID:     clientId,
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := os.WriteFile(cm.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// LoadToken retrieves the OAuth2 token from disk
func (cm *credentialManager) LoadToken() (*oauth2.Token, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	data, err := os.ReadFile(cm.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return &oauth2.Token{
		AccessToken:  creds.AccessToken,
		TokenType:    creds.TokenType,
		RefreshToken: creds.RefreshToken,
		Expiry:       creds.Expiry,
	}, nil
}

func (cm *credentialManager) LoadClientID() (string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	data, err := os.ReadFile(cm.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", fmt.Errorf("failed to unmarshal credentials: %w", err)
	}
	log.Println("creds:", creds)

	return creds.ClientID, nil
}

// ClearToken removes the stored credentials
func (cm *credentialManager) ClearToken() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err := os.Remove(cm.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove credentials file: %w", err)
	}

	return nil
}

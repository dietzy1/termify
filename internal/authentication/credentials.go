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
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

// CredentialManager handles the storage and retrieval of authentication credentials and client ID
type credentialManager struct {
	configClientID string
	configFilePath string
	mu             sync.RWMutex
}

func NewCredentialManager(configClientID string, configDir string) (*credentialManager, error) {

	return &credentialManager{
		configClientID: configClientID,
		configFilePath: filepath.Join(configDir, "spotify_credentials.json"),
		mu:             sync.RWMutex{},
	}, nil
}

// SaveToken stores the OAuth2 token to disk
func (cm *credentialManager) saveToken(token *oauth2.Token) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	creds := credentials{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := os.WriteFile(cm.configFilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// LoadToken retrieves the OAuth2 token from disk
func (cm *credentialManager) loadToken() (*oauth2.Token, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	data, err := os.ReadFile(cm.configFilePath)
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

func (cm *credentialManager) saveClientID(clientID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	clientIDFilePath := filepath.Join(filepath.Dir(cm.configFilePath), "spotify_client_id.json")

	data, err := json.MarshalIndent(map[string]string{"client_id": clientID}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal client ID: %w", err)
	}

	if err := os.WriteFile(clientIDFilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write client ID file: %w", err)
	}

	return nil
}

func (cm *credentialManager) loadClientID() (string, error) {

	if cm.configClientID != "" {
		log.Printf("Using client ID from config: %s", cm.configClientID)
		return cm.configClientID, nil
	}

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	clientIDFilePath := filepath.Join(filepath.Dir(cm.configFilePath), "spotify_client_id.json")

	data, err := os.ReadFile(clientIDFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No client ID file exists yet
		}
		return "", fmt.Errorf("failed to read client ID file: %w", err)
	}

	var clientIDData map[string]string
	if err := json.Unmarshal(data, &clientIDData); err != nil {
		return "", fmt.Errorf("failed to unmarshal client ID: %w", err)
	}

	clientID, ok := clientIDData["client_id"]
	if !ok {
		return "", fmt.Errorf("client ID not found in file")
	}

	return clientID, nil
}

// ClearToken removes the stored credentials
func (cm *credentialManager) clearToken() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err := os.Remove(cm.configFilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove credentials file: %w", err)
	}

	return nil
}

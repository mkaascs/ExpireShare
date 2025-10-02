package crypto

import (
	"crypto/rand"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
)

type HmacConfig struct {
	hmacSecret []byte
}

const hmacSecretLength = 64

func NewHmacConfig(envPath string) (*HmacConfig, error) {
	path := filepath.Join(envPath, ".env")
	if err := godotenv.Load(path); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	keyPath := os.Getenv("HMAC_SECRET_PATH")
	if keyPath == "" {
		return nil, fmt.Errorf("env variable HMAC_SECRET_PATH not found")
	}

	hmacConfig := &HmacConfig{}

	if err := hmacConfig.loadSecret(keyPath); err != nil {
		if err = hmacConfig.generateSecret(keyPath); err != nil {
			return nil, fmt.Errorf("failed to load or generate secret: %w", err)
		}
	}

	return hmacConfig, nil
}

func (hc *HmacConfig) GetHmacSecret() []byte {
	return hc.hmacSecret
}

func (hc *HmacConfig) loadSecret(secretPath string) error {
	data, err := os.ReadFile(secretPath)
	if err == nil {
		hc.hmacSecret = data
	}

	return err
}

func (hc *HmacConfig) generateSecret(secretPath string) error {
	secret := make([]byte, hmacSecretLength)
	if _, err := rand.Read(secret); err != nil {
		return err
	}

	hc.hmacSecret = secret
	return os.WriteFile(secretPath, hc.hmacSecret, 0600)
}

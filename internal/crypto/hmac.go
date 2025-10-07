package crypto

import (
	"crypto/rand"
	"fmt"
	"github.com/joho/godotenv"
	"maps"
	"os"
	"path/filepath"
)

type HmacConfig struct {
	accessTokenSecret  []byte
	refreshTokenSecret []byte
}

const hmacSecretLength = 32

func (hc *HmacConfig) GetAccessTokenSecret() []byte {
	return hc.accessTokenSecret
}

func (hc *HmacConfig) GetRefreshTokenSecret() []byte {
	return hc.refreshTokenSecret
}

func NewHmacConfig(envPath string) (*HmacConfig, error) {
	path := filepath.Join(envPath, ".env")
	if err := godotenv.Load(path); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	hmacConfig := &HmacConfig{}

	secrets := map[string]*[]byte{
		"ACCESS_TOKEN_SECRET_PATH":  &hmacConfig.accessTokenSecret,
		"REFRESH_TOKEN_SECRET_PATH": &hmacConfig.refreshTokenSecret,
	}

	for key := range maps.Keys(secrets) {
		keyPath := os.Getenv(key)
		if keyPath == "" {
			return nil, fmt.Errorf("env variable %s not found", key)
		}

		var err error
		*secrets[key], err = os.ReadFile(keyPath)
		if err != nil {
			if *secrets[key], err = generateSecret(keyPath); err != nil {
				return nil, fmt.Errorf("failed to load and generate secret: %w", err)
			}
		}
	}

	return hmacConfig, nil
}

func generateSecret(secretPath string) ([]byte, error) {
	secret := make([]byte, hmacSecretLength)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}

	if err := os.WriteFile(secretPath, secret, 0600); err != nil {
		return nil, err
	}

	return secret, nil
}

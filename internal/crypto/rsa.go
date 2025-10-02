package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
)

type RsaKeyConfig struct {
	privateKey *rsa.PrivateKey
}

func NewRsaKey(envPath string) (*RsaKeyConfig, error) {
	path := filepath.Join(envPath, ".env")
	if err := godotenv.Load(path); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	keyPath := os.Getenv("RSA_KEY_PATH")
	if keyPath == "" {
		return nil, fmt.Errorf("env variable RSA_KEY_PATH not found")
	}

	keyConfig := &RsaKeyConfig{}

	if err := keyConfig.loadKey(keyPath); err != nil {
		if err = keyConfig.generateKey(keyPath); err != nil {
			return nil, fmt.Errorf("failed to load or generate key: %w", err)
		}
	}

	return keyConfig, nil
}

func (kc *RsaKeyConfig) GetPrivateKey() *rsa.PrivateKey {
	return kc.privateKey
}

func (kc *RsaKeyConfig) loadKey(keyPath string) error {
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("failed to parse PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}

		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("failed to parse private key")
		}

		privateKey = rsaKey
	}

	kc.privateKey = privateKey
	return nil
}

func (kc *RsaKeyConfig) generateKey(keyPath string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	kc.privateKey = privateKey
	return kc.saveKey(keyPath)
}

func (kc *RsaKeyConfig) saveKey(keyPath string) error {
	dir := filepath.Dir(keyPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(kc.privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return os.WriteFile(keyPath, privateKeyPEM, 0600)
}

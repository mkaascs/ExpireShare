package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
)

type KeyManager struct {
	privateKey *rsa.PrivateKey
	keyPath    string
}

func MustLoad(envPath string) *KeyManager {
	km, err := Load(envPath)
	if err != nil {
		log.Fatal(err)
	}

	return km
}

func Load(envPath string) (*KeyManager, error) {
	path := filepath.Join(envPath, ".env")
	if err := godotenv.Load(path); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	keyPath := os.Getenv("RSA_KEY_PATH")
	if keyPath == "" {
		return nil, fmt.Errorf("env variable RSA_KEY_PATH not found")
	}

	km := &KeyManager{keyPath: keyPath}

	if err := km.loadKey(); err == nil {
		return km, nil
	}

	if err := km.generateKey(); err != nil {
		return nil, fmt.Errorf("error generating key: %v", err)
	}

	return km, nil
}

func (km *KeyManager) GetPrivateKey() *rsa.PrivateKey {
	return km.privateKey
}

func (km *KeyManager) GetPublicKey() *rsa.PublicKey {
	return &km.privateKey.PublicKey
}

func (km *KeyManager) loadKey() error {
	data, err := os.ReadFile(km.keyPath)
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

	km.privateKey = privateKey
	return nil
}

func (km *KeyManager) generateKey() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	km.privateKey = privateKey
	return km.saveKey()
}

func (km *KeyManager) saveKey() error {
	dir := filepath.Dir(km.keyPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(km.privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return os.WriteFile(km.keyPath, privateKeyPEM, 0600)
}

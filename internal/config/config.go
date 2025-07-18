package config

import (
	"expire-share/internal/lib/sizes"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	EnvironmentLocal = "local"
	EnvironmentDev   = "dev"
	EnvironmentProd  = "prod"
)

type Config struct {
	Environment      string `yaml:"environment" default:"local"`
	ConnectionString string `yaml:"connection_string" required:"true"`
	Storage          `yaml:"storage"`
	HttpServer       `yaml:"http_server"`
	Service          `yaml:"service"`
}

type Storage struct {
	Type               string `yaml:"type" default:"local"`
	Path               string `yaml:"path" required:"true"`
	MaxFileSize        string `yaml:"max_file_size" default:"100mb"`
	MaxFileSizeInBytes int64
}

type HttpServer struct {
	Address     string        `yaml:"address" required:"true"`
	Timeout     time.Duration `yaml:"timeout" default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" default:"60s"`
}

type Service struct {
	DefaultTtl      time.Duration `yaml:"default_ttl" default:"1h"`
	MaxDownloads    int16         `yaml:"default_max_downloads" default:"1"`
	AliasLength     int16         `yaml:"alias_length" default:"6"`
	FileWorkerDelay time.Duration `yaml:"file_worker_delay" default:"5m"`
}

func MustLoad(envPath string) *Config {
	cfg, err := Load(envPath)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func Load(envPath string) (*Config, error) {
	path := filepath.Join(envPath, ".env")
	if err := godotenv.Load(path); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		return nil, fmt.Errorf("env variable CONFIG_PATH not found")
	}

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file %s does not exist", cfgPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(cfgPath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	bytes, err := sizes.ToBytes(cfg.MaxFileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to parse max file size in config: %w", err)
	}

	cfg.MaxFileSizeInBytes = bytes
	return &cfg, nil
}

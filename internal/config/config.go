package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"os"
	"time"
)

type Config struct {
	Environment      string `yaml:"environment" default:"local"`
	ConnectionString string `yaml:"connection_string" required:"true"`
	Storage          `yaml:"storage"`
	HttpServer       `yaml:"http_server"`
	Service          `yaml:"service"`
}

type Storage struct {
	Type        string `yaml:"type" default:"local"`
	Path        string `yaml:"path" required:"true"`
	MaxFileSize string `yaml:"max_file_size" default:"100mb"`
}

type HttpServer struct {
	Address     string        `yaml:"address" required:"true"`
	Timeout     time.Duration `yaml:"timeout" default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" default:"60s"`
}

type Service struct {
	DefaultTtl   time.Duration `yaml:"default_ttl" default:"1h"`
	MaxDownloads int           `yaml:"default_max_downloads" default:"1"`
	AliasLength  int           `yaml:"alias_length" default:"6"`
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}

	return cfg
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
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

	return &cfg, nil
}

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App       AppConfig        `mapstructure:"app"`
	Server    ServerConfig     `mapstructure:"server"`
	Database  DatabaseConfig   `mapstructure:"database"`
	Redis     RedisConfig      `mapstructure:"redis"`
	Telegram  TelegramConfig   `mapstructure:"telegram"`
	Auth      AuthConfig       `mapstructure:"auth"`
	Ingestor  IngestorConfig   `mapstructure:"ingestor"`
	Exchanges []ExchangeConfig `mapstructure:"exchanges"`
}

type IngestorConfig struct {
	BatchSize       int `mapstructure:"batch_size"`
	FlushIntervalMS int `mapstructure:"flush_interval_ms"`
	BufferSize      int `mapstructure:"buffer_size"`
}

type ExchangeConfig struct {
	Name    string   `mapstructure:"name"`
	Enabled bool     `mapstructure:"enabled"`
	Market  string   `mapstructure:"market"`
	URL     string   `mapstructure:"url"`
	Symbols []string `mapstructure:"symbols"`
}

type AuthConfig struct {
	JWTSecret       string `mapstructure:"jwt_secret"`
	AccessTokenTTL  int    `mapstructure:"access_token_ttl"`  // minutes
	RefreshTokenTTL int    `mapstructure:"refresh_token_ttl"` // hours
}

type AppConfig struct {
	Env string `mapstructure:"env"`
}

type ServerConfig struct {
	Port         int `mapstructure:"port"`
	ReadTimeout  int `mapstructure:"read_timeout"`
	WriteTimeout int `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type TelegramConfig struct {
	Token     string `mapstructure:"token"`
	ChannelID string `mapstructure:"channel_id"`
}

// LoadConfig loads local values from an optional .env file. Existing process
// environment variables take precedence, so the same code works in Docker and
// production without a .env file.
func LoadConfig(envFile string) (*Config, error) {
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("load %s: %w", envFile, err)
		}
	}

	var config Config
	var err error
	config.App.Env = os.Getenv("APP_ENV")
	if config.Server.Port, err = envInt("SERVER_PORT"); err != nil {
		return nil, err
	}
	if config.Server.ReadTimeout, err = envInt("SERVER_READ_TIMEOUT"); err != nil {
		return nil, err
	}
	if config.Server.WriteTimeout, err = envInt("SERVER_WRITE_TIMEOUT"); err != nil {
		return nil, err
	}
	config.Database.Host = os.Getenv("DATABASE_HOST")
	if config.Database.Port, err = envInt("DATABASE_PORT"); err != nil {
		return nil, err
	}
	config.Database.User = os.Getenv("DATABASE_USER")
	config.Database.Password = os.Getenv("DATABASE_PASSWORD")
	config.Database.Name = os.Getenv("DATABASE_NAME")
	config.Database.SSLMode = os.Getenv("DATABASE_SSLMODE")
	config.Redis.Addr = os.Getenv("REDIS_ADDR")
	config.Redis.Password = os.Getenv("REDIS_PASSWORD")
	if config.Redis.DB, err = envInt("REDIS_DB"); err != nil {
		return nil, err
	}
	config.Telegram.Token = os.Getenv("TELEGRAM_TOKEN")
	config.Telegram.ChannelID = os.Getenv("TELEGRAM_CHANNEL_ID")
	config.Auth.JWTSecret = os.Getenv("AUTH_JWT_SECRET")
	if config.Auth.AccessTokenTTL, err = envInt("AUTH_ACCESS_TOKEN_TTL"); err != nil {
		return nil, err
	}
	if config.Auth.RefreshTokenTTL, err = envInt("AUTH_REFRESH_TOKEN_TTL"); err != nil {
		return nil, err
	}
	if config.Ingestor.BatchSize, err = envInt("INGESTOR_BATCH_SIZE"); err != nil {
		return nil, err
	}
	if config.Ingestor.FlushIntervalMS, err = envInt("INGESTOR_FLUSH_INTERVAL_MS"); err != nil {
		return nil, err
	}
	if config.Ingestor.BufferSize, err = envInt("INGESTOR_BUFFER_SIZE"); err != nil {
		return nil, err
	}
	if raw := os.Getenv("EXCHANGES"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &config.Exchanges); err != nil {
			return nil, fmt.Errorf("parse EXCHANGES: %w", err)
		}
	}
	config.applyDefaults()
	return &config, nil
}

func envInt(key string) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return value, nil
}

// Default returns a complete development configuration.
func Default() *Config { config := &Config{}; config.applyDefaults(); return config }

// applyDefaults fills in sensible fallbacks for optional settings so the app
// stays runnable in development with a minimal .env file.
func (c *Config) applyDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 9000
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 10
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 10
	}
	if c.Auth.JWTSecret == "" {
		c.Auth.JWTSecret = "dev-insecure-secret-change-me"
	}
	if c.Auth.AccessTokenTTL == 0 {
		c.Auth.AccessTokenTTL = 15
	}
	if c.Auth.RefreshTokenTTL == 0 {
		c.Auth.RefreshTokenTTL = 24 * 7
	}
	if c.Ingestor.BatchSize == 0 {
		c.Ingestor.BatchSize = 500
	}
	if c.Ingestor.FlushIntervalMS == 0 {
		c.Ingestor.FlushIntervalMS = 1000
	}
	if c.Ingestor.BufferSize == 0 {
		c.Ingestor.BufferSize = 10000
	}
}

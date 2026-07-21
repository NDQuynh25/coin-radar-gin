package config

import (
	"strings"

	"github.com/spf13/viper"
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

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	config.applyDefaults()
	return &config, nil
}

// Default returns a complete development configuration.
func Default() *Config { config := &Config{}; config.applyDefaults(); return config }

// applyDefaults fills in sensible fallbacks for optional settings so the app
// stays runnable in development even without a full config.yaml.
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

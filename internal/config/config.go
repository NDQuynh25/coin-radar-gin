package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Telegram TelegramConfig `mapstructure:"telegram"`
	Auth     AuthConfig     `mapstructure:"auth"`
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

// applyDefaults fills in sensible fallbacks for optional settings so the app
// stays runnable in development even without a full config.yaml.
func (c *Config) applyDefaults() {
	if c.Auth.JWTSecret == "" {
		c.Auth.JWTSecret = "dev-insecure-secret-change-me"
	}
	if c.Auth.AccessTokenTTL == 0 {
		c.Auth.AccessTokenTTL = 15 // 15 minutes
	}
	if c.Auth.RefreshTokenTTL == 0 {
		c.Auth.RefreshTokenTTL = 24 * 7 // 7 days
	}
}

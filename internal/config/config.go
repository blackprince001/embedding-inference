package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	TEI    TEIConfig    `mapstructure:"tei"`
	Client ClientConfig `mapstructure:"client"`
	GRPC   GRPCConfig   `mapstructure:"grpc"`
	Log    LogConfig    `mapstructure:"log"`
}

type GRPCConfig struct {
	Port int `mapstructure:"port"`
}

type TEIConfig struct {
	BaseURL        string        `mapstructure:"base_url" validate:"required,url"`
	Timeout        time.Duration `mapstructure:"timeout"`
	MaxRetries     int           `mapstructure:"max_retries"`
	RetryDelay     time.Duration `mapstructure:"retry_delay"`
	MaxConnections int           `mapstructure:"max_connections"`
}

type ClientConfig struct {
	Name           string        `mapstructure:"name"`
	Version        string        `mapstructure:"version"`
	DefaultTimeout time.Duration `mapstructure:"default_timeout"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("docker")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	setDefaults()
	setGRPCDefaults()

	viper.AutomaticEnv()
	viper.SetEnvPrefix("TEI_CLIENT")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	viper.SetDefault("tei.base_url", "http://text-embeddings-inference:8080")
	viper.SetDefault("tei.timeout", "30s")
	viper.SetDefault("tei.max_retries", 3)
	viper.SetDefault("tei.retry_delay", "1s")
	viper.SetDefault("tei.max_connections", 10)

	viper.SetDefault("client.name", "text-embeddings-client")
	viper.SetDefault("client.version", "1.0.0")
	viper.SetDefault("client.default_timeout", "30s")

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
}

func setGRPCDefaults() {
	// gRPC server defaults
	viper.SetDefault("grpc.port", "9090")
}

func (c *Config) Validate() error {
	if c.TEI.BaseURL == "" {
		return fmt.Errorf("tei.base_url is required")
	}

	if c.TEI.Timeout <= 0 {
		return fmt.Errorf("tei.timeout must be positive")
	}

	if c.TEI.MaxRetries < 0 {
		return fmt.Errorf("tei.max_retries must be non-negative")
	}

	if c.TEI.MaxConnections <= 0 {
		return fmt.Errorf("tei.max_connections must be positive")
	}

	return nil
}

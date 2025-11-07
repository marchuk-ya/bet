package configs

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server    ServerConfig
	RateLimit RateLimitConfig
}

type ServerConfig struct {
	Port         int
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

type RateLimitConfig struct {
	RequestsPerMinute int
}

type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("configuration error for field '%s': %s", e.Field, e.Message)
}

func LoadConfig() (*Config, error) {
	port, err := getEnvAsInt("SERVER_PORT", 8080)
	if err != nil {
		return nil, &ConfigError{
			Field:   "SERVER_PORT",
			Message: fmt.Sprintf("invalid port: %v", err),
		}
	}

	readTimeout, err := getEnvAsInt("SERVER_READ_TIMEOUT", 15)
	if err != nil {
		return nil, &ConfigError{
			Field:   "SERVER_READ_TIMEOUT",
			Message: fmt.Sprintf("invalid read timeout: %v", err),
		}
	}

	writeTimeout, err := getEnvAsInt("SERVER_WRITE_TIMEOUT", 15)
	if err != nil {
		return nil, &ConfigError{
			Field:   "SERVER_WRITE_TIMEOUT",
			Message: fmt.Sprintf("invalid write timeout: %v", err),
		}
	}

	idleTimeout, err := getEnvAsInt("SERVER_IDLE_TIMEOUT", 60)
	if err != nil {
		return nil, &ConfigError{
			Field:   "SERVER_IDLE_TIMEOUT",
			Message: fmt.Sprintf("invalid idle timeout: %v", err),
		}
	}

	rateLimit, err := getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 60)
	if err != nil {
		return nil, &ConfigError{
			Field:   "RATE_LIMIT_REQUESTS_PER_MINUTE",
			Message: fmt.Sprintf("invalid rate limit: %v", err),
		}
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:         port,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: rateLimit,
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if err := validateRange("SERVER_PORT", c.Server.Port, 1, 65535); err != nil {
		return err
	}

	if err := validateRange("SERVER_READ_TIMEOUT", c.Server.ReadTimeout, 1, 300); err != nil {
		return err
	}

	if err := validateRange("SERVER_WRITE_TIMEOUT", c.Server.WriteTimeout, 1, 300); err != nil {
		return err
	}

	if err := validateRange("SERVER_IDLE_TIMEOUT", c.Server.IdleTimeout, 1, 600); err != nil {
		return err
	}

	if err := validateRange("RATE_LIMIT_REQUESTS_PER_MINUTE", c.RateLimit.RequestsPerMinute, 1, 10000); err != nil {
		return err
	}

	return nil
}

func validateRange(field string, value, min, max int) error {
	if value < min {
		return &ConfigError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d, got: %d", min, value),
		}
	}

	if value > max {
		return &ConfigError{
			Field:   field,
			Message: fmt.Sprintf("must not exceed %d, got: %d", max, value),
		}
	}

	return nil
}

func getEnvAsInt(key string, defaultValue int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s as integer: %w", key, err)
	}

	return intValue, nil
}

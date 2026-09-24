package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config 集中保存服务运行所需配置，全部通过环境变量注入。
type Config struct {
	AppEnv     string `env:"APP_ENV" envDefault:"development"`
	ServerPort string `env:"SERVER_PORT" envDefault:"3000"`

	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     string `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER" envDefault:"postgres"`
	DBPassword string `env:"DB_PASSWORD" envDefault:"postgres"`
	DBName     string `env:"DB_NAME" envDefault:"renovation_budget"`
	DBSSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`

	RedisURL string `env:"REDIS_URL" envDefault:"redis://localhost:6379/0"`

	JWTSecret      string `env:"JWT_SECRET" envDefault:"dev-secret-change-me"`
	JWTExpireHours int    `env:"JWT_EXPIRE_HOURS" envDefault:"24"`

	RateLimitMax           int `env:"RATE_LIMIT_MAX" envDefault:"100"`
	RateLimitWindowSeconds int `env:"RATE_LIMIT_WINDOW_SECONDS" envDefault:"60"`
}

// Load 从环境变量解析配置。
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// DSN 返回 PostgreSQL 连接串。
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

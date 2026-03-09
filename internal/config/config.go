package config

import "go-starter/internal/env"

type Config struct {
	DbUrl    string
	RedisUrl string
	AppPort  string
	JwtSecret string
}

func NewConfig() *Config {
	return &Config{
		DbUrl: env.GetString("POSTGRES_URL", ""),
		RedisUrl: env.GetString("REDIS_URL", ""),
		AppPort: env.GetString("PORT", ""),
		JwtSecret: env.GetString("JWT_SECRET", ""),
	}
}
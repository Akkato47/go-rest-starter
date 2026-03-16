package config

import "go-starter/internal/common/env"

type Config struct {
	DbUrl            string
	DbHost           string
	DbPort           string
	DbUser           string
	DbPassword       string
	DbName           string
	DbSslMode        string
	RedisUrl         string
	AppPort          string
	AppURL           string
	ProductionStatus bool
	AllowedOrigins   []string
	JwtSecret        string
	CsrfSecret       string
}

func NewConfig() *Config {
	return &Config{
		DbUrl:            env.GetString("POSTGRES_URL", ""),
		DbHost:           env.GetString("POSTGRES_HOST", ""),
		DbPort:           env.GetString("POSTGRES_PORT", ""),
		DbUser:           env.GetString("POSTGRES_USER", ""),
		DbPassword:       env.GetString("POSTGRES_PASSWORD", ""),
		DbName:           env.GetString("POSTGRES_NAME", ""),
		DbSslMode:        env.GetString("POSTGRES_SSL_MODE", ""),
		RedisUrl:         env.GetString("REDIS_URL", ""),
		AppPort:          env.GetString("APP_PORT", ""),
		AppURL:           env.GetString("APP_URL", ""),
		AllowedOrigins:   env.GetStrings("ALLOWED_ORIGINS", []string{}),
		ProductionStatus: getProductionStatus(env.GetString("GIN_MODE", "test")),
		JwtSecret:        env.GetString("JWT_SECRET", ""),
		CsrfSecret:       env.GetString("CSRF_SECRET", ""),
	}
}

func getProductionStatus(mode string) bool {
	if mode == "realese" {
		return true
	}
	return false
}

package config

import (
	_ "github.com/joho/godotenv/autoload"
	"os"
	"strconv"
)

type Config struct {
	DB    PostgresConfig
	EMAIL EmailConfig
}

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}
type PostgresConfig struct {
	URL string
}

func LoadConfig() (*Config, error) {
	emailPort, _ := strconv.Atoi(os.Getenv("EMAIL_PORT"))

	cfg := &Config{
		DB: PostgresConfig{
			URL: os.Getenv("DATABASE_URL"),
		},
		EMAIL: EmailConfig{
			Host:     os.Getenv("EMAIL_HOST"),
			Port:     emailPort,
			Username: os.Getenv("EMAIL_USERNAME"),
			Password: os.Getenv("EMAIL_PASSWORD"),
		},
	}

	return cfg, nil
}

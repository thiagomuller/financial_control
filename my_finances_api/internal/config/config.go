package config

import "os"

type Config struct {
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	JWTSecret  string
	Port       string
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "my_finances"),
		DBUser:     getEnv("DB_USER", "finances_user"),
		DBPassword: getEnv("DB_PASSWORD", "finances_pass"),
		JWTSecret:  getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		Port:       getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

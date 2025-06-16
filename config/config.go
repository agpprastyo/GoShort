package config

import (
	"io"
	"time"
)

// AppConfig holds all application configuration
type AppConfig struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	Logger      LoggerConfig
	JWT         JWT
	SwaggerAuth SwaggerAuthConfig
	BasicAuth   BasicAuthConfig
}

type BasicAuthConfig struct {
	Username string
	Password string
}

type DatabaseConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	DBName      string
	SSLMode     string
	MaxOpenConn int
	MaxIdleConn int
	MaxLifetime time.Duration
}

type JWT struct {
	Secret   string
	Expire   time.Duration
	Issuer   string
	Audience string
}

type LoggerConfig struct {
	Level      string
	JSONFormat bool
	Output     io.Writer
}

// ServerConfig holds app configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	BaseURL      string
}

// RedisConfig Config holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

// SwaggerAuthConfig holds configuration for Swagger authentication
type SwaggerAuthConfig struct {
	Username string
	Password string
}

// Load reads configuration from environment variables with defaults
func Load() *AppConfig {
	return &AppConfig{
		BasicAuth: BasicAuthConfig{
			Username: getEnv("BASIC_AUTH_USERNAME", "admin"),
			Password: getEnv("BASIC_AUTH_PASSWORD", "admin123"),
		},
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			BaseURL:      getEnv("SERVER_BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			Host:        getEnv("DB_HOST", "localhost"),
			Port:        getEnv("DB_PORT", "5432"),
			User:        getEnv("DB_USER", "postgres"),
			Password:    getEnv("DB_PASSWORD", "postgres"),
			DBName:      getEnv("DB_NAME", "goshort"),
			SSLMode:     getEnv("DB_SSLMODE", "disable"),
			MaxOpenConn: getInt("DB_MAX_OPEN_CONN", 10),
			MaxIdleConn: getInt("DB_MAX_IDLE_CONN", 5),
			MaxLifetime: getDuration("DB_MAX_LIFETIME", 5*time.Minute),
		},
		JWT: JWT{
			Secret:   getEnv("JWT_SECRET", "defaultsecret"),
			Expire:   getDuration("JWT_EXPIRE", 24*time.Hour),
			Issuer:   getEnv("JWT_ISSUER", "goshort"),
			Audience: getEnv("JWT_AUDIENCE", "goshort"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getInt("REDIS_DB", 0),
			PoolSize: getInt("REDIS_POOL_SIZE", 10),
		},

		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			JSONFormat: getBool("LOG_JSON", false),
		},
		SwaggerAuth: SwaggerAuthConfig{
			Username: getEnv("SWAGGER_AUTH_USERNAME", "admin"),
			Password: getEnv("SWAGGER_AUTH_PASSWORD", "admin123"),
		},
	}
}

package configs

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	POSTGRES     PostgresConfig
	RabbitMQ     RabbitMqConfig
	JWT          JwtConfig
	App          AppConfig
	REDIS_URL    string
	FRONTEND_URL string
	COOKIE_NAME  string
}

type AppConfig struct {
	NAME    string
	VERSION string
	PORT    string
	GO_ENV  string
}

type PostgresConfig struct {
	POSTGRES_URL      string
	POSTGRES_HOST     string
	POSTGRES_PORT     string
	POSTGRES_DB       string
	POSTGRES_USER     string
	POSTGRES_PASSWORD string
}

type RabbitMqConfig struct {
	RABBITMQ_URL             string
	RABBITMQ_PORT            string
	RABBITMQ_MANAGEMENT_PORT string
	RABBITMQ_USER            string
	RABBITMQ_PASSWORD        string
}

type JwtConfig struct {
	JWT_SECRET                string
	REFRESH_JWT_SECRET        string
	JWT_EXPIRY_HOURS_REGISTER time.Duration
	JWT_EXPIRY_HOURS_LOGIN    time.Duration
	JWT_EXPIRY_HOURS_REFRESH  time.Duration
}

var Configs Config

func LoadConfig() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env.dev"
	}
	// In Docker/production, environment variables may be injected directly
	// and an env file may not exist inside the container filesystem.
	if err := godotenv.Load(envFile); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: failed to load env file %s: %v", envFile, err)
	}
	registerHours, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS_REGISTER", "15"))
	loginHours, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS_LOGIN", "7"))
	refreshHours, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS_REFRESH", "30"))
	jwtExpiryRegister := time.Duration(registerHours) * time.Minute  // 15 Minutes
	jwtExpiryLogin := time.Duration(loginHours) * 24 * time.Hour     // 7 Days
	jwtExpiryRefresh := time.Duration(refreshHours) * 24 * time.Hour // 30 Days

	Configs = Config{
		POSTGRES: PostgresConfig{
			POSTGRES_URL:      getEnv("POSTGRES_URL", "postgresql://taskio:taskio@localhost:5432/taskio"),
			POSTGRES_DB:       getEnv("POSTGRES_DB", "taskio"),
			POSTGRES_HOST:     getEnv("POSTGRES_HOST", "localhost"),
			POSTGRES_PORT:     getEnv("POSTGRES_PORT", "5432"),
			POSTGRES_USER:     getEnv("POSTGRES_USER", "taskio"),
			POSTGRES_PASSWORD: getEnv("POSTGRES_PASSWORD", "taskio"),
		},
		RabbitMQ: RabbitMqConfig{
			RABBITMQ_URL:             getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672"),
			RABBITMQ_PORT:            getEnv("RABBITMQ_PORT", "5672"),
			RABBITMQ_MANAGEMENT_PORT: getEnv("RABBITMQ_MANAGEMENT_PORT", "15672"),
			RABBITMQ_USER:            getEnv("RABBITMQ_USER", "guest"),
			RABBITMQ_PASSWORD:        getEnv("RABBITMQ_PASSWORD", "guest"),
		},
		App: AppConfig{
			NAME:    getEnv("NAME", "taskio"),
			PORT:    getEnv("PORT", ":3000"),
			VERSION: getEnv("VERSION", "1.0.0.0"),
			GO_ENV:  getEnv("GO_ENV", "development"),
		},
		JWT: JwtConfig{
			JWT_SECRET:                getEnv("JWT_SECRET", "super-safe-jwt-secret"),
			JWT_EXPIRY_HOURS_REGISTER: jwtExpiryRegister,
			JWT_EXPIRY_HOURS_LOGIN:    jwtExpiryLogin,
			JWT_EXPIRY_HOURS_REFRESH:  jwtExpiryRefresh,
		},
		REDIS_URL:    getEnv("REDIS_URL", "redis://redis:6379"),
		FRONTEND_URL: getEnv("FRONTEND_URL", "http://localhost:5173"),
		COOKIE_NAME:  getEnv("COOKIE_NAME", "taskio-cookie"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	} else {
		return defaultValue
	}
}

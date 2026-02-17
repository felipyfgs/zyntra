package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DefaultConfig returns config from environment variables
func DefaultConfig() *Config {
	return &Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "zyntra"),
		Password: getEnv("DB_PASSWORD", "zyntra_secret"),
		DBName:   getEnv("DB_NAME", "zyntra"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// ConnectionString returns the PostgreSQL connection string
func (c *Config) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// DSN returns the PostgreSQL DSN for pgx
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}

// DB wraps sql.DB with additional functionality
type DB struct {
	*sql.DB
	config *Config
}

// New creates a new database connection
func New(config *Config) (*DB, error) {
	if config == nil {
		config = DefaultConfig()
	}

	log.Printf("[Database] Connecting to %s:%s/%s...", config.Host, config.Port, config.DBName)

	db, err := sql.Open("pgx", config.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Test connection with retry
	var lastErr error
	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = db.PingContext(ctx)
		cancel()

		if err == nil {
			log.Printf("[Database] Connected successfully")
			return &DB{DB: db, config: config}, nil
		}

		lastErr = err
		log.Printf("[Database] Connection attempt %d failed: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect after 5 attempts: %w", lastErr)
}

// Config returns the database configuration
func (db *DB) Config() *Config {
	return db.config
}

// HealthCheck performs a health check on the database
func (db *DB) HealthCheck(ctx context.Context) error {
	return db.PingContext(ctx)
}

// Close closes the database connection
func (db *DB) Close() error {
	log.Printf("[Database] Closing connection...")
	return db.DB.Close()
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

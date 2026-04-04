package database

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the database connection configuration.
type Config struct {
	Driver   string
	Host     string
	Port     int
	Username string
	Password string
	DBName   string
	SSLMode  string
}

// BuildDSN generates the Data Source Name string for the database connection.
func (c *Config) BuildDSN() string {
	switch c.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.DBName, c.SSLMode)
	default:
		return ""
	}
}

// NewConfig loads the configuration from environment variables.
func NewConfig() *Config {
	portStr := os.Getenv("DB_PORT")
	port, _ := strconv.Atoi(portStr)

	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "postgres" // default to postgres if not specified
	}
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	return &Config{
		Driver:   driver,
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  sslmode,
	}
}

package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerHostPort string
	LogLevel       string
	DatabaseDsn    string
	IntegrationDsn string
}

func NewConfig() *Config {
	serverHostPort := flag.String("a", "localhost:8888", "server host:port")
	logLevel := flag.String("l", "info", "log level")
	databaseDsn := flag.String("d", "localhost:5432", "databse destination (host/host:port)")
	integrationDsn := flag.String("i", "localhost:2345", "integration destination (host:port)")
	flag.Parse()
	if envServerHostPort := os.Getenv("SERVER_ADDRESS"); envServerHostPort != "" {
		*serverHostPort = envServerHostPort
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		*logLevel = envLogLevel
	}
	if envDatabaseDsn := os.Getenv("DATABASE_DSN"); envDatabaseDsn != "" {
		*databaseDsn = envDatabaseDsn
	}
	if envIntegrationDsn := os.Getenv("INTEGRATION_DSN"); envIntegrationDsn != "" {
		*integrationDsn = envIntegrationDsn
	}
	conf := &Config{
		ServerHostPort: *serverHostPort,
		LogLevel:       *logLevel,
		DatabaseDsn:    *databaseDsn,
		IntegrationDsn: *integrationDsn,
	}
	return conf
}

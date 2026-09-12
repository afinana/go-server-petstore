package petstore

import (
	"os"
	"strings"
)

type Config struct {
	ServerAddr           string
	MongoURI             string
	MongoDatabase        string
	EnableCredentials    bool
	MongoUsername        string
	MongoPassword        string
	EnableAuthValidation bool
	GatewaySecret        string
}

func LoadConfig() *Config {
	serverAddr := os.Getenv("SERVER_ADDR")
	if serverAddr == "" {
		serverAddr = os.Getenv("serverAddr")
	}
	if serverAddr == "" {
		serverAddr = "localhost:8080"
	}

	mongoURI := os.Getenv("DATABASE_URI")
	if mongoURI == "" {
		mongoURI = os.Getenv("databaseURI")
	}
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDatabase := os.Getenv("DATABASE_NAME")
	if mongoDatabase == "" {
		mongoDatabase = os.Getenv("MONGODB_DATABASE")
	}
	if mongoDatabase == "" {
		mongoDatabase = "petstore"
	}

	enableCredentials := os.Getenv("ENABLE_CREDENTIALS") == "true"

	// Default value for auth validation is enabled (true)
	enableAuthVal := true
	if val := os.Getenv("ENABLE_AUTH_VALIDATION"); val != "" {
		lower := strings.ToLower(strings.TrimSpace(val))
		enableAuthVal = lower != "false" && lower != "0" && lower != "off" && lower != "no"
	} else if val := os.Getenv("ENABLE_HEADER_VALIDATION"); val != "" {
		lower := strings.ToLower(strings.TrimSpace(val))
		enableAuthVal = lower != "false" && lower != "0" && lower != "off" && lower != "no"
	}

	return &Config{
		ServerAddr:           serverAddr,
		MongoURI:             mongoURI,
		MongoDatabase:        mongoDatabase,
		EnableCredentials:    enableCredentials,
		MongoUsername:        os.Getenv("MONGODB_USERNAME"),
		MongoPassword:        os.Getenv("MONGODB_PASSWORD"),
		EnableAuthValidation: enableAuthVal,
		GatewaySecret:        os.Getenv("GATEWAY_SECRET"),
	}
}

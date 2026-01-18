package petstore

import (
	"os"
)

type Config struct {
	ServerAddr        string
	MongoURI          string
	MongoDatabase     string
	EnableCredentials bool
	MongoUsername     string
	MongoPassword     string
}

func LoadConfig() *Config {
	serverAddr := os.Getenv("serverAddr")
	if serverAddr == "" {
		serverAddr = "localhost:8080"
	}

	mongoURI := os.Getenv("databaseURI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDatabase := os.Getenv("MONGODB_DATABASE")
	if mongoDatabase == "" {
		mongoDatabase = "petstore"
	}

	enableCredentials := os.Getenv("ENABLE_CREDENTIALS") == "true"

	return &Config{
		ServerAddr:        serverAddr,
		MongoURI:          mongoURI,
		MongoDatabase:     mongoDatabase,
		EnableCredentials: enableCredentials,
		MongoUsername:     os.Getenv("MONGODB_USERNAME"),
		MongoPassword:     os.Getenv("MONGODB_PASSWORD"),
	}
}

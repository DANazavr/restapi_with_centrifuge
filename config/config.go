package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Addr        string `json:"addr"`
	LogLevel    string `json:"log_level"`
	DatabaseURL string `json:"database_url"`
}

func NewConfig() *Config {
	return &Config{
		Addr:        ":8080",
		LogLevel:    "debug",
		DatabaseURL: "postgres://user:password@localhost:5432/dbname?sslmode=disable",
	}
}

func ParseConfig(filepath string) *Config {
	var c Config
	data, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	if err := json.Unmarshal(data, &c); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	return &c
}

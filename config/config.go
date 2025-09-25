package config

import (
	"encoding/json"
	"log"
	"os"
)

// HOW TO USE
// conn := config.Config["connectionStr"].(string)

var Config map[string]interface{}

func init() {
	loadConfig()
}

func loadConfig() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&Config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}
}

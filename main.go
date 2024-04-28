package main

import (
	"encoding/json"
	"os"
)

func main() {
	config, err := readConfig()
	if err != nil {
		panic(err)
	}

	CreateAndShowTable(config)
}

func readConfig() (Configuration, error) {
	fileContent, err := os.ReadFile("config.json")
	if err != nil {
		return Configuration{}, err
	}
	var config = Configuration{}

	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return Configuration{}, err
	}
	return config, nil
}

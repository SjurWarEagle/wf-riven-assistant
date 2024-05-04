package main

import (
	"encoding/json"
	"github.com/rivo/tview"
	"os"
)

func main() {
	config, err := readConfig()
	if err != nil {
		panic(err)
	}
	market := NewWfMarket()
	//errRiven := createRivenTable(config, market)
	//if errRiven != nil {
	//	panic(errRiven)
	//}

	offersTable, errOffers := CreateOffersTable(config, market)
	if errOffers != nil {
		panic(errOffers)
	}

	if err := tview.NewApplication().
		SetRoot(offersTable, true).
		EnableMouse(true).
		Run(); err != nil {
		panic(err)
	}

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

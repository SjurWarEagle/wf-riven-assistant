package main

import (
	"encoding/json"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"os"
)

func main() {
	config, err := readConfig()
	if err != nil {
		panic(err)
	}
	market := NewWfMarket()
	rivenTable, errRiven := createRivenTable(config, market)
	if errRiven != nil {
		panic(errRiven)
	}

	offersTable, errOffers := CreateOffersTable(config, market)
	if errOffers != nil {
		panic(errOffers)
	}

	app := tview.NewApplication()
	pages := tview.NewPages()
	//app.SetRoot(offersTable, true)
	app.EnableMouse(true)

	pages.AddPage("Rivens", rivenTable, true, true)
	pages.AddPage("Offers", offersTable, true, false)
	app.SetRoot(pages, true)

	pages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			currentPage, _ := pages.GetFrontPage()
			if currentPage == "Rivens" {
				pages.SwitchToPage("Offers")
			}
			if currentPage == "Offers" {
				pages.SwitchToPage("Rivens")
			}
		}
		return event
	})

	errRun := app.Run()
	if errRun != nil {
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

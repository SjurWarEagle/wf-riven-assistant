package main

import (
	"encoding/json"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

var OfferColumnWeaponName = 0
var OfferColumnUserPrice = 1
var OfferColumnMarketPrice = 2
var OfferColumnDelta = 3

func (wfm WfMarket) FindLowestPriceForItem(itemUrl string, config Configuration) (int, error) {
	time.Sleep(100 * time.Millisecond)

	url := fmt.Sprintf("https://api.warframe.market/v1/items/%s/orders", itemUrl)
	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("Platform", config.Setup.Platform)
	req.Header.Set("accept", "application/json")
	req.Header.Set("Language", "en")
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var ordersResponse OrdersResponse
	err2 := json.Unmarshal(responseData, &ordersResponse)
	if err2 != nil {
		fmt.Println(string(responseData))
		log.Fatal(err2)
	}

	minPrice := 999
	for i := range ordersResponse.Payload.Orders {
		order := ordersResponse.Payload.Orders[i]
		if order.User.Status == "offline" {
			continue
		}
		if order.User.Status == "online" {
			continue
		}
		if order.OrderType != "sell" {
			continue
		}
		if order.Platinum > 0 && order.Platinum < minPrice {
			minPrice = order.Platinum
		}
	}

	return minPrice, nil
}

func (wfm WfMarket) GetOffersOfUser(config Configuration) ([]PersonalOffer, error) {
	offers := make([]PersonalOffer, 0)
	user := config.Setup.WfMarketAccount
	//url := "https://api.warframe.market/v1/profile/SjurWarEagle/orders?include=profile"
	url := fmt.Sprintf("https://api.warframe.market/v1/profile/%s/orders", user)
	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("Platform", config.Setup.Platform)
	req.Header.Set("accept", "application/json")
	req.Header.Set("Language", "en")
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var offerResponse OfferResponse
	err2 := json.Unmarshal(responseData, &offerResponse)
	if err2 != nil {
		fmt.Println(string(responseData))
		log.Fatal(err2)
	}
	log.Printf("start collecting item prices")
	//todo remove reduction
	for i := range offerResponse.Payload.Offers {
		log.Printf("\tcollecting item prices for '%s'", offerResponse.Payload.Offers[i].Item.En.ItemName)

		lowestMarketPrice, errMarketPrice := wfm.FindLowestPriceForItem(offerResponse.Payload.Offers[i].Item.UrlName, config)
		if errMarketPrice != nil {
			return offers, errMarketPrice
		}
		delta := offerResponse.Payload.Offers[i].Platinum - lowestMarketPrice
		offers = append(offers, PersonalOffer{ItemName: offerResponse.Payload.Offers[i].Item.En.ItemName, UserPrice: offerResponse.Payload.Offers[i].Platinum, MarketMinPrice: lowestMarketPrice, Delta: delta})
	}
	log.Printf("done collecting item prices")

	return offers, nil
}
func CreateOffersTable(config Configuration, market *WfMarket) (*tview.Table, error) {
	var offersOfUser []PersonalOffer
	offersOfUser, err := market.GetOffersOfUser(config)
	if err != nil {
		return nil, err
	}

	//convertPersonalOfferToStructForTable(offersOfUser)
	//
	tableData := OfferTableData{
		config:               Configuration{},
		TableContentReadOnly: tview.TableContentReadOnly{},
		sortedColumn:         OfferColumnDelta,
		sortAsc:              false,
		data:                 offersOfUser,
	}
	table := tview.NewTable().
		SetBorders(true).
		SetSelectable(true, true).
		SetContent(tableData)

	table.Select(0, OfferColumnDelta).
		SetFixed(1, 1).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				os.Exit(0)
			}
		})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			row, column := table.GetSelection()
			if row == 0 {
				SortOfferData(row, column, &tableData)
			}
		}
		return event
	})

	SortOfferData(0, OfferColumnDelta, &tableData)

	return table, nil
}

type OfferTableData struct {
	config Configuration
	tview.TableContentReadOnly
	sortedColumn int
	sortAsc      bool
	data         []PersonalOffer
}

func (o OfferTableData) GetCell(row, column int) *tview.TableCell {
	value := ""
	align := tview.AlignRight
	cellColor := tcell.ColorWhite

	if row == 0 {
		cellColor = tcell.ColorYellow
	}
	if column == 0 {
		cellColor = tcell.ColorYellow
	}

	switch column {
	case OfferColumnWeaponName:
		if row == 0 {
			value = "    Weapon Name    "
			align = tview.AlignCenter
		} else {
			value = fmt.Sprintf("%-40s", o.data[row-1].ItemName)
			align = tview.AlignLeft
		}
		break
	case OfferColumnUserPrice:
		if row == 0 {
			value = "User"
			align = tview.AlignCenter
		} else {
			value = strconv.Itoa(o.data[row-1].UserPrice)
			align = tview.AlignRight
		}
		break
	case OfferColumnMarketPrice:
		if row == 0 {
			value = "Market"
			align = tview.AlignCenter
		} else {
			value = strconv.Itoa(o.data[row-1].MarketMinPrice)
			align = tview.AlignRight
		}
		break
	case OfferColumnDelta:
		if row == 0 {
			value = "Delta"
			align = tview.AlignCenter
		} else {
			marker := o.determinePriceMarker(row)

			delta := o.data[row-1].Delta
			value = fmt.Sprintf("%s %3d", marker, delta)
			align = tview.AlignRight
			if delta >= 5 {
				cellColor = tcell.ColorGreen
			} else if delta <= -5 {
				cellColor = tcell.ColorRed
			}
		}
		break
	}
	return tview.NewTableCell(fmt.Sprintf("%s", value)).SetTextColor(cellColor).SetAlign(align)
}

func (o OfferTableData) determinePriceMarker(row int) string {
	var marker = " "
	delta := o.data[row-1].Delta
	if delta >= 5 {
		marker = "▲"
	}
	if delta <= -5 {
		marker = "▼"
	}

	return marker
}

func (o OfferTableData) GetRowCount() int {
	//+1 due to header
	return len(o.data) + 1
}

func (o OfferTableData) GetColumnCount() int {
	return 4
}

func SortOfferData(row int, column int, data *OfferTableData) {
	if row != 0 {
		return
	}
	dataToSort := data.data
	if data.sortedColumn == column {
		data.sortAsc = !data.sortAsc
	} else {
		data.sortAsc = true
	}
	data.sortedColumn = column

	sort.Slice(dataToSort, func(i, j int) bool {
		if column == OfferColumnWeaponName {
			if data.sortAsc {
				return dataToSort[i].ItemName < dataToSort[j].ItemName
			}
			return dataToSort[i].ItemName > dataToSort[j].ItemName
		} else if column == OfferColumnUserPrice {
			if data.sortAsc {
				return dataToSort[i].UserPrice < dataToSort[j].UserPrice
			}
			return dataToSort[i].UserPrice > dataToSort[j].UserPrice
		} else if column == OfferColumnMarketPrice {
			if data.sortAsc {
				return dataToSort[i].MarketMinPrice < dataToSort[j].MarketMinPrice
			}
			return dataToSort[i].MarketMinPrice > dataToSort[j].MarketMinPrice
		} else if column == OfferColumnDelta {
			if data.sortAsc {
				return dataToSort[i].Delta < dataToSort[j].Delta
			}
			return dataToSort[i].Delta > dataToSort[j].Delta
		}
		return false
	})

	data.data = dataToSort
}

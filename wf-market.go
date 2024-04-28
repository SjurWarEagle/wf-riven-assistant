package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type WfMarket struct {
	items []Item
}

func NewWfMarket() *WfMarket {
	rc := WfMarket{}

	//rc.items = rc.fillKnownItems()

	return &rc

}

type RivenPrices struct {
	LowerSection_Cnt     int
	LowerSection_Average int
	LowerSection_Min     int
	LowerSection_Max     int
	All_Cnt              int
	all_Sum              int
	All_Average          int
	All_Min              int
	All_Max              int
}

type AuctionsResponse struct {
	Payload struct {
		Auction []Auction `json:"auctions"`
	} `json:"payload"`
}
type ItemResponse struct {
	Payload struct {
		Item []Item `json:"items"`
	} `json:"payload"`
}

type Auction struct {
	MinimalReputation int    `json:"minimal_reputation"`
	StartingPrice     int    `json:"starting_price"`
	BuyoutPrice       int    `json:"buyout_price"`
	Visible           bool   `json:"visible"`
	Note              string `json:"note"`
	Item              struct {
		WeaponUrlName string `json:"weapon_url_name"`
		Polarity      string `json:"polarity"`
		Name          string `json:"name"`
		Type          string `json:"type"`
		ReRolls       int    `json:"re_rolls"`
		ModRank       int    `json:"mod_rank"`
		Attributes    []struct {
			Value    float64 `json:"value"`
			Positive bool    `json:"positive"`
			UrlName  string  `json:"url_name"`
		} `json:"attributes"`
		MasteryLevel int `json:"mastery_level"`
	} `json:"item"`
	Owner struct {
		IngameName string      `json:"ingame_name"`
		LastSeen   time.Time   `json:"last_seen"`
		Reputation int         `json:"reputation"`
		Locale     string      `json:"locale"`
		Status     string      `json:"status"`
		Id         string      `json:"id"`
		Region     string      `json:"region"`
		Avatar     interface{} `json:"avatar"`
	} `json:"owner"`
	Platform          string      `json:"platform"`
	Closed            bool        `json:"closed"`
	TopBid            int         `json:"top_bid"`
	Winner            interface{} `json:"winner"`
	IsMarkedFor       interface{} `json:"is_marked_for"`
	MarkedOperationAt interface{} `json:"marked_operation_at"`
	Created           time.Time   `json:"created"`
	Updated           time.Time   `json:"updated"`
	NoteRaw           string      `json:"note_raw"`
	IsDirectSell      bool        `json:"is_direct_sell"`
	Id                string      `json:"id"`
	Private           bool        `json:"private"`
}
type Item struct {
	Thumb    string `json:"thumb"`
	Id       string `json:"id"`
	UrlName  string `json:"url_name"`
	ItemName string `json:"item_name"`
}

func (wfm WfMarket) requestAuctionPrice(itemName string, config Configuration) (RivenPrices, error) {

	itemUrl, err := wfm.getItemUrlByName(itemName)
	if err != nil {
		return RivenPrices{}, err
	}

	client := &http.Client{}
	// sorting by price to get cheap ones first
	// max number of results seems to be 500
	url := fmt.Sprintf("https://api.warframe.market/v1/auctions/search?type=riven&sort_by=price_asc&buyout_policy=direct&weapon_url_name=%s", itemUrl)
	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("Platform", config.Setup.Platform)
	req.Header.Set("accept", "application/json")
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	response, err := client.Do(req)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var auctionResponse AuctionsResponse
	err2 := json.Unmarshal(responseData, &auctionResponse)
	if err2 != nil {
		fmt.Println(string(responseData))
		log.Fatal(err2)
	}

	relevantActions := wfm.filterOnlyRelevant(auctionResponse.Payload.Auction)
	rivePrices := wfm.consolidatePrices(relevantActions)

	auctionItemsLower := extractLowerSection(relevantActions, 10)
	rivePricesLower := wfm.consolidatePrices(auctionItemsLower)

	rivePrices.LowerSection_Cnt = rivePricesLower.All_Cnt
	rivePrices.LowerSection_Min = rivePricesLower.All_Min
	rivePrices.LowerSection_Max = rivePricesLower.All_Max
	rivePrices.LowerSection_Average = rivePricesLower.All_Average

	return rivePrices, nil
}

func extractLowerSection(auctions []Auction, minCnt int) []Auction {
	//too few items, just do not filter
	if len(auctions) <= minCnt {
		return auctions
	}

	var distribution = make(map[int]int)
	for idx := range auctions {
		auction := auctions[idx]
		distribution[auction.StartingPrice]++
	}

	totals := 0
	lowerPriceBorder := 0
	//TODO there must be a nicer way to do this
	for i := range 10_000 {
		if distribution[i] != 0 {
			totals += distribution[i]
			lowerPriceBorder = i
			if totals > minCnt {
				break
			}
		}
	}

	if lowerPriceBorder == 0 {
		panic("lowerPriceBorder=0!")
	}
	// now collect all auctions lower than the price
	var rc []Auction
	for idx := range auctions {
		auction := auctions[idx]
		if auction.StartingPrice > lowerPriceBorder {
			continue
		}
		rc = append(rc, auction)
	}

	return rc
}

func (wfm WfMarket) consolidatePrices(auctions []Auction) RivenPrices {
	rivePrices := RivenPrices{}
	rivePrices.All_Min = 9999

	for idx := range auctions {
		auction := auctions[idx]

		if !relevant(auction) {
			continue
		}

		if auction.StartingPrice > 0 {
			rivePrices.All_Cnt++
			rivePrices.all_Sum += auction.StartingPrice

			if rivePrices.All_Min > auction.StartingPrice {
				rivePrices.All_Min = auction.StartingPrice
			}
			if rivePrices.All_Max < auction.StartingPrice {
				rivePrices.All_Max = auction.StartingPrice
			}
		}
	}
	if rivePrices.All_Cnt > 0 {
		rivePrices.All_Average = rivePrices.all_Sum / rivePrices.All_Cnt
	}

	return rivePrices
}

func relevant(auction Auction) bool {
	if strings.Compare(auction.Owner.Status, "ingame") != 0 && strings.Compare(auction.Owner.Status, "online") != 0 {
		return false
	}
	if !auction.IsDirectSell {
		return false
	}
	if auction.Closed {
		return false
	}

	return true
}

func (wfm WfMarket) getItemUrlByName(name string) (string, error) {
	converted := name
	converted = strings.ToLower(converted)
	converted = strings.ReplaceAll(converted, " ", "_")
	return converted, nil
}

func (wfm WfMarket) fillKnownItems() []Item {
	response, err := http.Get("https://api.warframe.market/v1/items")

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := io.ReadAll(response.Body)
	itemResponse := ItemResponse{}
	err2 := json.Unmarshal(responseData, &itemResponse)
	if err2 != nil {
		log.Fatal(err2)

	}
	return itemResponse.Payload.Item
}

func (wfm WfMarket) filterOnlyRelevant(auctions []Auction) []Auction {
	var rc []Auction
	for idx := range auctions {
		auction := auctions[idx]
		if relevant(auction) {
			rc = append(rc, auction)
		}
	}
	return rc
}

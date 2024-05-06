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

	return &rc
}

type RivenPrices struct {
	LowersectionCnt     int
	LowersectionAverage int
	LowersectionMin     int
	LowersectionMax     int
	AllCnt              int
	allSum              int
	AllAverage          int
	AllMin              int
	AllMax              int
}

type AuctionsResponse struct {
	Payload struct {
		Auction []Auction `json:"auctions"`
	} `json:"payload"`
}
type OrdersResponse struct {
	Payload struct {
		Orders []Order `json:"orders"`
	} `json:"payload"`
}
type OfferResponse struct {
	Payload struct {
		Offers []Offer `json:"sell_orders"`
	} `json:"payload"`
}
type ItemResponse struct {
	Payload struct {
		Item []Item `json:"items"`
	} `json:"payload"`
}

type PersonalOffer struct {
	ItemName       string
	UserPrice      int
	MarketMinPrice int
	Delta          int
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

type Order struct {
	Platinum  int    `json:"platinum"`
	OrderType string `json:"order_type"`
	Quantity  int    `json:"quantity"`
	Id        string `json:"id"`
	Visible   bool   `json:"visible"`
	User      struct {
		Status string `json:"status"`
	} `json:"user"`
}

type Offer struct {
	Id        string `json:"id"`
	Platinum  int    `json:"platinum"`
	Quantity  int    `json:"quantity"`
	Visible   bool   `json:"visible"`
	OrderType string `json:"order_type"`
	Item      struct {
		UrlName string `json:"url_name"`
		En      struct {
			ItemName string `json:"item_name"`
		} `json:"en"`
		De struct {
			ItemName string `json:"item_name"`
		} `json:"de"`
	} `json:"item"`
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

	rivePrices.LowersectionCnt = rivePricesLower.AllCnt
	rivePrices.LowersectionMin = rivePricesLower.AllMin
	rivePrices.LowersectionMax = rivePricesLower.AllMax
	rivePrices.LowersectionAverage = rivePricesLower.AllAverage

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
	rivePrices.AllMin = 9999

	for idx := range auctions {
		auction := auctions[idx]

		if !relevant(auction) {
			continue
		}

		if auction.StartingPrice > 0 {
			rivePrices.AllCnt++
			rivePrices.allSum += auction.StartingPrice

			if rivePrices.AllMin > auction.StartingPrice {
				rivePrices.AllMin = auction.StartingPrice
			}
			if rivePrices.AllMax < auction.StartingPrice {
				rivePrices.AllMax = auction.StartingPrice
			}
		}
	}
	if rivePrices.AllCnt > 0 {
		rivePrices.AllAverage = rivePrices.allSum / rivePrices.AllCnt
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

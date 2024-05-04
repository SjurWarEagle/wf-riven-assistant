package main

type Riven struct {
	Weapon     string `json:"weapon"`
	Attributes string `json:"attributes"`
}
type Configuration struct {
	Rivens []Riven `json:"rivens"`
	Setup  struct {
		WfMarketAccount                       string `json:"wf-market-account"`
		Platform                              string `json:"platform"`
		LowerSectionAverageHighlightThreshold int    `json:"lowerSectionAverageHighlightThreshold"`
	} `json:"setup"`
}

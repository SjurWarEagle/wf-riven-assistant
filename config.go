package main

type Riven struct {
	Weapon     string `json:"weapon"`
	Attributes string `json:"attributes"`
}
type Configuration struct {
	Rivens []Riven `json:"rivens"`
	Setup  struct {
		Platform                              string `json:"platform"`
		LowerSectionAverageHighlightThreshold int    `json:"lowerSectionAverageHighlightThreshold"`
	} `json:"setup"`
}

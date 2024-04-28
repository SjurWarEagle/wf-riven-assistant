package main

type Riven struct {
	Weapon     string `json:"weapon"`
	Attributes string `json:"attributes"`
}
type Configuration struct {
	Rivens []Riven `json:"rivens"`
}

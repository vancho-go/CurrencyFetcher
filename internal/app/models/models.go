package models

type Valute struct {
	CharCode string `xml:"CharCode"`
	Nominal  int    `xml:"Nominal"`
	Name     string `xml:"Name"`
	Value    string `xml:"Value"`
}

type ValuteCurs struct {
	Valutes []Valute `xml:"Valute"`
}

type APIGetValuteResponse struct {
	Name     string `json:"name"`
	CharCode string `json:"char_code"`
	Date     string `json:"date"`
	Value    string `json:"value"`
	Nominal  int    `json:"nominal"`
}

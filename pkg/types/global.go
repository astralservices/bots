package types

type Region struct {
	ID         string  `json:"id"`
	Flag       string  `json:"flag"`
	IP         string  `json:"ip"`
	City       string  `json:"city"`
	Country    string  `json:"country"`
	Region     string  `json:"region"`
	PrettyName string  `json:"prettyName"`
	Lat        float64 `json:"lat"`
	Long       float64 `json:"long"`
	MaxBots    int     `json:"maxBots"`
	Status     string  `json:"status"`

	Bots int `json:"bots"`
}

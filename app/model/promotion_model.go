package model

type Promotion struct {
	Id         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Tiers      []float64 `json:"tiers"`
	ProductIds []string  `json:"product_ids"`
}

type Promotions struct {
	Promotions []Promotion `json:"promotions"`
}

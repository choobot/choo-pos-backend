package model

type Order struct {
	Id       string      `json:"id"`
	Items    []OrderItem `json:"items"`
	Total    float64     `json:"total"`
	Subtotal float64     `json:"subtotal"`
}

type OrderItem struct {
	Id      string  `json:"id"`
	Product Product `json:"product"`
	Price   float64 `json:"price"`
}

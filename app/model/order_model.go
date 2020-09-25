package model

type Order struct {
	Id        string      `json:"id"`
	Items     []OrderItem `json:"items" validate:"required"`
	Total     float64     `json:"total"`
	Subtotal  float64     `json:"subtotal"`
	Cash      float64     `json:"cash"`
	CreatedAt string      `json:"created_at"`
}

type OrderItem struct {
	Id      string  `json:"id"`
	Product Product `json:"product"`
	Price   float64 `json:"price"`
}

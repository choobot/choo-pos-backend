package model

type Product struct {
	Cover     string  `json:"cover"`
	Price     float32 `json:"price"`
	Title     string  `json:"title"`
	Id        string  `json:"id"`
	Status    int     `json:"status"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

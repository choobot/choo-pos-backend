package model

type Product struct {
	Id        string  `json:"id"`
	Cover     string  `json:"cover"`
	Price     float64 `json:"price"`
	Title     string  `json:"title"`
	Status    int     `json:"status"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type Books struct {
	Books []Product `json:"books"`
}

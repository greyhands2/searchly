package model

import "time"

type Product struct {
	ProductId   string    `json:"product_id"`
	ProductName string    `json:"product_name"`
	Price       float32   `json:"price"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

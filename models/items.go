package models

import (
	"time"
)

type Item struct {
	ID          int       `json:"id,omitempty"`
	UserID      int       `json:"user_id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

type StockTransaction struct {
	ID        int       `json:"id"`
	ItemID    int       `json:"item_id"`
	UserID    int       `json:"user_id"`
	Quantity  int       `json:"quantity"`
	Type      string    `json:"type"`
	ItemName  string    `json:"item_name"`
	CreatedAt time.Time `json:"created_at" swaggertype:"string"`
}

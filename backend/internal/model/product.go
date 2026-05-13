package model

import "time"

type Category struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	SortOrder int    `json:"sort_order"`
}

type ProductImage struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	ImagePath string `json:"image_path"`
	IsPrimary bool   `json:"is_primary"`
	SortOrder int    `json:"sort_order"`
}

type ProductSize struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	Size      string `json:"size"`
	StockQty  int    `json:"stock_qty"`
}

type Product struct {
	ID          int64          `json:"id"`
	CategoryID  int64          `json:"category_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Price       float64        `json:"price"`
	IsActive    bool           `json:"is_active"`
	IsOnSale    bool           `json:"is_on_sale"`
	CreatedAt   time.Time      `json:"created_at"`
	Images      []ProductImage `json:"images,omitempty"`
	Sizes       []ProductSize  `json:"sizes,omitempty"`
}

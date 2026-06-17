package model

import "time"

// Category is a top-level product grouping shown in the catalogue navigation.
type Category struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`       // URL-friendly identifier used in catalogue routes
	SortOrder int    `json:"sort_order"` // manual ordering for display in menus
	TypeName  string `json:"type_name"`  // singular display type, e.g. "Футболка" for category "Футболки"
}

// ProductImage is one image belonging to a product gallery.
type ProductImage struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	ImagePath string `json:"image_path"`
	IsPrimary bool   `json:"is_primary"` // marks the thumbnail/cover image for the product
	SortOrder int    `json:"sort_order"` // gallery display order
}

// ProductSize is a single size variant of a product and its current stock.
type ProductSize struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	Size      string `json:"size"`      // e.g. "S", "M", "42"
	StockQty  int    `json:"stock_qty"` // units currently in stock for this size
}

// Product is a catalogue item with its variants and gallery.
type Product struct {
	ID          int64          `json:"id"`
	CategoryID  int64          `json:"category_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Price       float64        `json:"price"`     // base price in rubles (float64)
	IsActive    bool           `json:"is_active"` // false hides the product from the storefront
	IsOnSale    bool           `json:"is_on_sale"`
	TypeName    string         `json:"type_name,omitempty"` // category's singular type, for "<тип> <название>" display
	CreatedAt   time.Time      `json:"created_at"`
	Images      []ProductImage `json:"images,omitempty"` // hydrated on demand; omitted when not loaded
	Sizes       []ProductSize  `json:"sizes,omitempty"`  // hydrated on demand; omitted when not loaded
}

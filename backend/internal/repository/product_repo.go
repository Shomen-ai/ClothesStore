// Package repository: data-access layer for products, categories, sizes and images.
package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"clothes-store/internal/model"
)

// ProductRepo provides queries over products and their related categories,
// product_sizes and product_images tables.
type ProductRepo struct{ db *sql.DB }

// NewProductRepo constructs a ProductRepo bound to db.
func NewProductRepo(db *sql.DB) *ProductRepo { return &ProductRepo{db: db} }

// GetCategories returns all categories in their configured display order.
func (r *ProductRepo) GetCategories() ([]model.Category, error) {
	rows, err := r.db.Query(`SELECT id,name,slug,sort_order,COALESCE(type_name,'') FROM categories ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cats := make([]model.Category, 0)
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.SortOrder, &c.TypeName); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cats, nil
}

// CreateCategory inserts a category and writes the generated id back into c.
func (r *ProductRepo) CreateCategory(c *model.Category) error {
	return r.db.QueryRow(
		`INSERT INTO categories(name,slug,sort_order) VALUES($1,$2,$3) RETURNING id`,
		c.Name, c.Slug, c.SortOrder,
	).Scan(&c.ID)
}

// UpdateCategory updates name, slug and sort order of an existing category.
func (r *ProductRepo) UpdateCategory(c *model.Category) error {
	_, err := r.db.Exec(`UPDATE categories SET name=$1,slug=$2,sort_order=$3 WHERE id=$4`,
		c.Name, c.Slug, c.SortOrder, c.ID)
	return err
}

// DeleteCategory removes a category by id.
func (r *ProductRepo) DeleteCategory(id int64) error {
	_, err := r.db.Exec(`DELETE FROM categories WHERE id=$1`, id)
	return err
}

// ProductFilter carries the optional catalog filters, sort key and pagination
// window for List. Zero/empty fields mean "no constraint".
type ProductFilter struct {
	CategoryID int64
	Size       string
	Search     string
	Sort       string
	PriceMin   float64
	PriceMax   float64
	Sale       bool
	Page       int
	PageSize   int
}

// List returns one page of active products matching the filter. The WHERE
// clauses, ORDER BY and LIMIT/OFFSET are all built dynamically: each optional
// predicate appends its own $N placeholder (the i counter tracks the next free
// index) so positional args stay aligned, and LIMIT/OFFSET take the final two
// placeholders ($i and $i+1). Images are batch-loaded afterwards via attachImages.
func (r *ProductRepo) List(f ProductFilter) ([]model.Product, error) {
	if f.Page < 1 {
		f.Page = 1 // default to first page
	}
	if f.PageSize < 1 {
		f.PageSize = 24 // default page size
	}
	offset := (f.Page - 1) * f.PageSize // page → SQL OFFSET

	where := []string{"p.is_active=true"} // always restrict to active products
	args := []any{}
	i := 1 // next positional-parameter index ($1, $2, …)

	if f.CategoryID > 0 {
		where = append(where, fmt.Sprintf("p.category_id=$%d", i))
		args = append(args, f.CategoryID)
		i++
	}
	if f.Search != "" {
		// Case-insensitive match on "<type> <name>", the bare name, and the category
		// name, so e.g. "Футболка", "Valor" and "Футболка Valor" all find the product.
		where = append(where, fmt.Sprintf(
			"(LOWER(COALESCE(c.type_name,'') || ' ' || p.name) LIKE $%d OR LOWER(p.name) LIKE $%d OR LOWER(c.name) LIKE $%d)", i, i, i))
		args = append(args, "%"+strings.ToLower(f.Search)+"%")
		i++
	}
	if f.Size != "" {
		// Keep only products that have the requested size in stock (correlated EXISTS).
		where = append(where, fmt.Sprintf(
			"EXISTS(SELECT 1 FROM product_sizes ps WHERE ps.product_id=p.id AND ps.size=$%d AND ps.stock_qty>0)", i))
		args = append(args, f.Size)
		i++
	}
	if f.PriceMin > 0 {
		where = append(where, fmt.Sprintf("p.price >= $%d", i))
		args = append(args, f.PriceMin)
		i++
	}
	if f.PriceMax > 0 {
		where = append(where, fmt.Sprintf("p.price <= $%d", i))
		args = append(args, f.PriceMax)
		i++
	}
	if f.Sale {
		where = append(where, "p.is_on_sale = true") // constant predicate, no placeholder
	}

	// Default ordering groups by category sort order then newest-first; the
	// price_asc/price_desc sort keys override it.
	orderBy := "c.sort_order ASC, p.created_at DESC"
	switch f.Sort {
	case "price_asc":
		orderBy = "p.price ASC"
	case "price_desc":
		orderBy = "p.price DESC"
	}

	// LEFT JOIN categories so products with no category still appear; the join
	// also exposes c.sort_order for the default ordering.
	query := fmt.Sprintf(
		`SELECT p.id,p.category_id,p.name,p.description,p.price,p.is_active,p.is_on_sale,COALESCE(c.type_name,''),p.created_at
		 FROM products p
		 LEFT JOIN categories c ON c.id = p.category_id
		 WHERE %s ORDER BY %s
		 LIMIT $%d OFFSET $%d`,
		strings.Join(where, " AND "), orderBy, i, i+1,
	)
	args = append(args, f.PageSize, offset) // bind the two trailing LIMIT/OFFSET placeholders

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]model.Product, 0)
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description,
			&p.Price, &p.IsActive, &p.IsOnSale, &p.TypeName, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.attachImages(products); err != nil {
		return nil, err
	}
	return products, nil
}

// attachImages fills the Images slice on each product in a single batched query.
func (r *ProductRepo) attachImages(products []model.Product) error {
	if len(products) == 0 {
		return nil
	}
	ids := make([]int64, len(products))
	idx := make(map[int64]int, len(products)) // product id → slice position, for O(1) fan-out
	for i, p := range products {
		ids[i] = p.ID
		idx[p.ID] = i
	}
	q, args := buildInQuery(
		`SELECT id,product_id,image_path,is_primary,sort_order FROM product_images WHERE product_id IN (%s) ORDER BY product_id, sort_order`,
		ids,
	)
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var img model.ProductImage
		if err := rows.Scan(&img.ID, &img.ProductID, &img.ImagePath, &img.IsPrimary, &img.SortOrder); err != nil {
			return err
		}
		if i, ok := idx[img.ProductID]; ok {
			products[i].Images = append(products[i].Images, img)
		}
	}
	return rows.Err()
}

// buildInQuery expands an "... IN (%s)" template into a parameterised IN list:
// it generates $1..$N placeholders for the ids and returns the finished SQL plus
// the matching args slice, keeping the query injection-safe.
func buildInQuery(template string, ids []int64) (string, []any) {
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	return fmt.Sprintf(template, strings.Join(placeholders, ",")), args
}

// GetByID loads a single product together with all of its sizes and images
// (three queries) and returns the fully assembled model.Product.
func (r *ProductRepo) GetByID(id int64) (*model.Product, error) {
	p := &model.Product{}
	err := r.db.QueryRow(
		`SELECT p.id,p.category_id,p.name,p.description,p.price,p.is_active,p.is_on_sale,COALESCE(c.type_name,''),p.created_at
		 FROM products p LEFT JOIN categories c ON c.id=p.category_id WHERE p.id=$1`, id,
	).Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.IsActive, &p.IsOnSale, &p.TypeName, &p.CreatedAt)
	if err != nil {
		return nil, err
	}

	sizeRows, err := r.db.Query(`SELECT id,product_id,size,stock_qty FROM product_sizes WHERE product_id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer sizeRows.Close()
	for sizeRows.Next() {
		var s model.ProductSize
		if err := sizeRows.Scan(&s.ID, &s.ProductID, &s.Size, &s.StockQty); err != nil {
			return nil, err
		}
		p.Sizes = append(p.Sizes, s)
	}
	if err := sizeRows.Err(); err != nil {
		return nil, err
	}

	imgRows, err := r.db.Query(
		`SELECT id,product_id,image_path,is_primary,sort_order FROM product_images WHERE product_id=$1 ORDER BY sort_order`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer imgRows.Close()
	for imgRows.Next() {
		var img model.ProductImage
		if err := imgRows.Scan(&img.ID, &img.ProductID, &img.ImagePath, &img.IsPrimary, &img.SortOrder); err != nil {
			return nil, err
		}
		p.Images = append(p.Images, img)
	}
	if err := imgRows.Err(); err != nil {
		return nil, err
	}
	return p, nil
}

// Create inserts a product and writes the generated id back into p.
func (r *ProductRepo) Create(p *model.Product) error {
	return r.db.QueryRow(
		`INSERT INTO products(category_id,name,description,price,is_active,is_on_sale)
		 VALUES($1,$2,$3,$4,$5,$6) RETURNING id`,
		p.CategoryID, p.Name, p.Description, p.Price, p.IsActive, p.IsOnSale,
	).Scan(&p.ID)
}

// Update overwrites the editable fields of an existing product.
func (r *ProductRepo) Update(p *model.Product) error {
	_, err := r.db.Exec(
		`UPDATE products SET category_id=$1,name=$2,description=$3,price=$4,is_active=$5,is_on_sale=$6 WHERE id=$7`,
		p.CategoryID, p.Name, p.Description, p.Price, p.IsActive, p.IsOnSale, p.ID,
	)
	return err
}

// Delete removes a product by id.
func (r *ProductRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM products WHERE id=$1`, id)
	return err
}

// UpsertSize inserts a size row or, if (product_id, size) already exists,
// overwrites its stock quantity with the proposed value. Returns the row id.
func (r *ProductRepo) UpsertSize(s *model.ProductSize) error {
	return r.db.QueryRow(
		`INSERT INTO product_sizes(product_id, size, stock_qty)
		 VALUES($1, $2, $3)
		 ON CONFLICT (product_id, size) DO UPDATE SET stock_qty = EXCLUDED.stock_qty
		 RETURNING id`,
		s.ProductID, s.Size, s.StockQty,
	).Scan(&s.ID)
}

// AddImage attaches an image to a product and writes the generated id back into img.
func (r *ProductRepo) AddImage(img *model.ProductImage) error {
	return r.db.QueryRow(
		`INSERT INTO product_images(product_id,image_path,is_primary,sort_order)
		 VALUES($1,$2,$3,$4) RETURNING id`,
		img.ProductID, img.ImagePath, img.IsPrimary, img.SortOrder,
	).Scan(&img.ID)
}

// DeleteImage removes an image owned by productID and returns its stored file
// path (read before the delete) so the caller can clean up the file on disk.
func (r *ProductRepo) DeleteImage(id, productID int64) (string, error) {
	var path string
	err := r.db.QueryRow(
		`SELECT image_path FROM product_images WHERE id=$1 AND product_id=$2`, id, productID,
	).Scan(&path)
	if err != nil {
		return "", err
	}
	_, err = r.db.Exec(`DELETE FROM product_images WHERE id=$1`, id)
	return path, err
}

// GetProductPriceForSize resolves the owning product id and current price for a
// product_sizes row, but only if the product is active and at least minQty units
// are in stock — used to validate a line item at checkout. Returns sql.ErrNoRows
// when the size is missing, inactive, or understocked.
func (r *ProductRepo) GetProductPriceForSize(sizeID int64, minQty int) (productID int64, price float64, err error) {
	err = r.db.QueryRow(
		`SELECT p.id, p.price FROM products p
		 JOIN product_sizes ps ON ps.product_id=p.id
		 WHERE ps.id=$1 AND p.is_active=true AND ps.stock_qty>=$2`,
		sizeID, minQty,
	).Scan(&productID, &price)
	return
}

// GetFeatured returns the two homepage rails: hits (top 5 best-sellers by total
// units ordered) and newest (5 most recently created active products). Images
// are attached to both slices before returning.
func (r *ProductRepo) GetFeatured() (hits []model.Product, newest []model.Product, err error) {
	hits = make([]model.Product, 0)
	newest = make([]model.Product, 0)
	// Join products against per-product ordered-quantity totals; ordering by that
	// summed quantity descending yields the best-sellers.
	hitRows, err := r.db.Query(
		`SELECT p.id,p.category_id,p.name,p.description,p.price,p.is_active,p.is_on_sale,COALESCE(c.type_name,''),p.created_at
		 FROM products p
		 JOIN (SELECT product_id, SUM(quantity) qty FROM order_items GROUP BY product_id) oi
		   ON oi.product_id=p.id
		 LEFT JOIN categories c ON c.id=p.category_id
		 WHERE p.is_active=true
		 ORDER BY oi.qty DESC
		 LIMIT 5`,
	)
	if err != nil {
		return nil, nil, err
	}
	defer hitRows.Close()
	for hitRows.Next() {
		var p model.Product
		if err := hitRows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description,
			&p.Price, &p.IsActive, &p.IsOnSale, &p.TypeName, &p.CreatedAt); err != nil {
			return nil, nil, err
		}
		hits = append(hits, p)
	}
	if err := hitRows.Err(); err != nil {
		return nil, nil, err
	}

	newRows, err := r.db.Query(
		`SELECT p.id,p.category_id,p.name,p.description,p.price,p.is_active,p.is_on_sale,COALESCE(c.type_name,''),p.created_at
		 FROM products p LEFT JOIN categories c ON c.id=p.category_id
		 WHERE p.is_active=true ORDER BY p.created_at DESC LIMIT 5`,
	)
	if err != nil {
		return nil, nil, err
	}
	defer newRows.Close()
	for newRows.Next() {
		var p model.Product
		if err := newRows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description,
			&p.Price, &p.IsActive, &p.IsOnSale, &p.TypeName, &p.CreatedAt); err != nil {
			return nil, nil, err
		}
		newest = append(newest, p)
	}
	if err := newRows.Err(); err != nil {
		return nil, nil, err
	}
	if err := r.attachImages(hits); err != nil {
		return nil, nil, err
	}
	if err := r.attachImages(newest); err != nil {
		return nil, nil, err
	}
	return hits, newest, nil
}

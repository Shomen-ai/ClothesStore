package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"clothes-store/internal/model"
)

type ProductRepo struct{ db *sql.DB }

func NewProductRepo(db *sql.DB) *ProductRepo { return &ProductRepo{db: db} }

func (r *ProductRepo) GetCategories() ([]model.Category, error) {
	rows, err := r.db.Query(`SELECT id,name,slug,sort_order FROM categories ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cats := make([]model.Category, 0)
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.SortOrder); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cats, nil
}

func (r *ProductRepo) CreateCategory(c *model.Category) error {
	return r.db.QueryRow(
		`INSERT INTO categories(name,slug,sort_order) VALUES($1,$2,$3) RETURNING id`,
		c.Name, c.Slug, c.SortOrder,
	).Scan(&c.ID)
}

func (r *ProductRepo) UpdateCategory(c *model.Category) error {
	_, err := r.db.Exec(`UPDATE categories SET name=$1,slug=$2,sort_order=$3 WHERE id=$4`,
		c.Name, c.Slug, c.SortOrder, c.ID)
	return err
}

func (r *ProductRepo) DeleteCategory(id int64) error {
	_, err := r.db.Exec(`DELETE FROM categories WHERE id=$1`, id)
	return err
}

type ProductFilter struct {
	CategoryID int64
	Size       string
	Search     string
	Sort       string
	Page       int
	PageSize   int
}

func (r *ProductRepo) List(f ProductFilter) ([]model.Product, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 24
	}
	offset := (f.Page - 1) * f.PageSize

	where := []string{"p.is_active=true"}
	args := []any{}
	i := 1

	if f.CategoryID > 0 {
		where = append(where, fmt.Sprintf("p.category_id=$%d", i))
		args = append(args, f.CategoryID)
		i++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("LOWER(p.name) LIKE $%d", i))
		args = append(args, "%"+strings.ToLower(f.Search)+"%")
		i++
	}
	if f.Size != "" {
		where = append(where, fmt.Sprintf(
			"EXISTS(SELECT 1 FROM product_sizes ps WHERE ps.product_id=p.id AND ps.size=$%d AND ps.stock_qty>0)", i))
		args = append(args, f.Size)
		i++
	}

	orderBy := "p.created_at DESC"
	switch f.Sort {
	case "price_asc":
		orderBy = "p.price ASC"
	case "price_desc":
		orderBy = "p.price DESC"
	}

	query := fmt.Sprintf(
		`SELECT p.id,p.category_id,p.name,p.description,p.price,p.is_active,p.created_at
		 FROM products p WHERE %s ORDER BY %s
		 LIMIT $%d OFFSET $%d`,
		strings.Join(where, " AND "), orderBy, i, i+1,
	)
	args = append(args, f.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]model.Product, 0)
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description,
			&p.Price, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepo) GetByID(id int64) (*model.Product, error) {
	p := &model.Product{}
	err := r.db.QueryRow(
		`SELECT id,category_id,name,description,price,is_active,created_at FROM products WHERE id=$1`, id,
	).Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.IsActive, &p.CreatedAt)
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

func (r *ProductRepo) Create(p *model.Product) error {
	return r.db.QueryRow(
		`INSERT INTO products(category_id,name,description,price,is_active)
		 VALUES($1,$2,$3,$4,$5) RETURNING id`,
		p.CategoryID, p.Name, p.Description, p.Price, p.IsActive,
	).Scan(&p.ID)
}

func (r *ProductRepo) Update(p *model.Product) error {
	_, err := r.db.Exec(
		`UPDATE products SET category_id=$1,name=$2,description=$3,price=$4,is_active=$5 WHERE id=$6`,
		p.CategoryID, p.Name, p.Description, p.Price, p.IsActive, p.ID,
	)
	return err
}

func (r *ProductRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM products WHERE id=$1`, id)
	return err
}

func (r *ProductRepo) UpsertSize(s *model.ProductSize) error {
	return r.db.QueryRow(
		`INSERT INTO product_sizes(product_id, size, stock_qty)
		 VALUES($1, $2, $3)
		 ON CONFLICT (product_id, size) DO UPDATE SET stock_qty = EXCLUDED.stock_qty
		 RETURNING id`,
		s.ProductID, s.Size, s.StockQty,
	).Scan(&s.ID)
}

func (r *ProductRepo) AddImage(img *model.ProductImage) error {
	return r.db.QueryRow(
		`INSERT INTO product_images(product_id,image_path,is_primary,sort_order)
		 VALUES($1,$2,$3,$4) RETURNING id`,
		img.ProductID, img.ImagePath, img.IsPrimary, img.SortOrder,
	).Scan(&img.ID)
}

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

func (r *ProductRepo) GetProductPriceForSize(sizeID int64, minQty int) (productID int64, price float64, err error) {
	err = r.db.QueryRow(
		`SELECT p.id, p.price FROM products p
		 JOIN product_sizes ps ON ps.product_id=p.id
		 WHERE ps.id=$1 AND p.is_active=true AND ps.stock_qty>=$2`,
		sizeID, minQty,
	).Scan(&productID, &price)
	return
}

func (r *ProductRepo) GetFeatured() (hits []model.Product, newest []model.Product, err error) {
	hits = make([]model.Product, 0)
	newest = make([]model.Product, 0)
	hitRows, err := r.db.Query(
		`SELECT p.id,p.category_id,p.name,p.description,p.price,p.is_active,p.created_at
		 FROM products p
		 JOIN (SELECT product_id, SUM(quantity) qty FROM order_items GROUP BY product_id) oi
		   ON oi.product_id=p.id
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
			&p.Price, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, nil, err
		}
		hits = append(hits, p)
	}
	if err := hitRows.Err(); err != nil {
		return nil, nil, err
	}

	newRows, err := r.db.Query(
		`SELECT id,category_id,name,description,price,is_active,created_at FROM products
		 WHERE is_active=true ORDER BY created_at DESC LIMIT 5`,
	)
	if err != nil {
		return nil, nil, err
	}
	defer newRows.Close()
	for newRows.Next() {
		var p model.Product
		if err := newRows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description,
			&p.Price, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, nil, err
		}
		newest = append(newest, p)
	}
	if err := newRows.Err(); err != nil {
		return nil, nil, err
	}
	return hits, newest, nil
}

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
		`INSERT INTO categories(name,slug,sort_order) VALUES(:1,:2,:3) RETURNING id INTO :4`,
		c.Name, c.Slug, c.SortOrder, &c.ID,
	).Err()
}

func (r *ProductRepo) UpdateCategory(c *model.Category) error {
	_, err := r.db.Exec(`UPDATE categories SET name=:1,slug=:2,sort_order=:3 WHERE id=:4`,
		c.Name, c.Slug, c.SortOrder, c.ID)
	return err
}

func (r *ProductRepo) DeleteCategory(id int64) error {
	_, err := r.db.Exec(`DELETE FROM categories WHERE id=:1`, id)
	return err
}

type ProductFilter struct {
	CategoryID int64
	Size       string
	Search     string
	Sort       string // price_asc | price_desc | newest
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

	where := []string{"p.is_active=1"}
	args := []any{}
	i := 1

	if f.CategoryID > 0 {
		where = append(where, fmt.Sprintf("p.category_id=:%d", i))
		args = append(args, f.CategoryID)
		i++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("LOWER(p.name) LIKE :%d", i))
		args = append(args, "%"+strings.ToLower(f.Search)+"%")
		i++
	}
	if f.Size != "" {
		where = append(where, fmt.Sprintf("EXISTS(SELECT 1 FROM product_sizes ps WHERE ps.product_id=p.id AND ps.size=:%d AND ps.stock_qty>0)", i))
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
		 OFFSET :%d ROWS FETCH NEXT :%d ROWS ONLY`,
		strings.Join(where, " AND "), orderBy, i, i+1,
	)
	args = append(args, offset, f.PageSize)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]model.Product, 0)
	for rows.Next() {
		var p model.Product
		var isActive int
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description,
			&p.Price, &isActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.IsActive = isActive == 1
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepo) GetByID(id int64) (*model.Product, error) {
	p := &model.Product{}
	var isActive int
	err := r.db.QueryRow(
		`SELECT id,category_id,name,description,price,is_active,created_at FROM products WHERE id=:1`, id,
	).Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &isActive, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	p.IsActive = isActive == 1

	sizeRows, err := r.db.Query(`SELECT id,product_id,size,stock_qty FROM product_sizes WHERE product_id=:1`, id)
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
		`SELECT id,product_id,image_path,is_primary,sort_order FROM product_images WHERE product_id=:1 ORDER BY sort_order`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer imgRows.Close()
	for imgRows.Next() {
		var img model.ProductImage
		var isPrimary int
		if err := imgRows.Scan(&img.ID, &img.ProductID, &img.ImagePath, &isPrimary, &img.SortOrder); err != nil {
			return nil, err
		}
		img.IsPrimary = isPrimary == 1
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
		 VALUES(:1,:2,:3,:4,:5) RETURNING id INTO :6`,
		p.CategoryID, p.Name, p.Description, p.Price, boolToInt(p.IsActive), &p.ID,
	).Err()
}

func (r *ProductRepo) Update(p *model.Product) error {
	_, err := r.db.Exec(
		`UPDATE products SET category_id=:1,name=:2,description=:3,price=:4,is_active=:5 WHERE id=:6`,
		p.CategoryID, p.Name, p.Description, p.Price, boolToInt(p.IsActive), p.ID,
	)
	return err
}

func (r *ProductRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM products WHERE id=:1`, id)
	return err
}

func (r *ProductRepo) UpsertSize(s *model.ProductSize) error {
	_, err := r.db.Exec(
		`MERGE INTO product_sizes dst
		 USING (SELECT :1 AS product_id, :2 AS size FROM dual) src
		   ON (dst.product_id = src.product_id AND dst.size = src.size)
		 WHEN MATCHED THEN UPDATE SET dst.stock_qty = :3
		 WHEN NOT MATCHED THEN INSERT (product_id, size, stock_qty) VALUES (:4, :5, :6)`,
		s.ProductID, s.Size, s.StockQty, s.ProductID, s.Size, s.StockQty,
	)
	if err != nil {
		return err
	}
	return r.db.QueryRow(
		`SELECT id FROM product_sizes WHERE product_id=:1 AND size=:2`, s.ProductID, s.Size,
	).Scan(&s.ID)
}

func (r *ProductRepo) AddImage(img *model.ProductImage) error {
	return r.db.QueryRow(
		`INSERT INTO product_images(product_id,image_path,is_primary,sort_order)
		 VALUES(:1,:2,:3,:4) RETURNING id INTO :5`,
		img.ProductID, img.ImagePath, boolToInt(img.IsPrimary), img.SortOrder, &img.ID,
	).Err()
}

func (r *ProductRepo) DeleteImage(id, productID int64) (string, error) {
	var path string
	err := r.db.QueryRow(`SELECT image_path FROM product_images WHERE id=:1 AND product_id=:2`, id, productID).Scan(&path)
	if err != nil {
		return "", err
	}
	_, err = r.db.Exec(`DELETE FROM product_images WHERE id=:1`, id)
	return path, err
}

func (r *ProductRepo) GetFeatured() (hits []model.Product, newest []model.Product, err error) {
	hits = make([]model.Product, 0)
	newest = make([]model.Product, 0)
	hitRows, err := r.db.Query(
		`SELECT p.id,p.category_id,p.name,p.description,p.price,p.is_active,p.created_at
		 FROM products p
		 JOIN (SELECT product_id, SUM(quantity) qty FROM order_items GROUP BY product_id) oi
		   ON oi.product_id=p.id
		 WHERE p.is_active=1
		 ORDER BY oi.qty DESC
		 FETCH FIRST 5 ROWS ONLY`,
	)
	if err != nil {
		return nil, nil, err
	}
	defer hitRows.Close()
	for hitRows.Next() {
		var p model.Product
		var isActive int
		if err := hitRows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description,
			&p.Price, &isActive, &p.CreatedAt); err != nil {
			return nil, nil, err
		}
		p.IsActive = isActive == 1
		hits = append(hits, p)
	}
	if err := hitRows.Err(); err != nil {
		return nil, nil, err
	}

	newRows, err := r.db.Query(
		`SELECT id,category_id,name,description,price,is_active,created_at FROM products
		 WHERE is_active=1 ORDER BY created_at DESC FETCH FIRST 5 ROWS ONLY`,
	)
	if err != nil {
		return nil, nil, err
	}
	defer newRows.Close()
	for newRows.Next() {
		var p model.Product
		var isActive int
		if err := newRows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description,
			&p.Price, &isActive, &p.CreatedAt); err != nil {
			return nil, nil, err
		}
		p.IsActive = isActive == 1
		newest = append(newest, p)
	}
	if err := newRows.Err(); err != nil {
		return nil, nil, err
	}
	return hits, newest, nil
}

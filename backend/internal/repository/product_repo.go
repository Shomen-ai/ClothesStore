package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"clothes-store/internal/model"
)

type ProductRepo struct{ db *sql.DB }

func NewProductRepo(db *sql.DB) *ProductRepo { return &ProductRepo{db: db} }

// ── Categories ──────────────────────────────────────────────────────────────

func (r *ProductRepo) GetCategories() ([]model.Category, error) {
	rows, err := r.db.Query(
		`SELECT id, name, slug, sort_order FROM categories ORDER BY sort_order, name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.SortOrder); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *ProductRepo) CreateCategory(c *model.Category) error {
	return r.db.QueryRow(
		`INSERT INTO categories(name, slug, sort_order) VALUES(:1,:2,:3) RETURNING id INTO :4`,
		c.Name, c.Slug, c.SortOrder, &c.ID,
	).Err()
}

func (r *ProductRepo) UpdateCategory(c *model.Category) error {
	_, err := r.db.Exec(
		`UPDATE categories SET name=:1, slug=:2, sort_order=:3 WHERE id=:4`,
		c.Name, c.Slug, c.SortOrder, c.ID,
	)
	return err
}

func (r *ProductRepo) DeleteCategory(id int64) error {
	_, err := r.db.Exec(`DELETE FROM categories WHERE id=:1`, id)
	return err
}

// ── Products ─────────────────────────────────────────────────────────────────

type ProductFilter struct {
	CategoryID *int64
	MinPrice   *float64
	MaxPrice   *float64
	Search     string
	Page       int
	PageSize   int
}

func (r *ProductRepo) List(f ProductFilter) ([]model.Product, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 20
	}

	where := []string{"p.is_active = 1"}
	args := []interface{}{}
	idx := 1

	if f.CategoryID != nil {
		where = append(where, fmt.Sprintf("p.category_id = :%d", idx))
		args = append(args, *f.CategoryID)
		idx++
	}
	if f.MinPrice != nil {
		where = append(where, fmt.Sprintf("p.price >= :%d", idx))
		args = append(args, *f.MinPrice)
		idx++
	}
	if f.MaxPrice != nil {
		where = append(where, fmt.Sprintf("p.price <= :%d", idx))
		args = append(args, *f.MaxPrice)
		idx++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("UPPER(p.name) LIKE UPPER(:%d)", idx))
		args = append(args, "%"+f.Search+"%")
		idx++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// total count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM products p %s`, whereClause)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.PageSize
	listArgs := append(args, f.PageSize, offset)
	listQuery := fmt.Sprintf(
		`SELECT p.id, p.category_id, p.name, p.description, p.price, p.is_active, p.created_at
		 FROM products p %s ORDER BY p.created_at DESC
		 OFFSET :%d ROWS FETCH NEXT :%d ROWS ONLY`,
		whereClause, idx, idx+1,
	)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		var active int
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &active, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		p.IsActive = active == 1
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// attach images and sizes
	for i := range products {
		if err := r.loadImages(&products[i]); err != nil {
			return nil, 0, err
		}
		if err := r.loadSizes(&products[i]); err != nil {
			return nil, 0, err
		}
	}

	return products, total, nil
}

func (r *ProductRepo) GetByID(id int64) (*model.Product, error) {
	p := &model.Product{}
	var active int
	err := r.db.QueryRow(
		`SELECT id, category_id, name, description, price, is_active, created_at
		 FROM products WHERE id=:1`,
		id,
	).Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &active, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	p.IsActive = active == 1
	if err := r.loadImages(p); err != nil {
		return nil, err
	}
	if err := r.loadSizes(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProductRepo) Create(p *model.Product) error {
	return r.db.QueryRow(
		`INSERT INTO products(category_id, name, description, price, is_active)
		 VALUES(:1,:2,:3,:4,:5) RETURNING id INTO :6`,
		p.CategoryID, p.Name, p.Description, p.Price, boolToInt(p.IsActive), &p.ID,
	).Err()
}

func (r *ProductRepo) Update(p *model.Product) error {
	_, err := r.db.Exec(
		`UPDATE products SET category_id=:1, name=:2, description=:3, price=:4, is_active=:5
		 WHERE id=:6`,
		p.CategoryID, p.Name, p.Description, p.Price, boolToInt(p.IsActive), p.ID,
	)
	return err
}

func (r *ProductRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM products WHERE id=:1`, id)
	return err
}

// ── Sizes ─────────────────────────────────────────────────────────────────────

func (r *ProductRepo) UpsertSize(s *model.ProductSize) error {
	var existing int64
	err := r.db.QueryRow(
		`SELECT id FROM product_sizes WHERE product_id=:1 AND size=:2`,
		s.ProductID, s.Size,
	).Scan(&existing)
	if err == sql.ErrNoRows {
		return r.db.QueryRow(
			`INSERT INTO product_sizes(product_id, size, stock_qty) VALUES(:1,:2,:3) RETURNING id INTO :4`,
			s.ProductID, s.Size, s.StockQty, &s.ID,
		).Err()
	}
	if err != nil {
		return err
	}
	s.ID = existing
	_, err = r.db.Exec(
		`UPDATE product_sizes SET stock_qty=:1 WHERE id=:2`,
		s.StockQty, s.ID,
	)
	return err
}

// ── Images ────────────────────────────────────────────────────────────────────

func (r *ProductRepo) AddImage(img *model.ProductImage) error {
	return r.db.QueryRow(
		`INSERT INTO product_images(product_id, image_path, is_primary, sort_order)
		 VALUES(:1,:2,:3,:4) RETURNING id INTO :5`,
		img.ProductID, img.ImagePath, boolToInt(img.IsPrimary), img.SortOrder, &img.ID,
	).Err()
}

func (r *ProductRepo) DeleteImage(id int64) error {
	_, err := r.db.Exec(`DELETE FROM product_images WHERE id=:1`, id)
	return err
}

// ── Featured ──────────────────────────────────────────────────────────────────

func (r *ProductRepo) GetFeatured(limit int) ([]model.Product, error) {
	if limit <= 0 {
		limit = 8
	}
	rows, err := r.db.Query(
		`SELECT p.id, p.category_id, p.name, p.description, p.price, p.is_active, p.created_at
		 FROM products p WHERE p.is_active = 1
		 ORDER BY p.created_at DESC
		 FETCH FIRST :1 ROWS ONLY`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		var active int
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &active, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.IsActive = active == 1
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range products {
		if err := r.loadImages(&products[i]); err != nil {
			return nil, err
		}
		if err := r.loadSizes(&products[i]); err != nil {
			return nil, err
		}
	}
	return products, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (r *ProductRepo) loadImages(p *model.Product) error {
	rows, err := r.db.Query(
		`SELECT id, product_id, image_path, is_primary, sort_order
		 FROM product_images WHERE product_id=:1 ORDER BY sort_order`,
		p.ID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var img model.ProductImage
		var primary int
		if err := rows.Scan(&img.ID, &img.ProductID, &img.ImagePath, &primary, &img.SortOrder); err != nil {
			return err
		}
		img.IsPrimary = primary == 1
		p.Images = append(p.Images, img)
	}
	return rows.Err()
}

func (r *ProductRepo) loadSizes(p *model.Product) error {
	rows, err := r.db.Query(
		`SELECT id, product_id, size, stock_qty
		 FROM product_sizes WHERE product_id=:1 ORDER BY size`,
		p.ID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var s model.ProductSize
		if err := rows.Scan(&s.ID, &s.ProductID, &s.Size, &s.StockQty); err != nil {
			return err
		}
		p.Sizes = append(p.Sizes, s)
	}
	return rows.Err()
}

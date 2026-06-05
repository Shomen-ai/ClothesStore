// admin_product_handler.go serves the admin catalogue-management endpoints
// (categories, products, sizes and product images) under /api/admin. All routes
// sit behind AuthRequired + AdminRequired, so the caller is assumed to be an
// authenticated admin.
package handler

import (
	"net/http"
	"strconv"
	"clothes-store/internal/model"
	"clothes-store/internal/repository"
	"clothes-store/internal/service"
	"github.com/gin-gonic/gin"
)

// AdminProductHandler manages categories and products; uploadSvc handles
// persisting and removing uploaded image files on disk.
type AdminProductHandler struct {
	repo      *repository.ProductRepo
	uploadSvc *service.UploadService
}

// NewAdminProductHandler constructs an AdminProductHandler from the product repo
// and upload service.
func NewAdminProductHandler(repo *repository.ProductRepo, uploadSvc *service.UploadService) *AdminProductHandler {
	return &AdminProductHandler{repo: repo, uploadSvc: uploadSvc}
}

// ListCategories serves GET /api/admin/categories, returning all categories
// ordered by sort_order. 200 on success, 500 on a repo error.
func (h *AdminProductHandler) ListCategories(c *gin.Context) {
	cats, err := h.repo.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cats)
}

// CreateCategory serves POST /api/admin/categories, inserting the category from
// the JSON body. 201 with the created category (id populated), 400 on a bad
// body, 500 on a repo error.
func (h *AdminProductHandler) CreateCategory(c *gin.Context) {
	var cat model.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.CreateCategory(&cat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

// UpdateCategory serves PUT /api/admin/categories/:id. The path id is set on the
// model first, then overwritten by any "id" present in the JSON body. 200 on
// success, 400 on a bad body, 500 on a repo error.
func (h *AdminProductHandler) UpdateCategory(c *gin.Context) {
	var cat model.Category
	cat.ID, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.UpdateCategory(&cat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cat)
}

// DeleteCategory serves DELETE /api/admin/categories/:id. Returns 204 on
// success or 500 on a repo error.
func (h *AdminProductHandler) DeleteCategory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.DeleteCategory(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ListProducts serves GET /api/admin/products. Supports page and category query
// params; unlike the public catalogue this admin view uses a page size of 50.
// Note the repo's List filter only returns active products (is_active=true), so
// inactive products are not listed here. 200 on success, 500 on a repo error.
func (h *AdminProductHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	catID, _ := strconv.ParseInt(c.Query("category"), 10, 64)
	products, err := h.repo.List(repository.ProductFilter{CategoryID: catID, Page: page, PageSize: 50})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

// CreateProduct serves POST /api/admin/products. Name and a positive Price are
// required by the binding tags. After inserting the product it upserts each
// supplied size (stock per size). 201 with the created product, 400 on a bad
// body, 500 on a repo error. Note: if a size upsert fails mid-loop the product
// has already been created, so the operation is not atomic.
func (h *AdminProductHandler) CreateProduct(c *gin.Context) {
	var req struct {
		CategoryID  int64              `json:"category_id"`
		Name        string             `json:"name" binding:"required"`
		Description string             `json:"description"`
		Price       float64            `json:"price" binding:"required,gt=0"`
		IsActive    bool               `json:"is_active"`
		IsOnSale    bool               `json:"is_on_sale"`
		Sizes       []model.ProductSize `json:"sizes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := &model.Product{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		IsActive:    req.IsActive,
		IsOnSale:    req.IsOnSale,
	}
	if err := h.repo.Create(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, s := range req.Sizes {
		s.ProductID = p.ID
		if err := h.repo.UpsertSize(&s); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusCreated, p)
}

// UpdateProduct serves PUT /api/admin/products/:id. Replaces the product columns
// and upserts the supplied sizes. Unlike CreateProduct, no field is required, so
// omitted fields are written as their zero values (e.g. price 0, is_active
// false) — see findings. 200 on success, 400 on a bad body, 500 on a repo error.
func (h *AdminProductHandler) UpdateProduct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		CategoryID  int64              `json:"category_id"`
		Name        string             `json:"name"`
		Description string             `json:"description"`
		Price       float64            `json:"price"`
		IsActive    bool               `json:"is_active"`
		IsOnSale    bool               `json:"is_on_sale"`
		Sizes       []model.ProductSize `json:"sizes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := &model.Product{
		ID:          id,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		IsActive:    req.IsActive,
		IsOnSale:    req.IsOnSale,
	}
	if err := h.repo.Update(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, s := range req.Sizes {
		s.ProductID = id
		if err := h.repo.UpsertSize(&s); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, p)
}

// DeleteProduct serves DELETE /api/admin/products/:id. 204 on success, 500 on a
// repo error.
func (h *AdminProductHandler) DeleteProduct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// UploadImage serves POST /api/admin/products/:id/images. Reads a multipart form
// and saves every file under the "images" field via uploadSvc, recording each as
// a ProductImage with its index as sort_order. 201 with the saved image rows,
// 400 if the multipart form cannot be parsed, 500 on a save/repo error. Note:
// the product id is not validated to exist, so images can be attached to a
// non-existent product.
func (h *AdminProductHandler) UploadImage(c *gin.Context) {
	productID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	files := form.File["images"]
	saved := make([]model.ProductImage, 0, len(files))
	for i, f := range files {
		path, err := h.uploadSvc.Save(f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		img := &model.ProductImage{ProductID: productID, ImagePath: path, SortOrder: i}
		if err := h.repo.AddImage(img); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		saved = append(saved, *img)
	}
	c.JSON(http.StatusCreated, saved)
}

// DeleteImage serves DELETE /api/admin/products/:id/images/:img_id. The repo
// scopes the delete to the product id, so a mismatched pair returns 404. On
// success it also removes the underlying file from disk. 204 on success, 404 if
// no matching image row exists.
func (h *AdminProductHandler) DeleteImage(c *gin.Context) {
	productID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	imgID, _ := strconv.ParseInt(c.Param("img_id"), 10, 64)
	path, err := h.repo.DeleteImage(imgID, productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	h.uploadSvc.Delete(path)
	c.Status(http.StatusNoContent)
}

package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"clothes-store/internal/repository"
	"github.com/gin-gonic/gin"
)

// catalogue_handler.go serves the public, unauthenticated storefront catalogue
// endpoints (categories, product listing/detail and featured products) under
// /api. The repo only exposes active products to these routes.
type CatalogueHandler struct{ repo *repository.ProductRepo }

// NewCatalogueHandler constructs a CatalogueHandler backed by the product repo.
func NewCatalogueHandler(repo *repository.ProductRepo) *CatalogueHandler {
	return &CatalogueHandler{repo: repo}
}

// GetCategories serves GET /api/categories, returning all categories ordered by
// sort_order. 200 on success, 500 on a repo error.
func (h *CatalogueHandler) GetCategories(c *gin.Context) {
	cats, err := h.repo.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cats)
}

// ListProducts serves GET /api/products with paging (24 per page) and optional
// filters: category, size, q (name search), price_min/price_max, sale and sort.
// Unparsable numeric params fall back to zero, which the repo treats as "no
// filter". Filters are applied as parameterised SQL; sort is an allow-listed
// switch in the repo, not interpolated. 200 on success, 500 on a repo error.
func (h *CatalogueHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	catID, _ := strconv.ParseInt(c.Query("category"), 10, 64)
	priceMin, _ := strconv.ParseFloat(c.Query("price_min"), 64)
	priceMax, _ := strconv.ParseFloat(c.Query("price_max"), 64)
	f := repository.ProductFilter{
		CategoryID: catID,
		Size:       c.Query("size"),
		Search:     c.Query("q"),
		Sort:       c.Query("sort"),
		PriceMin:   priceMin,
		PriceMax:   priceMax,
		Sale:       c.Query("sale") == "1" || c.Query("sale") == "true",
		Page:       page,
		PageSize:   24,
	}
	products, err := h.repo.List(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

// GetProduct serves GET /api/products/:id, returning a single product with its
// sizes and images. 200 on success, 400 on a non-positive/invalid id, 404 if no
// such product, 500 on a repo error. Note: there is no is_active guard here, so
// an inactive product is still readable by id (see findings).
func (h *CatalogueHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	p, err := h.repo.GetByID(id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// GetFeatured serves GET /api/products/featured, returning {hits, new}: the
// top-5 best sellers (by total quantity ordered) and the 5 newest products. 200
// on success, 500 on a repo error.
func (h *CatalogueHandler) GetFeatured(c *gin.Context) {
	hits, newest, err := h.repo.GetFeatured()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"hits": hits, "new": newest})
}

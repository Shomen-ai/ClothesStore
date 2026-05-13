package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"clothes-store/internal/repository"
	"github.com/gin-gonic/gin"
)

type CatalogueHandler struct{ repo *repository.ProductRepo }

func NewCatalogueHandler(repo *repository.ProductRepo) *CatalogueHandler {
	return &CatalogueHandler{repo: repo}
}

func (h *CatalogueHandler) GetCategories(c *gin.Context) {
	cats, err := h.repo.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cats)
}

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

func (h *CatalogueHandler) GetFeatured(c *gin.Context) {
	hits, newest, err := h.repo.GetFeatured()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"hits": hits, "new": newest})
}

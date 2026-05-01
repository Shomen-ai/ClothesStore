package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"clothes-store/internal/repository"
	"github.com/gin-gonic/gin"
)

type CatalogueHandler struct {
	repo *repository.ProductRepo
}

func NewCatalogueHandler(repo *repository.ProductRepo) *CatalogueHandler {
	return &CatalogueHandler{repo: repo}
}

func (h *CatalogueHandler) GetCategories(c *gin.Context) {
	cats, err := h.repo.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": cats})
}

func (h *CatalogueHandler) ListProducts(c *gin.Context) {
	f := repository.ProductFilter{}

	if v := c.Query("category_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		f.CategoryID = &id
	}
	if v := c.Query("min_price"); v != "" {
		p, err := strconv.ParseFloat(v, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid min_price"})
			return
		}
		f.MinPrice = &p
	}
	if v := c.Query("max_price"); v != "" {
		p, err := strconv.ParseFloat(v, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid max_price"})
			return
		}
		f.MaxPrice = &p
	}
	f.Search = c.Query("search")

	if v := c.Query("page"); v != "" {
		page, err := strconv.Atoi(v)
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
			return
		}
		f.Page = page
	}
	if v := c.Query("page_size"); v != "" {
		ps, err := strconv.Atoi(v)
		if err != nil || ps < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_size"})
			return
		}
		f.PageSize = ps
	}

	products, total, err := h.repo.List(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"total":    total,
		"page":     f.Page,
		"page_size": f.PageSize,
	})
}

func (h *CatalogueHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	p, err := h.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *CatalogueHandler) GetFeatured(c *gin.Context) {
	limit := 8
	if v := c.Query("limit"); v != "" {
		l, err := strconv.Atoi(v)
		if err != nil || l < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
		limit = l
	}
	products, err := h.repo.GetFeatured(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"products": products})
}

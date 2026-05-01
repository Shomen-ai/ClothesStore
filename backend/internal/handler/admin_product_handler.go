package handler

import (
	"net/http"
	"strconv"
	"clothes-store/internal/model"
	"clothes-store/internal/repository"
	"clothes-store/internal/service"
	"github.com/gin-gonic/gin"
)

type AdminProductHandler struct {
	repo      *repository.ProductRepo
	uploadSvc *service.UploadService
}

func NewAdminProductHandler(repo *repository.ProductRepo, uploadSvc *service.UploadService) *AdminProductHandler {
	return &AdminProductHandler{repo: repo, uploadSvc: uploadSvc}
}

func (h *AdminProductHandler) ListCategories(c *gin.Context) {
	cats, err := h.repo.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cats)
}

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

func (h *AdminProductHandler) DeleteCategory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.DeleteCategory(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

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

func (h *AdminProductHandler) CreateProduct(c *gin.Context) {
	var req struct {
		CategoryID  int64              `json:"category_id"`
		Name        string             `json:"name" binding:"required"`
		Description string             `json:"description"`
		Price       float64            `json:"price" binding:"required,gt=0"`
		IsActive    bool               `json:"is_active"`
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
	}
	if err := h.repo.Create(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, s := range req.Sizes {
		s.ProductID = p.ID
		h.repo.UpsertSize(&s)
	}
	c.JSON(http.StatusCreated, p)
}

func (h *AdminProductHandler) UpdateProduct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		CategoryID  int64              `json:"category_id"`
		Name        string             `json:"name"`
		Description string             `json:"description"`
		Price       float64            `json:"price"`
		IsActive    bool               `json:"is_active"`
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
	}
	if err := h.repo.Update(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, s := range req.Sizes {
		s.ProductID = id
		h.repo.UpsertSize(&s)
	}
	c.JSON(http.StatusOK, p)
}

func (h *AdminProductHandler) DeleteProduct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

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
		h.repo.AddImage(img)
		saved = append(saved, *img)
	}
	c.JSON(http.StatusCreated, saved)
}

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

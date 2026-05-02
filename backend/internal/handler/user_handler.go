package handler

import (
	"net/http"
	"strconv"
	"clothes-store/internal/model"
	"clothes-store/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct{ repo *repository.UserRepo }

func NewUserHandler(repo *repository.UserRepo) *UserHandler { return &UserHandler{repo: repo} }

func (h *UserHandler) GetProfile(c *gin.Context) {
	id := c.GetInt64("userID")
	u, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	id := c.GetInt64("userID")
	var req struct {
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u := &model.User{ID: id, Name: req.Name, Phone: req.Phone}
	if err := h.repo.Update(u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := h.repo.UpdatePassword(id, string(hash)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, u)
}

func (h *UserHandler) GetAddresses(c *gin.Context) {
	id := c.GetInt64("userID")
	addrs, err := h.repo.GetAddresses(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, addrs)
}

func (h *UserHandler) CreateAddress(c *gin.Context) {
	id := c.GetInt64("userID")
	var a model.Address
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a.UserID = id
	if err := h.repo.CreateAddress(&a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if a.IsDefault {
		h.repo.SetDefaultAddress(a.ID, id)
	}
	c.JSON(http.StatusCreated, a)
}

func (h *UserHandler) UpdateAddress(c *gin.Context) {
	userID := c.GetInt64("userID")
	addrID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || addrID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address id"})
		return
	}
	var a model.Address
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a.ID = addrID
	a.UserID = userID
	if a.IsDefault {
		h.repo.SetDefaultAddress(addrID, userID)
	}
	if err := h.repo.UpdateAddress(&a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *UserHandler) DeleteAddress(c *gin.Context) {
	userID := c.GetInt64("userID")
	addrID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || addrID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address id"})
		return
	}
	if err := h.repo.DeleteAddress(addrID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

package handler

import (
	"net/http"
	"strconv"
	"clothes-store/internal/model"
	"clothes-store/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// user_handler.go serves the authenticated user's profile and address-book
// endpoints under /api/user. All routes sit behind AuthRequired; the acting user
// id comes from the JWT (c.GetInt64("userID")), and address operations are scoped
// to that user in the repo's SQL.
type UserHandler struct{ repo *repository.UserRepo }

// NewUserHandler constructs a UserHandler backed by the user repo.
func NewUserHandler(repo *repository.UserRepo) *UserHandler { return &UserHandler{repo: repo} }

// GetProfile serves GET /api/user/profile, returning the authenticated user's
// record (the password hash is excluded via the model's json:"-" tag). 200 on
// success, 404 if the user no longer exists.
func (h *UserHandler) GetProfile(c *gin.Context) {
	id := c.GetInt64("userID")
	u, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

// UpdateProfile serves PUT /api/user/profile. It updates name and phone, and if
// a non-empty password is supplied, bcrypts and stores it. 200 with the updated
// user, 400 on a bad body, 500 on a repo error. Note: the returned object
// reflects only name/phone (not other persisted columns), and password length is
// not validated here (see findings).
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

// GetAddresses serves GET /api/user/addresses, returning the user's addresses
// with the default first. 200 on success, 500 on a repo error.
func (h *UserHandler) GetAddresses(c *gin.Context) {
	id := c.GetInt64("userID")
	addrs, err := h.repo.GetAddresses(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, addrs)
}

// CreateAddress serves POST /api/user/addresses. The owner is forced to the JWT
// user id (ignoring any user_id in the body). If is_default is set, it is
// promoted to the sole default afterwards. 201 with the created address, 400 on
// a bad body, 500 on a repo error.
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

// UpdateAddress serves PUT /api/user/addresses/:id. The update is scoped to the
// JWT user id in SQL, so a user cannot edit another's address. 200 with the
// address, 400 on a bad id/body, 500 on a repo error. Note: SetDefaultAddress
// runs before UpdateAddress, and neither return value nor the not-found case is
// checked (see findings).
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

// DeleteAddress serves DELETE /api/user/addresses/:id, scoped to the JWT user id.
// 204 on success (also when nothing matched), 400 on a bad id, 500 on a repo
// error.
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

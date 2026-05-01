package main

import (
	"log"
	"os"
	"clothes-store/internal/config"
	"clothes-store/internal/db"
	"clothes-store/internal/handler"
	"clothes-store/internal/repository"
	"clothes-store/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	if cfg.DBConnStr == "" {
		log.Fatal("DB_CONN_STR required")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET required")
	}
	if err := os.MkdirAll(cfg.UploadsDir, 0755); err != nil {
		log.Fatalf("uploads dir: %v", err)
	}

	database, err := db.Connect(cfg.DBConnStr)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	// Repos
	userRepo := repository.NewUserRepo(database)
	productRepo := repository.NewProductRepo(database)

	// Services
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	catalogueH := handler.NewCatalogueHandler(productRepo)

	r := gin.Default()
	r.Static("/uploads", cfg.UploadsDir)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.POST("/refresh", authH.Refresh)

		api.GET("/categories", catalogueH.GetCategories)
		api.GET("/products/featured", catalogueH.GetFeatured)
		api.GET("/products", catalogueH.ListProducts)
		api.GET("/products/:id", catalogueH.GetProduct)
	}

	log.Printf("Listening on :%s", cfg.Port)
	r.Run(":" + cfg.Port)
}

package main

import (
	"log"
	"os"
	"clothes-store/internal/config"
	"clothes-store/internal/db"
	"clothes-store/internal/handler"
	"clothes-store/internal/middleware"
	"clothes-store/internal/repository"
	"clothes-store/internal/service"
	"github.com/gin-contrib/cors"
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
	userRepo     := repository.NewUserRepo(database)
	productRepo  := repository.NewProductRepo(database)
	promoRepo    := repository.NewPromoRepo(database)
	orderRepo    := repository.NewOrderRepo(database)
	wishlistRepo := repository.NewWishlistRepo(database)
	statsRepo    := repository.NewStatsRepo(database)
	reportsRepo  := repository.NewReportsRepo(database)

	// Services
	authSvc   := service.NewAuthService(userRepo, cfg.JWTSecret)
	orderSvc  := service.NewOrderService(orderRepo, productRepo, promoRepo)
	uploadSvc := service.NewUploadService(cfg.UploadsDir)

	// Handlers
	authH         := handler.NewAuthHandler(authSvc)
	catalogueH    := handler.NewCatalogueHandler(productRepo)
	orderH        := handler.NewOrderHandler(orderSvc, promoRepo, orderRepo)
	wishlistH     := handler.NewWishlistHandler(wishlistRepo)
	userH         := handler.NewUserHandler(userRepo)
	adminProductH := handler.NewAdminProductHandler(productRepo, uploadSvc)
	adminOrderH   := handler.NewAdminOrderHandler(orderRepo)
	adminPromoH   := handler.NewAdminPromoHandler(promoRepo)
	adminStatsH   := handler.NewAdminStatsHandler(statsRepo)
	adminReportH  := handler.NewAdminReportHandler(reportsRepo)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
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

		protected := api.Group("", middleware.AuthRequired(cfg.JWTSecret))
		protected.POST("/orders", orderH.Create)
		protected.POST("/promo/validate", orderH.ValidatePromo)

		user := api.Group("/user", middleware.AuthRequired(cfg.JWTSecret))
		user.GET("/orders", orderH.GetUserOrders)
		user.GET("/orders/:id", orderH.GetUserOrder)
		user.GET("/wishlist", wishlistH.Get)
		user.POST("/wishlist/:product_id", wishlistH.Add)
		user.DELETE("/wishlist/:product_id", wishlistH.Remove)
		user.GET("/profile", userH.GetProfile)
		user.PUT("/profile", userH.UpdateProfile)
		user.GET("/addresses", userH.GetAddresses)
		user.POST("/addresses", userH.CreateAddress)
		user.PUT("/addresses/:id", userH.UpdateAddress)
		user.DELETE("/addresses/:id", userH.DeleteAddress)

		admin := api.Group("/admin", middleware.AuthRequired(cfg.JWTSecret), middleware.AdminRequired())
		admin.GET("/products", adminProductH.ListProducts)
		admin.POST("/products", adminProductH.CreateProduct)
		admin.PUT("/products/:id", adminProductH.UpdateProduct)
		admin.DELETE("/products/:id", adminProductH.DeleteProduct)
		admin.POST("/products/:id/images", adminProductH.UploadImage)
		admin.DELETE("/products/:id/images/:img_id", adminProductH.DeleteImage)
		admin.GET("/categories", adminProductH.ListCategories)
		admin.POST("/categories", adminProductH.CreateCategory)
		admin.PUT("/categories/:id", adminProductH.UpdateCategory)
		admin.DELETE("/categories/:id", adminProductH.DeleteCategory)
		admin.GET("/orders", adminOrderH.List)
		admin.GET("/orders/:id", adminOrderH.Get)
		admin.PUT("/orders/:id/status", adminOrderH.UpdateStatus)
		admin.GET("/promo-codes", adminPromoH.List)
		admin.POST("/promo-codes", adminPromoH.Create)
		admin.PUT("/promo-codes/:id/deactivate", adminPromoH.Deactivate)
		admin.DELETE("/promo-codes/:id", adminPromoH.Delete)
		admin.GET("/stats/revenue", adminStatsH.Revenue)
		admin.GET("/stats/orders", adminStatsH.Orders)
		admin.GET("/stats/promo-codes", adminStatsH.Promos)
		admin.GET("/reports/excel", adminReportH.Excel)
	}

	log.Printf("Listening on :%s", cfg.Port)
	r.Run(":" + cfg.Port)
}

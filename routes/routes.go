package routes

import (
	"github.com/BilyHakim/go-walet/config"
	"github.com/BilyHakim/go-walet/controllers"
	"github.com/BilyHakim/go-walet/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, rmq *config.RabbitMQ) {
	userController := controllers.NewUserController(db)
	walletController := controllers.NewWalletController(db, rmq)

	public := r.Group("/api")
	{
		public.POST("/register", userController.Register)
		public.POST("/login", userController.Login)
	}

	protected := r.Group("/api")
	protected.Use(middleware.JWTAuth())
	{
		protected.PUT("/update-profile", userController.UpdateProfile)
		protected.POST("/get-user", userController.GetUserByPhone)
		protected.POST("/topup", walletController.TopUp)
		protected.POST("/payments", walletController.Payment)
		protected.POST("/transfers", walletController.Transfer)
		protected.GET("/transactions", walletController.GetTransactions)
	}
}

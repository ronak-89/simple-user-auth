package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ronak-89/simple-user-auth/config"
	"github.com/ronak-89/simple-user-auth/internal/handlers"
	"github.com/ronak-89/simple-user-auth/internal/services"
)

func main() {
	client := config.MongoConnect()

	userService := services.NewUserService(client)
	userHandler := handlers.NewUserHandler(userService)

	router := gin.Default()

	router.POST("/login", userHandler.Login)
	router.POST("/user", userHandler.RegisterUser)
	router.POST("/verify-otp", userHandler.VerifyOtp)

	// router.POST("/test", services.GetData)
	// router.GET("/users", services.GetUsers)
	// router.GET("/user/:id", services.GetUserById)
	// router.PUT("/user/:id", services.UpdateUser)
	// router.PATCH("/user/:id", services.PatchUser)
	// router.DELETE("/user/:id", services.DeleteUser)

	err := router.Run("localhost:8000")
	if err != nil {
		return
	}
}

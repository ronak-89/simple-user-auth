package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/ronak-89/simple-user-auth/internal/models"
	"github.com/ronak-89/simple-user-auth/internal/services"
	"github.com/ronak-89/simple-user-auth/utils"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type UserHandler struct {
	UserService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

func (h *UserHandler) RegisterUser(c *gin.Context) {

	var newUser models.User
	var validationErrors []models.ErrorResponse
	var validationFieldErrors []models.ErrorResponse

	if err := c.ShouldBindJSON(&newUser); err != nil {
		validationErrors = utils.ValidateBinding(err)

		if len(validationErrors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"errors": validationErrors,
			})
			return
		}
	}

	if err := validate.Struct(newUser); err != nil {
		validationFieldErrors = utils.Validatefields(err)

		if len(validationFieldErrors) > 0 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": "error",
				"errors": validationFieldErrors,
			})
			return
		}
	}

	message, err := h.UserService.PostUser(&newUser)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": message})

}

func (h *UserHandler) VerifyOtp(c *gin.Context) {
	var requestData models.EmailOtp

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
		return
	}

	message, statusCode, err := h.UserService.VerifyOtp(&requestData)

	if err != nil {
		c.JSON(statusCode, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": message})

}

func (h *UserHandler) Login(c *gin.Context) {
	var user models.UserLogin
	var validationErrors []models.ErrorResponse

	if err := c.ShouldBindJSON(&user); err != nil {
		validationErrors = utils.ValidateBinding(err)
		if len(validationErrors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "errors": validationErrors})
			return
		}
	}

	message, statusCode, err := h.UserService.Login(&user)
	if err != nil {
		c.JSON(statusCode, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": message})
}

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
func RegisterUser(c *gin.Context) {

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

	message, err := services.PostUser(&newUser)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": message})

}

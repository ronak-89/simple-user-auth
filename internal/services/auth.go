package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/ronak-89/simple-user-auth/config"
	"github.com/ronak-89/simple-user-auth/internal/models"
	"github.com/ronak-89/simple-user-auth/internal/repositories"
	"github.com/ronak-89/simple-user-auth/utils"
)

var db = config.DbConnection()
var validate *validator.Validate

func init() {
	validate = validator.New()
}

func PostUser(newUser *models.User) (string, error) {

	existingUser, _ := repositories.GetUserByEmail(newUser.Email)
	if existingUser != nil {
		return "", errors.New("user with this email already exists")
	}

	hashPassword := utils.HashPassword(newUser.Password)
	newUser.Password = hashPassword

	if err := repositories.CreateUser(newUser); err != nil {
		return "", err
	}

	otp := utils.GenerateOtp()

	if err := repositories.CreateEmail(newUser.Email, otp); err != nil {
		return "", err
	}

	go utils.SendMail(otp)

	return "User registered successfully. OTP sent to email.", nil

}

func GetUsers(c *gin.Context) {
	var users []models.User
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	result := db.Limit(pageSize).Offset(offset).Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}
	var total int64
	db.Model(&models.User{}).Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"page":     page,
		"pageSize": pageSize,
		"total":    total,
		"users":    users,
	})
}

func GetUserById(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var updatedUser models.User

	var validationErrors []models.ErrorResponse

	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		var unmarshalTypeErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &unmarshalTypeErr):
			field := unmarshalTypeErr.Field
			expectedType := unmarshalTypeErr.Type.String()
			validationErrors = append(validationErrors, models.ErrorResponse{
				Field:   field,
				Message: fmt.Sprintf("%s must be of type %s", field, expectedType),
			})
		case strings.Contains(err.Error(), "required"):
			fieldName := strings.Split(err.Error(), "'")[1]
			validationErrors = append(validationErrors, models.ErrorResponse{
				Field:   fieldName,
				Message: fmt.Sprintf("%s field is required", fieldName),
			})
		default:
			validationErrors = append(validationErrors, models.ErrorResponse{
				Field:   "request",
				Message: "Invalid JSON format",
			})
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"errors": validationErrors,
		})
		return
	}

	if err := validate.Struct(updatedUser); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			for _, e := range ve {
				validationErrors = append(validationErrors, models.ErrorResponse{
					Field:   e.Field(),
					Message: utils.GetValidationMessage(e.Field(), e.Tag(), e.Param()),
				})
			}
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": "error",
				"errors": validationErrors,
			})
			return
		}
	}

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	user.Name = updatedUser.Name
	user.Email = updatedUser.Email
	user.Number = updatedUser.Number
	user.Gender = updatedUser.Gender

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error updating user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func PatchUser(c *gin.Context) {
	id := c.Param("id")
	var patchUser models.PatchUser

	var validationErrors []models.ErrorResponse

	if err := c.ShouldBindJSON(&patchUser); err != nil {
		var unmarshalTypeErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &unmarshalTypeErr):
			field := unmarshalTypeErr.Field
			expectedType := unmarshalTypeErr.Type.String()
			validationErrors = append(validationErrors, models.ErrorResponse{
				Field:   field,
				Message: fmt.Sprintf("%s must be of type %s", field, expectedType),
			})
		case strings.Contains(err.Error(), "required"):
			fieldName := strings.Split(err.Error(), "'")[1]
			validationErrors = append(validationErrors, models.ErrorResponse{
				Field:   fieldName,
				Message: fmt.Sprintf("%s field is required", fieldName),
			})
		default:
			validationErrors = append(validationErrors, models.ErrorResponse{
				Field:   "request",
				Message: "Invalid JSON format",
			})
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"errors": validationErrors,
		})
		return
	}

	if err := validate.Struct(patchUser); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			for _, e := range ve {
				validationErrors = append(validationErrors, models.ErrorResponse{
					Field:   e.Field(),
					Message: utils.GetValidationMessage(e.Field(), e.Tag(), e.Param()),
				})
			}
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": "error",
				"errors": validationErrors,
			})
			return
		}
	}

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if patchUser.Name != nil {
		user.Name = *patchUser.Name
	}
	if patchUser.Email != nil {
		user.Email = *patchUser.Email
	}
	if patchUser.Number != nil {
		user.Number = *patchUser.Number
	}
	if patchUser.Gender != nil {
		user.Gender = *patchUser.Gender
	}

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error updating user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if err := db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error deleting user"})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "User deleted"})
}

func Login(c *gin.Context) {
	var user models.UserLogin
	var validationErrors []models.ErrorResponse

	if err := c.ShouldBindJSON(&user); err != nil {
		var unmarshalTypeErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &unmarshalTypeErr):
			field := unmarshalTypeErr.Field
			expectedType := unmarshalTypeErr.Type.String()
			validationErrors = append(validationErrors, models.ErrorResponse{
				Field:   field,
				Message: fmt.Sprintf("%s must be of type %s", field, expectedType),
			})
		case strings.Contains(err.Error(), "required"):
			fieldName := strings.Split(err.Error(), "'")[1]
			validationErrors = append(validationErrors, models.ErrorResponse{
				Field:   fieldName,
				Message: fmt.Sprintf("%s field is required", fieldName),
			})
		default:
			validationErrors = append(validationErrors, models.ErrorResponse{
				Field:   "request",
				Message: "Invalid JSON format",
			})
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"errors": validationErrors,
		})
		return
	}

	var existingUser models.User

	if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "User not found"})
		return
	}
	if !utils.CheckPasswordHash(user.Password, existingUser.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid email or password"})
		return
	}

	if !existingUser.IsVerified {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "User is not verified"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Logged in successfully",
	})

}

func VerifyOtp(c *gin.Context) {
	var requestData models.EmailOtp

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
	}

	var storedOtp models.EmailOtp
	if err := db.Where("email = ?", requestData.Email).First(&storedOtp).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "OTP not found or expired"})
		return
	}

	if requestData.Otp != storedOtp.Otp {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid OTP"})
		return
	}

	var user models.User
	if err := db.Where("email = ?", requestData.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "User not found"})
		return
	}

	user.IsVerified = true
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to update user verification status"})
		return
	}

	if err := db.Where("email = ?", requestData.Email).Delete(&models.EmailOtp{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to delete OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "OTP verified successfully. Your account is now verified.",
	})
}

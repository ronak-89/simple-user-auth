package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/ronak-89/simple-user-auth/models"
	"github.com/ronak-89/simple-user-auth/utils"
	"net/http"
	"strconv"
	"strings"
)

var db = DbConnection()
var validate *validator.Validate

func init() {
	validate = validator.New()
}

func postUser(c *gin.Context) {
	var newUser models.User
	var validationErrors []models.ErrorResponse
	//var errors []models.ErrorResponse
	//
	//if err := c.ShouldBindJSON(&newUser); err != nil {
	//	if strings.Contains(err.Error(), "required") {
	//		fmt.Println(err.Error())
	//		fieldName := strings.Split(err.Error(), "'")[1]
	//		validationErrors = append(validationErrors, models.ErrorResponse{
	//			Field:   fieldName,
	//			Message: fmt.Sprintf("%s field is required", fieldName),
	//		})
	//	} else if strings.Contains(err.Error(), "cannot unmarshal") {
	//		errStr := err.Error()
	//		fmt.Println(errStr)
	//		field := strings.Split(strings.Split(errStr, "field ")[1], " ")[0]
	//		expectedType := strings.Split(strings.Split(errStr, "type ")[1], " ")[0]
	//
	//		validationErrors = append(validationErrors, models.ErrorResponse{
	//			Field:   field,
	//			Message: fmt.Sprintf("%s must be a %s", field, expectedType),
	//		})
	//	} else {
	//		validationErrors = append(validationErrors, models.ErrorResponse{
	//			Field:   "request",
	//			Message: "Invalid JSON format",
	//		})
	//	}
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"status": "error",
	//		"errors": validationErrors,
	//	})
	//	return
	//}

	if err := c.ShouldBindJSON(&newUser); err != nil {
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
	}
	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"errors": validationErrors,
		})
		return
	}

	if err := validate.Struct(newUser); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			for _, e := range ve {
				validationErrors = append(validationErrors, models.ErrorResponse{
					Field:   e.Field(),
					Message: utils.GetValidationMessage(e.Field(), e.Tag(), e.Param()),
				})
			}
		}
	}
	if len(validationErrors) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "error",
			"errors": validationErrors,
		})
		return
	}

	var existingUser models.User
	if err := db.Where("email = ?", newUser.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": "User with this email already exists",
		})
		return
	}
	hashPassword := utils.HashPassword(newUser.Password)
	newUser.Password = hashPassword
	if err := db.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create user: " + err.Error(),
		})
		return
	}

	otp := utils.GenerateOtp()

	emailOtp := models.EmailOtp{
		Email: newUser.Email,
		Otp:   otp,
	}
	if err := db.Create(&emailOtp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to save OTP: " + err.Error(),
		})
		return
	}

	go utils.SendMail(otp)

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "User registered successfully. OTP sent to email.",
	})
}

func getUsers(c *gin.Context) {
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

func getUserById(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func updateUser(c *gin.Context) {
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

func patchUser(c *gin.Context) {
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

func deleteUser(c *gin.Context) {
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

func login(c *gin.Context) {
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

func verifyOtp(c *gin.Context) {
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

func main() {

	err := db.AutoMigrate(&models.User{}, &models.EmailOtp{})
	if err != nil {
		return
	}

	router := gin.Default()

	router.POST("/login", login)
	router.POST("/user", postUser)
	router.POST("/verify-otp", verifyOtp)
	router.GET("/users", getUsers)
	router.GET("/user/:id", getUserById)
	router.PUT("/user/:id", updateUser)
	router.PATCH("/user/:id", patchUser)
	router.DELETE("/user/:id", deleteUser)

	err = router.Run("localhost:8000")
	if err != nil {
		return
	}
}

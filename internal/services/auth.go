package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ronak-89/simple-user-auth/internal/models"
	"github.com/ronak-89/simple-user-auth/internal/repositories"
	"github.com/ronak-89/simple-user-auth/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserService struct {
	UserRepo  *repositories.UserRepository
	EmailRepo *repositories.EmailRepository
}

func NewUserService(client *mongo.Client) *UserService {
	return &UserService{
		UserRepo:  repositories.NewUserRepository(client),
		EmailRepo: repositories.NewEmailRepository(client),
	}
}

func (s *UserService) Login(user *models.UserLogin) (string, int, error) {
	existingUser, err := s.UserRepo.GetUserByEmail(user.Email)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	if existingUser == nil {
		return "", http.StatusNotFound, errors.New("user not found")
	}

	if !utils.CheckPasswordHash(user.Password, existingUser.Password) {
		return "", http.StatusUnauthorized, errors.New("invalid email or password")
	}

	if !existingUser.IsVerified {
		return "", http.StatusUnauthorized, errors.New("user is not verified")
	}

	return "User logged in successfully", http.StatusOK, nil
}

func (s *UserService) PostUser(newUser *models.User) (string, error) {

	existingUser, err := s.UserRepo.GetUserByEmail(newUser.Email)

	if err != nil {
		return "", err
	}

	if existingUser != nil {
		return "", errors.New("user with this email already exists")
	}

	hashPassword := utils.HashPassword(newUser.Password)
	newUser.Password = hashPassword

	if err := s.UserRepo.CreatUser(newUser); err != nil {
		return "", err
	}

	otp := utils.GenerateOtp()

	if err := s.EmailRepo.CreateEmail(newUser.Email, otp); err != nil {
		return "", err
	}

	go utils.SendMail(otp, newUser.Email)

	return "User registered successfully. OTP sent to email.", nil

}

func (s *UserService) VerifyOtp(requestData *models.EmailOtp) (string, int, error) {

	var emailObj models.EmailOtp

	err := s.UserRepo.EmailCollection.FindOne(context.TODO(), bson.D{{Key: "email", Value: requestData.Email}}).Decode(&emailObj)

	if err == mongo.ErrNoDocuments {
		return "", http.StatusNotFound, errors.New("OTP record not found")
	} else if err != nil {
		return "", http.StatusInternalServerError, err
	}

	fmt.Println(emailObj)
	fmt.Println(emailObj.Otp)

	if requestData.Otp != emailObj.Otp {
		return "", http.StatusUnauthorized, errors.New("invalid OTP")
	}

	existingUser, err := s.UserRepo.GetUserByEmail(requestData.Email)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	if existingUser == nil {
		return "", http.StatusNotFound, errors.New("user not found")
	}

	filter := bson.D{{Key: "email", Value: requestData.Email}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "is_verified", Value: true}}}}

	_, err = s.UserRepo.UserCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return "", http.StatusInternalServerError, errors.New("failed to update user verification status")
	}

	_, err = s.EmailRepo.EmailCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return "", http.StatusInternalServerError, errors.New("failed to delete OTP record")
	}
	return "OTP verified successfully. Your account is now verified.", http.StatusOK, nil

}

// func GetUsers(c *gin.Context) {
// 	var users []models.User
// 	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
// 	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

// 	if page < 1 {
// 		page = 1
// 	}
// 	if pageSize < 1 {
// 		pageSize = 10
// 	}

// 	offset := (page - 1) * pageSize
// 	result := db.Limit(pageSize).Offset(offset).Find(&users)
// 	if result.Error != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
// 		return
// 	}
// 	var total int64
// 	db.Model(&models.User{}).Count(&total)

// 	c.JSON(http.StatusOK, gin.H{
// 		"page":     page,
// 		"pageSize": pageSize,
// 		"total":    total,
// 		"users":    users,
// 	})
// }

// func GetUserById(c *gin.Context) {
// 	id := c.Param("id")

// 	var user models.User
// 	if err := db.First(&user, id).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, user)
// }

// func UpdateUser(c *gin.Context) {
// 	id := c.Param("id")
// 	var updatedUser models.User

// 	var validationErrors []models.ErrorResponse

// 	if err := c.ShouldBindJSON(&updatedUser); err != nil {
// 		var unmarshalTypeErr *json.UnmarshalTypeError
// 		switch {
// 		case errors.As(err, &unmarshalTypeErr):
// 			field := unmarshalTypeErr.Field
// 			expectedType := unmarshalTypeErr.Type.String()
// 			validationErrors = append(validationErrors, models.ErrorResponse{
// 				Field:   field,
// 				Message: fmt.Sprintf("%s must be of type %s", field, expectedType),
// 			})
// 		case strings.Contains(err.Error(), "required"):
// 			fieldName := strings.Split(err.Error(), "'")[1]
// 			validationErrors = append(validationErrors, models.ErrorResponse{
// 				Field:   fieldName,
// 				Message: fmt.Sprintf("%s field is required", fieldName),
// 			})
// 		default:
// 			validationErrors = append(validationErrors, models.ErrorResponse{
// 				Field:   "request",
// 				Message: "Invalid JSON format",
// 			})
// 		}
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"status": "error",
// 			"errors": validationErrors,
// 		})
// 		return
// 	}

// 	if err := validate.Struct(updatedUser); err != nil {
// 		var ve validator.ValidationErrors
// 		if errors.As(err, &ve) {
// 			for _, e := range ve {
// 				validationErrors = append(validationErrors, models.ErrorResponse{
// 					Field:   e.Field(),
// 					Message: utils.GetValidationMessage(e.Field(), e.Tag(), e.Param()),
// 				})
// 			}
// 			c.JSON(http.StatusUnprocessableEntity, gin.H{
// 				"status": "error",
// 				"errors": validationErrors,
// 			})
// 			return
// 		}
// 	}

// 	var user models.User
// 	if err := db.First(&user, id).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
// 		return
// 	}

// 	user.Name = updatedUser.Name
// 	user.Email = updatedUser.Email
// 	user.Number = updatedUser.Number
// 	user.Gender = updatedUser.Gender

// 	if err := db.Save(&user).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error updating user"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, user)
// }

// func PatchUser(c *gin.Context) {
// 	id := c.Param("id")
// 	var patchUser models.PatchUser

// 	var validationErrors []models.ErrorResponse

// 	if err := c.ShouldBindJSON(&patchUser); err != nil {
// 		var unmarshalTypeErr *json.UnmarshalTypeError
// 		switch {
// 		case errors.As(err, &unmarshalTypeErr):
// 			field := unmarshalTypeErr.Field
// 			expectedType := unmarshalTypeErr.Type.String()
// 			validationErrors = append(validationErrors, models.ErrorResponse{
// 				Field:   field,
// 				Message: fmt.Sprintf("%s must be of type %s", field, expectedType),
// 			})
// 		case strings.Contains(err.Error(), "required"):
// 			fieldName := strings.Split(err.Error(), "'")[1]
// 			validationErrors = append(validationErrors, models.ErrorResponse{
// 				Field:   fieldName,
// 				Message: fmt.Sprintf("%s field is required", fieldName),
// 			})
// 		default:
// 			validationErrors = append(validationErrors, models.ErrorResponse{
// 				Field:   "request",
// 				Message: "Invalid JSON format",
// 			})
// 		}
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"status": "error",
// 			"errors": validationErrors,
// 		})
// 		return
// 	}

// 	if err := validate.Struct(patchUser); err != nil {
// 		var ve validator.ValidationErrors
// 		if errors.As(err, &ve) {
// 			for _, e := range ve {
// 				validationErrors = append(validationErrors, models.ErrorResponse{
// 					Field:   e.Field(),
// 					Message: utils.GetValidationMessage(e.Field(), e.Tag(), e.Param()),
// 				})
// 			}
// 			c.JSON(http.StatusUnprocessableEntity, gin.H{
// 				"status": "error",
// 				"errors": validationErrors,
// 			})
// 			return
// 		}
// 	}

// 	var user models.User
// 	if err := db.First(&user, id).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
// 		return
// 	}

// 	if patchUser.Name != nil {
// 		user.Name = *patchUser.Name
// 	}
// 	if patchUser.Email != nil {
// 		user.Email = *patchUser.Email
// 	}
// 	if patchUser.Number != nil {
// 		user.Number = *patchUser.Number
// 	}
// 	if patchUser.Gender != nil {
// 		user.Gender = *patchUser.Gender
// 	}

// 	if err := db.Save(&user).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error updating user"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, user)
// }

// func DeleteUser(c *gin.Context) {
// 	id := c.Param("id")

// 	var user models.User
// 	if err := db.First(&user, id).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
// 		return
// 	}

// 	if err := db.Delete(&user).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error deleting user"})
// 		return
// 	}

// 	c.JSON(http.StatusNoContent, gin.H{"message": "User deleted"})
// }

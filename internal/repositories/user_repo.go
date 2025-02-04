package repositories

import (
	"github.com/ronak-89/simple-user-auth/config"
	"github.com/ronak-89/simple-user-auth/internal/models"
	"gorm.io/gorm"
)

var db = config.DbConnection()

func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := db.Where("email = ?", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func CreateUser(user *models.User) error {
	return db.Create(user).Error

}

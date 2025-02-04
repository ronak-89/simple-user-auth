package repositories

import (
	"github.com/ronak-89/simple-user-auth/internal/models"
)

// var db = config.DbConnection()

func CreateEmail(email string, otp string) error {
	emailOtp := models.EmailOtp{
		Email: email,
		Otp:   otp,
	}
	return db.Create(emailOtp).Error

}

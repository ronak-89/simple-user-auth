package utils

import (
	"fmt"
	"github.com/ronak-89/simple-user-auth/config"
	"golang.org/x/crypto/bcrypt"
	gomail "gopkg.in/mail.v2"
	"math/rand"
)

var cnf, _ = config.LoadConfig()

func SendMail(otp string) {
	message := gomail.NewMessage()

	message.SetHeader("From", "ronak285.rejoice@gmail.com")
	message.SetHeader("To", "ronak@appscrip.co")
	message.SetHeader("Subject", "testing")

	message.SetBody("text/plain", fmt.Sprintf("Your OTP is: %s", otp))

	dialer := gomail.NewDialer(cnf.EMAIL.Host, cnf.EMAIL.Port, cnf.EMAIL.Username, cnf.EMAIL.Password)

	if err := dialer.DialAndSend(message); err != nil {
		fmt.Println("Error:", err)
		panic(err)
	} else {
		fmt.Println("Email sent successfully!")
	}
}

func GenerateOtp() string {
	otp := rand.Intn(900000) + 100000
	return fmt.Sprintf("%06v", otp)
}

func GetValidationMessage(field string, tag string, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return "Please provide a valid email address"
	case "min":
		if field == "Name" {
			return fmt.Sprintf("%s must be at least %s characters long", field, param)
		}
		if field == "Number" {
			return "Phone number must be 10 digits"
		}
	case "max":
		if field == "Name" {
			return fmt.Sprintf("%s cannot be longer than %s characters", field, param)
		}
		if field == "Number" {
			return "Phone number must be 10 digits"
		}
	case "oneof":
		return fmt.Sprintf("%s must be either 'Male' or 'Female'", field)
	}
	return fmt.Sprintf("%s failed validation for %s", field, tag)
}

func HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil

}

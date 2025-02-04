package models

type User struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"not null;size:10	" json:"name" binding:"required" validate:"min=3,max=20"`
	Email      string `gorm:"not null;size:100" json:"email" binding:"required" validate:"email"`
	Password   string `gorm:"not null" json:"password" binding:"required" validate:"min=8,max=20"`
	Number     int64  `gorm:"not null" json:"number" binding:"required" validate:"min=1000000000,max=9999999999"`
	Gender     string `gorm:"not null;size:10" json:"gender" binding:"required" validate:"oneof=Male Female"`
	IsVerified bool   `gorm:"not null, default:False" json:"is_verified"`
}

type PatchUser struct {
	Name   *string `json:"name" validate:"omitempty,min=3,max=20"`
	Email  *string `json:"email" validate:"omitempty,email"`
	Number *int64  `json:"number" validate:"omitempty,min=1000000000,max=9999999999"`
	Gender *string `json:"gender" validate:"omitempty,oneof=Male Female"`
}

type EmailOtp struct {
	Email string `json:"email" validate:"required,email"`
	Otp   string `json:"otp"`
}

type UserLogin struct {
	Email    string `json:"email" binding:"required" validate:"email"`
	Password string `json:"password" binding:"required" validate:"password"`
}

type ErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

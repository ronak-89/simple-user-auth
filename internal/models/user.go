package models

type User struct {
	Name       string `bson:"name" json:"name" binding:"required" validate:"min=3"`
	Email      string `bson:"email" json:"email" binding:"required" validate:"email"`
	Password   string `bson:"password" json:"password" binding:"required" validate:"min=6"`
	Number     int64  `bson:"number" json:"number" binding:"required" validate:"min=1000000000,max=9999999999"`
	Gender     string `bson:"gender" json:"gender" binding:"required" validate:"oneof=Male Female"`
	IsVerified bool   `bson:"is_verified" json:"is_verified"`
}

type EmailOtp struct {
	Email string `bson:"email" json:"email" validate:"required,email"`
	Otp   string `bson:"otp" json:"otp" validate:"required"`
}

type UserLogin struct {
	Email    string `bson:"email" json:"email" binding:"required" validate:"email"`
	Password string `bson:"password" json:"password" binding:"required" validate:"password"`
}

type PatchUser struct {
	Name   *string `json:"name" validate:"omitempty,min=3,max=20"`
	Email  *string `json:"email" validate:"omitempty,email"`
	Number *int64  `json:"number" validate:"omitempty,min=1000000000,max=9999999999"`
	Gender *string `json:"gender" validate:"omitempty,oneof=Male Female"`
}

type ErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

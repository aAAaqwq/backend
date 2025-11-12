package model

import "time"

const (
	RoleAdmin = "admin"
	RoleUser = "user"
)

type User struct{
	UID int64 `json:"uid"`
	Role string `json:"role"`
	Username string `json:"username"`
	Email string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	PasswordHash string `json:"password_hash"`
	CreateAt time.Time `json:"create_at"`
	UpdateAt time.Time `json:"update_at"`
}

type ChangePasswordReq struct{
	UID int64 `json:"uid"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
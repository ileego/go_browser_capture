package models

import "time"

type BaseModel struct {
	ID        uint      `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type User struct {
	BaseModel
	Username string `json:"username" db:"username" binding:"required"`
	Email    string `json:"email" db:"email" binding:"required,email"`
	Password string `json:"-" db:"password" binding:"required"`
}

type ChromeExtension struct {
	BaseModel
	Name        string `json:"name" db:"name" binding:"required"`
	Version     string `json:"version" db:"version" binding:"required"`
	Description string `json:"description" db:"description"`
	Manifest    string `json:"manifest" db:"manifest" binding:"required"`
	UserID      uint   `json:"user_id" db:"user_id" binding:"required"`
}

package auth

import (
	"errors"
	"time"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

var ErrEmailTaken = errors.New("email already taken")
var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrTokenExpired = errors.New("refresh token expired")

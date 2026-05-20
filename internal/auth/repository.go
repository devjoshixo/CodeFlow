package auth

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByID(ctx context.Context, id string) (*User, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	FindToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteToken(ctx context.Context, token string) error
}

package postgres

import (
	"codeflow/internal/auth"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

type RefreshTokenRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func NewRefreshTokenRepo(db *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (r *UserRepo) FindUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	user := &auth.User{}
	err := r.db.QueryRow(ctx,
		"SELECT id,email,password_hash,created_at,updated_at FROM users WHERE email=$1", email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepo) FindUserByID(ctx context.Context, id string) (*auth.User, error) {
	user := &auth.User{}
	err := r.db.QueryRow(ctx, "SELECT id,email,password_hash,created_at,updated_at FROM users WHERE id=$1", id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepo) Create(ctx context.Context, user *auth.User) error {
	_, err := r.db.Exec(ctx, "INSERT INTO users (id, email, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5);", user.ID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("error in creating user")
	}

	return nil

}

func (r *RefreshTokenRepo) Create(ctx context.Context, refreshToken *auth.RefreshToken) error {
	_, err := r.db.Exec(ctx, "INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4, $5);", refreshToken.ID, refreshToken.UserID, refreshToken.Token, refreshToken.ExpiresAt, refreshToken.CreatedAt)
	if err != nil {
		return fmt.Errorf("Failed create refresh token")
	}
	return nil
}

func (r *RefreshTokenRepo) FindToken(ctx context.Context, token string) (*auth.RefreshToken, error) {
	refreshToken := &auth.RefreshToken{}
	err := r.db.QueryRow(ctx, "SELECT id,user_id,token,expires_at,created_at FROM refresh_tokens WHERE token=$1", token).Scan(&refreshToken.ID, &refreshToken.UserID, &refreshToken.Token, &refreshToken.ExpiresAt, &refreshToken.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("couldn't find the token row: %w", err)
	}

	return refreshToken, nil
}

func (r *RefreshTokenRepo) DeleteToken(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM refresh_tokens WHERE token=$1", token)
	if err != nil {
		return fmt.Errorf("error in deleting token row: %w", err)
	}
	return nil
}

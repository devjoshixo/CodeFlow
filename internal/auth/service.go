package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  UserRepository
	tokenRepo RefreshTokenRepository
	jwtSecret string
}

func NewAuthService(userRepo UserRepository, tokenRepo RefreshTokenRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, email string, password string) (*User, error) {
	existing, _ := s.userRepo.FindUserByEmail(ctx, email)
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hashed),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (string, string, error) {
	existing, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(existing.PasswordHash), []byte(password))
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"user_id": existing.ID,
		"email":   existing.Email,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign token: %w", err)
	}

	bytes := make([]byte, 32)
	rand.Read(bytes)
	refreshToken := hex.EncodeToString(bytes)

	refreshUser := &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    existing.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().AddDate(0, 0, 7),
		CreatedAt: time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, refreshUser); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return signed, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, token string) (string, string, error) {
	existingToken, err := s.tokenRepo.FindToken(ctx, token)
	if err != nil {
		return "", "", fmt.Errorf("Token is invalid: %w", err)
	}

	if existingToken.ExpiresAt.Before(time.Now()) {
		return "", "", ErrTokenExpired
	}

	if err := s.tokenRepo.DeleteToken(ctx, token); err != nil {
		return "", "", fmt.Errorf("couldn't delete the refresh token: %w", err)
	}

	claims := jwt.MapClaims{
		"user_id": existingToken.UserID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := newToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign token: %w", err)
	}

	bytes := make([]byte, 32)
	rand.Read(bytes)
	refreshToken := hex.EncodeToString(bytes)

	refreshUser := &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    existingToken.UserID,
		Token:     refreshToken,
		ExpiresAt: time.Now().AddDate(0, 0, 7),
		CreatedAt: time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, refreshUser); err != nil {
		return "", "", fmt.Errorf("error in creating new token: %w", err)
	}

	return signed, refreshToken, nil

}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	if err := s.tokenRepo.DeleteToken(ctx, token); err != nil {
		return fmt.Errorf("could not delete the token: %w", err)
	}

	return nil

}

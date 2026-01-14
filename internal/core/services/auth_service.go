package services

import (
	"context"
	"errors"
	"time"

	"github.com/damantine/multi-tenant-hosting/internal/core/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db        *gorm.DB
	secretKey []byte
}

func NewAuthService(db *gorm.DB, secret string) *AuthService {
	return &AuthService{
		db:        db,
		secretKey: []byte(secret),
	}
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) error {
	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := domain.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hashed),
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return err
	}
	return nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	var user domain.User
	if err := s.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *AuthService) ValidateToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	}, jwt.WithLeeway(5*time.Second))

	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sub, _ := claims.GetSubject()
		return uuid.Parse(sub)
	}

	return uuid.Nil, errors.New("invalid token")
}

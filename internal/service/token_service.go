package service

import (
	"auth-service/internal/domain"
	"auth-service/pkg/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	AuthService *AuthService
	Config      *config.Config
}

func NewTokenService(authService *AuthService) *TokenService {
	return &TokenService{AuthService: authService}
}

func (s *TokenService) CreateTokens(user *domain.User) (string, string, error) {
	AccessClaims := domain.Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.Config.JWTExpiration * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "token-service",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, AccessClaims)
	accessToken, err := token.SignedString([]byte(s.Config.JWTSecret))
	if err != nil {
		return "", "", err
	}

	RefreshClaims := domain.Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.Config.JWTRefreshExpiration * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "token-service",
		},
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, RefreshClaims)
	refreshToken, err := refreshTokenObj.SignedString([]byte(s.Config.JWTSecret))
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (s *TokenService) ValidateToken(token string) (*domain.Claims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &domain.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.Config.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := parsedToken.Claims.(*domain.Claims); ok && parsedToken.Valid {
		return claims, nil
	}
	return nil, jwt.ErrInvalidKey
}

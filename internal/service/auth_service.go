package service

import (
	"auth-service/internal/domain"
	"auth-service/internal/repository"
	"auth-service/pkg/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	UserRepo repository.UserRepository
	Config   *config.Config
}

func NewAuthService(userRepo repository.UserRepository, config *config.Config) *AuthService {
	return &AuthService{
		UserRepo: userRepo,
		Config:   config,
	}
}

func (s *AuthService) Login() {}

func (s *AuthService) Register() {}

func (s *AuthService) RefreshToken() {}

func (s *AuthService) Logout() {}

// CreateTokens generates a new access token and refresh token for the given user.
func (s *AuthService) CreateTokens(user *domain.User) (string, string, error) {
	AccessClaims := domain.Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.Config.JWTExpiration * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-service",
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
			Issuer:    "auth-service",
		},
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, RefreshClaims)
	refreshToken, err := refreshTokenObj.SignedString([]byte(s.Config.JWTSecret))
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// ValidateToken validates the given JWT token and returns the claims if the token is valid.
func (s *AuthService) ValidateToken(token string) (*domain.Claims, error) {
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

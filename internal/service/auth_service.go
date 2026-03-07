package service

import (
	"auth-service/internal/domain"
	"auth-service/internal/repository"
)

type AuthService struct {
	UserRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{
		UserRepo: userRepo,
	}
}

func (s *AuthService) Login() {}

func (s *AuthService) Register() {}

func (s *AuthService) ValidateToken(token string) (*domain.Claims, error) {

}

func (s *AuthService) RefreshToken() {}

func (s *AuthService) Logout() {}

func CreateTokens(user *domain.User) (string, string, error) {
	claims := domain.Claims{
		UserID: user.ID,
		Email:  user.Email,
	}

}

package service

import (
	"auth-service/internal/repository"
	"auth-service/pkg/config"
)

type AuthService struct {
	UserRepo       repository.UserRepository
	TokenService   *TokenService
	HashService    *HashService
	RoleService    *RoleService
	SessionService *SessionService
	Config         *config.Config
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenService *TokenService,
	hashService *HashService,
	roleService *RoleService,
	sessionService *SessionService,
	config *config.Config,
) *AuthService {
	return &AuthService{
		UserRepo:       userRepo,
		TokenService:   tokenService,
		HashService:    hashService,
		RoleService:    roleService,
		SessionService: sessionService,
		Config:         config,
	}
}

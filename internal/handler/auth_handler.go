package handler

import (
	"auth-service/internal/service"
)

type AuthHandler struct {
	AuthService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

// Login handles user login requests.
func (h *AuthHandler) Login() {}

// Register handles user registration requests.
func (h *AuthHandler) Register() {}

// RefreshToken handles token refresh requests.
func (h *AuthHandler) RefreshToken() {}

// Logout handles user logout requests.
func (h *AuthHandler) Logout() {}

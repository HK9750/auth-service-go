package service

import (
	"auth-service/internal/repository"
)

type SessionService struct {
	SessionRepo repository.SessionRepository
}

func NewSessionService(sessionRepo repository.SessionRepository) *SessionService {
	return &SessionService{SessionRepo: sessionRepo}
}

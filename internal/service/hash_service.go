package service

import (
	"golang.org/x/crypto/bcrypt"
)

type HashService struct {
	SaltRounds int
}

func NewHashService(saltRounds int) *HashService {
	return &HashService{SaltRounds: saltRounds}
}

func (s *HashService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.SaltRounds)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (s *HashService) ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

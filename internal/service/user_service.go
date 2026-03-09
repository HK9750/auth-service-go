package service

import "auth-service/internal/repository"

type UserService struct {
	UserRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		UserRepo: userRepo,
	}
}

func (s *UserService) GetUserByID() {}

func (s *UserService) GetUserByEmail() {}

func (s *UserService) CreateUser() {}

func (s *UserService) UpdateUser() {}

func (s *UserService) DeleteUser() {}

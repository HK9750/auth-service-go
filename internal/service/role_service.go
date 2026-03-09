package service

import (
	"auth-service/internal/repository"
)

type RoleService struct {
	RoleRepo repository.RoleRepository
}

func NewRoleService(roleRepo repository.RoleRepository) *RoleService {
	return &RoleService{RoleRepo: roleRepo}
}

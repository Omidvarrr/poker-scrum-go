package services

import (
	repositories "awesomeProject1/internal/repositories"
)

type AuthService struct {
	userRepository repositories.UserRepository
}

func NewAuthService(userRepo repositories.UserRepository) *AuthService {
	return &AuthService{
		userRepository: userRepo,
	}
}

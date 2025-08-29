package services

import (
	"awesomeProject1/internal/dto"
	"awesomeProject1/internal/models"
	repositories "awesomeProject1/internal/repositories"
	"awesomeProject1/internal/utils"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepository repositories.UserRepository
}

func NewAuthService(userRepo repositories.UserRepository) *AuthService {
	return &AuthService{
		userRepository: userRepo,
	}
}

func (as *AuthService) Login(request dto.LoginRequest) (*dto.LoginResponse, error) {
	var user models.User
	var err error

	// Try to find existing user
	user, err = as.userRepository.GetUserByEmail(request.Email)
	if err != nil {
		// If user not found, create a new one
		if errors.Is(err, repositories.ErrNotFound) {
			user, err = as.createNewUser(request)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// Verify password for existing user
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
			return nil, errors.New("invalid credentials")
		}
	}

	// Generate tokens
	accessToken, err := utils.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		NeedCompleteProfile: !user.ProfileCompleted,
		Access:              accessToken,
		Refresh:             refreshToken,
	}, nil
}

func (as *AuthService) createNewUser(request dto.LoginRequest) (models.User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	// Create new user
	newUser := models.User{
		Email:            request.Email,
		PasswordHash:     string(hashedPassword),
		ProfileCompleted: false,
	}

	// Save user to database
	createdUser, err := as.userRepository.CreateUser(newUser)
	if err != nil {
		return models.User{}, err
	}

	return createdUser, nil
}

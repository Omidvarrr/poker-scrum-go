package services

import (
	"awesomeProject1/internal/dto"
	"awesomeProject1/internal/repositories"
	"awesomeProject1/internal/utils"
)

type ProfileService struct {
	userRepo repositories.UserRepository
}

func NewProfileService(userRepo repositories.UserRepository) *ProfileService {
	return &ProfileService{
		userRepo: userRepo,
	}
}

func (ps *ProfileService) CompleteProfile(userID int, request dto.CompleteProfileRequest) error {
	return ps.userRepo.CompleteUserProfile(userID, request.Name, request.Avatar)
}

func (ps *ProfileService) UpdateProfile(userID int, request dto.UpdateProfileRequest) error {
	return ps.userRepo.UpdateUserProfile(userID, request.Name, request.Avatar)
}

func (ps *ProfileService) GetProfile(userID int, baseURL string) (dto.ProfileResponse, error) {
	user, err := ps.userRepo.GetUserById(userID)
	if err != nil {
		return dto.ProfileResponse{}, err
	}

	return dto.ProfileResponse{
		ID:               user.ID,
		Email:            user.Email,
		Name:             user.Name,
		Avatar:           utils.ConvertToURL(user.Avatar, baseURL),
		ProfileCompleted: user.ProfileCompleted,
	}, nil
}

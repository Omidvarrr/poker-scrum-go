package dto

type CompleteProfileRequest struct {
	Name   string `json:"name" form:"name"`
	Avatar string `json:"avatar" form:"avatar"`
}

type UpdateProfileRequest struct {
	Name   string `json:"name" form:"name"`
	Avatar string `json:"avatar" form:"avatar"`
}

type ProfileResponse struct {
	ID               int    `json:"id"`
	Email            string `json:"email"`
	Name             string `json:"name"`
	Avatar           string `json:"avatar"`
	ProfileCompleted bool   `json:"profile_completed"`
}

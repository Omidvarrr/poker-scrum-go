package dto

type CompleteProfileRequest struct {
	FirstName string `json:"first_name" form:"first_name"`
	LastName  string `json:"last_name" form:"last_name"`
	Avatar    string `json:"avatar" form:"avatar"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name" form:"first_name"`
	LastName  string `json:"last_name" form:"last_name"`
	Avatar    string `json:"avatar" form:"avatar"`
}

type ProfileResponse struct {
	ID               int    `json:"id"`
	Email            string `json:"email"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Avatar           string `json:"avatar"`
	ProfileCompleted bool   `json:"profile_completed"`
}

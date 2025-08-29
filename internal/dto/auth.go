package dto

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	NeedCompleteProfile bool   `json:"NeedCompleteProfile"`
	Access              string `json:"access"`
	Refresh             string `json:"refresh"`
}

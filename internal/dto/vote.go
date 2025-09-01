package dto

import "awesomeProject1/internal/models"

type VoteRequest struct {
	VoteValue string `json:"vote_value"`
}

type VoteResponse struct {
	SessionID  string        `json:"session_id"`
	IsRevealed bool          `json:"is_revealed"`
	Votes      []models.Vote `json:"votes,omitempty"`
}

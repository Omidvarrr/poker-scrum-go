package sqlite

import (
	"awesomeProject1/internal/models"
	repositories "awesomeProject1/internal/repositories"
	"gorm.io/gorm"
)

type voteRepository struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) repositories.VoteRepository {
	return &voteRepository{db: db}
}

func (r *voteRepository) CreateVoteSession(session models.VoteSession) (models.VoteSession, error) {
	result := r.db.Create(&session)
	if result.Error != nil {
		return models.VoteSession{}, result.Error
	}
	return session, nil
}

func (r *voteRepository) GetVoteSession(sessionID string) (models.VoteSession, error) {
	var session models.VoteSession
	result := r.db.Preload("Votes").Preload("Votes.User").First(&session, "id = ?", sessionID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return models.VoteSession{}, repositories.ErrNotFound
		}
		return models.VoteSession{}, result.Error
	}
	return session, nil
}

func (r *voteRepository) GetActiveVoteSession(roomID string) (models.VoteSession, error) {
	var session models.VoteSession
	result := r.db.Preload("Votes").Preload("Votes.User").
		Order("created_at DESC").
		First(&session, "room_id = ?", roomID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return models.VoteSession{}, repositories.ErrNotFound
		}
		return models.VoteSession{}, result.Error
	}
	return session, nil
}

func (r *voteRepository) RevealVoteSession(sessionID string) error {
	result := r.db.Model(&models.VoteSession{}).
		Where("id = ?", sessionID).
		Update("is_revealed", true)
	return result.Error
}

func (r *voteRepository) CastVote(vote models.Vote) error {
	// Use UPSERT to handle vote updates
	var existingVote models.Vote
	result := r.db.Where("room_id = ? AND user_id = ? AND session_id = ?",
		vote.RoomID, vote.UserID, vote.SessionID).First(&existingVote)

	if result.Error == gorm.ErrRecordNotFound {
		// No existing vote, create new one
		return r.db.Create(&vote).Error
	} else if result.Error != nil {
		return result.Error
	} else {
		// Update existing vote
		existingVote.VoteValue = vote.VoteValue
		existingVote.VotedAt = vote.VotedAt
		return r.db.Save(&existingVote).Error
	}
}

func (r *voteRepository) RemoveVote(userID int, roomID string, sessionID string) error {
	result := r.db.Delete(&models.Vote{}, "room_id = ? AND user_id = ? AND session_id = ?",
		roomID, userID, sessionID)
	return result.Error
}

func (r *voteRepository) GetVotesBySession(sessionID string) ([]models.Vote, error) {
	var votes []models.Vote
	result := r.db.Preload("User").Find(&votes, "session_id = ?", sessionID)
	return votes, result.Error
}

func (r *voteRepository) DeleteVotesBySession(sessionID string) error {
	result := r.db.Delete(&models.Vote{}, "session_id = ?", sessionID)
	return result.Error
}

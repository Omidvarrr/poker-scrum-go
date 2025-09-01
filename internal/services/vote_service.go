package services

import (
	"awesomeProject1/internal/dto"
	"awesomeProject1/internal/models"
	"awesomeProject1/internal/repositories"
	"errors"
	"github.com/google/uuid"
)

type VoteService struct {
	voteRepo repositories.VoteRepository
	roomRepo repositories.RoomRepository
}

func NewVoteService(voteRepo repositories.VoteRepository, roomRepo repositories.RoomRepository) *VoteService {
	return &VoteService{
		voteRepo: voteRepo,
		roomRepo: roomRepo,
	}
}

func (vs *VoteService) CastVote(userID int, roomID string, request dto.VoteRequest) error {
	session, err := vs.voteRepo.GetActiveVoteSession(roomID)
	if err != nil {
		if err == repositories.ErrNotFound {
			sessionID := uuid.New().String()
			newSession := models.VoteSession{
				ID:         sessionID,
				RoomID:     roomID,
				IsRevealed: false,
			}
			session, err = vs.voteRepo.CreateVoteSession(newSession)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// If vote value is empty, remove the user's vote
	if request.VoteValue == "" {
		return vs.voteRepo.RemoveVote(userID, roomID, session.ID)
	}

	vote := models.Vote{
		RoomID:    roomID,
		UserID:    userID,
		VoteValue: request.VoteValue,
		SessionID: session.ID,
	}

	return vs.voteRepo.CastVote(vote)
}

func (vs *VoteService) GetVotes(userID int, roomID string) (dto.VoteResponse, error) {
	session, err := vs.voteRepo.GetActiveVoteSession(roomID)
	if err != nil {
		if err == repositories.ErrNotFound {
			return dto.VoteResponse{
				SessionID:  "",
				IsRevealed: false,
				Votes:      []models.Vote{},
			}, nil
		}
		return dto.VoteResponse{}, err
	}

	response := dto.VoteResponse{
		SessionID:  session.ID,
		IsRevealed: session.IsRevealed,
		Votes:      session.Votes,
	}

	return response, nil
}

func (vs *VoteService) RevealVotes(userID int, roomID string) error {
	isAdmin, err := vs.isOwnerOrAdmin(userID, roomID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("only admins can reveal votes")
	}

	session, err := vs.voteRepo.GetActiveVoteSession(roomID)
	if err != nil {
		return err
	}

	return vs.voteRepo.RevealVoteSession(session.ID)
}

func (vs *VoteService) ResetVotes(userID int, roomID string) error {
	isAdmin, err := vs.isOwnerOrAdmin(userID, roomID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("only admins can reset votes")
	}

	session, err := vs.voteRepo.GetActiveVoteSession(roomID)
	if err != nil {
		if err == repositories.ErrNotFound {
			return nil
		}
		return err
	}

	err = vs.voteRepo.DeleteVotesBySession(session.ID)
	if err != nil {
		return err
	}

	sessionID := uuid.New().String()
	newSession := models.VoteSession{
		ID:         sessionID,
		RoomID:     roomID,
		IsRevealed: false,
	}

	_, err = vs.voteRepo.CreateVoteSession(newSession)
	return err
}

func (vs *VoteService) isOwnerOrAdmin(userID int, roomID string) (bool, error) {
	isOwner, err := vs.roomRepo.IsUserOwner(roomID, userID)
	if err != nil {
		return false, err
	}
	if isOwner {
		return true, nil
	}

	return false, nil
}

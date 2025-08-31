package sqlite

import (
	"awesomeProject1/internal/models"
	repositories "awesomeProject1/internal/repositories"
	"time"

	"gorm.io/gorm"
)

type joinRequestRepository struct {
	db *gorm.DB
}

func NewJoinRequestRepository(db *gorm.DB) repositories.JoinRequestRepository {
	return &joinRequestRepository{db: db}
}

func (r *joinRequestRepository) CreateJoinRequest(request models.JoinRequest) (models.JoinRequest, error) {
	result := r.db.Create(&request)
	if result.Error != nil {
		return models.JoinRequest{}, result.Error
	}
	return request, nil
}

func (r *joinRequestRepository) GetJoinRequestsByRoom(roomID string) ([]models.JoinRequest, error) {
	var requests []models.JoinRequest
	result := r.db.Preload("User").Find(
		&requests, "room_id = ? AND status = ?", roomID, models.JoinRequestStatusPending,
	)
	return requests, result.Error
}

func (r *joinRequestRepository) GetJoinRequestsByRooms(roomIDs []string) ([]models.JoinRequest, error) {
	var requests []models.JoinRequest
	result := r.db.Preload("User").Preload("Room").Find(
		&requests, "room_id IN (?) AND status = ?", roomIDs, models.JoinRequestStatusPending,
	)
	return requests, result.Error
}

func (r *joinRequestRepository) GetJoinRequestsByUser(userID int) ([]models.JoinRequest, error) {
	var requests []models.JoinRequest
	result := r.db.Preload("Room").Find(&requests, "user_id = ?", userID)
	return requests, result.Error
}

func (r *joinRequestRepository) GetJoinRequestById(id int) (models.JoinRequest, error) {
	var request models.JoinRequest
	result := r.db.Preload("User").Preload("Room").First(&request, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return models.JoinRequest{}, repositories.ErrNotFound
		}
		return models.JoinRequest{}, result.Error
	}
	return request, nil
}

func (r *joinRequestRepository) UpdateJoinRequestStatus(id int, status models.JoinRequestStatus, handledBy int) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     status,
		"handled_at": now,
		"handled_by": handledBy,
	}
	result := r.db.Model(&models.JoinRequest{}).Where("id = ?", id).Updates(updates)
	return result.Error
}

func (r *joinRequestRepository) HasPendingRequest(roomID string, userID int) (bool, error) {
	var count int64
	result := r.db.Model(&models.JoinRequest{}).
		Where("room_id = ? AND user_id = ? AND status = ?", roomID, userID, models.JoinRequestStatusPending).
		Count(&count)
	return count > 0, result.Error
}

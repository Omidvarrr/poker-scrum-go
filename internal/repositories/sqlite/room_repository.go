package sqlite

import (
	"awesomeProject1/internal/models"
	repositories "awesomeProject1/internal/repositories"

	"gorm.io/gorm"
)

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) repositories.RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) CreateRoom(room models.Room) (models.Room, error) {
	result := r.db.Create(&room)
	if result.Error != nil {
		return models.Room{}, result.Error
	}
	return room, nil
}

func (r *roomRepository) GetRoomById(id string) (models.Room, error) {
	var room models.Room
	result := r.db.First(&room, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return models.Room{}, repositories.ErrNotFound
		}
		return models.Room{}, result.Error
	}
	return room, nil
}

func (r *roomRepository) GetRoomsByOwner(ownerID int) ([]models.Room, error) {
	var rooms []models.Room
	result := r.db.Find(&rooms, "owner_id = ?", ownerID)
	return rooms, result.Error
}

func (r *roomRepository) GetAllRooms() ([]models.Room, error) {
	var rooms []models.Room
	result := r.db.Find(&rooms)
	return rooms, result.Error
}

func (r *roomRepository) GetRoomsByIDs(ids []string) ([]models.Room, error) {
	var rooms []models.Room
	result := r.db.Where("id IN ?", ids).Find(&rooms)
	return rooms, result.Error
}

func (r *roomRepository) UpdateRoom(id string, name string, avatar string) error {
	updates := map[string]interface{}{
		"name": name,
	}

	// Only update avatar if it's not empty
	if avatar != "" {
		updates["avatar"] = avatar
	}

	result := r.db.Model(&models.Room{}).Where("id = ?", id).Updates(updates)
	return result.Error
}

func (r *roomRepository) DeleteRoom(id string) error {
	result := r.db.Delete(&models.Room{}, "id = ?", id)
	return result.Error
}

func (r *roomRepository) IsUserOwner(roomID string, userID int) (bool, error) {
	var count int64
	result := r.db.Model(&models.Room{}).
		Where("id = ? AND owner_id = ?", roomID, userID).
		Count(&count)
	return count > 0, result.Error
}

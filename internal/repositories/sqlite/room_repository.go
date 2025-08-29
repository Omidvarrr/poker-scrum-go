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

func (r *roomRepository) GetRoomsByUser(userID int) ([]models.Room, error) {
	var rooms []models.Room
	result := r.db.Table("rooms").
		Joins("JOIN room_members ON rooms.id = room_members.room_id").
		Where("room_members.user_id = ?", userID).
		Find(&rooms)
	return rooms, result.Error
}

func (r *roomRepository) UpdateRoom(id string, name string, avatar string) error {
	updates := map[string]interface{}{
		"name":   name,
		"avatar": avatar,
	}
	result := r.db.Model(&models.Room{}).Where("id = ?", id).Updates(updates)
	return result.Error
}

func (r *roomRepository) DeleteRoom(id string) error {
	result := r.db.Delete(&models.Room{}, "id = ?", id)
	return result.Error
}

func (r *roomRepository) GetRoomMembers(roomID string) ([]models.RoomMember, error) {
	var members []models.RoomMember
	result := r.db.Find(&members, "room_id = ?", roomID)
	return members, result.Error
}

func (r *roomRepository) AddRoomMember(roomID string, userID int, role models.Role) error {
	member := models.RoomMember{
		RoomID: roomID,
		UserID: userID,
		Role:   role,
	}
	result := r.db.Create(&member)
	return result.Error
}

func (r *roomRepository) RemoveRoomMember(roomID string, userID int) error {
	result := r.db.Delete(&models.RoomMember{}, "room_id = ? AND user_id = ?", roomID, userID)
	return result.Error
}

func (r *roomRepository) UpdateMemberRole(roomID string, userID int, role models.Role) error {
	result := r.db.Model(&models.RoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Update("role", role)
	return result.Error
}

func (r *roomRepository) IsUserInRoom(roomID string, userID int) (bool, error) {
	var count int64
	result := r.db.Model(&models.RoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Count(&count)
	return count > 0, result.Error
}

func (r *roomRepository) IsUserAdmin(roomID string, userID int) (bool, error) {
	var count int64
	result := r.db.Model(&models.RoomMember{}).
		Where("room_id = ? AND user_id = ? AND role = ?", roomID, userID, models.RoleAdmin).
		Count(&count)
	return count > 0, result.Error
}

func (r *roomRepository) IsUserOwner(roomID string, userID int) (bool, error) {
	var count int64
	result := r.db.Model(&models.Room{}).
		Where("id = ? AND owner_id = ?", roomID, userID).
		Count(&count)
	return count > 0, result.Error
}

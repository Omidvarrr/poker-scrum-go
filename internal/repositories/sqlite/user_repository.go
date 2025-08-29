package sqlite

import (
	"awesomeProject1/internal/models"
	repositories "awesomeProject1/internal/repositories"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new SQLite implementation of UserRepository
func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUserById(id int) (models.User, error) {
	var user models.User
	result := r.db.First(&user, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return models.User{}, repositories.ErrNotFound
		}
		return models.User{}, result.Error
	}
	return user, nil
}

func (r *userRepository) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	result := r.db.First(&user, "email = ?", email)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return models.User{}, repositories.ErrNotFound
		}
		return models.User{}, result.Error
	}
	return user, nil
}

func (r *userRepository) CreateUser(user models.User) (models.User, error) {
	result := r.db.Create(&user)
	if result.Error != nil {
		return models.User{}, result.Error
	}
	return user, nil
}

func (r *userRepository) UpdateUserProfile(id int, name string, avatar string) error {
	updates := map[string]interface{}{
		"name":   name,
		"avatar": avatar,
	}
	result := r.db.Model(&models.User{}).Where("id = ?", id).Updates(updates)
	return result.Error
}

func (r *userRepository) CompleteUserProfile(id int, name string, avatar string) error {
	updates := map[string]interface{}{
		"name":              name,
		"avatar":            avatar,
		"profile_completed": true,
	}
	result := r.db.Model(&models.User{}).Where("id = ?", id).Updates(updates)
	return result.Error
}

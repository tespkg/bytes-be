package dao

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type User struct {
	Id        int64           `json:"id" gorm:"column:id"`
	Uuid      string          `json:"uuid" gorm:"column:uuid"`
	Email     *string         `json:"email" gorm:"column:email"`
	Phone     *string         `json:"phone" gorm:"column:phone"`
	Password  []byte          `json:"password" gorm:"column:password"`
	IsEnabled bool            `json:"isEnabled" gorm:"column:is_enabled"`
	CreatedAt *time.Time      `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt *time.Time      `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt *gorm.DeletedAt `json:"deletedAt" gorm:"column:deleted_at"`

	Roles []Role `json:"roles" gorm:"many2many:user_roles;"`

	UserRoles []UserRole `json:"userRoles" gorm:"foreignKey:user_id;"`

	Customers *Customer `json:"customers" gorm:"foreignKey:user_id;"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) Save(db *gorm.DB) error {
	return db.Save(u).Error
}

func GetUserById(db *gorm.DB, id int64) (*User, error) {
	var user *User
	if err := db.Model(&User{}).Where("id = ?", id).
		Preload("Roles").Preload("UserRoles").Preload("Customers").
		First(&user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return user, nil
}

func GetUserByPhoneAndRole(db *gorm.DB, phone, role string) (*User, error) {
	var user *User
	if err := db.Model(&User{}).
		Joins("LEFT JOIN user_roles ur ON users.id = ur.user_id").
		Joins("LEFT JOIN roles r ON ur.role_id = r.id").
		Where("users.phone = ? AND r.name = ?", phone, role).
		Distinct("users.*").
		Preload("Roles").Preload("UserRoles").Preload("Customers").
		First(&user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return user, nil
}

func GetUserByEmailAndRole(db *gorm.DB, email, role string) (*User, error) {
	var user *User
	if err := db.Model(&User{}).
		Joins("LEFT JOIN user_roles ur ON users.id = ur.user_id").
		Joins("LEFT JOIN roles r ON ur.role_id = r.id").
		Where("users.email = ? AND r.name = ?", email, role).
		Distinct("users.*").
		Preload("Roles").Preload("UserRoles").Preload("Customers").
		First(&user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return user, nil
}

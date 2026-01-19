package dao

import (
	"gorm.io/gorm"
	"time"
)

type UserRole struct {
	Id        int64           `json:"id" gorm:"column:id"`
	UserId    int64           `json:"userId" gorm:"column:user_id"`
	RoleId    int64           `json:"roleId" gorm:"column:role_id"`
	CreatedAt *time.Time      `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt *time.Time      `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt *gorm.DeletedAt `json:"deletedAt" gorm:"column:deleted_at"`
}

func (u *UserRole) TableName() string {
	return "user_roles"
}

func (u *UserRole) Save(db *gorm.DB) error {
	return db.Save(u).Error
}

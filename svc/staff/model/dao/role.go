package dao

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	RoleAdmin    = "admin"
	RoleCustomer = "customer"
	RoleMerchant = "merchant"
	RoleDriver   = "driver"
)

type Role struct {
	Id   int64  `json:"id" gorm:"column:id"`
	Role string `json:"role" gorm:"column:role"`
}

func (r *Role) TableName() string {
	return "roles"
}

func (r *Role) Save(db *gorm.DB) error {
	return db.Save(r).Error
}

func GetRoleByName(db *gorm.DB, roleName string) (*Role, error) {
	var role *Role
	if err := db.Model(&Role{}).
		Where("role = ?", roleName).First(&role).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return role, nil
}

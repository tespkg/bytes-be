package dao

import "gorm.io/gorm"

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

package dao

import (
	"gorm.io/gorm"
	"time"
)

const (
	CustomerGenderMale   = 0
	CustomerGenderFemale = 1
)

type Customer struct {
	Id          int64           `json:"id" gorm:"column:id"`
	UserId      int64           `json:"userId" gorm:"column:user_id"`
	FirstName   *string         `json:"firstName" gorm:"column:first_name"`
	LastName    *string         `json:"lastName" gorm:"column:last_name"`
	Gender      *int            `json:"gender" gorm:"column:gender"`
	Birthday    *string         `json:"birthday" gorm:"column:birthday"`
	Nationality *string         `json:"nationality" gorm:"column:nationality"`
	CreatedAt   *time.Time      `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   *time.Time      `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt   *gorm.DeletedAt `json:"deletedAt" gorm:"column:deleted_at"`

	CustomerAddresses []CustomerAddress `json:"customerAddresses" gorm:"foreignKey:customer_id;"`
}

func (c *Customer) TableName() string {
	return "customers"
}

func (c *Customer) Save(db *gorm.DB) error {
	return db.Save(c).Error
}

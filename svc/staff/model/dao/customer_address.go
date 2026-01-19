package dao

import (
	"gorm.io/gorm"
	"time"
)

type CustomerAddress struct {
	Id         int64           `json:"id" gorm:"column:id"`
	CustomerId int64           `json:"customerId" gorm:"column:customer_id"`
	Country    *string         `json:"country" gorm:"column:country"`
	State      *string         `json:"state" gorm:"column:state"`
	City       *string         `json:"city" gorm:"column:city"`
	Street     *string         `json:"street" gorm:"column:street"`
	ZipCode    *string         `json:"zipCode" gorm:"column:zip_code"`
	Address    string          `json:"address" gorm:"column:address"`
	Longitude  float64         `json:"longitude" gorm:"column:longitude"`
	Latitude   float64         `json:"latitude" gorm:"column:latitude"`
	CreatedAt  *time.Time      `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt  *time.Time      `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt  *gorm.DeletedAt `json:"deletedAt" gorm:"column:deleted_at"`
}

func (c *CustomerAddress) TableName() string {
	return "customer_addresses"
}

func (c *CustomerAddress) Save(db *gorm.DB) error {
	return db.Save(c).Error
}

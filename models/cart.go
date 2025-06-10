package models

import (
	"gorm.io/gorm"
)

type Cart struct {
	gorm.Model
	UserID *uint      `json:"user_id"` // nullable for guests
	User   *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Items  []CartItem `json:"items"`
}

type CartItem struct {
	gorm.Model
	CartID    uint     `json:"cart_id"`
	Cart      *Cart    `json:"-" gorm:"foreignKey:CartID"`
	ProductID uint     `json:"product_id"`
	Product   *Product `json:"product" gorm:"foreignKey:ProductID"`
	Quantity  int      `json:"quantity"`
}

package entity

import "gorm.io/gorm"

type ItemCategory struct {
	gorm.Model

	Name string `gorm:"type:text;size:255;not null"`
}

func (ItemCategory) TableName() string {
	return "item_categories"
}

package entity

import (
	"time"

	"gorm.io/gorm"
)

type Item struct {
	gorm.Model

	ReleasedAt time.Time     `gorm:"type:date"`
	StartsAt   time.Time     `gorm:"type:timestamptz"`
	Category   *ItemCategory `gorm:"foreignKey:CategoryID;constraint:OnDelete:RESTRICT"`
	Name       string        `gorm:"type:text;size:255;not null"`
	Notes      string        `gorm:"type:text"`
	Contact    string        `gorm:"type:text;size:255"`
	Stock      uint          `gorm:"type:integer;not null;default:0"`
	Price      uint          `gorm:"type:bigint;not null;default:0"`
	CategoryID uint          `gorm:"index;not null"`
	Active     bool          `gorm:"type:boolean;not null;default:false"`
}

func (Item) TableName() string {
	return "items"
}

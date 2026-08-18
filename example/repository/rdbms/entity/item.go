package entity

import (
	"time"

	"gorm.io/gorm"
)

type Item struct {
	gorm.Model

	ReleasedAt time.Time     `gorm:"type:date;not null"`
	StartsAt   time.Time     `gorm:"type:timestamptz;not null"`
	Category   *ItemCategory `gorm:"foreignKey:CategoryID;constraint:OnDelete:RESTRICT"`
	Name       string        `gorm:"type:text;size:255;not null"`
	Notes      string        `gorm:"type:text"`
	Contact    string        `gorm:"type:text;size:255"`
	Stock      uint          `gorm:"type:integer;default:0"`
	Price      uint          `gorm:"type:bigint;default:0;not null"`
	CategoryID uint          `gorm:"index;not null"`
	Active     bool          `gorm:"type:boolean;default:false"`
}

func (Item) TableName() string {
	return "items"
}

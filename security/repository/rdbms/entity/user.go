package entity

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Username string         `gorm:"uniqueIndex:idx_user_username;column:username;type:text;not null;size:100"`
	Password string         `gorm:"column:password;type:text;not null;size:255"`
	Roles    pq.StringArray `gorm:"column:roles;type:text[]"`
}

func (User) TableName() string {
	return "users"
}

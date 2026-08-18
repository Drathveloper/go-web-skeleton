package rdbms

import "gorm.io/gorm"

const (
	minPage         = 1
	minPageSize     = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(database *gorm.DB) *gorm.DB {
		if page < minPage {
			page = minPage
		}
		if pageSize > maxPageSize {
			pageSize = maxPageSize
		} else if pageSize < minPageSize {
			pageSize = defaultPageSize
		}

		offset := (page - 1) * pageSize
		return database.Offset(offset).Limit(pageSize)
	}
}

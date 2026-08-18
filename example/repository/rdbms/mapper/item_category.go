package mapper

import (
	"gorm.io/gorm"

	"github.com/Drathveloper/go-web-skeleton/example/domain"
	"github.com/Drathveloper/go-web-skeleton/example/repository/rdbms/entity"
)

func EntityItemCategoriesToDomainItemCategories(entities []entity.ItemCategory) []domain.ItemCategory {
	if len(entities) == 0 {
		return []domain.ItemCategory{}
	}
	itemCategories := make([]domain.ItemCategory, 0, len(entities))
	for _, e := range entities {
		itemCategories = append(itemCategories, *EntityItemCategoryToDomainItemCategory(&e))
	}
	return itemCategories
}

func EntityItemCategoryToDomainItemCategory(itemCategory *entity.ItemCategory) *domain.ItemCategory {
	if itemCategory == nil {
		return &domain.ItemCategory{}
	}
	return &domain.ItemCategory{
		ID:   itemCategory.ID,
		Name: itemCategory.Name,
	}
}

func DomainItemCategoryToEntityItemCategory(itemCategory *domain.ItemCategory) *entity.ItemCategory {
	return &entity.ItemCategory{
		Model: gorm.Model{ID: itemCategory.ID},
		Name:  itemCategory.Name,
	}
}

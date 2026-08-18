package mapper

import (
	"gorm.io/gorm"

	"github.com/Drathveloper/go-web-skeleton/example/domain"
	"github.com/Drathveloper/go-web-skeleton/example/repository/rdbms/entity"
)

func EntityItemsToDomainItems(entities []entity.Item) []domain.Item {
	if len(entities) == 0 {
		return []domain.Item{}
	}
	items := make([]domain.Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, *EntityItemToDomainItem(&e))
	}
	return items
}

func EntityItemToDomainItem(item *entity.Item) *domain.Item {
	if item == nil {
		return &domain.Item{}
	}
	return &domain.Item{
		ID:         item.ID,
		Name:       item.Name,
		Notes:      item.Notes,
		Stock:      item.Stock,
		Price:      item.Price,
		Contact:    item.Contact,
		ReleasedAt: item.ReleasedAt,
		StartsAt:   item.StartsAt,
		CategoryID: item.CategoryID,
		Active:     item.Active,
		Category:   entityItemCategoryOrNil(item.Category),
	}
}

func DomainItemToEntityItem(item *domain.Item) *entity.Item {
	return &entity.Item{
		Model:      gorm.Model{ID: item.ID},
		Name:       item.Name,
		Notes:      item.Notes,
		Stock:      item.Stock,
		Price:      item.Price,
		Contact:    item.Contact,
		ReleasedAt: item.ReleasedAt,
		StartsAt:   item.StartsAt,
		CategoryID: item.CategoryID,
		Active:     item.Active,
	}
}

// entityItemCategoryOrNil keeps an absent relation absent. The plain mapper
// returns a zero value for nil, which would make the view render a item category
// named "" with id 0 instead of showing nothing.
func entityItemCategoryOrNil(itemCategory *entity.ItemCategory) *domain.ItemCategory {
	if itemCategory == nil {
		return nil
	}
	return EntityItemCategoryToDomainItemCategory(itemCategory)
}

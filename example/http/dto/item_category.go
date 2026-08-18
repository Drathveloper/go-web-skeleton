package dto

type ItemCategory struct {
	Name string `binding:"required" form:"name"`
	ID   uint   `form:"id"`
}

type ItemCategoriesResponse struct {
	ItemCategories []ItemCategory
}

type ItemCategoryResponse struct {
	ItemCategory *ItemCategory
}

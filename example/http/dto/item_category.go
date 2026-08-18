package dto

// ItemCategory is the form payload. Tags are `binding:` and never `validate:` — gin's
// binder only reads the former, so a `validate:` tag here would look like
// validation while silently never running.
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

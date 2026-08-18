package dto

// Item is the form payload. Tags are `binding:` and never `validate:` — gin's
// binder only reads the former, so a `validate:` tag here would look like
// validation while silently never running.
//
// Price is a string, not a number: money arrives as "12.34" and is converted to
// cents by the mapper, with `decimal2` enforcing the two-decimal shape.
type Item struct {
	Name       string `binding:"required"          form:"name"`
	Notes      string `form:"notes"`
	Price      string `binding:"required,decimal2" form:"price"`
	Contact    string `binding:"omitempty,email"   form:"contact"`
	ReleasedAt string `binding:"required"          form:"released_at"`
	StartsAt   string `binding:"required"          form:"starts_at"`
	ID         uint   `form:"id"`
	Stock      uint   `binding:"gte=0"             form:"stock"`
	CategoryID uint   `binding:"required"          form:"category_id"`
	Active     bool   `form:"active"`
}

type ItemsResponse struct {
	Items []Item
}

type ItemResponse struct {
	Item *Item
}

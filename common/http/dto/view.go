package dto

type ViewResponse[T any] struct {
	Data        *T
	Language    string
	CSRFToken   string
	User        string
	Msgs        []AlertMessage
	Breadcrumbs []string
	IsLogged    bool
}

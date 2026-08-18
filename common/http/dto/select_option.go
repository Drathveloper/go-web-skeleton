package dto

type SelectOption struct {
	Value any
	Label string
}

type MultiSelectOption struct {
	Value    any
	Label    string
	Selected bool
}

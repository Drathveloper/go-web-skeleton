package dto

// Field types understood by templates/files/components/form/input.gohtml. They
// line up one to one with the field types the scaffold generator accepts,
// because a generated form is nothing but a sequence of these.
const (
	// FieldTypeText renders <input type="text">, for a string field.
	FieldTypeText = "text"
	// FieldTypeTextarea renders <textarea>, for a text field.
	FieldTypeTextarea = "textarea"
	// FieldTypeNumber renders <input type="number">, for int and uint fields.
	FieldTypeNumber = "number"
	// FieldTypeCheckbox renders <input type="checkbox" value="true">, for bool.
	FieldTypeCheckbox = "checkbox"
	// FieldTypeDate renders <input type="date">, value formatted 2006-01-02.
	FieldTypeDate = "date"
	// FieldTypeDateTime renders <input type="datetime-local">, value formatted
	// 2006-01-02T15:04.
	FieldTypeDateTime = "datetime"
	// FieldTypeMoney renders <input type="text" inputmode="decimal">, because
	// the money binding is decimal2 over a string, not a browser number.
	FieldTypeMoney = "money"
	// FieldTypeEmail renders <input type="email">.
	FieldTypeEmail = "email"
	// FieldTypeSelect renders a <select> fed by Options, for a ref field.
	FieldTypeSelect = "select"
	// FieldTypeMultiSelect renders a multiple <select>, for a ref list.
	FieldTypeMultiSelect = "multiselect"
	// FieldTypePassword renders <input type="password">.
	FieldTypePassword = "password"
	// FieldTypeHidden renders <input type="hidden">, label and hint ignored.
	FieldTypeHidden = "hidden"
	// FieldTypeCheckboxGroup renders one checkbox per StringOptions entry, all
	// sharing Name, so the form posts a []string.
	FieldTypeCheckboxGroup = "checkboxgroup"
)

// FormView drives templates/files/components/form/modal.gohtml, the modal card
// every create and update form is made of. The generated module contributes the
// Fields; the card, the error banner, the CSRF token and the footer are the
// component's business.
//
// On a validation failure the handler answers 422 with the very same view, with
// Error set to a localized, user-safe message and the submitted values still in
// the fields.
type FormView struct {
	Title string
	// Action is the URL the form posts to over HTMX.
	Action string
	// SubmitLabel falls back to the core key actions.save.
	SubmitLabel string
	// Error is a localized, user-safe message shown above the fields. It is
	// never an err.Error(): that goes to the log.
	Error     string
	Language  string
	CSRFToken string
	Fields    []FormField
}

// FormField is one labelled control. Value, Checked and Selected are what gets
// rendered back after a failed submit, so the user never retypes a form.
type FormField struct {
	// Name is the form field name and the DOM id of the control.
	Name  string
	Label string
	// Type is one of the FieldType* constants.
	Type string
	// Value is the already formatted string value of the control.
	Value       string
	Placeholder string
	// Hint is a small note under the control, hidden while Error is set.
	Hint string
	// Error is a localized per-field message rendered under the control.
	Error string
	// Options feeds FieldTypeSelect and FieldTypeMultiSelect.
	Options []SelectOption
	// StringOptions feeds FieldTypeCheckboxGroup, whose members are bare
	// values with no separate label — a role list, a set of flags.
	StringOptions []string
	// Selected holds the option values, stringified, that are selected.
	Selected []string
	// Rows is the textarea height. Zero falls back to the component default.
	Rows     int
	Required bool
	Checked  bool
	Disabled bool
}

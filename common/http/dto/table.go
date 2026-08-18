package dto

// TableView is the complete parameter set of the shared table components under
// templates/files/components/table. Every listing in the application renders
// through them, so anything that differs between two listings has to be a field
// here. A value that lives in a page template instead is a value that has to be
// edited once per generated module the day the table design changes.
//
// Strings that reach the screen arrive already localized: whoever builds the
// view has the locale, the component does not. Language is carried anyway for
// the handful of generic strings (delete confirmation, empty state, search
// placeholder) that belong to the core catalog rather than to the module.
type TableView struct {
	// ID is the plural slug of the entity, "items". The components derive
	// "<ID>-table", "<ID>-tbody" and "<ID>-table-search" from it, and the
	// out-of-band swap a create handler answers with targets "#<ID>-tbody".
	ID string
	// Title and Subtitle head the page above the table.
	Title    string
	Subtitle string
	// SearchPlaceholder falls back to the core key table.search_text.
	SearchPlaceholder string
	// NewURL is fetched into #modal by the primary button. Empty hides it.
	NewURL   string
	NewLabel string
	// EmptyText falls back to the core key table.no_results.
	EmptyText string
	Language  string
	Columns   []TableColumn
	Rows      []TableRow
}

// TableColumn describes one <th>. The actions column is a column like any
// other: give it the localized actions.* label and Align "right".
type TableColumn struct {
	Label string
	// SortKey turns the header into a client side sort control. Empty leaves
	// the column unsorted; the value itself is only a DOM identifier.
	SortKey string
	// Align is "", "right" or "center".
	Align string
}

// TableRow is both a row of the listing and the unit of the HTMX row swap: the
// same value renders inside the table on a full page load and on its own
// through the fragments/table/row template after a create or an update.
type TableRow struct {
	// ID is the DOM id of the <tr>, "item-row-42". Delete targets it.
	ID string
	// EditURL is fetched into #modal. DeleteURL is posted with the CSRF token
	// and swaps the row out. Either may be empty to drop that button.
	EditURL   string
	DeleteURL string
	// ConfirmDelete falls back to the core key table.confirm_delete.
	ConfirmDelete string
	CSRFToken     string
	Language      string
	// OOB drives the out-of-band swap when the row is rendered standalone:
	// "" inside a full listing, "true" to replace the row in place after an
	// update, "afterbegin:#items-tbody" to prepend it after a create.
	OOB   string
	Cells []TableCell
}

// TableCell is a single <td>. It is deliberately not a string of HTML: the
// generator fills these in, and a template.HTML here would turn every generated
// mapper into an injection site.
type TableCell struct {
	Text string
	// Secondary is a smaller muted line under Text: the "#42" id line under a
	// name, a phone under a contact. Empty renders a single line cell.
	Secondary string
	// Badge renders Text as a badge of that variant: "brand", "neutral",
	// "success", "warning", "danger" or "info". Empty renders plain text.
	Badge string
	// Align is "", "right" or "center".
	Align string
	// Mono renders Text with the tabular numeric font, for amounts and codes.
	Mono bool
	// SecondaryMono does the same for Secondary, as the id line does.
	SecondaryMono bool
	// Strong emphasises Text. Conventionally the first column of a listing.
	Strong bool
}

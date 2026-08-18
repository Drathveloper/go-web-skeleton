package templates_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/http/dto"
	"github.com/Drathveloper/go-web-skeleton/common/http/templates"
	"github.com/Drathveloper/go-web-skeleton/common/i18n"
)

// A template that compiles into the binary but blows up on render is worse than
// a missing one, and html/template only reports that at request time. Every
// template the skeleton ships is executed here with representative data.

func newEngine(t *testing.T) *gin.Engine {
	t.Helper()
	require.NoError(t, i18n.InitializeI18n())
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	require.NoError(t, templates.InitializeTemplateRenderer(engine))
	return engine
}

// render executes one registered template through the real renderer.
//
// The status code is not enough to tell a good render from a bad one: gin
// writes the header before executing, so a template that blows up half way
// still answers 200 with a truncated body. gin pushes the execution error onto
// c.Errors, and that is what actually has to be empty.
func render(t *testing.T, engine *gin.Engine, status int, name string, data any) string {
	t.Helper()
	var renderErrors []string
	engine.GET("/"+name, func(c *gin.Context) {
		c.HTML(status, name, data)
		for _, ginErr := range c.Errors {
			renderErrors = append(renderErrors, ginErr.Error())
		}
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/"+name, nil)
	engine.ServeHTTP(recorder, request)
	require.Empty(t, renderErrors, "template %q failed to execute", name)
	require.Equal(t, status, recorder.Code)
	require.NotEmpty(t, recorder.Body.String())
	return recorder.Body.String()
}

func tableRow(oob string) dto.TableRow {
	return dto.TableRow{
		ID:            "item-row-42",
		EditURL:       "/item/42/edit",
		DeleteURL:     "/item/42/delete",
		ConfirmDelete: "Delete this item?",
		CSRFToken:     "csrf-token-value",
		Language:      "en",
		OOB:           oob,
		Cells: []dto.TableCell{
			{Text: "Regulator", Secondary: "#42", Strong: true, SecondaryMono: true},
			{Text: "Rental", Badge: "brand"},
			{Text: "", Badge: "neutral"},
			{Text: "129.90", Align: "right", Mono: true},
		},
	}
}

func tableView(rows []dto.TableRow) dto.TableView {
	return dto.TableView{
		ID:       "items",
		Title:    "Items",
		Subtitle: "Everything you rent out",
		NewURL:   "/item/new",
		NewLabel: "New item",
		Language: "en",
		Columns: []dto.TableColumn{
			{Label: "Name", SortKey: "name"},
			{Label: "Category"},
			{Label: "Certification"},
			{Label: "Price", SortKey: "price", Align: "right"},
			{Label: "Actions", Align: "right"},
		},
		Rows: rows,
	}
}

func formView(errMsg string) dto.FormView {
	return dto.FormView{
		Title:       "New item",
		Action:      "/item/new",
		SubmitLabel: "Create item",
		Error:       errMsg,
		Language:    "en",
		CSRFToken:   "csrf-token-value",
		Fields: []dto.FormField{
			{Name: "name", Label: "Name", Type: dto.FieldTypeText, Value: "Regulator", Required: true},
			{Name: "notes", Label: "Notes", Type: dto.FieldTypeTextarea, Value: "Serviced in May", Rows: 3},
			{Name: "units", Label: "Units", Type: dto.FieldTypeNumber, Value: "7"},
			{Name: "active", Label: "Active", Type: dto.FieldTypeCheckbox, Checked: true},
			{Name: "bought_on", Label: "Bought on", Type: dto.FieldTypeDate, Value: "2026-08-18"},
			{Name: "checked_at", Label: "Checked at", Type: dto.FieldTypeDateTime, Value: "2026-08-18T09:30"},
			{Name: "price", Label: "Price", Type: dto.FieldTypeMoney, Value: "129.90", Hint: "Two decimals"},
			{Name: "contact", Label: "Contact", Type: dto.FieldTypeEmail, Value: "not-an-email", Error: "Enter a valid email address."},
			{
				Name: "category_id", Label: "Category", Type: dto.FieldTypeSelect,
				Options:  []dto.SelectOption{{Value: uint(1), Label: "Rental"}, {Value: uint(2), Label: "Sale"}},
				Selected: []string{"2"},
			},
			{
				Name: "branch_ids", Label: "Branches", Type: dto.FieldTypeMultiSelect,
				Options:  []dto.SelectOption{{Value: uint(7), Label: "Pier"}, {Value: uint(8), Label: "Marina"}},
				Selected: []string{"7"},
			},
			{Name: "password", Label: "Password", Type: dto.FieldTypePassword},
			{Name: "id", Type: dto.FieldTypeHidden, Value: "42"},
		},
	}
}

func TestHomePageRenders(t *testing.T) {
	t.Parallel()
	body := render(t, newEngine(t), http.StatusOK, "home/home", dto.ViewResponse[struct{}]{
		Language: "en",
		User:     "admin",
		IsLogged: true,
		Msgs:     []dto.AlertMessage{dto.NewSuccessMsg("Saved", "The item was saved.")},
	})
	require.Contains(t, body, "<title>Home</title>")
	require.Contains(t, body, "Manage users")
	require.Contains(t, body, `class="sidebar"`)
	require.NotContains(t, body, "home.heading", "an untranslated key leaked into the page")
}

func TestTableFragmentRenders(t *testing.T) {
	t.Parallel()
	body := render(t, newEngine(t), http.StatusOK, "fragments/table/table", tableView([]dto.TableRow{tableRow("")}))
	require.Contains(t, body, `id="items-table"`)
	require.Contains(t, body, `id="items-tbody"`)
	require.Contains(t, body, `id="items-table-search"`)
	require.Contains(t, body, `data-sort-key="name"`)
	require.Contains(t, body, `<script>initDataTable("items-table");</script>`)
	require.Contains(t, body, `<span class="badge badge-brand">Rental</span>`)
	require.Contains(t, body, `<td class="text-right num">`)
	require.Contains(t, body, `hx-swap="outerHTML swap:200ms"`)
	require.Contains(t, body, `hx-confirm="Delete this item?"`)
	require.Contains(t, body, "<tr data-empty-row hidden>")
}

func TestTableFragmentRendersEmptyState(t *testing.T) {
	t.Parallel()
	body := render(t, newEngine(t), http.StatusOK, "fragments/table/table", tableView(nil))
	require.Contains(t, body, "No results available")
	require.Contains(t, body, "<tr data-empty-row>")
}

func TestRowFragmentSpeaksTheHTMXContract(t *testing.T) {
	t.Parallel()
	engine := newEngine(t)

	created := render(t, engine, http.StatusOK, "fragments/table/row", tableRow("afterbegin:#items-tbody"))
	require.Contains(t, created, `<tbody hx-swap-oob="afterbegin:#items-tbody">`)
	require.Contains(t, created, `<tr id="item-row-42">`)

	engine = newEngine(t)
	updated := render(t, engine, http.StatusOK, "fragments/table/row", tableRow("true"))
	require.Contains(t, updated, `<tr id="item-row-42" hx-swap-oob="true">`)
	require.NotContains(t, updated, "<tbody")
}

func TestFormFragmentRendersEveryFieldType(t *testing.T) {
	t.Parallel()
	body := render(t, newEngine(t), http.StatusOK, "fragments/form/modal", formView(""))
	for _, want := range []string{
		`type="text"`, `<textarea`, `type="number"`, `type="checkbox"`,
		`type="date"`, `type="datetime-local"`, `inputmode="decimal"`,
		`type="email"`, `type="password"`, `type="hidden"`,
		`<select id="category_id"`, `<select id="branch_ids"`, "multiple",
		`<option value="2" selected>Sale</option>`,
		`<option value="7" selected>Pier</option>`,
		`name="csrf_token" value="csrf-token-value"`,
		`hx-post="/item/new"`, `hx-target="#modal"`,
		"Hold Ctrl or Cmd to select several.",
	} {
		require.Contains(t, body, want)
	}
	require.NotContains(t, body, "form.multiselect_hint")
}

func TestFormFragmentRendersValidationErrorInlineWith422(t *testing.T) {
	t.Parallel()
	body := render(t, newEngine(t), http.StatusUnprocessableEntity, "fragments/form/modal",
		formView("The submitted data is not valid."))
	require.Contains(t, body, `class="alert alert-danger"`)
	require.Contains(t, body, "The submitted data is not valid.")
	require.Contains(t, body, `<p class="form-error">Enter a valid email address.</p>`)
	require.Contains(t, body, `aria-invalid="true"`)
	require.Contains(t, body, `value="Regulator"`, "a rejected submit must not empty the form")
}

func TestInputFragmentRendersOnItsOwn(t *testing.T) {
	t.Parallel()
	body := render(t, newEngine(t), http.StatusOK, "fragments/form/input", map[string]any{
		"Language": "en",
		"Field":    dto.FormField{Name: "name", Label: "Name", Type: dto.FieldTypeText, Required: true},
	})
	require.Contains(t, body, `class="form-label req"`)
	require.Contains(t, body, `id="name"`)
}

func TestErrorTemplateRendersWithTheRealStatus(t *testing.T) {
	t.Parallel()
	body := render(t, newEngine(t), http.StatusInternalServerError, "error", dto.ViewResponse[struct{}]{
		Language: "es",
		Msgs: []dto.AlertMessage{dto.NewErrorMsg(http.StatusInternalServerError, "Error",
			"Ha ocurrido un error inesperado. Inténtalo de nuevo más tarde.")},
	})
	require.Contains(t, body, "Ha ocurrido un error inesperado")
	require.NotContains(t, strings.ToLower(body), "sql", "no internal detail may reach the page")
}

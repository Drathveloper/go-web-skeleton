package mapper

import (
	"fmt"
	"strconv"
	"time"

	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	commondto "github.com/Drathveloper/go-web-skeleton/common/http/dto"
	commonmapper "github.com/Drathveloper/go-web-skeleton/common/http/mapper"
	"github.com/Drathveloper/go-web-skeleton/example/domain"
	"github.com/Drathveloper/go-web-skeleton/example/http/dto"
)

const (
	itemBasePath  = "/item"
	itemTableID   = "items"
	itemRowPrefix = "item-row-"

	// dateLayout and dateTimeLayout are the wire formats of <input type="date">
	// and <input type="datetime-local">. They are not display formats: what the
	// user sees is the browser's locale rendering of these values.
	dateLayout     = "2006-01-02"
	dateTimeLayout = "2006-01-02T15:04"
)

func DomainItemsToItemsViewResponse(
	session *commondomain.Session, items []domain.Item) *commondto.ViewResponse[commondto.TableView] {
	data := DomainItemsToTableView(session, items)
	return commonmapper.MapDataToViewResponse(&data, getItemBreadcrumb(session), session)
}

func DomainItemsToTableView(session *commondomain.Session, items []domain.Item) commondto.TableView {
	rows := make([]commondto.TableRow, 0, len(items))
	for i := range items {
		rows = append(rows, DomainItemToTableRow(session, &items[i], ""))
	}
	return commondto.TableView{
		ID:       itemTableID,
		Title:    localize(session, "item.title"),
		Subtitle: localize(session, "item.subtitle"),
		NewURL:   itemBasePath + "/new",
		NewLabel: localize(session, "item.actions.new"),
		Language: session.Language,
		Columns: []commondto.TableColumn{
			{Label: localize(session, "item.fields.name"), SortKey: "name"},
			{Label: localize(session, "item.fields.category_id"), SortKey: "category"},
			{Label: localize(session, "item.fields.stock"), SortKey: "stock", Align: "right"},
			{Label: localize(session, "item.fields.price"), SortKey: "price", Align: "right"},
			{Label: localize(session, "item.fields.released_at"), SortKey: "released_at"},
			{Label: localize(session, "item.fields.active"), SortKey: "active", Align: "center"},
			{Label: localize(session, "actions.title"), Align: "right"},
		},
		Rows: rows,
	}
}

func DomainItemToTableRow(
	session *commondomain.Session, item *domain.Item, oob string) commondto.TableRow {
	itemID := strconv.FormatUint(uint64(item.ID), 10)
	return commondto.TableRow{
		ID:        itemRowPrefix + itemID,
		EditURL:   itemBasePath + "/" + itemID + "/edit",
		DeleteURL: itemBasePath + "/" + itemID + "/delete",
		CSRFToken: session.CSRFToken,
		Language:  session.Language,
		OOB:       oob,
		Cells: []commondto.TableCell{
			{Text: item.Name, Secondary: "#" + itemID, SecondaryMono: true, Strong: true},
			{Text: itemCategoryName(item)},
			{Text: strconv.FormatUint(uint64(item.Stock), 10), Align: "right", Mono: true},
			{Text: *commonmapper.UintToDecimalString(item.Price), Align: "right", Mono: true},
			{Text: formatDate(item.ReleasedAt)},
			{Text: activeLabel(session, item.Active), Badge: activeBadge(item.Active), Align: "center"},
		},
	}
}

// DomainItemToFormView renders create and update alike.
//
// generator supports appears here exactly once: text, textarea, number,
// checkbox, date, datetime, money, email and a select fed by a lookup.
//
//nolint:funlen // a declarative list with one literal per module field Every field type the
func DomainItemToFormView(
	session *commondomain.Session,
	item *dto.Item,
	categories []domain.ItemCategory,
	isEdit bool,
	errMsg string) commondto.FormView {
	action := itemBasePath + "/new"
	title := localize(session, "item.actions.new")
	if isEdit && item != nil {
		itemID := strconv.FormatUint(uint64(item.ID), 10)
		action = itemBasePath + "/" + itemID + "/edit"
		title = localize(session, "item.actions.edit")
	}
	return commondto.FormView{
		Title:     title,
		Action:    action,
		Error:     errMsg,
		Language:  session.Language,
		CSRFToken: session.CSRFToken,
		Fields: []commondto.FormField{
			{
				Name:     "name",
				Label:    localize(session, "item.fields.name"),
				Type:     commondto.FieldTypeText,
				Value:    itemField(item, func(i *dto.Item) string { return i.Name }),
				Required: true,
			},
			{
				Name:  "notes",
				Label: localize(session, "item.fields.notes"),
				Type:  commondto.FieldTypeTextarea,
				Value: itemField(item, func(i *dto.Item) string { return i.Notes }),
			},
			{
				Name:  "stock",
				Label: localize(session, "item.fields.stock"),
				Type:  commondto.FieldTypeNumber,
				Value: itemField(item, func(i *dto.Item) string {
					return strconv.FormatUint(uint64(i.Stock), 10)
				}),
			},
			{
				Name:     "price",
				Label:    localize(session, "item.fields.price"),
				Type:     commondto.FieldTypeMoney,
				Value:    itemField(item, func(i *dto.Item) string { return i.Price }),
				Required: true,
			},
			{
				Name:  "contact",
				Label: localize(session, "item.fields.contact"),
				Type:  commondto.FieldTypeEmail,
				Value: itemField(item, func(i *dto.Item) string { return i.Contact }),
			},
			{
				Name:     "released_at",
				Label:    localize(session, "item.fields.released_at"),
				Type:     commondto.FieldTypeDate,
				Value:    itemField(item, func(i *dto.Item) string { return i.ReleasedAt }),
				Required: true,
			},
			{
				Name:     "starts_at",
				Label:    localize(session, "item.fields.starts_at"),
				Type:     commondto.FieldTypeDateTime,
				Value:    itemField(item, func(i *dto.Item) string { return i.StartsAt }),
				Required: true,
			},
			{
				Name:     "category_id",
				Label:    localize(session, "item.fields.category_id"),
				Type:     commondto.FieldTypeSelect,
				Options:  itemCategoryOptions(categories),
				Selected: itemSelectedCategory(item),
				Required: true,
			},
			{
				Name:    "active",
				Label:   localize(session, "item.fields.active"),
				Type:    commondto.FieldTypeCheckbox,
				Checked: item != nil && item.Active,
			},
		},
	}
}

// DTOItemToDomainItem converts the submitted form. Money arrives as a decimal
// string and is stored in cents; the dates arrive in the <input> wire formats.
// A malformed value is an error, never a silently zeroed field.
func DTOItemToDomainItem(item *dto.Item, itemID uint) (*domain.Item, error) {
	price, err := commonmapper.ParseDecimalToUint(item.Price)
	if err != nil {
		return nil, fmt.Errorf("parse item price failed: %w", err)
	}
	releasedAt, err := time.Parse(dateLayout, item.ReleasedAt)
	if err != nil {
		return nil, fmt.Errorf("parse item released_at failed: %w", err)
	}
	startsAt, err := time.Parse(dateTimeLayout, item.StartsAt)
	if err != nil {
		return nil, fmt.Errorf("parse item starts_at failed: %w", err)
	}
	return &domain.Item{
		ID:         itemID,
		Name:       item.Name,
		Notes:      item.Notes,
		Stock:      item.Stock,
		Active:     item.Active,
		ReleasedAt: releasedAt,
		StartsAt:   startsAt,
		Price:      *price,
		Contact:    item.Contact,
		CategoryID: item.CategoryID,
	}, nil
}

func DomainItemToDTOItem(item *domain.Item) *dto.Item {
	return &dto.Item{
		ID:         item.ID,
		Name:       item.Name,
		Notes:      item.Notes,
		Stock:      item.Stock,
		Active:     item.Active,
		ReleasedAt: formatDate(item.ReleasedAt),
		StartsAt:   formatDateTime(item.StartsAt),
		Price:      *commonmapper.UintToDecimalString(item.Price),
		Contact:    item.Contact,
		CategoryID: item.CategoryID,
	}
}

func DomainItemsToDTOItems(items []domain.Item) []dto.Item {
	result := make([]dto.Item, 0, len(items))
	for i := range items {
		result = append(result, *DomainItemToDTOItem(&items[i]))
	}
	return result
}

func itemCategoryOptions(categories []domain.ItemCategory) []commondto.SelectOption {
	options := make([]commondto.SelectOption, 0, len(categories))
	for i := range categories {
		options = append(options, commondto.SelectOption{
			Value: strconv.FormatUint(uint64(categories[i].ID), 10),
			Label: categories[i].Name,
		})
	}
	return options
}

func itemSelectedCategory(item *dto.Item) []string {
	if item == nil || item.CategoryID == 0 {
		return nil
	}
	return []string{strconv.FormatUint(uint64(item.CategoryID), 10)}
}

func itemCategoryName(item *domain.Item) string {
	if item.Category == nil {
		return ""
	}
	return item.Category.Name
}

func itemField(item *dto.Item, get func(*dto.Item) string) string {
	if item == nil {
		return ""
	}
	return get(item)
}

// formatDate and formatDateTime keep a zero time rendering as an empty control
// rather than as year 1, which is what a naive Format would put on screen.
func formatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(dateLayout)
}

func formatDateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(dateTimeLayout)
}

func activeLabel(session *commondomain.Session, active bool) string {
	if active {
		return localize(session, "item.active.yes")
	}
	return localize(session, "item.active.no")
}

func activeBadge(active bool) string {
	if active {
		return "success"
	}
	return "neutral"
}

func getItemBreadcrumb(session *commondomain.Session) []string {
	return []string{localize(session, "example.title"), localize(session, "item.title")}
}

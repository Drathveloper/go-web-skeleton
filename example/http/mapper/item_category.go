package mapper

import (
	"strconv"

	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	commondto "github.com/Drathveloper/go-web-skeleton/common/http/dto"
	commonmapper "github.com/Drathveloper/go-web-skeleton/common/http/mapper"
	"github.com/Drathveloper/go-web-skeleton/common/i18n"
	"github.com/Drathveloper/go-web-skeleton/example/domain"
	"github.com/Drathveloper/go-web-skeleton/example/http/dto"
)

const (
	itemCategoryBasePath  = "/item-category"
	itemCategoryTableID   = "item-categories"
	itemCategoryRowPrefix = "item-category-row-"
)

func DomainItemCategoriesToItemCategoriesViewResponse(
	session *commondomain.Session,
	itemCategories []domain.ItemCategory) *commondto.ViewResponse[commondto.TableView] {
	data := DomainItemCategoriesToTableView(session, itemCategories)
	return commonmapper.MapDataToViewResponse(&data, getItemCategoryBreadcrumb(session), session)
}

// DomainItemCategoriesToTableView builds the whole listing. The page template
// is a single call into the shared table component, so a change to the table
// design is one edit here and not one per generated module.
func DomainItemCategoriesToTableView(
	session *commondomain.Session, itemCategories []domain.ItemCategory) commondto.TableView {
	rows := make([]commondto.TableRow, 0, len(itemCategories))
	for i := range itemCategories {
		rows = append(rows, DomainItemCategoryToTableRow(session, &itemCategories[i], ""))
	}
	return commondto.TableView{
		ID:       itemCategoryTableID,
		Title:    localize(session, "item_category.title"),
		Subtitle: localize(session, "item_category.subtitle"),
		NewURL:   itemCategoryBasePath + "/new",
		NewLabel: localize(session, "item_category.actions.new"),
		Language: session.Language,
		Columns: []commondto.TableColumn{
			{Label: localize(session, "item_category.fields.name"), SortKey: "name"},
			{Label: localize(session, "actions.title"), Align: "right"},
		},
		Rows: rows,
	}
}

func DomainItemCategoryToTableRow(
	session *commondomain.Session, itemCategory *domain.ItemCategory, oob string) commondto.TableRow {
	itemCategoryID := strconv.FormatUint(uint64(itemCategory.ID), 10)
	return commondto.TableRow{
		ID:        itemCategoryRowPrefix + itemCategoryID,
		EditURL:   itemCategoryBasePath + "/" + itemCategoryID + "/edit",
		DeleteURL: itemCategoryBasePath + "/" + itemCategoryID + "/delete",
		CSRFToken: session.CSRFToken,
		Language:  session.Language,
		OOB:       oob,
		Cells: []commondto.TableCell{
			{Text: itemCategory.Name, Secondary: "#" + itemCategoryID, SecondaryMono: true, Strong: true},
		},
	}
}

// DomainItemCategoryToFormView renders both create and update: the only
// difference is the action URL and the title, so there is one form, not two.
func DomainItemCategoryToFormView(
	session *commondomain.Session,
	itemCategory *dto.ItemCategory,
	isEdit bool,
	errMsg string) commondto.FormView {
	action := itemCategoryBasePath + "/new"
	title := localize(session, "item_category.actions.new")
	if isEdit && itemCategory != nil {
		itemCategoryID := strconv.FormatUint(uint64(itemCategory.ID), 10)
		action = itemCategoryBasePath + "/" + itemCategoryID + "/edit"
		title = localize(session, "item_category.actions.edit")
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
				Label:    localize(session, "item_category.fields.name"),
				Type:     commondto.FieldTypeText,
				Value:    itemCategoryFieldName(itemCategory),
				Required: true,
			},
		},
	}
}

func itemCategoryFieldName(itemCategory *dto.ItemCategory) string {
	if itemCategory == nil {
		return ""
	}
	return itemCategory.Name
}

func DomainItemCategoriesToDTOItemCategories(itemCategories []domain.ItemCategory) []dto.ItemCategory {
	result := make([]dto.ItemCategory, 0, len(itemCategories))
	for i := range itemCategories {
		result = append(result, *DomainItemCategoryToDTOItemCategory(&itemCategories[i]))
	}
	return result
}

func DTOItemCategoryToDomainItemCategory(itemCategory *dto.ItemCategory, id uint) *domain.ItemCategory {
	return &domain.ItemCategory{
		ID:   id,
		Name: itemCategory.Name,
	}
}

func DomainItemCategoryToDTOItemCategory(itemCategory *domain.ItemCategory) *dto.ItemCategory {
	return &dto.ItemCategory{
		ID:   itemCategory.ID,
		Name: itemCategory.Name,
	}
}

func getItemCategoryBreadcrumb(session *commondomain.Session) []string {
	return []string{localize(session, "example.title"), localize(session, "item_category.title")}
}

// localize resolves a catalog key in the language of the current session. Every
// string a mapper puts into a view has already been through here: the shared
// components receive text, never keys.
func localize(session *commondomain.Session, messageID string) string {
	return i18n.LocalizeMessage(session.Language, messageID)
}

package helper

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	commonmapper "github.com/Drathveloper/go-web-skeleton/common/http/mapper"
	"github.com/Drathveloper/go-web-skeleton/common/i18n"
	"github.com/Drathveloper/go-web-skeleton/common/log"
)

// Identifiers of the only text a failing request is allowed to put on screen.
// The catalog lives in common/i18n/files/core.<lang>.json. Nothing derived from
// an error value is ever rendered: a raw err.Error() leaks driver messages, SQL
// and infrastructure addresses to whoever submitted the form.
const (
	ErrorTitleMessageID      = "errors.title"
	UnexpectedErrorMessageID = "errors.unexpected"
	InvalidFormDataMessageID = "errors.invalid_form_data"
	NotFoundErrorMessageID   = "errors.not_found"
)

// These live here rather than next to the handlers that use them because every
// generated module needs exactly the same three calls. Keeping them shared is
// what stops `scaffold module` from emitting a copy of this file per module,
// and what makes a change to the error contract a single edit.

// LogError sends the cause of a failure to the request logger, which is where
// the detail belongs and the only place it exists.
func LogError(c *gin.Context, logMsg string, err error) {
	log.ContextLogger(c.Request.Context()).Error(logMsg, slog.String("error", err.Error()))
}

// FlashError logs the cause and leaves a generic, localized alert in the
// session, so the page the caller is about to be redirected to can show it.
func FlashError(
	c *gin.Context, session *commondomain.Session, status int, logMsg, messageID string, err error) {
	LogError(c, logMsg, err)
	session.AddAlertMessages(commondomain.NewErrorAlertMessage(
		status,
		i18n.LocalizeMessage(session.Language, ErrorTitleMessageID),
		i18n.LocalizeMessage(session.Language, messageID)))
}

// RenderErrorPage answers with the error page and the real HTTP status.
// Rendering a failure as 200, as the original handlers did, makes browsers
// cache it, monitoring count it as a success and HTMX treat it as a good swap.
func RenderErrorPage(
	c *gin.Context, session *commondomain.Session, status int, logMsg, messageID string, err error) {
	FlashError(c, session, status, logMsg, messageID, err)
	c.HTML(status, "error", commonmapper.MapDataToViewResponse[any](nil, nil, session))
}

// LocalizedMessage resolves a message identifier in the language of the current
// session, for error text rendered inline inside a form fragment.
func LocalizedMessage(session *commondomain.Session, messageID string) string {
	return i18n.LocalizeMessage(session.Language, messageID)
}

package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	commonmapper "github.com/Drathveloper/go-web-skeleton/common/http/mapper"
	"github.com/Drathveloper/go-web-skeleton/common/i18n"
	"github.com/Drathveloper/go-web-skeleton/common/log"
)

// Identifiers of the only text a failing request is allowed to put on screen. The
// catalog lives in common/i18n/files/core.<lang>.json. Nothing derived from an error
// value is ever rendered: a raw err.Error() leaks driver messages, SQL and
// infrastructure detail to whoever submitted the form.
const (
	errorTitleMessageID         = "errors.title"
	unexpectedErrorMessageID    = "errors.unexpected"
	invalidFormDataMessageID    = "errors.invalid_form_data"
	invalidCredentialsMessageID = "errors.invalid_credentials"
	notFoundErrorMessageID      = "errors.not_found"
)

const invalidFormDataErrMsg = "invalid form data"

// logError sends the cause of a failure to the request logger, which is where the
// detail belongs and the only place it exists.
func logError(c *gin.Context, logMsg string, err error) {
	log.ContextLogger(c.Request.Context()).Error(logMsg, slog.String("error", err.Error()))
}

// localizedMessage resolves a message identifier in the language of the current
// session, falling back to the default locale when the request never went through
// the language middleware.
func localizedMessage(session *commondomain.Session, messageID string) string {
	lang := session.Language
	if lang == "" {
		lang = i18n.DefaultLanguage
	}
	return i18n.LocalizeMessage(lang, messageID)
}

// flashError logs the cause and leaves a generic, localized alert in the session so
// the page the caller is about to redirect to can show it.
func flashError(c *gin.Context, session *commondomain.Session, status int, logMsg, messageID string, err error) {
	logError(c, logMsg, err)
	session.AddAlertMessages(commondomain.NewErrorAlertMessage(
		status,
		localizedMessage(session, errorTitleMessageID),
		localizedMessage(session, messageID)))
}

// renderErrorPage answers with the error page and the real HTTP status. Rendering a
// failure as 200, as the original handlers did, makes browsers cache it, monitoring
// count it as a success and HTMX treat it as a good swap.
func renderErrorPage(
	c *gin.Context, session *commondomain.Session, status int, logMsg, messageID string, err error) {
	logError(c, logMsg, err)
	response := commonmapper.MapDomainErrorToViewResponse(session.Language, status, errorTitleMessageID, messageID)
	c.HTML(status, "error", response)
}

package dto

const (
	alertMessageErrorType   = "error"
	alertMessageWarningType = "warning"
	alertMessageInfoType    = "info"
	alertMessageSuccessType = "success"
)

type AlertMessage struct {
	Title   string
	Message string
	Type    string
	Code    int
}

// NewErrorMsg builds a user-facing error alert.
//
// message must be safe to display: a generic, already-localized explanation.
// It deliberately does not accept an error. Passing err.Error() straight to
// the screen leaks driver text, SQL fragments, hostnames and internal paths
// to whoever triggered the failure; the error belongs in the log, this string
// belongs to the user.
func NewErrorMsg(code int, title, message string) AlertMessage {
	return AlertMessage{
		Code:    code,
		Title:   title,
		Message: message,
		Type:    alertMessageErrorType,
	}
}

// NewWarningMsg builds a user-facing warning alert. As with NewErrorMsg,
// message must be safe to display and is not derived from an error.
func NewWarningMsg(title, message string) AlertMessage {
	return AlertMessage{
		Title:   title,
		Message: message,
		Type:    alertMessageWarningType,
	}
}

func NewInfoMsg(title string, message string) AlertMessage {
	return AlertMessage{
		Title:   title,
		Message: message,
		Type:    alertMessageInfoType,
	}
}

func NewSuccessMsg(title string, message string) AlertMessage {
	return AlertMessage{
		Title:   title,
		Message: message,
		Type:    alertMessageSuccessType,
	}
}

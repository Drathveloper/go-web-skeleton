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

func NewErrorMsg(code int, title string, err error) AlertMessage {
	message := ""
	if err != nil {
		message = err.Error()
	}
	return AlertMessage{
		Code:    code,
		Title:   title,
		Message: message,
		Type:    alertMessageErrorType,
	}
}

func NewWarningMsg(title string, err error) AlertMessage {
	message := ""
	if err != nil {
		message = err.Error()
	}
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

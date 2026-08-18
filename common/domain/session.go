package domain

type Session struct {
	ID                   string
	Username             string
	CSRFToken            string
	Language             string
	Roles                []Role
	AlertMessages        AlertMessages
	UserID               uint
	IsLanguageOverridden bool
}

func (s *Session) AddAlertMessages(alertMessages ...AlertMessage) {
	if s.AlertMessages == nil {
		s.AlertMessages = make(AlertMessages, 0, len(alertMessages))
	}
	s.AlertMessages = append(s.AlertMessages, alertMessages...)
}

func (s *Session) FlushAlertMessages() {
	s.AlertMessages = nil
}

type AlertMessage struct {
	Title   string
	Message string
	Type    string
	Code    int
}

func NewAlertMessage(code int, title, message, messageType string) AlertMessage {
	return AlertMessage{
		Code:    code,
		Title:   title,
		Message: message,
		Type:    messageType,
	}
}

func NewSuccessAlertMessage(title, message string) AlertMessage {
	return NewAlertMessage(0, title, message, "success")
}

func NewErrorAlertMessage(code int, title, message string) AlertMessage {
	return NewAlertMessage(code, title, message, "error")
}

func NewWarningAlertMessage(title, message string) AlertMessage {
	return NewAlertMessage(0, title, message, "warning")
}

func NewInfoAlertMessage(title, message string) AlertMessage {
	return NewAlertMessage(0, title, message, "info")
}

type AlertMessages []AlertMessage

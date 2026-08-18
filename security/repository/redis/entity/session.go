package entity

type Session struct {
	Username             string          `json:"username,omitempty"`
	CSRFToken            string          `json:"csrf_token,omitempty"`
	Language             string          `json:"language,omitempty"`
	Roles                []string        `json:"roles,omitempty"`
	AlertMessages        []AlertMessages `json:"alert_messages,omitempty"`
	UserID               uint            `json:"user_id,omitempty"`
	IsLanguageOverridden bool            `json:"is_language_overridden,omitempty"`
}

type AlertMessages struct {
	Title   string `json:"title,omitempty"`
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
	Code    int    `json:"code,omitempty"`
}

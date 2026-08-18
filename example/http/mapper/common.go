package mapper

import (
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/i18n"
)

// This file is created once per bounded context, not once per module: every
// generated mapper in the context shares it. Putting localize in a module file
// would make the second generated module redeclare it.

// localize resolves a catalog key in the language of the current session. Every
// string a mapper puts into a view has already been through here: the shared
// components receive text, never keys.
func localize(session *commondomain.Session, messageID string) string {
	return i18n.LocalizeMessage(session.Language, messageID)
}

package dto

// AllEventNames lists every security event topic. The audit logger subscribes
// to exactly this list, so every new event declared in this package must be
// added here or it will never be audited.
func AllEventNames() []string {
	return []string{
		LoginEventName,
		LogoutEventName,
		UserCreatedEventName,
		UserUpdatedEventName,
	}
}

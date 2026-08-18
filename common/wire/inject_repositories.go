package wire

// Markers below are insertion points for `scaffold module`. The generator
// appends above each marker, so the surrounding code must stay valid with
// zero entries: that is why `container` is named rather than blanked out,
// even while nothing reads it yet.

type RequiredRepositories struct {
	// scaffold:repositories:fields
}

func injectRDBMSRepositories(container *Container) error {
	// scaffold:repositories:init
	return nil
}

func injectRedisRepositories(container *Container) error {
	// scaffold:repositories:redis:init
	return nil
}

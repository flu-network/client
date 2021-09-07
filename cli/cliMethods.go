package cli

import "github.com/flu-network/client/catalogue"

// Methods is the 'public' (CLI-facing) interface of the client daemon process. These methods are
// imported and used directly by github.com/flu-network/cli over unix domain sockets, exposed
// by main.go
type Methods struct {
	// Used to expose catalogue information (e.g., download stats) to the CLI
	cat *catalogue.Cat
}

// NewMethods returns a NewMethods instance. cat is expected to be initialized by the caller.
func NewMethods(cat *catalogue.Cat) *Methods {
	return &Methods{cat: cat}
}

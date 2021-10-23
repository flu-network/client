package cli

// Printable is anything that returns a pretty-printed string to show an end user
type Printable interface {
	Sprintf() string
}

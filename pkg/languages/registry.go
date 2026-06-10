package languages

// Registry defines the language identifiers and runtime metadata used by goboxd.
type Registry struct {
	ID string
}

// NewRegistry returns an empty registry placeholder for future extension.
func NewRegistry() *Registry {
	return &Registry{}
}

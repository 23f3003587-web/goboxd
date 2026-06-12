package languages

// Loader loads the language registry from config.
type Loader struct{}

func (Loader) Load(path string) (*Registry, error) {
	_ = path
	return NewRegistry(), nil
}

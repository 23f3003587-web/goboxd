package runner

import (
	"sort"

	"github.com/thesouldev/goboxd/internal/config"
)

// Runner describes a language execution backend.
type Runner interface {
	Name() string
	Supports(language string) bool
}

// Registry provides the collection of language runners used by the service.
type Registry struct {
	runners map[string]Runner
}

func NewRegistry() *Registry {
	reg := &Registry{runners: make(map[string]Runner)}
	for id := range config.Languages {
		reg.runners[id] = stubRunner{name: id}
	}
	return reg
}

func (r *Registry) List() []string {
	ids := make([]string, 0, len(r.runners))
	for id := range r.runners {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (r *Registry) Has(language string) bool {
	_, ok := r.runners[language]
	return ok
}

func (r *Registry) Runner(language string) Runner {
	return r.runners[language]
}

type stubRunner struct {
	name string
}

func (s stubRunner) Name() string {
	return s.name
}

func (s stubRunner) Supports(language string) bool {
	return s.name == language
}

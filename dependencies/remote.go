package dependencies

type Remote struct {
	Url          string
	Dependencies map[string]Dependency
}

func (r *Remote) HasDependency(name string) bool {
	_, ok := r.Dependencies[name]
	return ok
}

type Dependency struct {
	Name string `json:"name"`
}

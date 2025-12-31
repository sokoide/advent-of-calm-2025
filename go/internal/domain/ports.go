package domain

// Renderer outputs a serialized representation of an architecture.
type Renderer interface {
	Render(*Architecture) (string, error)
}

// Parser builds an architecture from a serialized representation.
type Parser interface {
	Parse(string) (*Architecture, error)
}

// LayoutRepository manages the persistence of layout metadata.
type LayoutRepository interface {
	Load(id string) (*ArchitectureLayout, error)
	Save(id string, layout *ArchitectureLayout) error
}

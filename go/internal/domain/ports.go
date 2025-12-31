package domain

// Renderer outputs a serialized representation of an architecture.
type Renderer interface {
	Render(*Architecture) (string, error)
}

// Parser builds an architecture from a serialized representation.
type Parser interface {
	Parse(string) (*Architecture, error)
}

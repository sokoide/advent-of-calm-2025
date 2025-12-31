package domain

// Renderer outputs a serialized representation of an architecture.
type Renderer interface {
	Render(*Architecture) (string, error)
}

// Parser builds an architecture from a serialized representation.
type Parser interface {
	Parse(string) (*Architecture, error)
}

// ASTSyncer updates Go DSL sources based on model changes.
type ASTSyncer interface {
	SyncFromJSON(src, jsonStr string) (string, error)
	AddNode(src, nodeID, nodeType, name, desc string) (string, error)
	UpdateNodeProperty(src, nodeID, property, value string) (string, error)
	DeleteNode(src, nodeID string) (string, error)
}

// LayoutRepository manages the persistence of layout metadata.
type LayoutRepository interface {
	Load(id string) (*ArchitectureLayout, error)
	Save(id string, layout *ArchitectureLayout) error
}

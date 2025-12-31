package domain

// NodeLayout represents the visual position of a node in the diagram.
type NodeLayout struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ArchitectureLayout stores layout information for all nodes in an architecture.
// This is stored as a sidecar JSON file.
type ArchitectureLayout struct {
	Nodes map[string]NodeLayout `json:"nodes"`
}

// NewArchitectureLayout creates an initialized layout.
func NewArchitectureLayout() *ArchitectureLayout {
	return &ArchitectureLayout{
		Nodes: make(map[string]NodeLayout),
	}
}

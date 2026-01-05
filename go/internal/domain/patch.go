package domain

// PatchType indicates the type of modification.
type PatchType string

const (
	PatchUpdateNode  PatchType = "update-node"
	PatchDeleteNode  PatchType = "delete-node"
	PatchUpdateCount PatchType = "update-count" // For loop variables
)

// PatchOperation represents a single change to the source code.
type PatchOperation struct {
	Type     PatchType    `json:"type"`
	NodeID   string       `json:"nodeId,omitempty"`
	Origin   *PatchOrigin `json:"origin,omitempty"`
	Property string       `json:"property,omitempty"` // For update-node
	Value    interface{}  `json:"value,omitempty"`    // For update-node or update-count
}

// PatchOrigin carries the necessary info to locate the code.
type PatchOrigin struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	LoopVar string `json:"loopVar,omitempty"`
	LoopMax int    `json:"loopMax,omitempty"`
}

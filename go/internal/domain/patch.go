package domain

// PatchType indicates the type of modification.
type PatchType string

const (
	PatchUpdateNode         PatchType = "update-node"
	PatchDeleteNode         PatchType = "delete-node"
	PatchUpdateCount        PatchType = "update-count" // For loop variables
	PatchDeleteRelationship PatchType = "delete-relationship"
	PatchAddRelationship    PatchType = "add-relationship"
	PatchAddInterface       PatchType = "add-interface"
	PatchDeleteInterface    PatchType = "delete-interface"
	PatchAddFlow            PatchType = "add-flow"
	PatchUpdateFlow         PatchType = "update-flow"
	PatchDeleteFlow         PatchType = "delete-flow"
	PatchAddComposedOf      PatchType = "add-composed-of"
	PatchUpdateComposedOf   PatchType = "update-composed-of"
	PatchDeleteComposedOf   PatchType = "delete-composed-of"
	PatchAddControl         PatchType = "add-control"
	PatchUpdateControl      PatchType = "update-control"
	PatchDeleteControl      PatchType = "delete-control"
)

// PatchOperation represents a single change to the source code.
type PatchOperation struct {
	Type           PatchType    `json:"type"`
	NodeID         string       `json:"nodeId,omitempty"`
	Origin         *PatchOrigin `json:"origin,omitempty"`
	Property       string       `json:"property,omitempty"`       // For update-node
	Value          interface{}  `json:"value,omitempty"`          // For update-node or update-count
	SourceNode     string       `json:"sourceNode,omitempty"`     // For add-relationship
	TargetNode     string       `json:"targetNode,omitempty"`     // For add-relationship
	IsInteracts    bool         `json:"isInteracts,omitempty"`    // For add-relationship: use Interacts() instead of Connect()
	InterfaceID    string       `json:"interfaceId,omitempty"`    // For add/delete-interface
	Protocol       string       `json:"protocol,omitempty"`       // For add-interface
	FlowID         string       `json:"flowId,omitempty"`         // For delete-flow, update-flow
	FlowName       string       `json:"flowName,omitempty"`       // For add-flow, update-flow
	FlowDesc       string       `json:"flowDesc,omitempty"`       // For add-flow, update-flow
	FlowSteps      []string     `json:"flowSteps,omitempty"`      // For add-flow, update-flow (list of relationship IDs)
	ContainerID    string       `json:"containerId,omitempty"`    // For add-composed-of
	ChildNodeIDs   []string     `json:"childNodeIds,omitempty"`   // For add-composed-of
	ComposedOfID   string       `json:"composedOfId,omitempty"`   // For update/delete-composed-of
	ComposedOfDesc string       `json:"composedOfDesc,omitempty"` // For update-composed-of
	ControlID      string       `json:"controlId,omitempty"`      // For add/update/delete-control
	ControlDesc    string       `json:"controlDesc,omitempty"`    // For add/update-control
}

// PatchOrigin carries the necessary info to locate the code.
type PatchOrigin struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	LoopVar string `json:"loopVar,omitempty"`
	LoopMax int    `json:"loopMax,omitempty"`
}

package main

import "fmt"

// --- Architecture ---
func NewArchitecture(id, name, desc string) *Architecture {
	return &Architecture{
		Schema:      "https://calm.finos.org/release/1.1/meta/calm.json",
		UniqueID:    id,
		Name:        name,
		Description: desc,
		Controls:    make(map[string]*Control),
		Metadata:    make(map[string]any),
	}
}

func (a *Architecture) AddMeta(k string, v any) *Architecture {
	a.Metadata[k] = v
	return a
}

func (a *Architecture) AddControl(id string, desc string, reqs ...Requirement) *Architecture {
	a.Controls[id] = &Control{Description: desc, Requirements: reqs}
	return a
}

func (a *Architecture) Node(id string, ntype NodeType, name, desc string) *Node {
	n := &Node{
		UniqueID:    id,
		NodeType:    ntype,
		Name:        name,
		Description: desc,
		Metadata:    make(map[string]any),
		Controls:    make(map[string]*Control),
	}
	a.Nodes = append(a.Nodes, n)
	return n
}

// --- Node Methods ---
func (n *Node) Standard(cc, owner string) *Node {
	n.CostCenter = cc
	n.Owner = owner
	return n
}

func (n *Node) Interface(id, protocol string) *Interface {
	i := Interface{UniqueID: id, Protocol: protocol}
	n.Interfaces = append(n.Interfaces, i)
	return &n.Interfaces[len(n.Interfaces)-1]
}

func (n *Node) AddMeta(k string, v any) *Node {
	n.Metadata[k] = v
	return n
}

func (n *Node) AddControl(id string, desc string, reqs ...Requirement) *Node {
	n.Controls[id] = &Control{Description: desc, Requirements: reqs}
	return n
}

// --- Interface Methods ---
func (i *Interface) SetPort(p int) *Interface    { i.Port = p; return i }
func (i *Interface) SetHost(h string) *Interface { i.Host = h; return i }
func (i *Interface) SetPath(p string) *Interface { i.Path = p; return i }
func (i *Interface) SetName(n string) *Interface { i.Name = n; return i }
func (i *Interface) SetDesc(d string) *Interface { i.Description = d; return i }
func (i *Interface) SetDB(d string) *Interface   { i.Database = d; return i }

// --- Metadata ---
func NewMetadata() Metadata {
	return make(Metadata)
}

// --- Relationship ---
func (a *Architecture) Interacts(id, desc, actor, node string) *Relationship {
	r := &Relationship{
		UniqueID:    id,
		Description: desc,
		Metadata:    make(map[string]any),
		RelationshipType: RelationshipType{
			Interacts: map[string]any{"actor": actor, "nodes": []string{node}},
		},
	}
	a.Relationships = append(a.Relationships, r)
	return r
}

func (a *Architecture) Connect(id, desc, src, dst string) *Relationship {
	r := &Relationship{
		UniqueID:    id,
		Description: desc,
		Metadata:    make(map[string]any),
		RelationshipType: RelationshipType{
			Connects: &Connects{
				Source:      NodeInterface{Node: src},
				Destination: NodeInterface{Node: dst},
			},
		},
	}
	a.Relationships = append(a.Relationships, r)
	return r
}

func (r *Relationship) SrcIntf(intfs ...string) *Relationship {
	r.RelationshipType.Connects.Source.Interfaces = intfs
	return r
}

func (r *Relationship) DstIntf(intfs ...string) *Relationship {
	r.RelationshipType.Connects.Destination.Interfaces = intfs
	return r
}

func (r *Relationship) Data(class string, enc bool) *Relationship {
	r.DataClassification = class
	r.Encrypted = BoolPtr(enc)
	return r
}

func (r *Relationship) WithProtocol(p string) *Relationship { r.Protocol = p; return r }
func (r *Relationship) AddMeta(k string, v any) *Relationship {
	r.Metadata[k] = v
	return r
}

// --- ComposedOf ---
func (a *Architecture) ComposedOf(id, desc, container string, nodes []string) *Relationship {
	r := &Relationship{
		UniqueID:    id,
		Description: desc,
		RelationshipType: RelationshipType{
			ComposedOf: map[string]any{"container": container, "nodes": nodes},
		},
	}
	a.Relationships = append(a.Relationships, r)
	return r
}

// --- Flows ---
func (a *Architecture) Flow(id, name, desc string) *Flow {
	f := &Flow{UniqueID: id, Name: name, Description: desc, Metadata: make(map[string]any)}
	a.Flows = append(a.Flows, f)
	return f
}

func (f *Flow) AddMeta(k string, v any) *Flow {
	f.Metadata[k] = v
	return f
}

func (f *Flow) Step(relID string, seq int, desc string, dir string) *Flow {
	f.Transitions = append(f.Transitions, Transition{
		RelationshipID: relID, SequenceNumber: seq, Description: desc, Direction: dir,
	})
	return f
}

// --- Global Helpers ---
func BoolPtr(b bool) *bool { return &b }

func NewRequirement(url string, config any) Requirement {
	return Requirement{RequirementURL: url, Config: config}
}

func NewRequirementURL(url, configURL string) Requirement {
	return Requirement{RequirementURL: url, ConfigURL: configURL}
}

func NewSecurityConfig(algo, scope string) map[string]any {
	return map[string]any{"algorithm": algo, "scope": scope}
}

func NewPerformanceConfig(p99, p95 int) map[string]any {
	return map[string]any{"p99-latency-ms": p99, "p95-latency-ms": p95}
}

func NewAvailabilityConfig(uptime float64, interval int) map[string]any {
	return map[string]any{"uptime-percentage": uptime, "monitoring-interval-seconds": interval}
}

func NewFailoverConfig(rto, rpo int, auto bool) map[string]any {
	return map[string]any{"rto-minutes": rto, "rpo-minutes": rpo, "automatic-failover": auto}
}

func NewCircuitBreakerConfig(threshold, wait, minCalls int) map[string]any {
	return map[string]any{
		"failure-threshold-percentage": threshold,
		"wait-duration-seconds":        wait,
		"minimum-calls-before-opening": minCalls,
	}
}

// --- NEW EXPERIMENTAL DSL (Fun Coding Experience) ---

// NodeOption defines a functional option for configuring a Node.
type NodeOption func(*Node)

// WithMeta merges the provided map into the Node's metadata.
func WithMeta(meta map[string]any) NodeOption {
	return func(n *Node) {
		for k, v := range meta {
			n.Metadata[k] = v
		}
	}
}

// WithOwner sets the owner and cost center.
func WithOwner(owner, costCenter string) NodeOption {
	return func(n *Node) {
		n.Owner = owner
		n.CostCenter = costCenter
	}
}

// WithTags adds tags to the metadata.
func WithTags(tags ...string) NodeOption {
	return func(n *Node) {
		n.Metadata["tags"] = tags
	}
}

// WithControl adds a control requirement.
func WithControl(id, desc string, reqs ...Requirement) NodeOption {
	return func(n *Node) {
		n.Controls[id] = &Control{Description: desc, Requirements: reqs}
	}
}

// DefineNode creates a new node with functional options.
// It replaces the verbose Node() method calls with a cleaner config style.
func (a *Architecture) DefineNode(id string, ntype NodeType, name, desc string, opts ...NodeOption) *Node {
	n := &Node{
		arch:        a,
		UniqueID:    id,
		NodeType:    ntype,
		Name:        name,
		Description: desc,
		Metadata:    make(map[string]any),
		Controls:    make(map[string]*Control),
	}
	for _, opt := range opts {
		opt(n)
	}
	a.Nodes = append(a.Nodes, n)
	return n
}

// ConnectTo establishes a relationship from this node to another.
func (n *Node) ConnectTo(dest *Node, desc string) *ConnectionBuilder {
	relID := fmt.Sprintf("%s-connects-%s", n.UniqueID, dest.UniqueID)
	rel := &Relationship{
		UniqueID:    relID,
		Description: desc,
		Metadata:    make(map[string]any),
		RelationshipType: RelationshipType{
			Connects: &Connects{
				Source:      NodeInterface{Node: n.UniqueID},
				Destination: NodeInterface{Node: dest.UniqueID},
			},
		},
	}
	// Auto-register with the architecture
	if n.arch != nil {
		n.arch.Relationships = append(n.arch.Relationships, rel)
	}
	return &ConnectionBuilder{rel: rel}
}

// WithID overrides the auto-generated relationship ID.
func (cb *ConnectionBuilder) WithID(id string) *ConnectionBuilder {
	cb.rel.UniqueID = id
	return cb
}

// Encrypted sets the encrypted flag.
func (cb *ConnectionBuilder) Encrypted(e bool) *ConnectionBuilder {
	cb.rel.Encrypted = BoolPtr(e)
	return cb
}

// Via specifies the interfaces for the connection.
func (cb *ConnectionBuilder) Via(srcIntf, dstIntf string) *ConnectionBuilder {
	if srcIntf != "" {
		cb.rel.RelationshipType.Connects.Source.Interfaces = []string{srcIntf}
	} else {
		cb.rel.RelationshipType.Connects.Source.Interfaces = nil
	}
	if dstIntf != "" {
		cb.rel.RelationshipType.Connects.Destination.Interfaces = []string{dstIntf}
	} else {
		cb.rel.RelationshipType.Connects.Destination.Interfaces = nil
	}
	return cb
}

// Tag adds metadata to the relationship.
func (cb *ConnectionBuilder) Tag(key string, val any) *ConnectionBuilder {
	cb.rel.Metadata[key] = val
	return cb
}

// Protocol sets the relationship protocol.
func (cb *ConnectionBuilder) Protocol(p string) *ConnectionBuilder {
	cb.rel.Protocol = p
	return cb
}

// Is sets the data classification.
func (cb *ConnectionBuilder) Is(classification string) *ConnectionBuilder {
	cb.rel.DataClassification = classification
	return cb
}

// GetID returns the relationship ID, useful for Flow definitions.
func (cb *ConnectionBuilder) GetID() string {
	return cb.rel.UniqueID
}

// FlowFromIds constructs a flow from a list of relationship IDs.
func (a *Architecture) FlowFromIds(id, name, desc string, relIDs ...string) *Flow {
	f := &Flow{UniqueID: id, Name: name, Description: desc, Metadata: make(map[string]any)}
	for i, rid := range relIDs {
		f.Transitions = append(f.Transitions, Transition{
			RelationshipID: rid,
			SequenceNumber: i + 1,
			Description:    "Step " + rid,
			Direction:      "source-to-destination",
		})
	}
	a.Flows = append(a.Flows, f)
	return f
}

// FlowBuilder helps build complex flows with explicit step metadata
type FlowBuilder struct {
	flow *Flow
}

func (a *Architecture) DefineFlow(id, name, desc string) *FlowBuilder {
	f := &Flow{UniqueID: id, Name: name, Description: desc, Metadata: make(map[string]any)}
	a.Flows = append(a.Flows, f)
	return &FlowBuilder{flow: f}
}

func (fb *FlowBuilder) Step(relID string, desc string) *FlowBuilder {
	fb.flow.Transitions = append(fb.flow.Transitions, Transition{
		RelationshipID: relID,
		SequenceNumber: len(fb.flow.Transitions) + 1,
		Description:    desc,
		Direction:      "source-to-destination",
	})
	return fb
}

func (fb *FlowBuilder) StepEx(relID, desc, dir string) *FlowBuilder {
	fb.flow.Transitions = append(fb.flow.Transitions, Transition{
		RelationshipID: relID,
		SequenceNumber: len(fb.flow.Transitions) + 1,
		Description:    desc,
		Direction:      dir,
	})
	return fb
}

func (fb *FlowBuilder) Meta(k string, v any) *FlowBuilder {
	fb.flow.Metadata[k] = v
	return fb
}

func (fb *FlowBuilder) MetaMap(m map[string]any) *FlowBuilder {
	for k, v := range m {
		fb.flow.Metadata[k] = v
	}
	return fb
}

// StepSpec defines a single step in a flow.
type StepSpec struct {
	ID   string
	Desc string
	Dir  string // Optional: defaults to "source-to-destination"
}

func (fb *FlowBuilder) Steps(specs ...StepSpec) *FlowBuilder {
	for _, s := range specs {
		dir := s.Dir
		if dir == "" {
			dir = "source-to-destination"
		}
		fb.StepEx(s.ID, s.Desc, dir)
	}
	return fb
}

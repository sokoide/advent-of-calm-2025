package main

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

func (a *Architecture) Control(id string, desc string, reqs ...Requirement) *Architecture {
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

func (n *Node) Control(id string, desc string, reqs ...Requirement) *Node {
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

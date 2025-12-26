package main

// --- Architecture ---
func NewArchitecture(id, name, desc string) *Architecture {
	return &Architecture{
		Schema:      "https://calm.finos.org/release/1.1/meta/calm.json",
		UniqueID:    id,
		Name:        name,
		Description: desc,
		Controls:    make(map[string]Control),
	}
}

// --- Node ---
func NewNode(id string, ntype NodeType, name, desc string) Node {
	return Node{
		UniqueID:    id,
		NodeType:    ntype,
		Name:        name,
		Description: desc,
		Metadata:    make(Metadata),
		Controls:    make(map[string]Control),
		Interfaces:  make([]Interface, 0),
	}
}

func (n Node) SetStandards(cc, owner string) Node {
	n.CostCenter = cc
	n.Owner = owner
	return n
}

func (n Node) WithMetadata(m Metadata) Node {
	for k, v := range m {
		n.Metadata[k] = v
	}
	return n
}

func (n Node) WithInterfaces(i ...Interface) Node {
	n.Interfaces = append(n.Interfaces, i...)
	return n
}

func (n Node) WithControl(id string, c Control) Node {
	n.Controls[id] = c
	return n
}

// --- Interface ---
func NewInterface(id, protocol string) Interface {
	return Interface{UniqueID: id, Protocol: protocol}
}

func (i Interface) AtPort(p int) Interface { i.Port = p; return i }
func (i Interface) OnHost(h string) Interface { i.Host = h; return i }
func (i Interface) WithPath(p string) Interface { i.Path = p; return i }
func (i Interface) WithName(n string) Interface { i.Name = n; return i }
func (i Interface) WithDesc(d string) Interface { i.Description = d; return i }
func (i Interface) WithDB(d string) Interface { i.Database = d; return i }

// --- Metadata ---
func NewMetadata() Metadata {
	return make(Metadata)
}

func (m Metadata) Add(k string, v any) Metadata {
	m[k] = v
	return m
}

// --- Relationship ---
func NewRel(id, desc string) Relationship {
	return Relationship{UniqueID: id, Description: desc, Metadata: make(Metadata)}
}

func Interacts(id, desc, actor, node string) Relationship {
	r := NewRel(id, desc)
	r.RelationshipType = RelationshipType{Interacts: map[string]any{"actor": actor, "nodes": []string{node}}}
	return r
}

func ConnectIntf(id, desc, src, dst string, srcIntfs, dstIntfs []string) Relationship {
	r := NewRel(id, desc)
	r.RelationshipType = RelationshipType{Connects: &Connects{
		Source:      NodeInterface{Node: src, Interfaces: srcIntfs},
		Destination: NodeInterface{Node: dst, Interfaces: dstIntfs},
	}}
	return r
}

func ComposedOf(id, desc, container string, nodes []string) Relationship {
	r := NewRel(id, desc)
	r.RelationshipType = RelationshipType{ComposedOf: map[string]any{"container": container, "nodes": nodes}}
	return r
}

func (r Relationship) WithData(class string, encrypted bool) Relationship {
	r.DataClassification = class
	r.Encrypted = BoolPtr(encrypted)
	return r
}

func (r Relationship) WithProtocol(p string) Relationship { r.Protocol = p; return r }
func (r Relationship) WithMetadata(m Metadata) Relationship {
	for k, v := range m {
		r.Metadata[k] = v
	}
	return r
}

// --- Control & Requirement ---
func NewControl(desc string, reqs ...Requirement) Control {
	return Control{Description: desc, Requirements: reqs}
}

func NewRequirement(url string, config any) Requirement {
	return Requirement{RequirementURL: url, Config: config}
}

func NewRequirementWithURL(url, configURL string) Requirement {
	return Requirement{RequirementURL: url, ConfigURL: configURL}
}

// --- Flow & Transition ---
func NewFlow(id, name, desc string) Flow {
	return Flow{UniqueID: id, Name: name, Description: desc, Metadata: make(Metadata)}
}

func (f Flow) WithMetadata(m Metadata) Flow {
	for k, v := range m {
		f.Metadata[k] = v
	}
	return f
}

func (f Flow) WithTransitions(t ...Transition) Flow {
	f.Transitions = append(f.Transitions, t...)
	return f
}

func NewTransition(relID string, seq int, desc string, dir string) Transition {
	return Transition{RelationshipID: relID, SequenceNumber: seq, Description: desc, Direction: dir}
}

// --- Configs ---
func NewSecurityConfig(algo, scope string) SecurityConfig {
	return SecurityConfig{Algorithm: algo, Scope: scope}
}

func NewPerformanceConfig(p99, p95 int) PerformanceConfig {
	return PerformanceConfig{P99LatencyMS: p99, P95LatencyMS: p95}
}

func NewAvailabilityConfig(uptime float64, interval int) AvailabilityConfig {
	return AvailabilityConfig{UptimePercentage: uptime, MonitoringIntervalSeconds: interval}
}

func NewFailoverConfig(rto, rpo int, auto bool) FailoverConfig {
	return FailoverConfig{RTOMinutes: rto, RPOMinutes: rpo, AutomaticFailover: auto}
}

func NewCircuitBreakerConfig(threshold, wait, minCalls int) CircuitBreakerConfig {
	return CircuitBreakerConfig{FailureThresholdPercentage: threshold, WaitDurationSeconds: wait, MinimumCallsBeforeOpening: minCalls}
}

func BoolPtr(b bool) *bool { return &b }
package main

import (
	"encoding/json"
)

// --- CALM Core Types ---
type NodeType string

const (
	Actor     NodeType = "actor"
	Service   NodeType = "service"
	Database  NodeType = "database"
	System    NodeType = "system"
	Queue     NodeType = "queue"
	WebClient NodeType = "webclient"
)

type Architecture struct {
	Schema        string             `json:"$schema"`
	ADRs          []string           `json:"adrs,omitempty"`
	UniqueID      string             `json:"unique-id"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	Metadata      Metadata           `json:"metadata,omitempty"`
	Controls      map[string]Control `json:"controls,omitempty"`
	Flows         []Flow             `json:"flows,omitempty"`
	Nodes         []Node             `json:"nodes"`
	Relationships []Relationship     `json:"relationships"`
}

type Metadata map[string]any

type Control struct {
	Description  string        `json:"description"`
	Requirements []Requirement `json:"requirements"`
}

type Requirement struct {
	RequirementURL string `json:"requirement-url"`
	Config         any    `json:"config,omitempty"`
	ConfigURL      string `json:"config-url,omitempty"`
}

// --- Specific Config Types for Controls ---
type SecurityConfig struct {
	Algorithm string `json:"algorithm"`
	Scope     string `json:"scope"`
}

type PerformanceConfig struct {
	P99LatencyMS int `json:"p99-latency-ms,omitempty"`
	P95LatencyMS int `json:"p95-latency-ms,omitempty"`
}

type AvailabilityConfig struct {
	UptimePercentage          float64 `json:"uptime-percentage,omitempty"`
	MonitoringIntervalSeconds int     `json:"monitoring-interval-seconds,omitempty"`
}

type FailoverConfig struct {
	RTOMinutes        int  `json:"rto-minutes,omitempty"`
	RPOMinutes        int  `json:"rpo-minutes,omitempty"`
	AutomaticFailover bool `json:"automatic-failover,omitempty"`
}

type CircuitBreakerConfig struct {
	FailureThresholdPercentage int `json:"failure-threshold-percentage"`
	WaitDurationSeconds        int `json:"wait-duration-seconds"`
	MinimumCallsBeforeOpening  int `json:"minimum-calls-before-opening"`
}

// --- Flows & Nodes ---
type Flow struct {
	UniqueID    string       `json:"unique-id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Metadata    Metadata     `json:"metadata,omitempty"`
	Transitions []Transition `json:"transitions"`
}

type Transition struct {
	RelationshipID string `json:"relationship-unique-id"`
	SequenceNumber int    `json:"sequence-number"`
	Description    string `json:"description"`
	Direction      string `json:"direction"`
}

type Node struct {
	UniqueID    string             `json:"unique-id"`
	NodeType    NodeType           `json:"node-type"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	CostCenter  string             `json:"costCenter,omitempty"`
	Owner       string             `json:"owner,omitempty"`
	Metadata    Metadata           `json:"metadata,omitempty"`
	Controls    map[string]Control `json:"controls,omitempty"`
	Interfaces  []Interface        `json:"interfaces,omitempty"`
}

type Interface struct {
	UniqueID    string `json:"unique-id"`
	Name        string `json:"name,omitempty"`
	Protocol    string `json:"protocol"`
	Port        int    `json:"port,omitempty"`
	Host        string `json:"host,omitempty"`
	Path        string `json:"path,omitempty"`
	Description string `json:"description,omitempty"`
	Database    string `json:"database,omitempty"`
}

type NodeInterface struct {
	Node       string   `json:"node"`
	Interfaces []string `json:"interfaces,omitempty"`
}

type Connects struct {
	Source      NodeInterface `json:"source"`
	Destination NodeInterface `json:"destination"`
}

type RelationshipType struct {
	Connects   *Connects      `json:"connects,omitempty"`
	Interacts  map[string]any `json:"interacts,omitempty"`
	ComposedOf map[string]any `json:"composed-of,omitempty"`
}

type Relationship struct {
	UniqueID           string           `json:"unique-id"`
	Description        string           `json:"description"`
	DataClassification string           `json:"dataClassification,omitempty"`
	Encrypted          *bool            `json:"encrypted,omitempty"`
	Protocol           string           `json:"protocol,omitempty"`
	Metadata           Metadata         `json:"metadata,omitempty"`
	RelationshipType   RelationshipType `json:"relationship-type"`
}

// --- DSL Helpers ---

func (a *Architecture) AddNode(n Node) {
	a.Nodes = append(a.Nodes, n)
}

func (a *Architecture) AddRelationship(r Relationship) {
	a.Relationships = append(a.Relationships, r)
}

func (a *Architecture) AddFlow(f Flow) {
	a.Flows = append(a.Flows, f)
}

func (a *Architecture) ToJSON() string {
	out, _ := json.MarshalIndent(a, "", "    ")
	return string(out)
}

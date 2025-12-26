package main

import (
	"encoding/json"
	"fmt"
)

// CALM types
type NodeType string

const (
	Actor    NodeType = "actor"
	Service  NodeType = "service"
	Database NodeType = "database"
	System   NodeType = "system"
	Queue    NodeType = "queue"
)

type Architecture struct {
	Schema        string         `json:"$schema"`
	UniqueID      string         `json:"unique-id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	Nodes         []*Node        `json:"nodes"`
	Relationships []*Relationship `json:"relationships"`
}

type Node struct {
	UniqueID    string         `json:"unique-id"`
	NodeType    NodeType       `json:"node-type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	CostCenter  string         `json:"costCenter,omitempty"`
	Owner       string         `json:"owner,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Interfaces  []Interface    `json:"interfaces,omitempty"`
}

type Interface struct {
	UniqueID    string `json:"unique-id"`
	Name        string `json:"name,omitempty"`
	Protocol    string `json:"protocol"`
	Port        int    `json:"port,omitempty"`
	Host        string `json:"host,omitempty"`
	Path        string `json:"path,omitempty"`
	Description string `json:"description,omitempty"`
}

type Relationship struct {
	UniqueID         string         `json:"unique-id"`
	Description      string         `json:"description"`
	RelationshipType map[string]any `json:"relationship-type"`
}

// Fluent API Helpers
func NewArchitecture(id, name, desc string) *Architecture {
	return &Architecture{
		Schema:      "https://calm.finos.org/release/1.1/meta/calm.json",
		UniqueID:    id,
		Name:        name,
		Description: desc,
		Metadata:    make(map[string]any),
	}
}

func (a *Architecture) NewNode(id string, ntype NodeType, name, desc string) *Node {
	node := &Node{
		UniqueID:    id,
		NodeType:    ntype,
		Name:        name,
		Description: desc,
		Metadata:    make(map[string]any),
	}
	a.Nodes = append(a.Nodes, node)
	return node
}

func (n *Node) AddInterface(id, proto string, port int) *Node {
	n.Interfaces = append(n.Interfaces, Interface{
		UniqueID: id,
		Protocol: proto,
		Port:     port,
	})
	return n
}

func (a *Architecture) Connect(id, desc, srcNode, dstNode string) {
	rel := &Relationship{
		UniqueID:    id,
		Description: desc,
		RelationshipType: map[string]any{
			"connects": map[string]any{
				"source":      map[string]string{"node": srcNode},
				"destination": map[string]string{"node": dstNode},
			},
		},
	}
	a.Relationships = append(a.Relationships, rel)
}

func main() {
	arch := NewArchitecture("ecommerce-go", "E-Commerce Go-Defined", "Generated via Go DSL")

	// Nodes
	customer := arch.NewNode("customer", Actor, "Customer", "A web user")
	customer.Owner = "marketing"

	gateway := arch.NewNode("api-gateway", Service, "API Gateway", "Entry point")
	gateway.AddInterface("https-in", "HTTPS", 443)

	orderSvc := arch.NewNode("order-service", Service, "Order Service", "Handles orders")
	orderSvc.AddInterface("api-rest", "REST", 8080)

	// Relationships
	arch.Connect("cust-to-gw", "Customer uses gateway", "customer", "api-gateway")
	arch.Connect("gw-to-order", "Gateway routes to order", "api-gateway", "order-service")

	// Output
	out, _ := json.MarshalIndent(arch, "", "  ")
	fmt.Println(string(out))
}

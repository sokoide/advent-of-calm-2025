package parser

import (
	"testing"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

func TestParseRichD2(t *testing.T) {
	t.Run("should parse basic architecture info", func(t *testing.T) {
		d2 := `# @calm:id=ecommerce-platform
# @calm:description=E-commerce Platform Architecture
`
		arch, err := ParseRichD2(d2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if arch.UniqueID != "ecommerce-platform" {
			t.Errorf("expected id ecommerce-platform, got %s", arch.UniqueID)
		}
		if arch.Description != "E-commerce Platform Architecture" {
			t.Errorf("expected description, got %s", arch.Description)
		}
	})

	t.Run("should parse nodes and their annotations", func(t *testing.T) {
		d2 := `
customer: Customer {
  # @calm:type=actor
  # @calm:owner=marketing-team
  # @calm:costCenter=cc-1
  # @calm:description=A customer browsing the store
  # @calm:metadata={"tier": "1"}
  # @calm:interfaces=[{"unique-id": "web", "protocol": "https"}]
  # @calm:controls={"PCI": {"description": "desc"}}
}
`
		arch, err := ParseRichD2(d2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(arch.Nodes) != 1 {
			t.Fatalf("expected 1 node, got %d", len(arch.Nodes))
		}

		node := arch.Nodes[0]
		if node.UniqueID != "customer" {
			t.Errorf("expected id customer, got %s", node.UniqueID)
		}
		if node.NodeType != domain.Actor {
			t.Errorf("expected type actor, got %s", node.NodeType)
		}
		if node.Owner != "marketing-team" {
			t.Errorf("expected owner marketing-team, got %s", node.Owner)
		}

		if node.CostCenter != "cc-1" {
			t.Errorf("expected cost center cc-1, got %s", node.CostCenter)
		}

		if node.Metadata["tier"] != "1" {
			t.Errorf("expected metadata tier 1, got %v", node.Metadata["tier"])
		}

		if len(node.Interfaces) != 1 || node.Interfaces[0].UniqueID != "web" {
			t.Errorf("expected interface web, got %v", node.Interfaces)
		}

		if len(node.Controls) != 1 || node.Controls["PCI"].Description != "desc" {
			t.Errorf("expected control PCI, got %v", node.Controls)
		}
	})

	t.Run("should parse relationships", func(t *testing.T) {
		d2 := `
customer -> api-gateway {
  # @calm:id=customer-to-gateway
  # @calm:type=interacts
  # @calm:actor=customer
  # @calm:description=Customer interacts with API Gateway
  # @calm:encrypted=true
  # @calm:classification=confidential
  # @calm:metadata={"key": "val"}
}
`
		arch, err := ParseRichD2(d2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(arch.Relationships) != 1 {
			t.Fatalf("expected 1 relationship, got %d", len(arch.Relationships))
		}

		rel := arch.Relationships[0]
		if rel.RelationshipType.Interacts == nil {
			t.Errorf("expected interacts relationship type, got nil")
		} else if rel.RelationshipType.Interacts["actor"] != "customer" {
			t.Errorf("expected actor customer, got %v", rel.RelationshipType.Interacts["actor"])
		}

		if rel.Encrypted == nil || !*rel.Encrypted {
			t.Errorf("expected encrypted true")
		}
		if rel.DataClassification != "confidential" {
			t.Errorf("expected classification confidential, got %s", rel.DataClassification)
		}
	})

	t.Run("should parse connections with interfaces", func(t *testing.T) {
		d2 := `
order-service -> inventory-db {
  # @calm:id=order-to-inv
  # @calm:srcInterfaces=["out"]
  # @calm:dstInterfaces=["in"]
}
`
		arch, err := ParseRichD2(d2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		rel := arch.Relationships[0]
		if rel.RelationshipType.Connects.Source.Interfaces[0] != "out" {
			t.Errorf("expected src interface out")
		}
		if rel.RelationshipType.Connects.Destination.Interfaces[0] != "in" {
			t.Errorf("expected dst interface in")
		}
	})

	t.Run("should use RichD2Parser", func(t *testing.T) {
		parser := RichD2Parser{}
		arch, err := parser.Parse("# @calm:id=test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if arch.UniqueID != "test" {
			t.Errorf("expected id test")
		}
	})

	t.Run("should parse flows with metadata", func(t *testing.T) {
		d2 := `
# @calm:flow id=order-flow name=Order Processing Flow
# @calm:flow-metadata={"critical": true}
# @calm:flow-step seq=1 rel=rel1 dir=src-to-dst desc=Step 1
`
		arch, err := ParseRichD2(d2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(arch.Flows) != 1 {
			t.Fatalf("expected 1 flow, got %d", len(arch.Flows))
		}

		flow := arch.Flows[0]
		if flow.Metadata["critical"] != true {
			t.Errorf("expected metadata critical=true, got %v", flow.Metadata["critical"])
		}
	})

	t.Run("should parse composed-of relationships", func(t *testing.T) {
		d2 := `
# @calm:composed-of id=comp1 container=cont1 nodes=["node1", "node2"]
`
		arch, err := ParseRichD2(d2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(arch.Relationships) != 1 {
			t.Fatalf("expected 1 relationship, got %d", len(arch.Relationships))
		}

		rel := arch.Relationships[0]
		comp := rel.RelationshipType.ComposedOf
		if comp == nil {
			t.Fatal("expected composed-of type")
		}
		if comp["container"] != "cont1" {
			t.Errorf("expected container cont1, got %v", comp["container"])
		}
	})

	t.Run("should parse global controls", func(t *testing.T) {
		d2 := `
# @calm:control id=PCI-DSS data={"description": "Secure payments"}
`
		arch, err := ParseRichD2(d2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(arch.Controls) != 1 {
			t.Fatalf("expected 1 control, got %d", len(arch.Controls))
		}

		ctrl := arch.Controls["PCI-DSS"]
		if ctrl.Description != "Secure payments" {
			t.Errorf("expected description, got %s", ctrl.Description)
		}
	})
}

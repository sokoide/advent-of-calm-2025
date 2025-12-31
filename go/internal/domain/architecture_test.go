package domain

import (
	"testing"
)

func TestArchitecture_DefineNode(t *testing.T) {
	arch := NewArchitecture("test-arch", "Test Architecture", "A test architecture")

	t.Run("should define a node with basic properties", func(t *testing.T) {
		nodeID := "test-node"
		nodeName := "Test Node"
		nodeDesc := "A test node description"
		nodeType := Service

		node := arch.DefineNode(nodeID, nodeType, nodeName, nodeDesc)

		if node.UniqueID != nodeID {
			t.Errorf("expected UniqueID %s, got %s", nodeID, node.UniqueID)
		}
		if node.NodeType != nodeType {
			t.Errorf("expected NodeType %s, got %s", nodeType, node.NodeType)
		}
		if node.Name != nodeName {
			t.Errorf("expected Name %s, got %s", nodeName, node.Name)
		}
		if node.Description != nodeDesc {
			t.Errorf("expected Description %s, got %s", nodeDesc, node.Description)
		}
		if node.Arch != arch {
			t.Errorf("node.Arch should point to the parent architecture")
		}
	})

	t.Run("should apply functional options", func(t *testing.T) {
		owner := "team-a"
		cc := "cc-123"
		
		node := arch.DefineNode("opt-node", Service, "Opt Node", "desc", 
			WithOwner(owner, cc),
			WithTags("tag1", "tag2"),
		)

		if node.Owner != owner {
			t.Errorf("expected Owner %s, got %s", owner, node.Owner)
		}
		if node.CostCenter != cc {
			t.Errorf("expected CostCenter %s, got %s", cc, node.CostCenter)
		}
		
		// Check metadata sync
		if node.Metadata["owner"] != owner {
			t.Errorf("expected metadata['owner'] %s, got %s", owner, node.Metadata["owner"])
		}

		tags, ok := node.Metadata["tags"].([]string)
		if !ok || len(tags) != 2 || tags[0] != "tag1" || tags[1] != "tag2" {
			t.Errorf("tags metadata not correctly set: %v", node.Metadata["tags"])
		}
	})
}

func TestNode_ConnectTo(t *testing.T) {
	arch := NewArchitecture("test-arch", "Test", "Desc")
	src := arch.DefineNode("src", Service, "Source", "desc")
	dst := arch.DefineNode("dst", Database, "Dest", "desc")

	t.Run("should create relationship via ConnectTo", func(t *testing.T) {
		relDesc := "Source to Dest"
		cb := src.ConnectTo(dst, relDesc)
		cb.Protocol("https").Encrypted(true).Is("confidential")

		relID := "src-connects-dst"
		if cb.GetID() != relID {
			t.Errorf("expected rel ID %s, got %s", relID, cb.GetID())
		}

		// Verify registration in architecture
		found := false
		for _, r := range arch.Relationships {
			if r.UniqueID == relID {
				found = true
				if r.Description != relDesc {
					t.Errorf("expected description %s, got %s", relDesc, r.Description)
				}
				if r.Protocol != "https" {
					t.Errorf("expected protocol https, got %s", r.Protocol)
				}
				if r.Encrypted == nil || !*r.Encrypted {
					t.Errorf("expected encrypted to be true")
				}
				if r.DataClassification != "confidential" {
					t.Errorf("expected classification confidential, got %s", r.DataClassification)
				}
			}
		}
		if !found {
			t.Errorf("relationship was not registered in architecture")
		}
	})
}

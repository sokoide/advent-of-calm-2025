package ast

import (
	"strings"
	"testing"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

func TestEcommerceFix(t *testing.T) {
	syncer := GoASTSyncer{}

	// Real content of ecommerce_architecture.go (simplified for test)
	src := `package usecase
import "github.com/sokoide/advent-of-calm-2025/internal/domain"

func (EcommerceBuilder) Build() *domain.Architecture {
	arch := domain.NewArchitecture("id", "name", "desc")
	// ...
	links := wireComponents(arch, nodes)
	return arch
}

func wireComponents(a *domain.Architecture, n *nodesContainer) *linksContainer {
	lc := &linksContainer{}
	// EXISTING LINKS
	return lc
}
`

	t.Run("Adding relationship should go to wireComponents and use 'a'", func(t *testing.T) {
		ops := []domain.PatchOperation{
			{
				Type:       domain.PatchAddRelationship,
				NodeID:     "new-rel-123",
				SourceNode: "src",
				TargetNode: "dst",
			},
		}

		got, err := syncer.ApplyPatch(src, ops)
		if err != nil {
			t.Fatalf("ApplyPatch failed: %v", err)
		}

		// Check if it's in wireComponents and uses 'a'
		wireStart := strings.Index(got, "func wireComponents")
		if wireStart == -1 {
			t.Fatal("wireComponents not found in output")
		}

		wireContent := got[wireStart:]
		if !strings.Contains(wireContent, `a.Connect`) || !strings.Contains(wireContent, `"new-rel-123"`) {
			t.Errorf("Expected a.Connect in wireComponents, but it was not found or used wrong receiver name.\nFull output around wireComponents:\n%s", wireContent)
		}

		// Verify it's NOT in Build()
		buildEnd := strings.Index(got, "func wireComponents")
		buildContent := got[:buildEnd]
		if strings.Contains(buildContent, `Connect("new-rel-123"`) {
			t.Errorf("Relationship was incorrectly added to Build() function instead of wireComponents().")
		}
	})

	t.Run("Fallback to Build and find 'arch'", func(t *testing.T) {
		src := `package usecase
func (EcommerceBuilder) Build() *domain.Architecture {
	arch := domain.NewArchitecture("id", "name", "desc")
	return arch
}
`
		ops := []domain.PatchOperation{
			{
				Type:       domain.PatchAddRelationship,
				NodeID:     "new-rel-456",
				SourceNode: "src",
				TargetNode: "dst",
			},
		}

		got, err := syncer.ApplyPatch(src, ops)
		if err != nil {
			t.Fatalf("ApplyPatch failed: %v", err)
		}

		if !strings.Contains(got, `arch.Connect("new-rel-456"`) {
			t.Errorf("Expected arch.Connect in Build, got:\n%s", got)
		}
	})
}

package ast

import (
	"strings"
	"testing"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

func TestApplyPatch(t *testing.T) {
	syncer := GoASTSyncer{}

	t.Run("UpdateNode property by line", func(t *testing.T) {
		src := `package main
func main() {
	arch.DefineNode("id1", domain.Service, "OldName", "OldDesc")
}
`
		// Line 3 is where DefineNode is.
		ops := []domain.PatchOperation{
			{
				Type:     domain.PatchUpdateNode,
				Origin:   &domain.PatchOrigin{Line: 3},
				Property: "name",
				Value:    "NewName",
			},
		}

		got, err := syncer.ApplyPatch(src, ops)
		if err != nil {
			t.Fatalf("ApplyPatch failed: %v", err)
		}

		if !strings.Contains(got, `"NewName"`) {
			t.Errorf("Expected NewName in output, got:\n%s", got)
		}
	})

	t.Run("DeleteNode explicit by line", func(t *testing.T) {
		src := `package main
func main() {
	arch.DefineNode("id1", domain.Service, "N1", "D1")
	arch.DefineNode("id2", domain.Service, "N2", "D2")
}
`
		// Delete the first node at line 3
		ops := []domain.PatchOperation{
			{
				Type:   domain.PatchDeleteNode,
				Origin: &domain.PatchOrigin{Line: 3},
			},
		}

		got, err := syncer.ApplyPatch(src, ops)
		if err != nil {
			t.Fatalf("ApplyPatch failed: %v", err)
		}

		if strings.Contains(got, "id1") {
			t.Errorf("Expected id1 to be removed, got:\n%s", got)
		}
		if !strings.Contains(got, "id2") {
			t.Errorf("Expected id2 to remain, got:\n%s", got)
		}
	})

	t.Run("Update loop variable", func(t *testing.T) {
		src := `package main
const numGateways = 5
func main() {
	for i := 0; i < numGateways; i++ {}
}
`
		ops := []domain.PatchOperation{
			{
				Type:   domain.PatchDeleteNode, // Delete node in loop -> decrement count
				Origin: &domain.PatchOrigin{LoopVar: "numGateways"},
			},
		}

		got, err := syncer.ApplyPatch(src, ops)
		if err != nil {
			t.Fatalf("ApplyPatch failed: %v", err)
		}

		if !strings.Contains(got, "const numGateways = 4") {
			t.Errorf("Expected numGateways = 4, got:\n%s", got)
		}
	})

	t.Run("Update loop variable inside var block", func(t *testing.T) {
		src := `package main
var (
	numGateways = 5
)
`
		ops := []domain.PatchOperation{
			{
				Type:   domain.PatchDeleteNode,
				Origin: &domain.PatchOrigin{LoopVar: "numGateways"},
			},
		}

		got, err := syncer.ApplyPatch(src, ops)
		if err != nil {
			t.Fatalf("ApplyPatch failed: %v", err)
		}

		if !strings.Contains(got, "numGateways = 4") {
			t.Errorf("Expected numGateways = 4, got:\n%s", got)
		}
	})
}

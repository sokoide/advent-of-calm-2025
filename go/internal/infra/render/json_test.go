package render

import (
	"encoding/json"
	"testing"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

func TestJSONRenderer_Render(t *testing.T) {
	arch := domain.NewArchitecture("test-arch", "Test Architecture", "Desc")
	arch.Node("node1", domain.Service, "Node 1", "desc1")

	renderer := JSONRenderer{}
	output, err := renderer.Render(arch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded domain.Architecture
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if decoded.UniqueID != arch.UniqueID {
		t.Errorf("expected id %s, got %s", arch.UniqueID, decoded.UniqueID)
	}
	if len(decoded.Nodes) != 1 || decoded.Nodes[0].UniqueID != "node1" {
		t.Errorf("node not correctly rendered")
	}
}

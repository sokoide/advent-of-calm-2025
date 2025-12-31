package render

import (
	"strings"
	"testing"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

func TestGoDSLRenderer_Render(t *testing.T) {
	arch := domain.NewArchitecture("test-arch", "Test Architecture", "Desc")
	arch.Node("node1", domain.Service, "Node 1", "desc1").Standard("cc1", "owner1")
	arch.AddMeta("global-key", "global-val")

	renderer := GoDSLRenderer{}
	output, err := renderer.Render(arch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "package main") {
		t.Errorf("expected package main")
	}
	if !strings.Contains(output, "arch.DefineNode") {
		t.Errorf("expected arch.DefineNode")
	}
	if !strings.Contains(output, "\"node1\", Service, \"Node 1\"") {
		t.Errorf("expected node parameters")
	}
	if !strings.Contains(output, "WithOwner(\"owner1\", \"cc1\")") {
		t.Errorf("expected WithOwner")
	}
}

func TestGoDSLRenderer_Metadata(t *testing.T) {
	arch := domain.NewArchitecture("test", "Test", "Desc")
	arch.DefineNode("node1", domain.Service, "Node 1", "desc",
		domain.WithMeta(map[string]any{
			"complex": map[string]any{"nested": "value"},
		}),
	)

	renderer := GoDSLRenderer{}
	output, err := renderer.Render(arch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// This is expected to fail or produce placeholder until fixed
	if strings.Contains(output, "map[string]any{...}") {
		t.Errorf("metadata should not be rendered as placeholder")
	}
}

func TestGoDSLRenderer_Full(t *testing.T) {
	arch := domain.NewArchitecture("full-arch", "Full Architecture", "Full Desc")
	arch.ADRs = []string{"adr1.md"}
	
	arch.DefineNode("n1", domain.Service, "Node 1", "desc1",
		domain.WithInterfaces(&domain.Interface{UniqueID: "i1", Name: "Intf 1", Protocol: "http", Port: 80}),
	)
	arch.DefineNode("n2", domain.Database, "Node 2", "desc2")
	
	arch.Connect("r1", "Rel 1", "n1", "n2").Data("internal", true).WithProtocol("grpc")
	arch.Interacts("r2", "Rel 2", "actor1", "n1")
	arch.ComposedOf("r3", "Rel 3", "sys1", []string{"n1"})
	
	arch.DefineFlow("f1", "Flow 1", "desc").Step("r1", "step1")
	
	arch.AddControl("c1", "Control 1", domain.NewRequirement("url1", nil))

	renderer := GoDSLRenderer{}
	output, err := renderer.Render(arch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"arch.ADRs = []string{",
		"\"adr1.md\"",
		"WithInterfaces(",
		"&Interface{UniqueID: \"i1\", Name: \"Intf 1\", Protocol: \"http\", Port: 80}",
		"arch.AddRelationship(&Relationship{",
		"UniqueID: \"r1\"",
		"DataClassification: \"internal\"",
		"Encrypted: BoolPtr(true)",
		"\"actor\": \"actor1\"",
		"\"container\": \"sys1\"",
		"arch.DefineFlow(\"f1\", \"Flow 1\", \"desc\")",
		".Step(\"r1\", \"step1\")",
		"arch.Controls[\"c1\"]",
	}

	for _, c := range checks {
		if !strings.Contains(output, c) {
			t.Errorf("expected output to contain %q", c)
		}
	}
}

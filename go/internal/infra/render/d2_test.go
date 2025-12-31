package render

import (
	"strings"
	"testing"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

func TestD2Renderer_Render(t *testing.T) {
	arch := domain.NewArchitecture("test-arch", "Test Architecture", "Desc")
	arch.Node("node1", domain.Service, "Node 1", "desc1").Standard("cc1", "owner1")
	arch.Node("node2", domain.Database, "Node 2", "desc2")
	arch.Connect("rel1", "Connects", "node1", "node2").Data("confidential", true).WithProtocol("https")

	renderer := D2Renderer{}
	output, err := renderer.Render(arch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Basic checks
	if !strings.Contains(output, "node1: Node 1") {
		t.Errorf("expected node1 in output")
	}
	if !strings.Contains(output, "class: service") {
		t.Errorf("expected class service in output")
	}
	if !strings.Contains(output, "tooltip: \"Owner: owner1\"") {
		t.Errorf("expected owner1 in tooltip")
	}
	if !strings.Contains(output, "node1 -> node2: https (confidential)") {
		t.Errorf("expected relationship with label in output")
	}
}

func TestD2Renderer_ComposedNodes(t *testing.T) {
	arch := domain.NewArchitecture("test-arch", "Test", "Desc")
	arch.Node("sys1", domain.System, "System 1", "desc")
	arch.Node("svc1", domain.Service, "Service 1", "desc")
	arch.ComposedOf("comp1", "composed", "sys1", []string{"svc1"})

	renderer := D2Renderer{}
	output, err := renderer.Render(arch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "sys1: System 1 {") {
		t.Errorf("expected sys1 container in output")
	}
	if !strings.Contains(output, "  svc1: Service 1 {") {
		t.Errorf("expected svc1 inside container in output")
	}
}

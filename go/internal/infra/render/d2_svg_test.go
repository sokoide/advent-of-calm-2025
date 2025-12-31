package render

import (
	"strings"
	"testing"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

func TestD2Renderer_RenderSVG(t *testing.T) {
	arch := domain.NewArchitecture("test-arch", "Test Architecture", "Desc")
	arch.Node("node1", domain.Service, "Node 1", "desc1")
	arch.Node("node2", domain.Database, "Node 2", "desc2")
	arch.Connect("rel1", "Connects", "node1", "node2")

	renderer := D2Renderer{}
	svg, err := renderer.RenderSVG(arch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Basic checks for SVG content
	if !strings.Contains(svg, "<svg") {
		t.Errorf("expected <svg tag in output")
	}
	if !strings.Contains(svg, "</svg>") {
		t.Errorf("expected </svg> tag in output")
	}
	// Check if it contains some of our content
	if !strings.Contains(svg, "Node 1") {
		t.Errorf("expected 'Node 1' in SVG output")
	}
}

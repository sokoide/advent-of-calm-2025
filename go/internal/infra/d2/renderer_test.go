package d2

import (
	"strings"
	"testing"
)

func TestD2Renderer(t *testing.T) {
	renderer := NewD2Renderer()

	d2Code := "x -> y"
	svg, err := renderer.RenderToSVG(d2Code)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	if !strings.Contains(svg, "<svg") {
		t.Errorf("expected SVG output, got: %s", svg)
	}
}

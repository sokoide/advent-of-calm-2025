package render

import (
	"strings"
	"testing"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

func TestRichD2Renderer_Render(t *testing.T) {
	arch := domain.NewArchitecture("test-arch", "Test Architecture", "Desc")
	arch.Node("node1", domain.Service, "Node 1", "desc1").Standard("cc1", "owner1")
	arch.Flow("flow1", "Flow 1", "desc").Step("rel1", 1, "step1", "src-to-dst")

	renderer := RichD2Renderer{}
	output, err := renderer.Render(arch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "# @calm:id=test-arch") {
		t.Errorf("expected architecture id annotation")
	}
	if !strings.Contains(output, "# @calm:owner=owner1") {
		t.Errorf("expected node owner annotation")
	}
	if !strings.Contains(output, "# @calm:flow id=flow1") {
		t.Errorf("expected flow annotation")
	}
	if !strings.Contains(output, "# @calm:flow-step seq=1") {
		t.Errorf("expected flow step annotation")
	}
}

func TestRichD2Renderer_SpecialChars(t *testing.T) {
	arch := domain.NewArchitecture("test", "Test", "Desc with\nnewline and = equals")

	renderer := RichD2Renderer{}
	output, err := renderer.Render(arch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Desc with\\nnewline and \\= equals") {
		t.Errorf("special characters not escaped correctly in architecture description")
	}
}

func TestRichD2Renderer_Full(t *testing.T) {
	arch := domain.NewArchitecture("rich-arch", "Rich Architecture", "Rich Desc")

	arch.DefineNode("sys1", domain.System, "System 1", "desc")
	arch.DefineNode("svc1", domain.Service, "Service 1", "desc")
	arch.ComposedOf("comp1", "composed", "sys1", []string{"svc1"})

	arch.Interacts("int1", "interacts", "actor1", "svc1")

	renderer := RichD2Renderer{}
	output, err := renderer.Render(arch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"sys1: System 1 {",
		"class: system",
		"svc1: Service 1 {",
		"class: service",
		"# @calm:composed-of id=comp1",
		"actor1 -> sys1.svc1 {",
		"# @calm:type=interacts",
	}

	for _, c := range checks {
		if !strings.Contains(output, c) {
			t.Errorf("expected output to contain %q", c)
		}
	}
}

func TestRichD2Renderer_FlowsAndControls(t *testing.T) {
	arch := domain.NewArchitecture("fc-arch", "Flows and Controls", "desc")
	arch.DefineNode("n1", domain.Service, "N1", "desc")
	arch.DefineNode("n2", domain.Service, "N2", "desc")
	rel := arch.Connect("r1", "desc", "n1", "n2")

	arch.DefineFlow("f1", "Flow 1", "Flow Desc").
		Meta("key", "val").
		Step(rel.UniqueID, "Step 1")

	arch.AddControl("c1", "Control 1")

	renderer := RichD2Renderer{}
	output, err := renderer.Render(arch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "# @calm:flow id=f1") {
		t.Errorf("expected flow annotation")
	}
	if !strings.Contains(output, "# @calm:flow-metadata={\"key\":\"val\"}") {
		t.Errorf("expected flow metadata")
	}
	if !strings.Contains(output, "# @calm:control id=c1") {
		t.Errorf("expected control annotation")
	}
}

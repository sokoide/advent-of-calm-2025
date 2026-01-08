package d2

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

// D2Renderer implements usecase.DiagramRenderer using the d2 CLI.
type D2Renderer struct{}

// NewD2Renderer creates a new D2Renderer.
func NewD2Renderer() *D2Renderer {
	return &D2Renderer{}
}

// RenderToSVG generates SVG from D2 source using the d2 CLI.
func (r *D2Renderer) RenderToSVG(d2Source string) (string, error) {
	if strings.TrimSpace(d2Source) == "" {
		return "", nil
	}

	d2Cmd := exec.Command("d2", "-", "-")
	d2Cmd.Stdin = strings.NewReader(d2Source)
	var svgOut, svgErr bytes.Buffer
	d2Cmd.Stdout = &svgOut
	d2Cmd.Stderr = &svgErr

	if err := d2Cmd.Run(); err != nil {
		log.Printf("❌ D2 SVG error: %v\n%s", err, svgErr.String())
		return "", err
	}

	return svgOut.String(), nil
}

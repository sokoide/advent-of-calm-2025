package handlers

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

// D2Runner abstracts D2 CLI execution for testability.
type D2Runner interface {
	RenderToSVG(d2Source string) (string, error)
}

// ExecD2Runner executes the d2 CLI.
type ExecD2Runner struct{}

// RenderToSVG generates SVG from D2 source using the d2 CLI.
func (ExecD2Runner) RenderToSVG(d2Source string) (string, error) {
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

// GenerateSVGFromD2 is a convenience function using ExecD2Runner.
func GenerateSVGFromD2(d2Source string) string {
	svg, _ := ExecD2Runner{}.RenderToSVG(d2Source)
	if svg != "" {
		log.Printf("🎨 D2 SVG generated (%d bytes)", len(svg))
	}
	return svg
}

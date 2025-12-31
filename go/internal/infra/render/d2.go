package render

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// D2Renderer renders CALM architectures into D2 source.
type D2Renderer struct{}

// Render generates D2 diagram source from the architecture.
func (D2Renderer) Render(a *domain.Architecture) (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("# CALM Architecture: " + a.Name + "\n")
	sb.WriteString("# Generated from Go DSL\n\n")

	// Direction
	sb.WriteString("direction: right\n\n")

	// Style definitions
	sb.WriteString("classes: {\n")
	sb.WriteString("  actor: {\n    shape: person\n    style.fill: \"#e1f5fe\"\n  }\n")
	sb.WriteString("  service: {\n    shape: rectangle\n    style.fill: \"#e8f5e9\"\n    style.border-radius: 8\n  }\n")
	sb.WriteString("  database: {\n    shape: cylinder\n    style.fill: \"#fff3e0\"\n  }\n")
	sb.WriteString("  queue: {\n    shape: queue\n    style.fill: \"#f3e5f5\"\n  }\n")
	sb.WriteString("  system: {\n    shape: rectangle\n    style.fill: \"#fafafa\"\n    style.stroke-dash: 3\n  }\n")
	sb.WriteString("}\n\n")

	// Track composed nodes
	composedNodes := make(map[string]string) // node -> container
	containers := make(map[string][]string)  // container -> nodes

	for _, rel := range a.Relationships {
		if rel.RelationshipType.ComposedOf != nil {
			container, _ := rel.RelationshipType.ComposedOf["container"].(string)
			if nodes, ok := rel.RelationshipType.ComposedOf["nodes"].([]string); ok {
				for _, n := range nodes {
					composedNodes[n] = container
				}
				containers[container] = append(containers[container], nodes...)
			}
		}
	}

	// Generate containers (systems) first
	for _, node := range a.Nodes {
		if node.NodeType == domain.System {
			if containedNodes, isContainer := containers[node.UniqueID]; isContainer {
				// This is a container with children
				sb.WriteString(fmt.Sprintf("%s: %s {\n", sanitizeID(node.UniqueID), node.Name))
				sb.WriteString("  class: system\n")

				// Add contained nodes
				for _, containedID := range containedNodes {
					for _, n := range a.Nodes {
						if n.UniqueID == containedID {
							writeNode(&sb, n, "  ")
							break
						}
					}
				}

				sb.WriteString("}\n\n")
			}
		}
	}

	// Generate standalone nodes (not in containers)
	for _, node := range a.Nodes {
		if _, inContainer := composedNodes[node.UniqueID]; !inContainer {
			// Skip systems that are containers (already handled)
			if node.NodeType == domain.System {
				if _, isContainer := containers[node.UniqueID]; isContainer {
					continue
				}
			}
			writeNode(&sb, node, "")
		}
	}

	sb.WriteString("\n# Relationships\n")

	// Generate relationships
	for _, rel := range a.Relationships {
		if rel.RelationshipType.Connects != nil {
			src := rel.RelationshipType.Connects.Source.Node
			dst := rel.RelationshipType.Connects.Destination.Node

			// Handle nodes in containers
			srcPath := getNodePath(src, composedNodes)
			dstPath := getNodePath(dst, composedNodes)

			label := ""
			if rel.Protocol != "" {
				label = rel.Protocol
			}
			if rel.DataClassification != "" {
				if label != "" {
					label += " "
				}
				label += "(" + rel.DataClassification + ")"
			}

			if label != "" {
				sb.WriteString(fmt.Sprintf("%s -> %s: %s\n", srcPath, dstPath, label))
			} else {
				sb.WriteString(fmt.Sprintf("%s -> %s\n", srcPath, dstPath))
			}
		}

		if rel.RelationshipType.Interacts != nil {
			actor, _ := rel.RelationshipType.Interacts["actor"].(string)
			if nodes, ok := rel.RelationshipType.Interacts["nodes"].([]string); ok {
				for _, n := range nodes {
					dstPath := getNodePath(n, composedNodes)
					sb.WriteString(fmt.Sprintf("%s -> %s\n", sanitizeID(actor), dstPath))
				}
			}
		}
	}

	return sb.String(), nil
}

// RenderSVG generates SVG from the architecture using the d2 CLI.
func (r D2Renderer) RenderSVG(a *domain.Architecture) (string, error) {
	d2Source, err := r.Render(a)
	if err != nil {
		return "", err
	}

	// Create a temporary file for D2 source
	tmpFile, err := os.CreateTemp("", "calm-*.d2")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(d2Source); err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpFile.Close()

	// Execute d2 CLI to generate SVG to stdout
	cmd := exec.Command("d2", "-", "-")
	cmd.Stdin = strings.NewReader(d2Source)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("d2 execution failed: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("d2 execution failed: %w", err)
	}

	return string(output), nil
}

func writeNode(sb *strings.Builder, node *domain.Node, indent string) {
	id := sanitizeID(node.UniqueID)
	className := strings.ToLower(string(node.NodeType))

	// Node with label
	sb.WriteString(fmt.Sprintf("%s%s: %s {\n", indent, id, node.Name))
	sb.WriteString(fmt.Sprintf("%s  class: %s\n", indent, className))

	// Add owner tooltip if available
	if node.Owner != "" {
		sb.WriteString(fmt.Sprintf("%s  tooltip: \"Owner: %s\"\n", indent, node.Owner))
	}

	sb.WriteString(indent + "}\n")
}

func getNodePath(nodeID string, composedNodes map[string]string) string {
	if container, ok := composedNodes[nodeID]; ok {
		return sanitizeID(container) + "." + sanitizeID(nodeID)
	}
	return sanitizeID(nodeID)
}

func sanitizeID(id string) string {
	// D2 IDs can contain hyphens, but we need to escape special characters
	return strings.ReplaceAll(id, " ", "-")
}

// WriteD2File writes D2 output to a file
func (D2Renderer) WriteD2(filename string, a *domain.Architecture) error {
	d2Source, _ := D2Renderer{}.Render(a)
	return writeFile(filename, d2Source)
}

func writeFile(filename, content string) error {
	return nil // Placeholder - would use os.WriteFile
}

package render

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// RichD2Renderer renders CALM architectures into Rich D2 source.
type RichD2Renderer struct{}

// Render generates D2 diagram source with embedded CALM metadata.
// The metadata is stored in comments with @calm: prefix for bidirectional editing.
func (RichD2Renderer) Render(a *domain.Architecture) (string, error) {
	var sb strings.Builder

	// Header with architecture metadata
	sb.WriteString("# CALM Architecture: " + a.Name + "\n")
	sb.WriteString("# @calm:id=" + a.UniqueID + "\n")
	sb.WriteString("# @calm:description=" + escapeD2String(a.Description) + "\n")
	if len(a.ADRs) > 0 {
		adrsJSON, _ := json.Marshal(a.ADRs)
		sb.WriteString("# @calm:adrs=" + string(adrsJSON) + "\n")
	}
	sb.WriteString("\n")

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
	composedNodes := make(map[string]string)
	containers := make(map[string][]string)

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

	// Generate containers (systems) with their contents
	for _, node := range a.Nodes {
		if node.NodeType == domain.System {
			if containedNodes, isContainer := containers[node.UniqueID]; isContainer {
				// Write container opening with contents inside
				writeRichContainerNode(&sb, node, containedNodes, a.Nodes)
				sb.WriteString("\n")
			}
		}
	}

	// Generate standalone nodes
	for _, node := range a.Nodes {
		if _, inContainer := composedNodes[node.UniqueID]; !inContainer {
			if node.NodeType == domain.System {
				if _, isContainer := containers[node.UniqueID]; isContainer {
					continue
				}
			}
			writeRichNode(&sb, node, "")
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n# Relationships\n")

	// Generate relationships with metadata
	for _, rel := range a.Relationships {
		if rel.RelationshipType.Connects != nil {
			src := rel.RelationshipType.Connects.Source.Node
			dst := rel.RelationshipType.Connects.Destination.Node

			srcPath := getNodePath(src, composedNodes)
			dstPath := getNodePath(dst, composedNodes)

			// Build label
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

			// Write relationship
			if label != "" {
				sb.WriteString(fmt.Sprintf("%s -> %s: %s", srcPath, dstPath, label))
			} else {
				sb.WriteString(fmt.Sprintf("%s -> %s", srcPath, dstPath))
			}

			// Write CALM metadata as block
			sb.WriteString(" {\n")
			sb.WriteString(fmt.Sprintf("  # @calm:id=%s\n", rel.UniqueID))
			sb.WriteString(fmt.Sprintf("  # @calm:description=%s\n", escapeD2String(rel.Description)))

			if rel.RelationshipType.Connects.Source.Interfaces != nil {
				sb.WriteString(fmt.Sprintf("  # @calm:srcInterfaces=%s\n", toJSON(rel.RelationshipType.Connects.Source.Interfaces)))
			}
			if rel.RelationshipType.Connects.Destination.Interfaces != nil {
				sb.WriteString(fmt.Sprintf("  # @calm:dstInterfaces=%s\n", toJSON(rel.RelationshipType.Connects.Destination.Interfaces)))
			}
			if rel.Encrypted != nil {
				sb.WriteString(fmt.Sprintf("  # @calm:encrypted=%t\n", *rel.Encrypted))
			}
			if rel.DataClassification != "" {
				sb.WriteString(fmt.Sprintf("  # @calm:classification=%s\n", rel.DataClassification))
			}
			if len(rel.Metadata) > 0 {
				sb.WriteString(fmt.Sprintf("  # @calm:metadata=%s\n", toJSON(rel.Metadata)))
			}

			sb.WriteString("}\n")
		}

		if rel.RelationshipType.Interacts != nil {
			actor, _ := rel.RelationshipType.Interacts["actor"].(string)
			if nodes, ok := rel.RelationshipType.Interacts["nodes"].([]string); ok {
				for _, n := range nodes {
					dstPath := getNodePath(n, composedNodes)
					sb.WriteString(fmt.Sprintf("%s -> %s {\n", sanitizeID(actor), dstPath))
					sb.WriteString(fmt.Sprintf("  # @calm:id=%s\n", rel.UniqueID))
					sb.WriteString(fmt.Sprintf("  # @calm:type=interacts\n"))
					sb.WriteString(fmt.Sprintf("  # @calm:actor=%s\n", actor))
					if rel.DataClassification != "" {
						sb.WriteString(fmt.Sprintf("  # @calm:classification=%s\n", rel.DataClassification))
					}
					sb.WriteString("}\n")
				}
			}
		}

		if rel.RelationshipType.ComposedOf != nil {
			container, _ := rel.RelationshipType.ComposedOf["container"].(string)
			nodes, _ := rel.RelationshipType.ComposedOf["nodes"].([]string)
			sb.WriteString(fmt.Sprintf("# @calm:composed-of id=%s container=%s nodes=%s\n",
				rel.UniqueID, container, toJSON(nodes)))
		}
	}

	// Generate flows
	if len(a.Flows) > 0 {
		sb.WriteString("\n# Flows\n")
		for _, flow := range a.Flows {
			sb.WriteString(fmt.Sprintf("# @calm:flow id=%s name=%s\n", flow.UniqueID, escapeD2String(flow.Name)))
			sb.WriteString(fmt.Sprintf("# @calm:flow-description=%s\n", escapeD2String(flow.Description)))
			if len(flow.Metadata) > 0 {
				sb.WriteString(fmt.Sprintf("# @calm:flow-metadata=%s\n", toJSON(flow.Metadata)))
			}
			for _, t := range flow.Transitions {
				sb.WriteString(fmt.Sprintf("# @calm:flow-step seq=%d rel=%s dir=%s desc=%s\n",
					t.SequenceNumber, t.RelationshipID, t.Direction, escapeD2String(t.Description)))
			}
		}
	}

	// Generate global controls
	if len(a.Controls) > 0 {
		sb.WriteString("\n# Global Controls\n")
		for id, ctrl := range a.Controls {
			ctrlJSON, _ := json.Marshal(ctrl)
			sb.WriteString(fmt.Sprintf("# @calm:control id=%s data=%s\n", id, string(ctrlJSON)))
		}
	}

	return sb.String(), nil
}

func writeRichNode(sb *strings.Builder, node *domain.Node, indent string) {
	id := sanitizeID(node.UniqueID)
	className := strings.ToLower(string(node.NodeType))

	sb.WriteString(fmt.Sprintf("%s%s: %s {\n", indent, id, node.Name))
	sb.WriteString(fmt.Sprintf("%s  class: %s\n", indent, className))

	// CALM metadata as comments
	sb.WriteString(fmt.Sprintf("%s  # @calm:id=%s\n", indent, node.UniqueID))
	sb.WriteString(fmt.Sprintf("%s  # @calm:type=%s\n", indent, node.NodeType))
	if node.Owner != "" {
		sb.WriteString(fmt.Sprintf("%s  # @calm:owner=%s\n", indent, node.Owner))
		sb.WriteString(fmt.Sprintf("%s  tooltip: \"Owner: %s\"\n", indent, node.Owner))
	}
	if node.CostCenter != "" {
		sb.WriteString(fmt.Sprintf("%s  # @calm:costCenter=%s\n", indent, node.CostCenter))
	}
	if node.Description != "" {
		sb.WriteString(fmt.Sprintf("%s  # @calm:description=%s\n", indent, escapeD2String(node.Description)))
	}

	// Metadata as JSON
	if len(node.Metadata) > 0 {
		sb.WriteString(fmt.Sprintf("%s  # @calm:metadata=%s\n", indent, toJSON(node.Metadata)))
	}

	// Interfaces as JSON
	if len(node.Interfaces) > 0 {
		sb.WriteString(fmt.Sprintf("%s  # @calm:interfaces=%s\n", indent, toJSON(node.Interfaces)))
	}

	// Controls as JSON
	if len(node.Controls) > 0 {
		sb.WriteString(fmt.Sprintf("%s  # @calm:controls=%s\n", indent, toJSON(node.Controls)))
	}

	sb.WriteString(indent + "}")
}

func escapeD2String(s string) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "=", "\\=")
	return s
}

func toJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// writeRichContainerNode writes a container node with its children in a single D2 block.
func writeRichContainerNode(sb *strings.Builder, container *domain.Node, childIDs []string, allNodes []*domain.Node) {
	id := sanitizeID(container.UniqueID)
	className := strings.ToLower(string(container.NodeType))

	// Start container block
	sb.WriteString(fmt.Sprintf("%s: %s {\n", id, container.Name))
	sb.WriteString(fmt.Sprintf("  class: %s\n", className))

	// Container metadata
	sb.WriteString(fmt.Sprintf("  # @calm:id=%s\n", container.UniqueID))
	sb.WriteString(fmt.Sprintf("  # @calm:type=%s\n", container.NodeType))
	if container.Owner != "" {
		sb.WriteString(fmt.Sprintf("  # @calm:owner=%s\n", container.Owner))
		sb.WriteString(fmt.Sprintf("  tooltip: \"Owner: %s\"\n", container.Owner))
	}
	if container.CostCenter != "" {
		sb.WriteString(fmt.Sprintf("  # @calm:costCenter=%s\n", container.CostCenter))
	}
	if container.Description != "" {
		sb.WriteString(fmt.Sprintf("  # @calm:description=%s\n", escapeD2String(container.Description)))
	}
	if len(container.Metadata) > 0 {
		sb.WriteString(fmt.Sprintf("  # @calm:metadata=%s\n", toJSON(container.Metadata)))
	}

	sb.WriteString("\n")

	// Write child nodes
	for _, childID := range childIDs {
		for _, n := range allNodes {
			if n.UniqueID == childID {
				writeRichNode(sb, n, "  ")
				sb.WriteString("\n")
				break
			}
		}
	}

	sb.WriteString("}\n")
}

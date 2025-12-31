package render

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// GoDSLRenderer renders CALM architectures into Go DSL source.
type GoDSLRenderer struct{}

// Render generates Go DSL code from a CALM Architecture.
// This enables the D2 → Go direction of bidirectional editing.
func (GoDSLRenderer) Render(a *domain.Architecture) (string, error) {
	var sb strings.Builder

	sb.WriteString("package main\n\n")
	sb.WriteString("func buildArchitecture() *Architecture {\n")
	sb.WriteString(fmt.Sprintf("\tarch := NewArchitecture(\n\t\t%q,\n\t\t%q,\n\t\t%q,\n\t)\n\n",
		a.UniqueID, a.Name, a.Description))

	// ADRs
	if len(a.ADRs) > 0 {
		sb.WriteString("\tarch.ADRs = []string{\n")
		for _, adr := range a.ADRs {
			sb.WriteString(fmt.Sprintf("\t\t%q,\n", adr))
		}
		sb.WriteString("\t}\n\n")
	}

	// Nodes
	sb.WriteString("\t// Define nodes\n")
	for _, node := range a.Nodes {
		generateNodeDSL(&sb, node)
	}

	// Relationships
	sb.WriteString("\n\t// Define relationships\n")
	for _, rel := range a.Relationships {
		generateRelDSL(&sb, rel)
	}

	// Flows
	if len(a.Flows) > 0 {
		sb.WriteString("\n\t// Define flows\n")
		for _, flow := range a.Flows {
			generateFlowDSL(&sb, flow)
		}
	}

	// Controls
	if len(a.Controls) > 0 {
		sb.WriteString("\n\t// Define global controls\n")
		for id, ctrl := range a.Controls {
			generateControlDSL(&sb, id, ctrl)
		}
	}

	sb.WriteString("\n\treturn arch\n")
	sb.WriteString("}\n")

	return sb.String(), nil
}

func generateNodeDSL(sb *strings.Builder, node *domain.Node) {
	nodeType := string(node.NodeType)
	caser := cases.Title(language.English)
	nodeType = caser.String(strings.ToLower(nodeType))

	sb.WriteString(fmt.Sprintf("\tarch.DefineNode(\n"))
	sb.WriteString(fmt.Sprintf("\t\t%q, %s, %q,\n", node.UniqueID, nodeType, node.Name))

	// Options
	if node.Owner != "" {
		sb.WriteString(fmt.Sprintf("\t\tWithOwner(%q),\n", node.Owner))
	}
	if node.CostCenter != "" {
		sb.WriteString(fmt.Sprintf("\t\tWithCostCenter(%q),\n", node.CostCenter))
	}
	if node.Description != "" {
		sb.WriteString(fmt.Sprintf("\t\tWithDescription(%q),\n", node.Description))
	}

	// Metadata
	if len(node.Metadata) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tWithMeta(map[string]any{\n"))
		keys := sortedKeys(node.Metadata)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("\t\t\t%q: %s,\n", k, formatValue(node.Metadata[k])))
		}
		sb.WriteString("\t\t}),\n")
	}

	// Interfaces
	if len(node.Interfaces) > 0 {
		sb.WriteString("\t\tWithInterfaces(\n")
		for _, iface := range node.Interfaces {
			generateInterfaceDSL(sb, iface)
		}
		sb.WriteString("\t\t),\n")
	}

	sb.WriteString("\t)\n\n")
}

func generateInterfaceDSL(sb *strings.Builder, iface domain.Interface) {
	sb.WriteString(fmt.Sprintf("\t\t\t&Interface{UniqueID: %q, Name: %q, Protocol: %q",
		iface.UniqueID, iface.Name, iface.Protocol))
	if iface.Port > 0 {
		sb.WriteString(fmt.Sprintf(", Port: %d", iface.Port))
	}
	if iface.Host != "" {
		sb.WriteString(fmt.Sprintf(", Host: %q", iface.Host))
	}
	sb.WriteString("},\n")
}

func generateRelDSL(sb *strings.Builder, rel *domain.Relationship) {
	if rel.RelationshipType.Connects != nil {
		src := rel.RelationshipType.Connects.Source.Node
		dst := rel.RelationshipType.Connects.Destination.Node

		sb.WriteString(fmt.Sprintf("\t// %s\n", rel.UniqueID))
		sb.WriteString(fmt.Sprintf("\tarch.AddRelationship(&Relationship{\n"))
		sb.WriteString(fmt.Sprintf("\t\tUniqueID: %q,\n", rel.UniqueID))
		sb.WriteString(fmt.Sprintf("\t\tDescription: %q,\n", rel.Description))

		if rel.Protocol != "" {
			sb.WriteString(fmt.Sprintf("\t\tProtocol: %q,\n", rel.Protocol))
		}
		if rel.Encrypted != nil && *rel.Encrypted {
			sb.WriteString("\t\tEncrypted: BoolPtr(true),\n")
		}
		if rel.DataClassification != "" {
			sb.WriteString(fmt.Sprintf("\t\tDataClassification: %q,\n", rel.DataClassification))
		}

		sb.WriteString("\t\tRelationshipType: RelationshipType{\n")
		sb.WriteString("\t\t\tConnects: &Connects{\n")
		sb.WriteString(fmt.Sprintf("\t\t\t\tSource: NodeInterface{Node: %q", src))
		if len(rel.RelationshipType.Connects.Source.Interfaces) > 0 {
			sb.WriteString(fmt.Sprintf(", Interfaces: %s", formatStringSlice(rel.RelationshipType.Connects.Source.Interfaces)))
		}
		sb.WriteString("},\n")
		sb.WriteString(fmt.Sprintf("\t\t\t\tDestination: NodeInterface{Node: %q", dst))
		if len(rel.RelationshipType.Connects.Destination.Interfaces) > 0 {
			sb.WriteString(fmt.Sprintf(", Interfaces: %s", formatStringSlice(rel.RelationshipType.Connects.Destination.Interfaces)))
		}
		sb.WriteString("},\n")
		sb.WriteString("\t\t\t},\n")
		sb.WriteString("\t\t},\n")
		sb.WriteString("\t})\n\n")
	}

	if rel.RelationshipType.Interacts != nil {
		actor, _ := rel.RelationshipType.Interacts["actor"].(string)
		nodes, _ := rel.RelationshipType.Interacts["nodes"].([]string)

		sb.WriteString(fmt.Sprintf("\t// %s (interacts)\n", rel.UniqueID))
		sb.WriteString(fmt.Sprintf("\tarch.AddRelationship(&Relationship{\n"))
		sb.WriteString(fmt.Sprintf("\t\tUniqueID: %q,\n", rel.UniqueID))
		sb.WriteString(fmt.Sprintf("\t\tDescription: %q,\n", rel.Description))
		sb.WriteString("\t\tRelationshipType: RelationshipType{\n")
		sb.WriteString("\t\t\tInteracts: map[string]any{\n")
		sb.WriteString(fmt.Sprintf("\t\t\t\t\"actor\": %q,\n", actor))
		sb.WriteString(fmt.Sprintf("\t\t\t\t\"nodes\": %s,\n", formatStringSlice(nodes)))
		sb.WriteString("\t\t\t},\n")
		sb.WriteString("\t\t},\n")
		sb.WriteString("\t})\n\n")
	}

	if rel.RelationshipType.ComposedOf != nil {
		container, _ := rel.RelationshipType.ComposedOf["container"].(string)
		nodes, _ := rel.RelationshipType.ComposedOf["nodes"].([]string)

		sb.WriteString(fmt.Sprintf("\t// %s (composed-of)\n", rel.UniqueID))
		sb.WriteString(fmt.Sprintf("\tarch.AddRelationship(&Relationship{\n"))
		sb.WriteString(fmt.Sprintf("\t\tUniqueID: %q,\n", rel.UniqueID))
		sb.WriteString("\t\tRelationshipType: RelationshipType{\n")
		sb.WriteString("\t\t\tComposedOf: map[string]any{\n")
		sb.WriteString(fmt.Sprintf("\t\t\t\t\"container\": %q,\n", container))
		sb.WriteString(fmt.Sprintf("\t\t\t\t\"nodes\": %s,\n", formatStringSlice(nodes)))
		sb.WriteString("\t\t\t},\n")
		sb.WriteString("\t\t},\n")
		sb.WriteString("\t})\n\n")
	}
}

func generateFlowDSL(sb *strings.Builder, flow *domain.Flow) {
	sb.WriteString(fmt.Sprintf("\tarch.DefineFlow(%q, %q,\n", flow.UniqueID, flow.Name))
	sb.WriteString(fmt.Sprintf("\t\tWithFlowDescription(%q),\n", flow.Description))

	for _, t := range flow.Transitions {
		sb.WriteString(fmt.Sprintf("\t\tWithStep(%d, %q, %q, %q),\n",
			t.SequenceNumber, t.RelationshipID, t.Direction, t.Description))
	}

	sb.WriteString("\t)\n\n")
}

func generateControlDSL(sb *strings.Builder, id string, ctrl *domain.Control) {
	sb.WriteString(fmt.Sprintf("\tarch.Controls[%q] = &Control{\n", id))
	sb.WriteString(fmt.Sprintf("\t\tDescription: %q,\n", ctrl.Description))
	if len(ctrl.Requirements) > 0 {
		sb.WriteString("\t\tRequirements: []*Requirement{\n")
		for _, req := range ctrl.Requirements {
			sb.WriteString(fmt.Sprintf("\t\t\t{RequirementURL: %q},\n", req.RequirementURL))
		}
		sb.WriteString("\t\t},\n")
	}
	sb.WriteString("\t}\n")
}

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		if val == float64(int(val)) {
			return fmt.Sprintf("%d", int(val))
		}
		return fmt.Sprintf("%v", val)
	case []any:
		var items []string
		for _, item := range val {
			items = append(items, formatValue(item))
		}
		return "[]any{" + strings.Join(items, ", ") + "}"
	case []string:
		return formatStringSlice(val)
	case map[string]any:
		var items []string
		keys := sortedKeys(val)
		for _, k := range keys {
			items = append(items, fmt.Sprintf("%q: %s", k, formatValue(val[k])))
		}
		return "map[string]any{" + strings.Join(items, ", ") + "}"
	default:
		return fmt.Sprintf("%#v", val)
	}
}

func formatStringSlice(ss []string) string {
	if len(ss) == 0 {
		return "[]string{}"
	}
	var items []string
	for _, s := range ss {
		items = append(items, fmt.Sprintf("%q", s))
	}
	return "[]string{" + strings.Join(items, ", ") + "}"
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

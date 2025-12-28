package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
)

type Architecture struct {
	UniqueID      string         `json:"unique-id"`
	Name          string         `json:"name"`
	Nodes         []Node         `json:"nodes"`
	Relationships []Relationship `json:"relationships"`
	Flows         []Flow         `json:"flows"`
	Controls      map[string]any `json:"controls"`
}

type Node struct {
	UniqueID string         `json:"unique-id"`
	Name     string         `json:"name"`
	NodeType string         `json:"node-type"`
	Owner    string         `json:"owner"`
	Metadata map[string]any `json:"metadata"`
	Controls map[string]any `json:"controls"`
}

type Relationship struct {
	UniqueID    string `json:"unique-id"`
	Description string `json:"description"`
}

type Flow struct {
	UniqueID    string `json:"unique-id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: arch-diff <old.json> <new.json>")
		fmt.Println("       arch-diff --git <old.json> <new.json>  (for git diff)")
		os.Exit(1)
	}

	oldPath := os.Args[1]
	newPath := os.Args[2]

	// Support git diff mode: git diff passes extra args
	if oldPath == "--git" && len(os.Args) >= 4 {
		oldPath = os.Args[2]
		newPath = os.Args[3]
	}

	oldArch, err := loadArchitecture(oldPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", oldPath, err)
		os.Exit(1)
	}

	newArch, err := loadArchitecture(newPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", newPath, err)
		os.Exit(1)
	}

	diff := compareArchitectures(oldArch, newArch)
	printDiff(diff)
}

func loadArchitecture(path string) (*Architecture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var arch Architecture
	if err := json.Unmarshal(data, &arch); err != nil {
		return nil, err
	}
	return &arch, nil
}

type ArchDiff struct {
	NodesAdded       []string
	NodesRemoved     []string
	NodesModified    []NodeChange
	RelsAdded        []string
	RelsRemoved      []string
	FlowsAdded       []string
	FlowsRemoved     []string
	ControlsAdded    []string
	ControlsRemoved  []string
	ControlsModified []string
}

type NodeChange struct {
	ID      string
	Changes []string
}

func compareArchitectures(old, new *Architecture) *ArchDiff {
	diff := &ArchDiff{}

	// Compare nodes
	oldNodes := make(map[string]Node)
	for _, n := range old.Nodes {
		oldNodes[n.UniqueID] = n
	}
	newNodes := make(map[string]Node)
	for _, n := range new.Nodes {
		newNodes[n.UniqueID] = n
	}

	for id := range newNodes {
		if _, exists := oldNodes[id]; !exists {
			diff.NodesAdded = append(diff.NodesAdded, id)
		}
	}
	for id := range oldNodes {
		if _, exists := newNodes[id]; !exists {
			diff.NodesRemoved = append(diff.NodesRemoved, id)
		}
	}
	for id, newNode := range newNodes {
		if oldNode, exists := oldNodes[id]; exists {
			changes := compareNodes(oldNode, newNode)
			if len(changes) > 0 {
				diff.NodesModified = append(diff.NodesModified, NodeChange{ID: id, Changes: changes})
			}
		}
	}

	// Compare relationships
	oldRels := make(map[string]bool)
	for _, r := range old.Relationships {
		oldRels[r.UniqueID] = true
	}
	newRels := make(map[string]bool)
	for _, r := range new.Relationships {
		newRels[r.UniqueID] = true
	}

	for id := range newRels {
		if !oldRels[id] {
			diff.RelsAdded = append(diff.RelsAdded, id)
		}
	}
	for id := range oldRels {
		if !newRels[id] {
			diff.RelsRemoved = append(diff.RelsRemoved, id)
		}
	}

	// Compare flows
	oldFlows := make(map[string]bool)
	for _, f := range old.Flows {
		oldFlows[f.UniqueID] = true
	}
	newFlows := make(map[string]bool)
	for _, f := range new.Flows {
		newFlows[f.UniqueID] = true
	}

	for id := range newFlows {
		if !oldFlows[id] {
			diff.FlowsAdded = append(diff.FlowsAdded, id)
		}
	}
	for id := range oldFlows {
		if !newFlows[id] {
			diff.FlowsRemoved = append(diff.FlowsRemoved, id)
		}
	}

	// Compare controls
	for id := range new.Controls {
		if _, exists := old.Controls[id]; !exists {
			diff.ControlsAdded = append(diff.ControlsAdded, id)
		}
	}
	for id := range old.Controls {
		if _, exists := new.Controls[id]; !exists {
			diff.ControlsRemoved = append(diff.ControlsRemoved, id)
		}
	}

	// Sort for consistent output
	sort.Strings(diff.NodesAdded)
	sort.Strings(diff.NodesRemoved)
	sort.Strings(diff.RelsAdded)
	sort.Strings(diff.RelsRemoved)
	sort.Strings(diff.FlowsAdded)
	sort.Strings(diff.FlowsRemoved)

	return diff
}

func compareNodes(old, new Node) []string {
	var changes []string

	if old.Name != new.Name {
		changes = append(changes, fmt.Sprintf("name: %q → %q", old.Name, new.Name))
	}
	if old.Owner != new.Owner {
		changes = append(changes, fmt.Sprintf("owner: %q → %q", old.Owner, new.Owner))
	}
	if old.NodeType != new.NodeType {
		changes = append(changes, fmt.Sprintf("type: %q → %q", old.NodeType, new.NodeType))
	}

	// Compare metadata keys
	for k := range new.Metadata {
		if _, exists := old.Metadata[k]; !exists {
			changes = append(changes, fmt.Sprintf("metadata +%s", k))
		}
	}
	for k := range old.Metadata {
		if _, exists := new.Metadata[k]; !exists {
			changes = append(changes, fmt.Sprintf("metadata -%s", k))
		}
	}

	// Compare controls
	for k := range new.Controls {
		if _, exists := old.Controls[k]; !exists {
			changes = append(changes, fmt.Sprintf("control +%s", k))
		}
	}
	for k := range old.Controls {
		if _, exists := new.Controls[k]; !exists {
			changes = append(changes, fmt.Sprintf("control -%s", k))
		}
	}

	return changes
}

func printDiff(diff *ArchDiff) {
	hasChanges := false

	if len(diff.NodesAdded) > 0 {
		hasChanges = true
		fmt.Printf("\n%s%s📦 Nodes Added:%s\n", colorBold, colorGreen, colorReset)
		for _, id := range diff.NodesAdded {
			fmt.Printf("  %s+ %s%s\n", colorGreen, id, colorReset)
		}
	}

	if len(diff.NodesRemoved) > 0 {
		hasChanges = true
		fmt.Printf("\n%s%s📦 Nodes Removed:%s\n", colorBold, colorRed, colorReset)
		for _, id := range diff.NodesRemoved {
			fmt.Printf("  %s- %s%s\n", colorRed, id, colorReset)
		}
	}

	if len(diff.NodesModified) > 0 {
		hasChanges = true
		fmt.Printf("\n%s%s📦 Nodes Modified:%s\n", colorBold, colorYellow, colorReset)
		for _, nc := range diff.NodesModified {
			fmt.Printf("  %s~ %s%s\n", colorYellow, nc.ID, colorReset)
			for _, c := range nc.Changes {
				fmt.Printf("      %s\n", c)
			}
		}
	}

	if len(diff.RelsAdded) > 0 {
		hasChanges = true
		fmt.Printf("\n%s%s🔗 Relationships Added:%s\n", colorBold, colorGreen, colorReset)
		for _, id := range diff.RelsAdded {
			fmt.Printf("  %s+ %s%s\n", colorGreen, id, colorReset)
		}
	}

	if len(diff.RelsRemoved) > 0 {
		hasChanges = true
		fmt.Printf("\n%s%s🔗 Relationships Removed:%s\n", colorBold, colorRed, colorReset)
		for _, id := range diff.RelsRemoved {
			fmt.Printf("  %s- %s%s\n", colorRed, id, colorReset)
		}
	}

	if len(diff.FlowsAdded) > 0 {
		hasChanges = true
		fmt.Printf("\n%s%s🌊 Flows Added:%s\n", colorBold, colorGreen, colorReset)
		for _, id := range diff.FlowsAdded {
			fmt.Printf("  %s+ %s%s\n", colorGreen, id, colorReset)
		}
	}

	if len(diff.FlowsRemoved) > 0 {
		hasChanges = true
		fmt.Printf("\n%s%s🌊 Flows Removed:%s\n", colorBold, colorRed, colorReset)
		for _, id := range diff.FlowsRemoved {
			fmt.Printf("  %s- %s%s\n", colorRed, id, colorReset)
		}
	}

	if len(diff.ControlsAdded) > 0 {
		hasChanges = true
		fmt.Printf("\n%s%s🛡️ Controls Added:%s\n", colorBold, colorGreen, colorReset)
		for _, id := range diff.ControlsAdded {
			fmt.Printf("  %s+ %s%s\n", colorGreen, id, colorReset)
		}
	}

	if len(diff.ControlsRemoved) > 0 {
		hasChanges = true
		fmt.Printf("\n%s%s🛡️ Controls Removed:%s\n", colorBold, colorRed, colorReset)
		for _, id := range diff.ControlsRemoved {
			fmt.Printf("  %s- %s%s\n", colorRed, id, colorReset)
		}
	}

	if !hasChanges {
		fmt.Printf("\n%s✅ No architecture changes detected.%s\n", colorCyan, colorReset)
	} else {
		fmt.Println()
	}
}

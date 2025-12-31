package parser

import (
	"bufio"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// ParseRichD2 parses a Rich D2 file and extracts CALM architecture data.
// This enables bidirectional editing: D2 → CALM Architecture → Go DSL
func ParseRichD2(content string) (*domain.Architecture, error) {
	arch := &domain.Architecture{
		Schema:   "https://calm.finos.org/release/1.1/meta/calm.json",
		Metadata: make(map[string]any),
		Controls: make(map[string]*domain.Control),
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentNode *domain.Node
	var currentRel *domain.Relationship
	var currentFlow *domain.Flow

	// Regex patterns for @calm: annotations
	calmPattern := regexp.MustCompile(`#\s*@calm:(\w+)=(.+)$`)
	flowPattern := regexp.MustCompile(`#\s*@calm:flow\s+id=(\S+)\s+name=(.+)$`)
	flowStepPattern := regexp.MustCompile(`#\s*@calm:flow-step\s+seq=(\d+)\s+rel=(\S+)\s+dir=(\S+)\s+desc=(.+)$`)
	composedPattern := regexp.MustCompile(`#\s*@calm:composed-of\s+id=(\S+)\s+container=(\S+)\s+nodes=(.+)$`)
	controlPattern := regexp.MustCompile(`#\s*@calm:control\s+id=(\S+)\s+data=(.+)$`)

	nodeStartPattern := regexp.MustCompile(`^\s*(\S+):\s*(.+?)\s*\{`)
	relPattern := regexp.MustCompile(`^\s*(\S+)\s*->\s*(\S+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Parse architecture-level @calm annotations
		if matches := calmPattern.FindStringSubmatch(line); matches != nil {
			key, value := matches[1], matches[2]

			// Check if we're inside a node or relationship block
			if currentNode != nil {
				parseNodeAnnotation(currentNode, key, value)
			} else if currentRel != nil {
				parseRelAnnotation(currentRel, key, value)
			} else if currentFlow != nil {
				parseFlowAnnotation(currentFlow, key, value)
			} else {
				// Top-level architecture annotation
				switch key {
				case "id":
					arch.UniqueID = value
				case "description":
					arch.Description = unescapeD2String(value)
				case "adrs":
					json.Unmarshal([]byte(value), &arch.ADRs)
				}
			}
			continue
		}

		// Parse flow definitions
		if matches := flowPattern.FindStringSubmatch(line); matches != nil {
			currentFlow = &domain.Flow{
				UniqueID: matches[1],
				Name:     unescapeD2String(matches[2]),
				Metadata: make(map[string]any),
			}
			arch.Flows = append(arch.Flows, currentFlow)
			continue
		}

		// Parse flow steps
		if matches := flowStepPattern.FindStringSubmatch(line); matches != nil && currentFlow != nil {
			var seq int
			json.Unmarshal([]byte(matches[1]), &seq)
			currentFlow.Transitions = append(currentFlow.Transitions, domain.Transition{
				SequenceNumber: seq,
				RelationshipID: matches[2],
				Direction:      matches[3],
				Description:    unescapeD2String(matches[4]),
			})
			continue
		}

		// Parse composed-of relationships
		if matches := composedPattern.FindStringSubmatch(line); matches != nil {
			var nodes []string
			json.Unmarshal([]byte(matches[3]), &nodes)
			rel := &domain.Relationship{
				UniqueID: matches[1],
				RelationshipType: domain.RelationshipType{
					ComposedOf: map[string]any{
						"container": matches[2],
						"nodes":     nodes,
					},
				},
			}
			arch.Relationships = append(arch.Relationships, rel)
			continue
		}

		// Parse global controls
		if matches := controlPattern.FindStringSubmatch(line); matches != nil {
			var ctrl domain.Control
			json.Unmarshal([]byte(matches[2]), &ctrl)
			arch.Controls[matches[1]] = &ctrl
			continue
		}

		// Parse node start
		if matches := nodeStartPattern.FindStringSubmatch(line); matches != nil {
			// End previous node if any
			if currentNode != nil && currentNode.UniqueID != "" {
				arch.Nodes = append(arch.Nodes, currentNode)
			}
			currentNode = &domain.Node{
				Arch:     arch,
				UniqueID: matches[1],
				Name:     matches[2],
				Metadata: make(map[string]any),
				Controls: make(map[string]*domain.Control),
			}
			currentRel = nil
			continue
		}

		// Parse relationship start
		if matches := relPattern.FindStringSubmatch(line); matches != nil {
			// End previous node if any
			if currentNode != nil && currentNode.UniqueID != "" {
				arch.Nodes = append(arch.Nodes, currentNode)
				currentNode = nil
			}
			currentRel = &domain.Relationship{
				Metadata: make(map[string]any),
				RelationshipType: domain.RelationshipType{
					Connects: &domain.Connects{
						Source:      domain.NodeInterface{Node: matches[1]},
						Destination: domain.NodeInterface{Node: matches[2]},
					},
				},
			}
			continue
		}

		// End of block
		if strings.TrimSpace(line) == "}" {
			if currentNode != nil && currentNode.UniqueID != "" {
				arch.Nodes = append(arch.Nodes, currentNode)
				currentNode = nil
			}
			if currentRel != nil && currentRel.UniqueID != "" {
				arch.Relationships = append(arch.Relationships, currentRel)
				currentRel = nil
			}
		}
	}

	// Add last node/rel if pending
	if currentNode != nil && currentNode.UniqueID != "" {
		arch.Nodes = append(arch.Nodes, currentNode)
	}
	if currentRel != nil && currentRel.UniqueID != "" {
		arch.Relationships = append(arch.Relationships, currentRel)
	}

	return arch, nil
}

// RichD2Parser implements the domain.Parser port for Rich D2 inputs.
type RichD2Parser struct{}

// Parse converts Rich D2 content into a CALM architecture model.
func (RichD2Parser) Parse(content string) (*domain.Architecture, error) {
	return ParseRichD2(content)
}

func parseNodeAnnotation(node *domain.Node, key, value string) {
	switch key {
	case "id":
		node.UniqueID = value
	case "type":
		node.NodeType = domain.NodeType(value)
	case "owner":
		node.Owner = value
	case "costCenter":
		node.CostCenter = value
	case "description":
		node.Description = unescapeD2String(value)
	case "metadata":
		json.Unmarshal([]byte(value), &node.Metadata)
	case "interfaces":
		json.Unmarshal([]byte(value), &node.Interfaces)
	case "controls":
		json.Unmarshal([]byte(value), &node.Controls)
	}
}

func parseRelAnnotation(rel *domain.Relationship, key, value string) {
	switch key {
	case "id":
		rel.UniqueID = value
	case "description":
		rel.Description = unescapeD2String(value)
	case "encrypted":
		if value == "true" {
			rel.Encrypted = domain.BoolPtr(true)
		} else {
			rel.Encrypted = domain.BoolPtr(false)
		}
	case "classification":
		rel.DataClassification = value
	case "srcInterfaces":
		var intfs []string
		json.Unmarshal([]byte(value), &intfs)
		if rel.RelationshipType.Connects != nil {
			rel.RelationshipType.Connects.Source.Interfaces = intfs
		}
	case "dstInterfaces":
		var intfs []string
		json.Unmarshal([]byte(value), &intfs)
		if rel.RelationshipType.Connects != nil {
			rel.RelationshipType.Connects.Destination.Interfaces = intfs
		}
	case "metadata":
		json.Unmarshal([]byte(value), &rel.Metadata)
	case "type":
		if value == "interacts" {
			// Convert to interacts type
			rel.RelationshipType.Connects = nil
		}
	case "actor":
		if rel.RelationshipType.Interacts == nil {
			rel.RelationshipType.Interacts = make(map[string]any)
		}
		rel.RelationshipType.Interacts["actor"] = value
	}
}

func parseFlowAnnotation(flow *domain.Flow, key, value string) {
	switch key {
	case "metadata":
		json.Unmarshal([]byte(value), &flow.Metadata)
	}
}

func unescapeD2String(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\=", "=")
	return s
}

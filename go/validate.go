package main

import (
	"fmt"
)

// ValidationRule defines a rule to check against the architecture
type ValidationRule interface {
	Name() string
	Validate(a *Architecture) []ValidationError
}

// ValidationError represents a validation failure
type ValidationError struct {
	Rule    string
	NodeID  string
	Message string
}

func (e ValidationError) String() string {
	if e.NodeID != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Rule, e.NodeID, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Rule, e.Message)
}

// Validate runs all rules against the architecture and returns errors
func (a *Architecture) Validate(rules ...ValidationRule) []ValidationError {
	var errors []ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(a)...)
	}
	return errors
}

// ValidateAndReport runs validation and prints colored output
func (a *Architecture) ValidateAndReport(rules ...ValidationRule) bool {
	errors := a.Validate(rules...)
	if len(errors) == 0 {
		fmt.Println("\033[32m✅ All validation rules passed\033[0m")
		return true
	}

	fmt.Printf("\033[31m❌ Validation failed with %d error(s):\033[0m\n", len(errors))
	for _, err := range errors {
		fmt.Printf("  \033[31m• %s\033[0m\n", err.String())
	}
	return false
}

// --- Built-in Validation Rules ---

// allNodesHaveOwner checks that every node has an owner
type allNodesHaveOwner struct{}

func AllNodesHaveOwner() ValidationRule { return allNodesHaveOwner{} }

func (r allNodesHaveOwner) Name() string { return "AllNodesHaveOwner" }

func (r allNodesHaveOwner) Validate(a *Architecture) []ValidationError {
	var errors []ValidationError
	for _, node := range a.Nodes {
		if node.Owner == "" {
			errors = append(errors, ValidationError{
				Rule:    r.Name(),
				NodeID:  node.UniqueID,
				Message: "missing owner",
			})
		}
	}
	return errors
}

// allServicesHaveHealthEndpoint checks that services have health-endpoint in metadata
type allServicesHaveHealthEndpoint struct{}

func AllServicesHaveHealthEndpoint() ValidationRule { return allServicesHaveHealthEndpoint{} }

func (r allServicesHaveHealthEndpoint) Name() string { return "AllServicesHaveHealthEndpoint" }

func (r allServicesHaveHealthEndpoint) Validate(a *Architecture) []ValidationError {
	var errors []ValidationError
	for _, node := range a.Nodes {
		if node.NodeType == Service {
			if _, ok := node.Metadata["health-endpoint"]; !ok {
				errors = append(errors, ValidationError{
					Rule:    r.Name(),
					NodeID:  node.UniqueID,
					Message: "service missing health-endpoint in metadata",
				})
			}
		}
	}
	return errors
}

// noDanglingRelationships checks that all relationship references valid nodes
type noDanglingRelationships struct{}

func NoDanglingRelationships() ValidationRule { return noDanglingRelationships{} }

func (r noDanglingRelationships) Name() string { return "NoDanglingRelationships" }

func (r noDanglingRelationships) Validate(a *Architecture) []ValidationError {
	nodeIDs := make(map[string]bool)
	for _, node := range a.Nodes {
		nodeIDs[node.UniqueID] = true
	}

	var errors []ValidationError
	for _, rel := range a.Relationships {
		rt := rel.RelationshipType

		if rt.Connects != nil {
			if !nodeIDs[rt.Connects.Source.Node] {
				errors = append(errors, ValidationError{
					Rule:    r.Name(),
					NodeID:  rel.UniqueID,
					Message: fmt.Sprintf("source node %q does not exist", rt.Connects.Source.Node),
				})
			}
			if !nodeIDs[rt.Connects.Destination.Node] {
				errors = append(errors, ValidationError{
					Rule:    r.Name(),
					NodeID:  rel.UniqueID,
					Message: fmt.Sprintf("destination node %q does not exist", rt.Connects.Destination.Node),
				})
			}
		}

		if rt.Interacts != nil {
			if actor, ok := rt.Interacts["actor"].(string); ok {
				if !nodeIDs[actor] {
					errors = append(errors, ValidationError{
						Rule:    r.Name(),
						NodeID:  rel.UniqueID,
						Message: fmt.Sprintf("actor %q does not exist", actor),
					})
				}
			}
			if nodes, ok := rt.Interacts["nodes"].([]string); ok {
				for _, n := range nodes {
					if !nodeIDs[n] {
						errors = append(errors, ValidationError{
							Rule:    r.Name(),
							NodeID:  rel.UniqueID,
							Message: fmt.Sprintf("target node %q does not exist", n),
						})
					}
				}
			}
		}

		if rt.ComposedOf != nil {
			if container, ok := rt.ComposedOf["container"].(string); ok {
				if !nodeIDs[container] {
					errors = append(errors, ValidationError{
						Rule:    r.Name(),
						NodeID:  rel.UniqueID,
						Message: fmt.Sprintf("container node %q does not exist", container),
					})
				}
			}
			if nodes, ok := rt.ComposedOf["nodes"].([]string); ok {
				for _, n := range nodes {
					if !nodeIDs[n] {
						errors = append(errors, ValidationError{
							Rule:    r.Name(),
							NodeID:  rel.UniqueID,
							Message: fmt.Sprintf("contained node %q does not exist", n),
						})
					}
				}
			}
		}
	}
	return errors
}

// allFlowsHaveValidTransitions checks flow transitions reference valid relationships
type allFlowsHaveValidTransitions struct{}

func AllFlowsHaveValidTransitions() ValidationRule { return allFlowsHaveValidTransitions{} }

func (r allFlowsHaveValidTransitions) Name() string { return "AllFlowsHaveValidTransitions" }

func (r allFlowsHaveValidTransitions) Validate(a *Architecture) []ValidationError {
	relIDs := make(map[string]bool)
	for _, rel := range a.Relationships {
		relIDs[rel.UniqueID] = true
	}

	var errors []ValidationError
	for _, flow := range a.Flows {
		for _, t := range flow.Transitions {
			if !relIDs[t.RelationshipID] {
				errors = append(errors, ValidationError{
					Rule:    r.Name(),
					NodeID:  flow.UniqueID,
					Message: fmt.Sprintf("transition references non-existent relationship %q", t.RelationshipID),
				})
			}
		}
	}
	return errors
}

// allDatabasesHaveBackupSchedule checks databases have backup-schedule metadata
type allDatabasesHaveBackupSchedule struct{}

func AllDatabasesHaveBackupSchedule() ValidationRule { return allDatabasesHaveBackupSchedule{} }

func (r allDatabasesHaveBackupSchedule) Name() string { return "AllDatabasesHaveBackupSchedule" }

func (r allDatabasesHaveBackupSchedule) Validate(a *Architecture) []ValidationError {
	var errors []ValidationError
	for _, node := range a.Nodes {
		if node.NodeType == Database {
			if _, ok := node.Metadata["backup-schedule"]; !ok {
				errors = append(errors, ValidationError{
					Rule:    r.Name(),
					NodeID:  node.UniqueID,
					Message: "database missing backup-schedule in metadata",
				})
			}
		}
	}
	return errors
}

// allTier1NodesHaveRunbook checks tier-1 nodes have runbook metadata
type allTier1NodesHaveRunbook struct{}

func AllTier1NodesHaveRunbook() ValidationRule { return allTier1NodesHaveRunbook{} }

func (r allTier1NodesHaveRunbook) Name() string { return "AllTier1NodesHaveRunbook" }

func (r allTier1NodesHaveRunbook) Validate(a *Architecture) []ValidationError {
	var errors []ValidationError
	for _, node := range a.Nodes {
		tier, _ := node.Metadata["tier"].(string)
		if tier == "tier-1" {
			if _, ok := node.Metadata["runbook"]; !ok {
				errors = append(errors, ValidationError{
					Rule:    r.Name(),
					NodeID:  node.UniqueID,
					Message: "tier-1 node missing runbook in metadata",
				})
			}
		}
	}
	return errors
}

// noUnusedNodes checks that all nodes are referenced by at least one relationship
type noUnusedNodes struct{}

func NoUnusedNodes() ValidationRule { return noUnusedNodes{} }

func (r noUnusedNodes) Name() string { return "NoUnusedNodes" }

func (r noUnusedNodes) Validate(a *Architecture) []ValidationError {
	usedNodes := make(map[string]bool)

	for _, rel := range a.Relationships {
		rt := rel.RelationshipType

		if rt.Connects != nil {
			usedNodes[rt.Connects.Source.Node] = true
			usedNodes[rt.Connects.Destination.Node] = true
		}
		if rt.Interacts != nil {
			if actor, ok := rt.Interacts["actor"].(string); ok {
				usedNodes[actor] = true
			}
			if nodes, ok := rt.Interacts["nodes"].([]string); ok {
				for _, n := range nodes {
					usedNodes[n] = true
				}
			}
		}
		if rt.ComposedOf != nil {
			if container, ok := rt.ComposedOf["container"].(string); ok {
				usedNodes[container] = true
			}
			if nodes, ok := rt.ComposedOf["nodes"].([]string); ok {
				for _, n := range nodes {
					usedNodes[n] = true
				}
			}
		}
	}

	var errors []ValidationError
	for _, node := range a.Nodes {
		if !usedNodes[node.UniqueID] {
			errors = append(errors, ValidationError{
				Rule:    r.Name(),
				NodeID:  node.UniqueID,
				Message: "node is not referenced by any relationship",
			})
		}
	}
	return errors
}

// DefaultValidationRules returns the standard set of validation rules
func DefaultValidationRules() []ValidationRule {
	return []ValidationRule{
		AllNodesHaveOwner(),
		AllServicesHaveHealthEndpoint(),
		NoDanglingRelationships(),
		AllFlowsHaveValidTransitions(),
		AllDatabasesHaveBackupSchedule(),
		AllTier1NodesHaveRunbook(),
	}
}

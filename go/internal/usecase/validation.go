package usecase

import "github.com/sokoide/advent-of-calm-2025/internal/domain"

// DefaultValidationRules returns the standard set of validation rules.
func DefaultValidationRules() []domain.ValidationRule {
	return []domain.ValidationRule{
		domain.AllNodesHaveOwner(),
		domain.AllServicesHaveHealthEndpoint(),
		domain.NoDanglingRelationships(),
		domain.AllFlowsHaveValidTransitions(),
		domain.AllDatabasesHaveBackupSchedule(),
		domain.AllTier1NodesHaveRunbook(),
	}
}

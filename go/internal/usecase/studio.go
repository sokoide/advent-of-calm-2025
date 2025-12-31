package usecase

import (
	"fmt"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// Layout is a use-case level alias for domain layouts.
type Layout = domain.ArchitectureLayout

// LayoutRepository is a use-case level alias for the domain port.
type LayoutRepository = domain.LayoutRepository

// ASTSyncer is a use-case level alias for the domain port.
type ASTSyncer = domain.ASTSyncer

// NodeAction represents a node update requested by the UI.
type NodeAction struct {
	Action   string
	NodeID   string
	NodeType string
	Name     string
	Desc     string
	Property string
	Value    string
}

// StudioService coordinates layout persistence and AST synchronization.
type StudioService struct {
	LayoutRepo LayoutRepository
	ASTSyncer  ASTSyncer
}

// NewStudioService builds a studio use case service.
func NewStudioService(layoutRepo LayoutRepository, astSyncer ASTSyncer) StudioService {
	return StudioService{LayoutRepo: layoutRepo, ASTSyncer: astSyncer}
}

// LoadLayout fetches the layout for a given architecture ID.
func (s StudioService) LoadLayout(id string) (*Layout, error) {
	return s.LayoutRepo.Load(id)
}

// SaveLayout persists the layout for a given architecture ID.
func (s StudioService) SaveLayout(id string, layout *Layout) error {
	return s.LayoutRepo.Save(id, layout)
}

// SyncFromJSON applies a JSON model to the Go DSL source.
func (s StudioService) SyncFromJSON(src, jsonStr string) (string, error) {
	return s.ASTSyncer.SyncFromJSON(src, jsonStr)
}

// ApplyNodeAction applies a node mutation to the Go DSL source.
func (s StudioService) ApplyNodeAction(src string, action NodeAction) (string, error) {
	switch action.Action {
	case "add":
		return s.ASTSyncer.AddNode(src, action.NodeID, action.NodeType, action.Name, action.Desc)
	case "update":
		return s.ASTSyncer.UpdateNodeProperty(src, action.NodeID, action.Property, action.Value)
	case "delete":
		return s.ASTSyncer.DeleteNode(src, action.NodeID)
	default:
		return "", fmt.Errorf("invalid action: %s", action.Action)
	}
}

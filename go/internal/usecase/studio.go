package usecase

import (
	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// Layout is a use-case level alias for domain layouts.
type Layout = domain.ArchitectureLayout

// LayoutRepository is a use-case level alias for the domain port.
type LayoutRepository = domain.LayoutRepository

// ASTSyncer is a use-case level alias for the domain port.
type ASTSyncer = domain.ASTSyncer

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

// ApplyPatch applies a list of patch operations to the Go DSL source.
func (s StudioService) ApplyPatch(src string, ops []domain.PatchOperation) (string, error) {
	return s.ASTSyncer.ApplyPatch(src, ops)
}

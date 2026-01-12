package usecase

import (
	"encoding/json"
	"fmt"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// CodeSyncUseCase handles the synchronization between different architecture representations.
type CodeSyncUseCase struct {
	dslRepo   DSLRepository
	renderer  DiagramRenderer
	parser    domain.Parser
	astSyncer domain.ASTSyncer
}

// NewCodeSyncUseCase creates a new CodeSyncUseCase.
func NewCodeSyncUseCase(
	dslRepo DSLRepository,
	renderer DiagramRenderer,
	parser domain.Parser,
	astSyncer domain.ASTSyncer,
) *CodeSyncUseCase {
	return &CodeSyncUseCase{
		dslRepo:   dslRepo,
		renderer:  renderer,
		parser:    parser,
		astSyncer: astSyncer,
	}
}

// ReadDSL reads the current Go DSL content.
func (u *CodeSyncUseCase) ReadDSL() (string, error) {
	return u.dslRepo.Read()
}

// WriteDSL writes new content to the Go DSL file.
func (u *CodeSyncUseCase) WriteDSL(content string) error {
	return u.dslRepo.Write(content)
}

// SyncD2ToGo parses the D2 code and updates the Go DSL file via AST manipulation.
func (u *CodeSyncUseCase) SyncD2ToGo(d2Code string) ([]string, error) {
	// 1. Parse D2 to Architecture Model
	arch, err := u.parser.Parse(d2Code)
	if err != nil {
		return nil, fmt.Errorf("failed to parse D2: %w", err)
	}

	// 2. Convert Architecture to JSON (intermediate format for ASTSyncer)
	// Note: ASTSyncer currently expects a JSON string to update the AST.
	// We construct a minimal JSON object with the nodes to be synced.
	type nodeUpdate struct {
		ID   string `json:"unique-id"`
		Name string `json:"name"`
	}
	type syncPayload struct {
		Nodes []nodeUpdate `json:"nodes"`
	}

	payload := syncPayload{}
	for _, n := range arch.Nodes {
		payload.Nodes = append(payload.Nodes, nodeUpdate{
			ID:   n.UniqueID,
			Name: n.Name,
		})
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sync payload: %w", err)
	}

	// 3. Read current Go Code
	goCode, err := u.dslRepo.Read()
	if err != nil {
		return nil, err
	}

	// 4. Apply changes via AST Syncer
	// Note: SyncFromJSON returns the *new* source code.
	newGoCode, err := u.astSyncer.SyncFromJSON(goCode, string(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("ast sync failed: %w", err)
	}

	// 5. Write back if changed
	if newGoCode != goCode {
		if err := u.dslRepo.Write(newGoCode); err != nil {
			return nil, err
		}
		return []string{"Synced D2 changes to Go AST"}, nil
	}

	return nil, nil
}

// GenerateSVG renders D2 source into SVG.
func (u *CodeSyncUseCase) GenerateSVG(d2Source string) (string, error) {
	return u.renderer.RenderToSVG(d2Source)
}

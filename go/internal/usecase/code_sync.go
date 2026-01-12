package usecase

import (
	"encoding/json"
	"fmt"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// NodeUpdate represents a single node property update for AST synchronization.
type NodeUpdate struct {
	ID   string `json:"unique-id"`
	Name string `json:"name"`
}

// SyncPayload is the intermediate data structure used to sync model changes back to Go DSL.
type SyncPayload struct {
	Nodes []NodeUpdate `json:"nodes"`
}

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
	// 1. Parse D2 and prepare the synchronization payload
	jsonPayload, err := u.prepareSyncPayload(d2Code)
	if err != nil {
		return nil, err
	}

	// 2. Read current Go Code
	goCode, err := u.dslRepo.Read()
	if err != nil {
		return nil, err
	}

	// 3. Apply changes via AST Syncer
	newGoCode, err := u.executeASTSync(goCode, jsonPayload)
	if err != nil {
		return nil, err
	}

	// 4. Write back if changed
	if newGoCode != goCode {
		if err := u.dslRepo.Write(newGoCode); err != nil {
			return nil, err
		}
		return []string{"Synced D2 changes to Go AST"}, nil
	}

	return nil, nil
}

// prepareSyncPayload parses the D2 code and constructs a JSON payload for the AST syncer.
func (u *CodeSyncUseCase) prepareSyncPayload(d2Code string) ([]byte, error) {
	arch, err := u.parser.Parse(d2Code)
	if err != nil {
		return nil, fmt.Errorf("failed to parse D2: %w", err)
	}

	payload := SyncPayload{}
	for _, n := range arch.Nodes {
		payload.Nodes = append(payload.Nodes, NodeUpdate{
			ID:   n.UniqueID,
			Name: n.Name,
		})
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sync payload: %w", err)
	}

	return jsonBytes, nil
}

// executeASTSync performs the actual AST synchronization.
func (u *CodeSyncUseCase) executeASTSync(goCode string, jsonPayload []byte) (string, error) {
	newGoCode, err := u.astSyncer.SyncFromJSON(goCode, string(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("ast sync failed: %w", err)
	}
	return newGoCode, nil
}

// GenerateSVG renders D2 source into SVG.
func (u *CodeSyncUseCase) GenerateSVG(d2Source string) (string, error) {
	return u.renderer.RenderToSVG(d2Source)
}

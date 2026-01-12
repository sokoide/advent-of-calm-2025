package usecase

import (
	"encoding/json"
	"testing"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// MockDSLRepository is a mock implementation of DSLRepository.
type MockDSLRepository struct {
	content string
	err     error
	written string
}

func (m *MockDSLRepository) Read() (string, error) {
	return m.content, m.err
}

func (m *MockDSLRepository) Write(content string) error {
	m.written = content
	return m.err
}

// MockDiagramRenderer is a mock implementation of DiagramRenderer.
type MockDiagramRenderer struct {
	svg string
	err error
}

func (m *MockDiagramRenderer) RenderToSVG(d2Code string) (string, error) {
	if d2Code == "error" {
		return "", m.err
	}
	return m.svg, m.err
}

// MockParser is a mock implementation of domain.Parser.
type MockParser struct {
	arch *domain.Architecture
	err  error
}

func (m *MockParser) Parse(string) (*domain.Architecture, error) {
	return m.arch, m.err
}

// MockASTSyncer is a mock implementation of domain.ASTSyncer.
type MockASTSyncer struct {
	result string
	err    error
}

func (m *MockASTSyncer) SyncFromJSON(src, jsonStr string) (string, error) {
	return m.result, m.err
}

func (m *MockASTSyncer) ApplyPatch(src string, ops []domain.PatchOperation) (string, error) {
	return m.result, m.err
}

func TestCodeSyncUseCase_SyncD2ToGo(t *testing.T) {
	initialGoCode := `package usecase
func DefineArchitecture() {
	DefineNode("node1", domain.Service, "Old Label")
}`
	d2CodeWithChange := `node1: New Label {
  @calm:id=node1
}`
	expectedGoCode := `package usecase
func DefineArchitecture() {
	DefineNode("node1", domain.Service, "New Label")
}`

	mockRepo := &MockDSLRepository{content: initialGoCode}
	mockRenderer := &MockDiagramRenderer{}
	mockParser := &MockParser{
		arch: &domain.Architecture{
			Nodes: []*domain.Node{{UniqueID: "node1", Name: "New Label"}},
		},
	}
	mockSyncer := &MockASTSyncer{result: expectedGoCode}

	uc := NewCodeSyncUseCase(mockRepo, mockRenderer, mockParser, mockSyncer)

	changes, err := uc.SyncD2ToGo(d2CodeWithChange)
	if err != nil {
		t.Fatalf("SyncD2ToGo failed: %v", err)
	}

	if len(changes) == 0 {
		t.Errorf("expected changes, got none")
	}

	if mockRepo.written != expectedGoCode {
		t.Errorf("expected Go code:\n%s\n\ngot:\n%s", expectedGoCode, mockRepo.written)
	}
}

func TestCodeSyncUseCase_GenerateSVG(t *testing.T) {
	expectedSVG := "<svg>test</svg>"
	mockRepo := &MockDSLRepository{}
	mockRenderer := &MockDiagramRenderer{svg: expectedSVG}
	mockParser := &MockParser{}
	mockSyncer := &MockASTSyncer{}
	uc := NewCodeSyncUseCase(mockRepo, mockRenderer, mockParser, mockSyncer)

	got, err := uc.GenerateSVG("some d2 code")
	if err != nil {
		t.Fatalf("GenerateSVG failed: %v", err)
	}

	if got != expectedSVG {
		t.Errorf("expected SVG %q, got %q", expectedSVG, got)
	}
}

func TestCodeSyncUseCase_PrepareSyncPayload(t *testing.T) {
	mockParser := &MockParser{
		arch: &domain.Architecture{
			Nodes: []*domain.Node{{UniqueID: "node1", Name: "New Label"}},
		},
	}
	uc := &CodeSyncUseCase{parser: mockParser}

	got, err := uc.prepareSyncPayload("some d2 code")
	if err != nil {
		t.Fatalf("prepareSyncPayload failed: %v", err)
	}

	var payload SyncPayload
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(payload.Nodes) != 1 || payload.Nodes[0].ID != "node1" || payload.Nodes[0].Name != "New Label" {
		t.Errorf("unexpected payload: %+v", payload)
	}
}

func TestCodeSyncUseCase_ExecuteASTSync(t *testing.T) {
	mockSyncer := &MockASTSyncer{result: "new code"}
	uc := &CodeSyncUseCase{astSyncer: mockSyncer}

	got, err := uc.executeASTSync("old code", []byte(`{"nodes":[]}`))
	if err != nil {
		t.Fatalf("executeASTSync failed: %v", err)
	}

	if got != "new code" {
		t.Errorf("expected \"new code\", got %q", got)
	}
}

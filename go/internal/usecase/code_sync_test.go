package usecase

import (
	"reflect"
	"testing"
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

func TestCodeSyncUseCase_SyncD2ToGo(t *testing.T) {
	initialGoCode := `package usecase
func DefineArchitecture() {
	DefineNode("node1", Service, "Old Label")
}`
	d2CodeWithChange := `node1: New Label {
  @calm:id=node1
}`
	expectedGoCode := `package usecase
func DefineArchitecture() {
	DefineNode("node1", Service, "New Label")
}`

	mockRepo := &MockDSLRepository{content: initialGoCode}
	mockRenderer := &MockDiagramRenderer{}
	uc := NewCodeSyncUseCase(mockRepo, mockRenderer)

	changes, err := uc.SyncD2ToGo(d2CodeWithChange)
	if err != nil {
		t.Fatalf("SyncD2ToGo failed: %v", err)
	}

	expectedChanges := []string{"node1: \"Old Label\" → \"New Label\""}
	if !reflect.DeepEqual(changes, expectedChanges) {
		t.Errorf("expected changes %v, got %v", expectedChanges, changes)
	}

	if mockRepo.written != expectedGoCode {
		t.Errorf("expected Go code:\n%s\n\ngot:\n%s", expectedGoCode, mockRepo.written)
	}
}

func TestCodeSyncUseCase_GenerateSVG(t *testing.T) {
	expectedSVG := "<svg>test</svg>"
	mockRepo := &MockDSLRepository{}
	mockRenderer := &MockDiagramRenderer{svg: expectedSVG}
	uc := NewCodeSyncUseCase(mockRepo, mockRenderer)

	got, err := uc.GenerateSVG("some d2 code")
	if err != nil {
		t.Fatalf("GenerateSVG failed: %v", err)
	}

	if got != expectedSVG {
		t.Errorf("expected SVG %q, got %q", expectedSVG, got)
	}
}

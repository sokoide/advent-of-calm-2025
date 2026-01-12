package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sokoide/advent-of-calm-2025/cmd/studio/handlers"
	"github.com/sokoide/advent-of-calm-2025/internal/infra/filesystem"
	"github.com/sokoide/advent-of-calm-2025/internal/usecase"
)

func TestRegenerate_ReadsGoDSL(t *testing.T) {
	// Setup temp directory to simulate project structure
	tmpDir, err := os.MkdirTemp("", "studio-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create dummy ecommerce_architecture.go
	dslPath := filepath.Join(tmpDir, "internal/usecase")
	if err := os.MkdirAll(dslPath, 0755); err != nil {
		t.Fatal(err)
	}

	expectedCode := "package usecase\nfunc Build() {}"
	if err := os.WriteFile(filepath.Join(dslPath, "ecommerce_architecture.go"), []byte(expectedCode), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a test State with necessary dependencies
	dslRepo := filesystem.NewFileSystemDSLRepository(
		filepath.Join(tmpDir, "internal/usecase/ecommerce_architecture.go"),
	)
	syncUseCase := usecase.NewCodeSyncUseCase(dslRepo, nil, nil, nil)
	testState := handlers.NewState(tmpDir, usecase.StudioService{}, syncUseCase)

	// Test file reading
	data, err := testState.SyncUseCase.ReadDSL()
	if err != nil {
		t.Errorf("expected to read Go DSL, got error: %v", err)
	}
	if data != expectedCode {
		t.Errorf("expected %s, got %s", expectedCode, data)
	}
}

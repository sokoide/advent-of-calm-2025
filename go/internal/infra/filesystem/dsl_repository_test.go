package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileSystemDSLRepository(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dsl-repo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "test.go")
	repo := NewFileSystemDSLRepository(path)

	content := "package main\n\nfunc main() {}"

	// Test Write
	if err := repo.Write(content); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Test Read
	got, err := repo.Read()
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	if got != content {
		t.Errorf("expected %q, got %q", content, got)
	}
}

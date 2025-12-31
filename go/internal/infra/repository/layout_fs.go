package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

type FSLayoutRepository struct {
	baseDir string
}

func NewFSLayoutRepository(baseDir string) *FSLayoutRepository {
	return &FSLayoutRepository{baseDir: baseDir}
}

func (r *FSLayoutRepository) Load(id string) (*domain.ArchitectureLayout, error) {
	path := r.getPath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.NewArchitectureLayout(), nil
		}
		return nil, err
	}

	var layout domain.ArchitectureLayout
	if err := json.Unmarshal(data, &layout); err != nil {
		return nil, err
	}

	return &layout, nil
}

func (r *FSLayoutRepository) Save(id string, layout *domain.ArchitectureLayout) error {
	path := r.getPath(id)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(layout, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (r *FSLayoutRepository) getPath(id string) string {
	return filepath.Join(r.baseDir, "layout", fmt.Sprintf("%s.layout.json", id))
}

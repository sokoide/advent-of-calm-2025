package filesystem

import (
	"os"
)

// FileSystemDSLRepository implements usecase.DSLRepository for the local file system.
type FileSystemDSLRepository struct {
	path string
}

// NewFileSystemDSLRepository creates a new FileSystemDSLRepository.
func NewFileSystemDSLRepository(path string) *FileSystemDSLRepository {
	return &FileSystemDSLRepository{
		path: path,
	}
}

// Read reads the DSL file content.
func (r *FileSystemDSLRepository) Read() (string, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Write writes the content to the DSL file.
func (r *FileSystemDSLRepository) Write(content string) error {
	return os.WriteFile(r.path, []byte(content), 0644)
}

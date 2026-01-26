package domain

import "fmt"

// PatchError represents an error that occurs during patch operations.
type PatchError struct {
	Operation PatchType         // The type of patch operation that failed
	Message   string            // Human-readable error message
	Details   map[string]string // Additional context about the error
}

// Error returns the error message.
func (e *PatchError) Error() string {
	if e.Details != nil && len(e.Details) > 0 {
		return fmt.Sprintf("%s: %s (details: %v)", e.Operation, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Operation, e.Message)
}

// NewPatchError creates a new PatchError.
func NewPatchError(operation PatchType, message string, details map[string]string) *PatchError {
	return &PatchError{
		Operation: operation,
		Message:   message,
		Details:   details,
	}
}

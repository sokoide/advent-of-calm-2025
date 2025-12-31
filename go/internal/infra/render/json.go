package render

import (
	"encoding/json"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// JSONRenderer renders CALM architectures into JSON.
type JSONRenderer struct{}

// Render marshals the architecture as indented JSON.
func (JSONRenderer) Render(a *domain.Architecture) (string, error) {
	out, err := json.MarshalIndent(a, "", "    ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

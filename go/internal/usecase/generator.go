package usecase

import (
	"fmt"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// Renderer is a use-case level alias for the domain renderer port.
type Renderer = domain.Renderer

// ValidationError is a use-case level alias for domain validation errors.
type ValidationError = domain.ValidationError

// OutputFormat defines supported renderer selections.
type OutputFormat string

const (
	FormatJSON   OutputFormat = "json"
	FormatD2     OutputFormat = "d2"
	FormatRichD2 OutputFormat = "rich-d2"
)

// Builder constructs an architecture model.
type Builder interface {
	Build() *domain.Architecture
}

// Validator evaluates an architecture against rule sets.
type Validator interface {
	Validate(*domain.Architecture) []domain.ValidationError
}

// RuleValidator validates using a fixed rule set.
type RuleValidator struct {
	Rules []domain.ValidationRule
}

// Validate runs the configured rules.
func (v RuleValidator) Validate(a *domain.Architecture) []domain.ValidationError {
	return a.Validate(v.Rules...)
}

// Generator orchestrates building, validating, and rendering architectures.
type Generator struct {
	Builder       Builder
	Renderers     map[OutputFormat]Renderer
	Validator     Validator
	DefaultFormat OutputFormat
}

// Generate builds the architecture and returns a rendered output.
func (g Generator) Generate(format OutputFormat, validate bool) (string, []ValidationError, error) {
	if g.Builder == nil {
		return "", nil, fmt.Errorf("builder is required")
	}

	arch := g.Builder.Build()

	if validate && g.Validator != nil {
		validationErrors := g.Validator.Validate(arch)
		if len(validationErrors) > 0 {
			return "", validationErrors, nil
		}
	}

	renderer := g.Renderers[format]
	if renderer == nil {
		renderer = g.Renderers[g.DefaultFormat]
	}
	if renderer == nil {
		return "", nil, fmt.Errorf("renderer not configured for %s", format)
	}

	output, err := renderer.Render(arch)
	if err != nil {
		return "", nil, err
	}

	return output, nil, nil
}

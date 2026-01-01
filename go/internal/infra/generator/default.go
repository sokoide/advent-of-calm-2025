package generator

import (
	"github.com/sokoide/advent-of-calm-2025/internal/infra/render"
	"github.com/sokoide/advent-of-calm-2025/internal/usecase"
)

// DefaultGenerator returns the standard CALM generator setup shared by CLI and Studio.
func DefaultGenerator() usecase.Generator {
	return usecase.Generator{
		Builder: usecase.EcommerceBuilder{},
		Renderers: map[usecase.OutputFormat]usecase.Renderer{
			usecase.FormatJSON:   render.JSONRenderer{},
			usecase.FormatD2:     render.D2Renderer{},
			usecase.FormatRichD2: render.RichD2Renderer{},
		},
		Validator:     usecase.RuleValidator{Rules: usecase.DefaultValidationRules()},
		DefaultFormat: usecase.FormatJSON,
	}
}

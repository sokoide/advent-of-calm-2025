package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sokoide/advent-of-calm-2025/internal/infra/generator"
	"github.com/sokoide/advent-of-calm-2025/internal/usecase"
)

const (
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

func main() {
	outputFormat := flag.String("format", "json", "Output format: json, d2, rich-d2")
	runValidation := flag.Bool("validate", false, "Run validation rules")
	flag.Parse()

	gen := generator.DefaultGenerator()

	output, validationErrors, err := gen.Generate(usecase.OutputFormat(*outputFormat), *runValidation)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *runValidation {
		if len(validationErrors) > 0 {
			printValidationErrors(validationErrors)
			os.Exit(1)
		}
		fmt.Printf("%s✅ All validation rules passed%s\n", colorGreen, colorReset)
		return
	}

	fmt.Println(output)
}

func printValidationErrors(errors []usecase.ValidationError) {
	fmt.Printf("%s❌ Validation failed with %d error(s):%s\n", colorRed, len(errors), colorReset)
	for _, err := range errors {
		fmt.Printf("  %s• %s%s\n", colorRed, err.String(), colorReset)
	}
}

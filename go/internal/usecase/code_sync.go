package usecase

import (
	"fmt"
	"regexp"
	"strings"
)

// CodeSyncUseCase handles the synchronization between different architecture representations.
type CodeSyncUseCase struct {
	dslRepo  DSLRepository
	renderer DiagramRenderer
}

// NewCodeSyncUseCase creates a new CodeSyncUseCase.
func NewCodeSyncUseCase(dslRepo DSLRepository, renderer DiagramRenderer) *CodeSyncUseCase {
	return &CodeSyncUseCase{
		dslRepo:  dslRepo,
		renderer: renderer,
	}
}

// ReadDSL reads the current Go DSL content.
func (u *CodeSyncUseCase) ReadDSL() (string, error) {
	return u.dslRepo.Read()
}

// WriteDSL writes new content to the Go DSL file.
func (u *CodeSyncUseCase) WriteDSL(content string) error {
	return u.dslRepo.Write(content)
}

// SyncD2ToGo parses the D2 code, finds label changes, and updates the DSL file via the repository.
func (u *CodeSyncUseCase) SyncD2ToGo(d2Code string) ([]string, error) {
	type nodeInfo struct {
		calmID string
		label  string
	}
	var nodes []nodeInfo

	lines := strings.Split(d2Code, "\n")
	var currentLabel string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, ": ") && strings.HasSuffix(trimmed, "{") {
			parts := strings.SplitN(trimmed, ": ", 2)
			if len(parts) == 2 {
				currentLabel = strings.TrimSuffix(strings.TrimSpace(parts[1]), " {")
			}
		}

		if strings.Contains(trimmed, "@calm:id=") {
			parts := strings.SplitN(trimmed, "@calm:id=", 2)
			if len(parts) == 2 {
				calmID := strings.TrimSpace(parts[1])
				if currentLabel != "" {
					nodes = append(nodes, nodeInfo{calmID: calmID, label: currentLabel})
				}
			}
		}
	}

	if len(nodes) == 0 {
		return nil, nil
	}

	goCode, err := u.dslRepo.Read()
	if err != nil {
		return nil, err
	}

	var changes []string

	for _, n := range nodes {
		pattern := fmt.Sprintf(`DefineNode\(\s*"%s"\s*,\s*[\w\.]+\s*,\s*"([^"]+)"`, regexp.QuoteMeta(n.calmID))
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(goCode)

		if len(matches) > 1 {
			oldLabel := matches[1]
			if oldLabel != n.label {
				fullMatch := matches[0]
				newMatch := strings.Replace(fullMatch, `"`+oldLabel+`"`, `"`+n.label+`"`, 1)
				goCode = strings.Replace(goCode, fullMatch, newMatch, 1)
				changes = append(changes, fmt.Sprintf("%s: %q → %q", n.calmID, oldLabel, n.label))
			}
		}
	}

	if len(changes) > 0 {
		if err := u.dslRepo.Write(goCode); err != nil {
			return nil, err
		}
	}

	return changes, nil
}

// GenerateSVG renders D2 source into SVG.
func (u *CodeSyncUseCase) GenerateSVG(d2Source string) (string, error) {
	return u.renderer.RenderToSVG(d2Source)
}

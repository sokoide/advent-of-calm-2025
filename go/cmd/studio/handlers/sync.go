package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// HandlePatch handles AST patching operations.
func (s *State) HandlePatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var ops []domain.PatchOperation
	if err := json.NewDecoder(r.Body).Decode(&ops); err != nil {
		log.Printf("❌ Patch decode error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("📥 Patch request: %d operations", len(ops))
	for i, op := range ops {
		log.Printf("  [%d] type=%s loopVar=%s line=%d", i, op.Type, op.Origin.LoopVar, op.Origin.Line)
	}

	src, err := os.ReadFile(s.DSLPath())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newCode, err := s.StudioSvc.ApplyPatch(string(src), ops)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(s.DSLPath(), []byte(newCode), 0644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Patch Applied: %d operations", len(ops))
	w.WriteHeader(http.StatusOK)
}

// HandlePreviewJSONSync previews changes from CALM JSON to Go DSL.
func (s *State) HandlePreviewJSONSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		JSON string `json:"json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	src, err := os.ReadFile(s.DSLPath())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newCode, err := s.StudioSvc.SyncFromJSON(string(src), req.JSON)
	if err != nil {
		log.Printf("❌ Sync Error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"newCode": newCode})
}

// HandleD2ToGo converts D2 source back to Go DSL code.
func (s *State) HandleD2ToGo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		D2Code string `json:"d2Code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	changes, err := s.applyD2ChangesToGo(req.D2Code)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"changes": changes,
		"success": len(changes) > 0,
	})
}

// applyD2ChangesToGo parses D2, finds label changes, and updates the DSL file.
func (s *State) applyD2ChangesToGo(d2Code string) ([]string, error) {
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

	content, err := os.ReadFile(s.DSLPath())
	if err != nil {
		return nil, err
	}

	goCode := string(content)
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
		if err := os.WriteFile(s.DSLPath(), []byte(goCode), 0644); err != nil {
			return nil, err
		}
	}

	return changes, nil
}

// WriteGoDSL writes Go code to the DSL file.
func (s *State) WriteGoDSL(content string) error {
	return os.WriteFile(s.DSLPath(), []byte(content), 0644)
}

// ReadGoDSL reads the Go DSL file content.
func (s *State) ReadGoDSL() (string, error) {
	data, err := os.ReadFile(s.DSLPath())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

package handlers

import (
	"encoding/json"
	"log"
	"net/http"

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

	src, err := s.SyncUseCase.ReadDSL()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newCode, err := s.StudioSvc.ApplyPatch(src, ops)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.SyncUseCase.WriteDSL(newCode); err != nil {
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

	src, err := s.SyncUseCase.ReadDSL()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newCode, err := s.StudioSvc.SyncFromJSON(src, req.JSON)
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

	changes, err := s.SyncUseCase.SyncD2ToGo(req.D2Code)
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

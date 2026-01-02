package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/sokoide/advent-of-calm-2025/internal/usecase"
)

// HandleLayout handles layout load and save.
func (s *State) HandleLayout(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing architecture id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		layout, err := s.StudioSvc.LoadLayout(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(layout)

	case http.MethodPost:
		var layout usecase.Layout
		if err := json.NewDecoder(r.Body).Decode(&layout); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.StudioSvc.SaveLayout(id, &layout); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

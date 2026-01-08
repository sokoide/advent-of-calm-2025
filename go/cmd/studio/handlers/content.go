package handlers

import (
	"encoding/json"
	"net/http"
)

// ServeContent returns the cached content as JSON.
func (s *State) ServeContent(w http.ResponseWriter, r *http.Request) {
	s.ContentMu.RLock()
	defer s.ContentMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.LastContent)
}

// ServeSVG returns only the SVG content.
func (s *State) ServeSVG(w http.ResponseWriter, r *http.Request) {
	s.ContentMu.RLock()
	d2Code := s.LastContent.D2Code
	svg := s.LastContent.SVG
	s.ContentMu.RUnlock()

	if svg == "" && d2Code != "" {
		svg, _ = s.SyncUseCase.GenerateSVG(d2Code)
		if svg != "" {
			s.ContentMu.Lock()
			s.LastContent.SVG = svg
			s.ContentMu.Unlock()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"svg": svg})
}

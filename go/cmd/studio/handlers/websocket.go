package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// HandleWebSocket handles WebSocket connections for real-time updates.
func (s *State) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			s.clientsMu.Lock()
			delete(s.clients, conn)
			s.clientsMu.Unlock()
			conn.Close()
			break
		}

		var update struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		}
		if json.Unmarshal(msg, &update) == nil {
			s.HandleClientUpdate(update.Type, update.Content)
		}
	}
}

// HandleUpdate handles HTTP POST updates.
func (s *State) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var update struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.HandleClientUpdate(update.Type, update.Content)
	w.WriteHeader(http.StatusOK)
}

// HandleClientUpdate processes client updates (Go or D2 code changes).
func (s *State) HandleClientUpdate(updateType, content string) {
	switch updateType {
	case "go":
		if err := s.SyncUseCase.WriteDSL(content); err != nil {
			log.Printf("Error writing Go file: %v", err)
		}
	case "d2":
		log.Println("📝 D2 update received, generating SVG...")
		s.ContentMu.Lock()
		s.LastContent.D2Code = content
		svg, err := s.SyncUseCase.GenerateSVG(content)
		if err == nil && svg != "" {
			s.LastContent.SVG = svg
			log.Println("✅ SVG generated successfully")
		} else if err != nil {
			log.Printf("❌ SVG generation error: %v", err)
		}
		s.ContentMu.Unlock()
		s.NotifyClients("refresh-svg")
	}
}

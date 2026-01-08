package handlers

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sokoide/advent-of-calm-2025/internal/usecase"
)

// State holds shared state for all handlers.
type State struct {
	GoDir       string
	StudioSvc   usecase.StudioService
	SyncUseCase *usecase.CodeSyncUseCase

	clients   map[*websocket.Conn]bool
	clientsMu sync.Mutex
	upgrader  websocket.Upgrader

	LastContent struct {
		GoCode string `json:"goCode"`
		D2Code string `json:"d2Code"`
		SVG    string `json:"svg"`
		JSON   string `json:"json"`
	}
	ContentMu sync.RWMutex
}

// NewState creates a new handler state.
func NewState(goDir string, studioSvc usecase.StudioService, syncUseCase *usecase.CodeSyncUseCase) *State {
	return &State{
		GoDir:       goDir,
		StudioSvc:   studioSvc,
		SyncUseCase: syncUseCase,
		clients:     make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// NotifyClients sends a message to all connected WebSocket clients.
func (s *State) NotifyClients(msg string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for client := range s.clients {
		if err := client.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			client.Close()
			delete(s.clients, client)
		}
	}
}

// SetContent updates all cached content fields.
func (s *State) SetContent(goCode, d2Code, svg, jsonStr string) {
	s.ContentMu.Lock()
	defer s.ContentMu.Unlock()
	s.LastContent.GoCode = goCode
	s.LastContent.D2Code = d2Code
	s.LastContent.SVG = svg
	s.LastContent.JSON = jsonStr
}

// UpdateGoCode updates only the Go code, preserving other fields.
func (s *State) UpdateGoCode(goCode string) {
	s.ContentMu.Lock()
	defer s.ContentMu.Unlock()
	s.LastContent.GoCode = goCode
}

// UpdateD2Content updates D2, SVG, and JSON, preserving GoCode.
func (s *State) UpdateD2Content(d2Code, svg, jsonStr string) {
	s.ContentMu.Lock()
	defer s.ContentMu.Unlock()
	s.LastContent.D2Code = d2Code
	s.LastContent.SVG = svg
	s.LastContent.JSON = jsonStr
}

// GetContent returns the cached content.
func (s *State) GetContent() (goCode, d2Code, svg, jsonStr string) {
	s.ContentMu.RLock()
	defer s.ContentMu.RUnlock()
	return s.LastContent.GoCode, s.LastContent.D2Code, s.LastContent.SVG, s.LastContent.JSON
}

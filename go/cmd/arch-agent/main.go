package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sokoide/advent-of-calm-2025/internal/infra/ast"
	"github.com/sokoide/advent-of-calm-2025/internal/infra/generator"
	"github.com/sokoide/advent-of-calm-2025/internal/infra/repository"
	"github.com/sokoide/advent-of-calm-2025/internal/usecase"
)

const dslRelativePath = "internal/usecase/ecommerce_architecture.go"

type contentSnapshot struct {
	GoCode string `json:"goCode"`
	D2Code string `json:"d2Code"`
	SVG    string `json:"svg"`
	JSON   string `json:"json"`
}

type server struct {
	goDir        string
	generateMode string
	studioSvc    usecase.StudioService
	contentMu    sync.RWMutex
	lastContent  contentSnapshot
}

func main() {
	port := flag.String("port", "8787", "Local agent port")
	dir := flag.String("dir", ".", "Repository root directory")
	mode := flag.String("mode", "gorun", "Generation mode: gorun or in-process")
	flag.Parse()

	goDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal(err)
	}

	layoutRepo := repository.NewFSLayoutRepository(filepath.Join(goDir, "architectures"))
	srv := &server{
		goDir:        goDir,
		generateMode: *mode,
		studioSvc:    usecase.NewStudioService(layoutRepo, ast.GoASTSyncer{}),
	}

	http.HandleFunc("/version", withCORS(srv.handleVersion))
	http.HandleFunc("/content", withCORS(srv.handleContent))
	http.HandleFunc("/svg", withCORS(srv.handleSVG))
	http.HandleFunc("/update", withCORS(srv.handleUpdate))
	http.HandleFunc("/sync-ast", withCORS(srv.handleASTSync))
	http.HandleFunc("/preview-json-sync", withCORS(srv.handlePreviewJSONSync))
	http.HandleFunc("/layout", withCORS(srv.handleLayout))

	addr := fmt.Sprintf("127.0.0.1:%s", *port)
	log.Printf("🧭 Arch Agent listening on http://%s (dir=%s, mode=%s)", addr, goDir, *mode)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func (s *server) handleVersion(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET /version from %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"name": "arch-agent",
		"mode": s.generateMode,
	})
}

func (s *server) handleContent(w http.ResponseWriter, r *http.Request) {
	includeSVG := r.URL.Query().Get("svg") == "1"
	log.Printf("GET /content?svg=%t from %s", includeSVG, r.RemoteAddr)
	snapshot, err := s.getContent(includeSVG)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

func (s *server) handleSVG(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET /svg from %s", r.RemoteAddr)
	snapshot, err := s.getContent(true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"svg": snapshot.SVG})
}

func (s *server) getContent(includeSVG bool) (contentSnapshot, error) {
	mainPath := filepath.Join(s.goDir, dslRelativePath)
	data, err := os.ReadFile(mainPath)
	if err != nil {
		return contentSnapshot{}, err
	}
	goCode := string(data)

	s.contentMu.Lock()
	defer s.contentMu.Unlock()

	changed := goCode != s.lastContent.GoCode
	if !changed {
		if includeSVG && s.lastContent.SVG == "" && s.lastContent.D2Code != "" {
			s.lastContent.SVG = generateSVGFromD2(s.lastContent.D2Code)
		}
		return s.lastContent, nil
	}

	jsonOut, d2Out, err := s.generateOutputs()
	if err != nil {
		return contentSnapshot{}, err
	}

	svg := ""
	if includeSVG {
		svg = generateSVGFromD2(d2Out)
	}

	s.lastContent = contentSnapshot{
		GoCode: goCode,
		D2Code: d2Out,
		SVG:    svg,
		JSON:   jsonOut,
	}

	return s.lastContent, nil
}

func (s *server) generateOutputs() (string, string, error) {
	if s.generateMode == "in-process" {
		gen := generator.DefaultGenerator()
		jsonOut, _, err := gen.Generate(usecase.FormatJSON, false)
		if err != nil {
			return "", "", err
		}

		d2Out, _, err := gen.Generate(usecase.FormatRichD2, false)
		if err != nil {
			return jsonOut, "", err
		}

		return jsonOut, d2Out, nil
	}

	cmdJSON := exec.Command("go", "run", "./cmd/arch-gen")
	cmdJSON.Dir = s.goDir
	var jsonOut bytes.Buffer
	cmdJSON.Stdout = &jsonOut
	cmdJSON.Stderr = &jsonOut
	if err := cmdJSON.Run(); err != nil {
		return "", "", fmt.Errorf("go run json failed: %w: %s", err, jsonOut.String())
	}

	cmdD2 := exec.Command("go", "run", "./cmd/arch-gen", "-format", "rich-d2")
	cmdD2.Dir = s.goDir
	var d2Out bytes.Buffer
	cmdD2.Stdout = &d2Out
	cmdD2.Stderr = &d2Out
	if err := cmdD2.Run(); err != nil {
		return jsonOut.String(), "", fmt.Errorf("go run rich-d2 failed: %w: %s", err, d2Out.String())
	}

	return jsonOut.String(), d2Out.String(), nil
}

func generateSVGFromD2(d2Source string) string {
	if strings.TrimSpace(d2Source) == "" {
		return ""
	}

	cmd := exec.Command("d2", "-", "-")
	cmd.Stdin = strings.NewReader(d2Source)
	var svgOut, svgErr bytes.Buffer
	cmd.Stdout = &svgOut
	cmd.Stderr = &svgErr
	if err := cmd.Run(); err != nil {
		log.Printf("❌ D2 SVG error: %v\n%s", err, svgErr.String())
		return ""
	}

	return svgOut.String()
}

func (s *server) handleUpdate(w http.ResponseWriter, r *http.Request) {
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

	if update.Type != "go" {
		http.Error(w, "Unsupported update type", http.StatusBadRequest)
		return
	}

	log.Printf("POST /update (go) bytes=%d from %s", len(update.Content), r.RemoteAddr)
	mainPath := filepath.Join(s.goDir, dslRelativePath)
	if err := os.WriteFile(mainPath, []byte(update.Content), 0644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) handlePreviewJSONSync(w http.ResponseWriter, r *http.Request) {
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

	log.Printf("POST /preview-json-sync bytes=%d from %s", len(req.JSON), r.RemoteAddr)
	mainPath := filepath.Join(s.goDir, dslRelativePath)
	src, err := os.ReadFile(mainPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newCode, err := s.studioSvc.SyncFromJSON(string(src), req.JSON)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"newCode": newCode})
}

func (s *server) handleASTSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Action   string `json:"action"`
		NodeID   string `json:"nodeId"`
		NodeType string `json:"nodeType,omitempty"`
		Name     string `json:"name,omitempty"`
		Desc     string `json:"desc,omitempty"`
		Property string `json:"property,omitempty"`
		Value    string `json:"value,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("POST /sync-ast action=%s node=%s from %s", req.Action, req.NodeID, r.RemoteAddr)
	mainPath := filepath.Join(s.goDir, dslRelativePath)
	src, err := os.ReadFile(mainPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newCode, err := s.studioSvc.ApplyNodeAction(string(src), usecase.NodeAction{
		Action:   req.Action,
		NodeID:   req.NodeID,
		NodeType: req.NodeType,
		Name:     req.Name,
		Desc:     req.Desc,
		Property: req.Property,
		Value:    req.Value,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(mainPath, []byte(newCode), 0644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) handleLayout(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing architecture id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		log.Printf("GET /layout?id=%s from %s", id, r.RemoteAddr)
		layout, err := s.studioSvc.LoadLayout(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(layout)

	case http.MethodPost:
		log.Printf("POST /layout?id=%s from %s", id, r.RemoteAddr)
		var layout usecase.Layout
		if err := json.NewDecoder(r.Body).Decode(&layout); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.studioSvc.SaveLayout(id, &layout); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

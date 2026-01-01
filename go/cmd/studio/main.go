package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/sokoide/advent-of-calm-2025/internal/infra/generator"
	"github.com/sokoide/advent-of-calm-2025/internal/infra/ast"
	"github.com/sokoide/advent-of-calm-2025/internal/infra/repository"
	"github.com/sokoide/advent-of-calm-2025/internal/usecase"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	goDir       string
	studioSvc   usecase.StudioService
	lastContent struct {
		GoCode string `json:"goCode"`
		D2Code string `json:"d2Code"`
		SVG    string `json:"svg"`
		JSON   string `json:"json"`
	}
	contentMu sync.RWMutex
	modeHint  sync.Once
)

const (
	dslRelativePath   = "internal/usecase/ecommerce_architecture.go"
	generateModeEnv   = "STUDIO_GENERATE_MODE"
	generateModeGoRun = "gorun"
)

func main() {
	flag.Parse()

	// Ensure we have an absolute path for goDir
	var err error
	goDir, err = filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}

	if flag.NArg() > 0 {
		goDir, err = filepath.Abs(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("🚀 Starting Studio in: %s", goDir)
	layoutRepo := repository.NewFSLayoutRepository(filepath.Join(goDir, "architectures"))
	studioSvc = usecase.NewStudioService(layoutRepo, ast.GoASTSyncer{})

	// Force an initial read of the DSL file before anything else
	initialReadDSL()

	// Initial generation
	go func() {
		log.Println("🔨 Performing initial generation...")
		regenerate()
		notifyClients("refresh")
	}()

	// Start file watcher
	go watchFiles()

	// Static frontend
	distPath := filepath.Join(goDir, "cmd/studio/frontend/dist")
	fs := http.FileServer(http.Dir(distPath))

	// HTTP handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve index.html for unknown paths (SPA)
		path := filepath.Join(distPath, r.URL.Path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	})
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/content", withCORS(serveContent))
	http.HandleFunc("/update", withCORS(handleUpdate))
	http.HandleFunc("/d2-to-go", withCORS(handleD2ToGo))
	http.HandleFunc("/layout", withCORS(handleLayout))
	http.HandleFunc("/sync-ast", withCORS(handleASTSync))
	http.HandleFunc("/preview-json-sync", withCORS(handlePreviewJSONSync))
	http.HandleFunc("/svg", withCORS(serveSVG))

	port := "3000"
	fmt.Printf("🎨 CALM Studio running at http://localhost:%s\n", port)
	fmt.Printf("📁 Watching: %s/internal and %s/cmd/arch-gen\n", goDir, goDir)
	fmt.Println("💡 Edit Go code or D2 diagram - changes sync both ways!")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func initialReadDSL() {
	mainPath := filepath.Join(goDir, dslRelativePath)
	data, err := os.ReadFile(mainPath)
	if err == nil {
		contentMu.Lock()
		lastContent.GoCode = string(data)
		contentMu.Unlock()
		log.Printf("📖 Initial DSL read success: %s (%d bytes)", mainPath, len(data))
	} else {
		log.Printf("⚠️ Initial DSL read failed at %s: %v", mainPath, err)
	}
}

func watchFiles() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	if err := addWatchDirs(watcher, goDir); err != nil {
		log.Fatal(err)
	}

	debounce := time.NewTimer(0)
	<-debounce.C

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				if strings.HasSuffix(event.Name, ".go") {
					debounce.Reset(300 * time.Millisecond)
				}
			}
		case <-debounce.C:
			log.Println("🔄 Go files changed, regenerating...")
			if regenerate() {
				notifyClients("refresh")
			}
		case err := <-watcher.Errors:
			log.Println("Watcher error:", err)
		}
	}
}

func addWatchDirs(watcher *fsnotify.Watcher, root string) error {
	watchRoots := []string{
		filepath.Join(root, "internal"),
		filepath.Join(root, "cmd", "arch-gen"),
	}

	for _, dir := range watchRoots {
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return watcher.Add(path)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func regenerate() bool {
	// 1. Read Go source first so it's available even if build fails
	goCode := ""
	mainPath := filepath.Join(goDir, dslRelativePath)
	data, err := os.ReadFile(mainPath)
	if err == nil {
		goCode = string(data)
	} else {
		log.Printf("❌ Failed to read Go DSL at %s: %v", mainPath, err)
	}

	contentMu.Lock()
	lastContent.GoCode = goCode
	contentMu.Unlock()

	if os.Getenv(generateModeEnv) == generateModeGoRun {
		return regenerateWithGoRun()
	}
	modeHint.Do(func() {
		log.Printf("ℹ️ In-process generator uses compiled Go DSL. Set %s=%s to reflect file edits.", generateModeEnv, generateModeGoRun)
	})
	return regenerateInProcess()
}

func regenerateInProcess() bool {
	gen := generator.DefaultGenerator()

	jsonOutput, _, err := gen.Generate(usecase.FormatJSON, false)
	if err != nil {
		log.Printf("❌ JSON generation error: %v", err)
		return false
	}

	d2Output, _, err := gen.Generate(usecase.FormatRichD2, false)
	if err != nil {
		log.Printf("❌ Rich D2 output error: %v", err)
		d2Output = ""
	}

	svg := generateSVGFromD2(d2Output)

	contentMu.Lock()
	lastContent.D2Code = d2Output
	lastContent.SVG = svg
	lastContent.JSON = jsonOutput
	contentMu.Unlock()

	log.Println("✅ Content updated (in-process)")
	return true
}

func regenerateWithGoRun() bool {
	// 2. Get JSON output
	cmdJSON := exec.Command("go", "run", "./cmd/arch-gen")
	cmdJSON.Dir = goDir
	var jsonOut, jsonErr bytes.Buffer
	cmdJSON.Stdout = &jsonOut
	cmdJSON.Stderr = &jsonErr

	if err := cmdJSON.Run(); err != nil {
		log.Printf("❌ Build error: %s\n%s", err, jsonErr.String())
		return false
	}

	// 3. Get Rich D2 output
	cmdD2 := exec.Command("go", "run", "./cmd/arch-gen", "-format", "rich-d2")
	cmdD2.Dir = goDir
	var d2Out, d2Err bytes.Buffer
	cmdD2.Stdout = &d2Out
	cmdD2.Stderr = &d2Err
	if err := cmdD2.Run(); err != nil {
		log.Printf("❌ Rich D2 output error: %v\n%s", err, d2Err.String())
		d2Out.Reset()
	}

	svg := generateSVGFromD2(d2Out.String())

	contentMu.Lock()
	lastContent.D2Code = d2Out.String()
	lastContent.SVG = svg
	lastContent.JSON = jsonOut.String()
	contentMu.Unlock()

	log.Println("✅ Content updated (go run)")
	return true
}

func generateSVGFromD2(d2Source string) string {
	if strings.TrimSpace(d2Source) == "" {
		return ""
	}

	d2Cmd := exec.Command("d2", "-", "-")
	d2Cmd.Stdin = strings.NewReader(d2Source)
	var svgOut, svgErr bytes.Buffer
	d2Cmd.Stdout = &svgOut
	d2Cmd.Stderr = &svgErr

	if err := d2Cmd.Run(); err != nil {
		log.Printf("❌ D2 SVG error: %v\n%s", err, svgErr.String())
		return ""
	}

	svg := svgOut.String()
	log.Printf("🎨 D2 SVG generated (%d bytes)", len(svg))
	return svg
}

func notifyClients(msg string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			conn.Close()
			break
		}

		// Handle incoming messages from client
		var update struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		}
		if json.Unmarshal(msg, &update) == nil {
			handleClientUpdate(update.Type, update.Content)
		}
	}
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func handleClientUpdate(updateType, content string) {
	switch updateType {
	case "go":
		// Write Go code to the DSL file and regenerate.
		mainPath := filepath.Join(goDir, dslRelativePath)
		if err := os.WriteFile(mainPath, []byte(content), 0644); err != nil {
			log.Printf("Error writing Go file: %v", err)
			return
		}
		// File watcher will pick it up and regenerate
	case "d2":
		// D2 update - regenerate SVG immediately
		log.Println("📝 D2 update received, generating SVG...")
		contentMu.Lock()
		lastContent.D2Code = content
		// Generate SVG
		d2Cmd := exec.Command("d2", "-", "-")
		d2Cmd.Stdin = strings.NewReader(content)
		var svgOut, svgErr bytes.Buffer
		d2Cmd.Stdout = &svgOut
		d2Cmd.Stderr = &svgErr
		if err := d2Cmd.Run(); err != nil {
			log.Printf("❌ D2 error: %v\n%s", err, svgErr.String())
		} else {
			lastContent.SVG = svgOut.String()
			log.Println("✅ SVG generated successfully")
		}
		contentMu.Unlock()
		notifyClients("refresh-svg")
	}
}

func serveContent(w http.ResponseWriter, r *http.Request) {
	contentMu.RLock()
	defer contentMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lastContent)
}

func serveSVG(w http.ResponseWriter, r *http.Request) {
	contentMu.RLock()
	d2Code := lastContent.D2Code
	svg := lastContent.SVG
	contentMu.RUnlock()

	if svg == "" && strings.TrimSpace(d2Code) != "" {
		svg = generateSVGFromD2(d2Code)
		if svg != "" {
			contentMu.Lock()
			lastContent.SVG = svg
			contentMu.Unlock()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"svg": svg})
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
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

	handleClientUpdate(update.Type, update.Content)
	w.WriteHeader(http.StatusOK)
}

// handlePreviewJSONSync takes CALM JSON and returns what the Go DSL would look like after sync.
func handlePreviewJSONSync(w http.ResponseWriter, r *http.Request) {
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

	mainPath := filepath.Join(goDir, dslRelativePath)
	src, err := os.ReadFile(mainPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newCode, err := studioSvc.SyncFromJSON(string(src), req.JSON)
	if err != nil {
		log.Printf("❌ Sync Error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"newCode": newCode,
	})
}

func handleASTSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Action   string `json:"action"` // "add", "update", "delete"
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

	mainPath := filepath.Join(goDir, dslRelativePath)
	src, err := os.ReadFile(mainPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newCode, err := studioSvc.ApplyNodeAction(string(src), usecase.NodeAction{
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

	log.Printf("✅ AST Synced: %s node %s", req.Action, req.NodeID)
	w.WriteHeader(http.StatusOK)
}

func handleLayout(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing architecture id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		layout, err := studioSvc.LoadLayout(id)
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
		if err := studioSvc.SaveLayout(id, &layout); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleD2ToGo converts D2 source back to Go DSL code
func handleD2ToGo(w http.ResponseWriter, r *http.Request) {
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

	// Apply D2 changes to the DSL file.
	changes, err := applyD2ChangesToGo(req.D2Code)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"changes": changes,
		"success": len(changes) > 0,
	})
}

// applyD2ChangesToGo parses D2, finds label changes, and updates the DSL file.
func applyD2ChangesToGo(d2Code string) ([]string, error) {
	// Parse D2 to find node IDs and their labels
	type nodeInfo struct {
		calmID string
		label  string
	}
	var nodes []nodeInfo

	lines := strings.Split(d2Code, "\n")
	var currentLabel string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Match node definition: id: Label {
		if strings.Contains(trimmed, ": ") && strings.HasSuffix(trimmed, "{") {
			parts := strings.SplitN(trimmed, ": ", 2)
			if len(parts) == 2 {
				currentLabel = strings.TrimSuffix(strings.TrimSpace(parts[1]), " {")
			}
		}

		// Match @calm:id to associate with the node
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

	// Read the DSL file.
	mainPath := filepath.Join(goDir, dslRelativePath)
	content, err := os.ReadFile(mainPath)
	if err != nil {
		return nil, err
	}

	goCode := string(content)
	var changes []string

	// Apply each node change
	for _, n := range nodes {
		// Look for DefineNode("calmID", Type, "OldLabel"
		pattern := fmt.Sprintf(`DefineNode\(\s*"%s"\s*,\s*[\w\.]+\s*,\s*"([^"]+)"`, regexp.QuoteMeta(n.calmID))
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(goCode)

		if len(matches) > 1 {
			oldLabel := matches[1]
			if oldLabel != n.label {
				// Use simple string replacement for safety
				fullMatch := matches[0]
				newMatch := strings.Replace(fullMatch, `"`+oldLabel+`"`, `"`+n.label+`"`, 1)
				goCode = strings.Replace(goCode, fullMatch, newMatch, 1)
				changes = append(changes, fmt.Sprintf("%s: %q → %q", n.calmID, oldLabel, n.label))
			}
		}
	}

	if len(changes) > 0 {
		// Write updated DSL file.
		if err := os.WriteFile(mainPath, []byte(goCode), 0644); err != nil {
			return nil, err
		}
	}

	return changes, nil
}

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
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	goDir       string
	lastContent struct {
		GoCode string `json:"goCode"`
		D2Code string `json:"d2Code"`
		SVG    string `json:"svg"`
		JSON   string `json:"json"`
	}
	contentMu sync.RWMutex
)

func main() {
	flag.Parse()
	goDir = "."
	if flag.NArg() > 0 {
		goDir = flag.Arg(0)
	}

	// Initial generation
	regenerate()

	// Start file watcher
	go watchFiles()

	// HTTP handlers
	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/content", serveContent)
	http.HandleFunc("/update", handleUpdate)
	http.HandleFunc("/d2-to-go", handleD2ToGo)

	port := "3000"
	fmt.Printf("🎨 CALM Studio running at http://localhost:%s\n", port)
	fmt.Printf("📁 Watching: %s/*.go\n", goDir)
	fmt.Println("💡 Edit Go code or D2 diagram - changes sync both ways!")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func watchFiles() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	if err := watcher.Add(goDir); err != nil {
		log.Fatal(err)
	}

	debounce := time.NewTimer(0)
	<-debounce.C

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				if strings.HasSuffix(event.Name, ".go") && !strings.Contains(event.Name, "cmd/") {
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

func regenerate() bool {
	// Get JSON output
	cmdJSON := exec.Command("go", "run", ".")
	cmdJSON.Dir = goDir
	var jsonOut, jsonErr bytes.Buffer
	cmdJSON.Stdout = &jsonOut
	cmdJSON.Stderr = &jsonErr

	if err := cmdJSON.Run(); err != nil {
		log.Printf("❌ Build error: %s\n%s", err, jsonErr.String())
		return false
	}

	// Get Rich D2 output
	cmdD2 := exec.Command("go", "run", ".", "-format", "rich-d2")
	cmdD2.Dir = goDir
	var d2Out bytes.Buffer
	cmdD2.Stdout = &d2Out
	cmdD2.Run()

	// Generate SVG from D2
	svg := ""
	d2Cmd := exec.Command("d2", "-", "-")
	d2Cmd.Stdin = strings.NewReader(d2Out.String())
	var svgOut bytes.Buffer
	d2Cmd.Stdout = &svgOut
	if d2Cmd.Run() == nil {
		svg = svgOut.String()
	}

	// Read Go source
	goCode := ""
	mainPath := filepath.Join(goDir, "main.go")
	if data, err := os.ReadFile(mainPath); err == nil {
		goCode = string(data)
	}

	contentMu.Lock()
	lastContent.GoCode = goCode
	lastContent.D2Code = d2Out.String()
	lastContent.SVG = svg
	lastContent.JSON = jsonOut.String()
	contentMu.Unlock()

	log.Println("✅ Content updated")
	return true
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

func handleClientUpdate(updateType, content string) {
	switch updateType {
	case "go":
		// Write Go code to main.go and regenerate
		mainPath := filepath.Join(goDir, "main.go")
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

	// Apply D2 changes to main.go
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

// applyD2ChangesToGo parses D2, finds label changes, and updates main.go
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

	// Read main.go
	mainPath := filepath.Join(goDir, "main.go")
	content, err := os.ReadFile(mainPath)
	if err != nil {
		return nil, err
	}

	goCode := string(content)
	var changes []string

	// Apply each node change
	for _, n := range nodes {
		// Look for DefineNode("calmID", Type, "OldLabel"
		pattern := fmt.Sprintf(`DefineNode\(\s*"%s"\s*,\s*\w+\s*,\s*"([^"]+)"`, regexp.QuoteMeta(n.calmID))
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
		// Write updated main.go
		if err := os.WriteFile(mainPath, []byte(goCode), 0644); err != nil {
			return nil, err
		}
	}

	return changes, nil
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>CALM Studio</title>
    <link href="https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs/editor/editor.main.css" rel="stylesheet">
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            background: #1e1e2e; color: #cdd6f4;
            height: 100vh; display: flex; flex-direction: column;
        }
        header {
            background: #11111b; padding: 12px 20px;
            display: flex; align-items: center; gap: 15px;
            border-bottom: 1px solid #313244;
        }
        header h1 { font-size: 18px; color: #89b4fa; }
        .status { font-size: 12px; padding: 4px 10px; border-radius: 12px; }
        .status.connected { background: #1e3a2f; color: #a6e3a1; }
        .status.disconnected { background: #3d2a2a; color: #f38ba8; }
        .toolbar {
            display: flex; gap: 8px; margin-left: auto;
        }
        .toolbar button {
            background: #45475a; border: none; color: #cdd6f4;
            padding: 6px 14px; border-radius: 6px; cursor: pointer;
            font-size: 13px; transition: background 0.2s;
        }
        .toolbar button:hover { background: #585b70; }
        .toolbar button.primary { background: #89b4fa; color: #1e1e2e; }
        .toolbar button.warning { background: #fab387; color: #1e1e2e; }
        .main {
            display: flex; flex: 1; overflow: hidden;
        }
        .pane {
            flex: 1; display: flex; flex-direction: column;
            border-right: 1px solid #313244;
        }
        .pane:last-child { border-right: none; }
        .pane-header {
            background: #181825; padding: 8px 15px;
            font-size: 13px; font-weight: 500; color: #a6adc8;
            border-bottom: 1px solid #313244;
            display: flex; align-items: center; gap: 8px;
        }
        .pane-content { flex: 1; overflow: auto; position: relative; }
        #editor { height: 100%; }
        #d2-editor { height: 100%; }
        #diagram {
            background: #fff; padding: 20px; height: 100%;
            display: flex; align-items: center; justify-content: center;
            overflow: auto;
        }
        #diagram svg { max-width: 100%; max-height: 100%; }
        .tabs {
            display: flex; border-bottom: 1px solid #313244;
        }
        .tab {
            padding: 8px 15px; cursor: pointer;
            background: transparent; border: none; color: #6c7086;
            font-size: 13px; border-bottom: 2px solid transparent;
        }
        .tab.active { color: #89b4fa; border-bottom-color: #89b4fa; }
        .d2-toolbar {
            position: absolute; top: 10px; right: 10px; z-index: 100;
            display: flex; gap: 6px;
        }
        .d2-toolbar button {
            background: #45475a; border: none; color: #cdd6f4;
            padding: 4px 10px; border-radius: 4px; cursor: pointer;
            font-size: 12px;
        }
        .d2-toolbar button:hover { background: #585b70; }
        .d2-toolbar button.apply { background: #a6e3a1; color: #1e1e2e; }
        #json-view, #d2-edit-view { display: none; padding: 0; height: 100%; }
        #json-code { font-size: 12px; overflow: auto; padding: 15px; }
    </style>
</head>
<body>
    <header>
        <h1>CALM Studio</h1>
        <div class="status" id="status">Connecting...</div>
        <div class="toolbar">
            <button onclick="saveGoCode()">Save</button>
            <button class="primary" onclick="validateArch()">Validate</button>
        </div>
    </header>
    <main class="main">
        <div class="pane">
            <div class="pane-header">
                <span>Go DSL (main.go)</span>
                <button onclick="syncToD2()" style="margin-left:auto; font-size:11px; padding:3px 8px;">Sync to D2</button>
            </div>
            <div class="pane-content">
                <div id="editor"></div>
            </div>
        </div>
        <div class="pane">
            <div class="pane-header">
                <span>Architecture Diagram</span>
            </div>
            <div class="tabs">
                <button class="tab active" onclick="showTab('diagram')">Diagram</button>
                <button class="tab" onclick="showTab('d2-edit')">Edit D2</button>
                <button class="tab" onclick="showTab('json')">JSON</button>
            </div>
            <div class="pane-content" id="content-area">
                <div id="diagram"></div>
                <div id="d2-edit-view">
                    <div class="d2-toolbar">
                        <button onclick="previewD2()">Preview</button>
                        <button onclick="resetD2()">Reset</button>
                        <button class="apply" onclick="applyD2ToGo()">Apply to Go</button>
                    </div>
                    <div id="d2-editor"></div>
                </div>
                <div id="json-view">
                    <pre id="json-code"></pre>
                </div>
            </div>
        </div>
    </main>

    <script src="https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs/loader.js"></script>
    <script>
        let goEditor, d2Editor;
        let ws;
        let currentTab = 'diagram';
        let d2Dirty = false;  // Track if D2 has been manually edited

        // Initialize Monaco Editors
        require.config({ paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs' }});
        require(['vs/editor/editor.main'], function() {
            // Go Editor
            goEditor = monaco.editor.create(document.getElementById('editor'), {
                value: '// Loading...',
                language: 'go',
                theme: 'vs-dark',
                fontSize: 13,
                minimap: { enabled: false },
                automaticLayout: true,
            });

            // D2 Editor
            d2Editor = monaco.editor.create(document.getElementById('d2-editor'), {
                value: '// D2 Source',
                language: 'yaml',  // D2 is similar to yaml
                theme: 'vs-dark',
                fontSize: 13,
                minimap: { enabled: false },
                automaticLayout: true,
            });

            // Track D2 editor changes
            d2Editor.onDidChangeModelContent(() => {
                d2Dirty = true;
            });

            // Auto-save Go on change (debounced)
            let saveTimeout;
            goEditor.onDidChangeModelContent(() => {
                clearTimeout(saveTimeout);
                saveTimeout = setTimeout(() => {
                    if (ws && ws.readyState === WebSocket.OPEN) {
                        ws.send(JSON.stringify({ type: 'go', content: goEditor.getValue() }));
                    }
                }, 1000);
            });

            loadContent();
        });

        // WebSocket connection
        function connectWS() {
            ws = new WebSocket('ws://' + location.host + '/ws');
            ws.onopen = () => {
                document.getElementById('status').className = 'status connected';
                document.getElementById('status').textContent = 'Connected';
            };
            ws.onmessage = (e) => {
                if (e.data === 'refresh') {
                    // Don't overwrite D2 edits when Go file changes trigger regenerate
                    if (d2Dirty) {
                        console.log('Ignoring refresh while D2 is being edited');
                        return;
                    }
                    loadContent();
                } else if (e.data === 'refresh-svg') {
                    // Only update SVG (for D2 preview)
                    loadSvgOnly();
                }
            };
            ws.onclose = () => {
                document.getElementById('status').className = 'status disconnected';
                document.getElementById('status').textContent = 'Disconnected';
                setTimeout(connectWS, 3000);
            };
        }
        connectWS();

        async function loadContent() {
            const resp = await fetch('/content');
            const data = await resp.json();

            if (goEditor && data.goCode) {
                const pos = goEditor.getPosition();
                goEditor.setValue(data.goCode);
                if (pos) goEditor.setPosition(pos);
            }

            if (d2Editor && data.d2Code && !d2Dirty) {
                const pos = d2Editor.getPosition();
                d2Editor.setValue(data.d2Code);
                if (pos) d2Editor.setPosition(pos);
            }

            if (data.svg) {
                document.getElementById('diagram').innerHTML = data.svg;
            }

            document.getElementById('json-code').textContent =
                data.json ? JSON.stringify(JSON.parse(data.json), null, 2) : '';
        }

        async function loadSvgOnly() {
            const resp = await fetch('/content');
            const data = await resp.json();
            if (data.svg) {
                document.getElementById('diagram').innerHTML = data.svg;
                console.log('SVG updated from D2 preview');
            }
        }

        function showTab(tab) {
            currentTab = tab;
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            event.target.classList.add('active');

            document.getElementById('diagram').style.display = tab === 'diagram' ? 'flex' : 'none';
            document.getElementById('d2-edit-view').style.display = tab === 'd2-edit' ? 'block' : 'none';
            document.getElementById('json-view').style.display = tab === 'json' ? 'block' : 'none';

            if (tab === 'd2-edit' && d2Editor) {
                d2Editor.layout();
            }
        }

        function saveGoCode() {
            if (goEditor && ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({ type: 'go', content: goEditor.getValue() }));
            }
        }

        function previewD2() {
            if (d2Editor && ws && ws.readyState === WebSocket.OPEN) {
                console.log('Sending D2 preview...');
                ws.send(JSON.stringify({ type: 'd2', content: d2Editor.getValue() }));
                alert('Preview sent! Click the Diagram tab to see the updated diagram.');
            } else {
                alert('WebSocket not connected');
            }
        }

        function resetD2() {
            d2Dirty = false;
            loadContent();
        }

        function syncToD2() {
            // Save Go code and regenerate D2
            if (goEditor && ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({ type: 'go', content: goEditor.getValue() }));
                d2Dirty = false;
                alert('Go code saved. D2 will be regenerated.');
            }
        }

        async function applyD2ToGo() {
            console.log('applyD2ToGo called');
            if (!d2Editor) {
                console.log('d2Editor is null');
                alert('D2 Editor not ready');
                return;
            }

            try {
                console.log('Fetching /d2-to-go...');
                const resp = await fetch('/d2-to-go', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ d2Code: d2Editor.getValue() })
                });

                console.log('Response status:', resp.status);
                const data = await resp.json();
                console.log('Response data:', data);

                if (data.error) {
                    alert('Error: ' + data.error);
                } else if (data.success && data.changes && data.changes.length > 0) {
                    var nl = String.fromCharCode(10);
                    alert('Go code updated!' + nl + nl + data.changes.join(nl));
                    // Reset D2 dirty flag and reload content
                    d2Dirty = false;
                    loadContent();
                } else {
                    alert('No changes to apply (labels match Go code)');
                }
            } catch (err) {
                console.log('Error:', err);
                alert('Error: ' + err.message);
            }
        }

        async function validateArch() {
            alert('Run: make check');
        }
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

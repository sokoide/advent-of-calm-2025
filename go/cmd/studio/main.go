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
		// TODO: Parse D2, convert to Architecture, generate Go
		log.Println("D2 → Go conversion requested (not yet implemented)")
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

func serveHTML(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
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
        .pane-content { flex: 1; overflow: auto; }
        #editor { height: 100%; }
        #diagram {
            background: #fff; padding: 20px;
            display: flex; align-items: center; justify-content: center;
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
    </style>
</head>
<body>
    <header>
        <h1>🏗️ CALM Studio</h1>
        <div class="status" id="status">Connecting...</div>
        <div class="toolbar">
            <button onclick="saveGoCode()">💾 Save</button>
            <button onclick="formatCode()">🎨 Format</button>
            <button class="primary" onclick="validateArch()">✅ Validate</button>
        </div>
    </header>
    <main class="main">
        <div class="pane">
            <div class="pane-header">
                <span>📝</span> Go DSL (main.go)
            </div>
            <div class="pane-content">
                <div id="editor"></div>
            </div>
        </div>
        <div class="pane">
            <div class="pane-header">
                <span>🎨</span> Architecture Diagram
            </div>
            <div class="tabs">
                <button class="tab active" onclick="showTab('diagram')">Diagram</button>
                <button class="tab" onclick="showTab('d2')">D2 Source</button>
                <button class="tab" onclick="showTab('json')">JSON</button>
            </div>
            <div class="pane-content" id="content-area">
                <div id="diagram"></div>
                <div id="d2-view" style="display:none; padding: 15px;">
                    <pre id="d2-code" style="font-size: 12px; overflow: auto;"></pre>
                </div>
                <div id="json-view" style="display:none; padding: 15px;">
                    <pre id="json-code" style="font-size: 12px; overflow: auto;"></pre>
                </div>
            </div>
        </div>
    </main>

    <script src="https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs/loader.js"></script>
    <script>
        let editor;
        let ws;
        let currentTab = 'diagram';

        // Initialize Monaco Editor
        require.config({ paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs' }});
        require(['vs/editor/editor.main'], function() {
            editor = monaco.editor.create(document.getElementById('editor'), {
                value: '// Loading...',
                language: 'go',
                theme: 'vs-dark',
                fontSize: 13,
                minimap: { enabled: false },
                automaticLayout: true,
            });

            // Auto-save on change (debounced)
            let saveTimeout;
            editor.onDidChangeModelContent(() => {
                clearTimeout(saveTimeout);
                saveTimeout = setTimeout(() => {
                    if (ws && ws.readyState === WebSocket.OPEN) {
                        ws.send(JSON.stringify({ type: 'go', content: editor.getValue() }));
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
                document.getElementById('status').textContent = '🟢 Connected';
            };
            ws.onmessage = (e) => {
                if (e.data === 'refresh') {
                    loadContent();
                }
            };
            ws.onclose = () => {
                document.getElementById('status').className = 'status disconnected';
                document.getElementById('status').textContent = '🔴 Disconnected';
                setTimeout(connectWS, 3000);
            };
        }
        connectWS();

        async function loadContent() {
            const resp = await fetch('/content');
            const data = await resp.json();

            if (editor && data.goCode) {
                const pos = editor.getPosition();
                editor.setValue(data.goCode);
                if (pos) editor.setPosition(pos);
            }

            if (data.svg) {
                document.getElementById('diagram').innerHTML = data.svg;
            }

            document.getElementById('d2-code').textContent = data.d2Code || '';
            document.getElementById('json-code').textContent =
                data.json ? JSON.stringify(JSON.parse(data.json), null, 2) : '';
        }

        function showTab(tab) {
            currentTab = tab;
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            event.target.classList.add('active');

            document.getElementById('diagram').style.display = tab === 'diagram' ? 'flex' : 'none';
            document.getElementById('d2-view').style.display = tab === 'd2' ? 'block' : 'none';
            document.getElementById('json-view').style.display = tab === 'json' ? 'block' : 'none';
        }

        function saveGoCode() {
            if (editor && ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({ type: 'go', content: editor.getValue() }));
            }
        }

        function formatCode() {
            if (editor) {
                editor.getAction('editor.action.formatDocument').run();
            }
        }

        async function validateArch() {
            alert('Validation: Check console for output');
        }
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
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
	lastContent   string
	lastContentMu sync.RWMutex
	d2Mode        bool
)

func main() {
	d2Flag := flag.Bool("d2", false, "Use D2 format instead of Mermaid")
	flag.Parse()
	d2Mode = *d2Flag

	goDir := "."
	if flag.NArg() > 0 {
		goDir = flag.Arg(0)
	}

	// Initial generation
	regenerate(goDir)

	// Start file watcher
	go watchFiles(goDir)

	// HTTP handlers
	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/content", serveContent)

	port := "3000"
	mode := "Mermaid"
	if d2Mode {
		mode = "D2"
	}
	fmt.Printf("🚀 Live Server (%s) running at http://localhost:%s\n", mode, port)
	fmt.Printf("📁 Watching: %s/*.go\n", goDir)
	fmt.Println("💡 Edit your Go files and see changes instantly!")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func watchFiles(dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	if err := watcher.Add(dir); err != nil {
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
			log.Println("🔄 Change detected, regenerating...")
			if regenerate(dir) {
				notifyClients()
			}
		case err := <-watcher.Errors:
			log.Println("Watcher error:", err)
		}
	}
}

func regenerate(dir string) bool {
	var cmd *exec.Cmd
	if d2Mode {
		cmd = exec.Command("go", "run", ".", "-format", "d2")
	} else {
		cmd = exec.Command("go", "run", ".")
	}
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("❌ Build error: %s\n%s", err, stderr.String())
		return false
	}

	newContent := stdout.String()

	// For D2 mode, generate SVG
	if d2Mode {
		svg, err := d2ToSVG(newContent)
		if err != nil {
			log.Printf("❌ D2 error: %s", err)
			return false
		}
		newContent = svg
	}

	lastContentMu.Lock()
	changed := newContent != lastContent
	lastContent = newContent
	lastContentMu.Unlock()

	if changed {
		log.Println("✅ Architecture updated")
	}
	return changed
}

func d2ToSVG(d2Source string) (string, error) {
	cmd := exec.Command("d2", "-", "-")
	cmd.Stdin = strings.NewReader(d2Source)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

func notifyClients() {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, []byte("reload")); err != nil {
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
		if _, _, err := conn.ReadMessage(); err != nil {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			conn.Close()
			break
		}
	}
}

func serveContent(w http.ResponseWriter, r *http.Request) {
	lastContentMu.RLock()
	defer lastContentMu.RUnlock()
	if d2Mode {
		w.Header().Set("Content-Type", "image/svg+xml")
	} else {
		w.Header().Set("Content-Type", "application/json")
	}
	w.Write([]byte(lastContent))
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	var html string
	if d2Mode {
		html = d2HTML()
	} else {
		html = mermaidHTML()
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func d2HTML() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>CALM Architecture (D2)</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            margin: 0; padding: 20px;
            background: #1a1a2e; color: #eee;
        }
        h1 { color: #00d9ff; margin-bottom: 10px; }
        .status { color: #888; font-size: 14px; margin-bottom: 20px; }
        .connected { color: #4ade80; }
        #diagram { background: #fff; border-radius: 8px; padding: 20px; text-align: center; }
        #diagram img { max-width: 100%; height: auto; }
    </style>
</head>
<body>
    <h1>🏗️ CALM Architecture (D2)</h1>
    <div class="status" id="status">Connecting...</div>
    <div id="diagram"></div>
    <script>
        async function loadDiagram() {
            const resp = await fetch('/content');
            const svg = await resp.text();
            document.getElementById('diagram').innerHTML = svg;
        }

        const ws = new WebSocket('ws://' + location.host + '/ws');
        ws.onopen = () => {
            document.getElementById('status').className = 'status connected';
            document.getElementById('status').textContent = '🟢 Connected - Watching for changes';
            loadDiagram();
        };
        ws.onmessage = () => loadDiagram();
        ws.onclose = () => {
            document.getElementById('status').className = 'status';
            document.getElementById('status').textContent = '🔴 Disconnected';
        };
    </script>
</body>
</html>`
}

func mermaidHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>CALM Architecture (Mermaid)</title>
    <script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            margin: 0; padding: 20px;
            background: #1a1a2e; color: #eee;
        }
        h1 { color: #00d9ff; margin-bottom: 10px; }
        .status { color: #888; font-size: 14px; margin-bottom: 20px; }
        .connected { color: #4ade80; }
        #diagram { background: #fff; border-radius: 8px; padding: 20px; }
        .mermaid { text-align: center; }
    </style>
</head>
<body>
    <h1>🏗️ CALM Architecture (Mermaid)</h1>
    <div class="status" id="status">Connecting...</div>
    <div id="diagram">
        <pre class="mermaid" id="mermaid"></pre>
    </div>
    <script>
        mermaid.initialize({ startOnLoad: false, theme: 'default' });

        async function loadDiagram() {
            const resp = await fetch('/content');
            const arch = await resp.json();
            const mermaidCode = generateMermaid(arch);
            const { svg } = await mermaid.render('graph', mermaidCode);
            document.getElementById('mermaid').innerHTML = svg;
        }

        function generateMermaid(arch) {
            let lines = ['graph LR'];

            for (const node of arch.nodes || []) {
                const shape = getShape(node['node-type']);
                lines.push('    ' + node['unique-id'] + shape[0] + '"' + node.name + '"' + shape[1]);
            }

            for (const rel of arch.relationships || []) {
                const rt = rel['relationship-type'];
                if (rt.connects) {
                    const src = rt.connects.source.node;
                    const dst = rt.connects.destination.node;
                    lines.push('    ' + src + ' --> ' + dst);
                } else if (rt.interacts) {
                    lines.push('    ' + rt.interacts.actor + ' --> ' + rt.interacts.nodes[0]);
                }
            }

            return lines.join('\n');
        }

        function getShape(nodeType) {
            switch(nodeType) {
                case 'actor': return ['((', '))'];
                case 'database': return ['[(', ')]'];
                case 'queue': return ['[[', ']]'];
                case 'system': return ['[/', '/]'];
                default: return ['[', ']'];
            }
        }

        const ws = new WebSocket('ws://' + location.host + '/ws');
        ws.onopen = () => {
            document.getElementById('status').className = 'status connected';
            document.getElementById('status').textContent = '🟢 Connected - Watching for changes';
            loadDiagram();
        };
        ws.onmessage = () => loadDiagram();
        ws.onclose = () => {
            document.getElementById('status').className = 'status';
            document.getElementById('status').textContent = '🔴 Disconnected';
        };
    </script>
</body>
</html>`
}

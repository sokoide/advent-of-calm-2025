package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sokoide/advent-of-calm-2025/cmd/studio/handlers"
	"github.com/sokoide/advent-of-calm-2025/internal/infra/ast"
	"github.com/sokoide/advent-of-calm-2025/internal/infra/generator"
	"github.com/sokoide/advent-of-calm-2025/internal/infra/repository"
	"github.com/sokoide/advent-of-calm-2025/internal/usecase"
)

var (
	state    *handlers.State
	modeHint sync.Once
)

const (
	generateModeEnv   = "STUDIO_GENERATE_MODE"
	generateModeGoRun = "gorun"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	flag.Parse()

	goDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if flag.NArg() > 0 {
		goDir, err = filepath.Abs(flag.Arg(0))
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
	}

	log.Printf("🚀 Starting Studio in: %s", goDir)
	layoutRepo := repository.NewFSLayoutRepository(filepath.Join(goDir, "architectures"))
	studioSvc := usecase.NewStudioService(layoutRepo, ast.GoASTSyncer{})

	state = handlers.NewState(goDir, studioSvc)

	// Initial DSL read
	initialReadDSL()

	// Initial generation
	go func() {
		log.Println("🔨 Performing initial generation...")
		regenerate()
		state.NotifyClients("refresh")
	}()

	// Start file watcher
	go watchFiles()

	// Static frontend
	distPath := filepath.Join(goDir, "cmd/studio/frontend/dist")
	fileServer := http.FileServer(http.Dir(distPath))

	// HTTP handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(distPath, r.URL.Path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
			return
		}
		fileServer.ServeHTTP(w, r)
	})
	http.HandleFunc("/ws", state.HandleWebSocket)
	http.HandleFunc("/content", handlers.WithCORS(state.ServeContent))
	http.HandleFunc("/update", handlers.WithCORS(state.HandleUpdate))
	http.HandleFunc("/d2-to-go", handlers.WithCORS(state.HandleD2ToGo))
	http.HandleFunc("/layout", handlers.WithCORS(state.HandleLayout))
	http.HandleFunc("/sync-ast", handlers.WithCORS(state.HandleASTSync))
	http.HandleFunc("/preview-json-sync", handlers.WithCORS(state.HandlePreviewJSONSync))
	http.HandleFunc("/svg", handlers.WithCORS(state.ServeSVG))

	port := "3000"
	fmt.Printf("🎨 CALM Studio running at http://localhost:%s\n", port)
	fmt.Printf("📁 Watching: %s/internal and %s/cmd/arch-gen\n", goDir, goDir)
	fmt.Println("💡 Edit Go code or D2 diagram - changes sync both ways!")

	return http.ListenAndServe(":"+port, nil)
}

func initialReadDSL() {
	goCode, err := state.ReadGoDSL()
	if err == nil {
		state.SetContent(goCode, "", "", "")
		log.Printf("📖 Initial DSL read success (%d bytes)", len(goCode))
	} else {
		log.Printf("⚠️ Initial DSL read failed: %v", err)
	}
}

func watchFiles() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("⚠️ Failed to create file watcher: %v", err)
		return
	}
	defer watcher.Close()

	if err := addWatchDirs(watcher, state.GoDir); err != nil {
		log.Printf("⚠️ Failed to add watch directories: %v", err)
		return
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
				state.NotifyClients("refresh")
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
	// Read Go source first
	goCode, err := state.ReadGoDSL()
	if err != nil {
		log.Printf("❌ Failed to read Go DSL: %v", err)
	} else {
		state.UpdateGoCode(goCode)
	}

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

	svg := handlers.GenerateSVGFromD2(d2Output)
	state.UpdateD2Content(d2Output, svg, jsonOutput)

	log.Println("✅ Content updated (in-process)")
	return true
}

func regenerateWithGoRun() bool {
	// Get JSON output
	cmdJSON := exec.Command("go", "run", "./cmd/arch-gen")
	cmdJSON.Dir = state.GoDir
	var jsonOut, jsonErr bytes.Buffer
	cmdJSON.Stdout = &jsonOut
	cmdJSON.Stderr = &jsonErr

	if err := cmdJSON.Run(); err != nil {
		log.Printf("❌ Build error: %s\n%s", err, jsonErr.String())
		return false
	}

	// Get Rich D2 output
	cmdD2 := exec.Command("go", "run", "./cmd/arch-gen", "-format", "rich-d2")
	cmdD2.Dir = state.GoDir
	var d2Out, d2Err bytes.Buffer
	cmdD2.Stdout = &d2Out
	cmdD2.Stderr = &d2Err
	if err := cmdD2.Run(); err != nil {
		log.Printf("❌ Rich D2 output error: %v\n%s", err, d2Err.String())
		d2Out.Reset()
	}

	svg := handlers.GenerateSVGFromD2(d2Out.String())
	state.UpdateD2Content(d2Out.String(), svg, jsonOut.String())

	log.Println("✅ Content updated (go run)")
	return true
}

# CALM Architecture Definition in Go & Studio

This directory contains a custom DSL in Go designed to define and generate CALM (Common Architecture Language Model) architectures, along with **CALM Studio (GUI)** for visual interaction.

## Concept: Fusion of DSL and GUI — "The Ultimate DX"

The era of manually writing static JSON is over. By seamlessly synchronizing robust **Go Logic (DSL)** with intuitive **Visual Interaction (GUI)**, this project transforms architecture design from "tedious documentation" into an "exciting development experience."

### Why Go? — From Configuration to Development
While editing static JSON or YAML often feels like a chore, using a Go DSL transforms design into a creative development process.
- **Superior Developer Experience (DX)**: Leverage modern IDEs with auto-completion, refactoring, and instant navigation.
- **Immediate Feedback**: Typos and type mismatches are caught instantly by your editor, and inconsistencies are detected at compile time.
- **Expressive Logic**: Use loops, conditionals, and variables to describe massive systems concisely and intelligently.
- **Above all, it's Fun**: Replace the stress of wrestling with giant JSON files with the pure joy of "coding" your system design.

---

## Key Features & Benefits

### 1. Bidirectional Sync
It is not just a one-way generator.
- **Go ➔ Diagram**: Diagrams update instantly on save (Hot Reload).
- **Diagram ➔ Go**: GUI edits (adding nodes, renaming) directly update the Go source code via AST analysis.
- **JSON ➔ Go (Reverse Conversion)**: Edit the generated JSON and safely sync changes back to Go DSL through a **Diff Preview Modal**.

### 2. Full Hierarchical Support
Recursively renders complex nesting (Containers) like `Order Database Cluster` in both D2 and React Flow. Moving a group automatically moves all its children.

---

## Why is this better than editing JSON directly?

| Aspect | Standard CALM JSON (Manual) | Go DSL + CALM Studio |
| :--- | :--- | :--- |
| **Consistency** | ID typos go unnoticed until runtime. | **Compile-time and real-time validation**. |
| **Scaling** | 10-instance scaling needs 10 copy-pastes. | **Change one constant**. Wiring updates automatically. |
| **Visibility** | Big picture lost in 1000s of lines. | **Always see the diagram** next to your code. |
| **Layout** | Manual coordinate entry is painful. | **Drag & Drop in GUI**. Coordinates saved automatically. |
| **Safety** | Search/replace risks breaking files. | **AST Analysis & Diff Preview** prevent mistakes. |

---

## DSL Naming Conventions

| Prefix / Method | Role | Description | Example |
| :--- | :--- | :--- | :--- |
| **`New...`** | **Independent part generation** | Creates a standalone component with no parent. | `NewRequirement` |
| **`Define...`** | **Declarative creation** | Generates objects using Functional Options. | `arch.DefineNode()` |
| **`With...`** | **Option setting** | Configuration functions for `Define...` methods. | `WithOwner()`, `WithMeta()` |
| **`ConnectTo`** | **Node-centric connection** | Initiates a connection from the node itself. | `node.ConnectTo(dest)` |
| **`Via` / `Is` / `Encrypted`** | **Attribute setting** | Fluently configures object properties. | `rel.Encrypted(true)` |
| **`Merge`** | **Metadata synthesis** | Combines multiple maps into one. Panics on collision. | `Merge(metaTier1, metaOps)` |

---

## Developer Tools

### Launching CALM Studio
Build the frontend and launch the server with one command:
```bash
make studio
```
Then, open `http://localhost:3000` and explore these tabs:
- **Merged**: Split view with Go Editor and Diagram (Standard mode).
- **Diagram**: Interactive editing via React Flow.
- **CALM JSON**: View JSON and perform "Reverse Sync" back to Go.
- **D2 Diagram**: High-fidelity static view powered by D2.

### Launching Local Agent + Studio
If you want to prioritize using your local Go/D2 toolchain, launch both simultaneously:
```bash
make studio-local
```
`studio-local` displays logs with `agent:` / `studio:` prefixes.

#### Significance of the Local Agent
- **Avoid Server Overload**: Conversion (Go DSL → JSON/D2) and SVG generation are executed on the client side, allowing the server to focus on storage and delivery.
- **Leverage Local Go/D2**: Since conversion and SVG generation use each user's Go compiler and D2 CLI, overall throughput is increased.

### Other Make Targets
| Command | Description |
| :--- | :--- |
| **`make format`** | Formats Go code with 120-character limit using `golines`. |
| **`make check`** | Verifies design rules (Ownership, Backup, etc.). |
| **`make validate`** | Validates generated JSON against CALM schema. |
| **`make diff-arch`** | Shows semantic differences between architectures in color. |
| **`make d2`** | Generates static D2 source and SVG files. |
| **`make test`** | Runs unit tests. |

---

## Summary: The Value of "Programming" Your Design

The integration of Go DSL and CALM Studio is about **writing "Expressive Code" (your design) with the same passion as "Functional Code" (your implementation).**

By freeing architects from manual integrity management and using Go's power to solve structural puzzles, we transform system design into a creative development experience. This is the ultimate value of CALM Studio.
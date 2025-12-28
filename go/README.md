# CALM Architecture Definition in Go

This directory contains a custom DSL in Go designed to define and generate CALM (Common Architecture Language Model) architectures. The goal is to move from manual JSON editing to "Architecture as Code," leveraging Go's type safety and composition patterns.

## Why Go? — From Configuration to Development

While editing static JSON or YAML often feels like a chore, using a Go DSL transforms architecture design into a creative development process.

- **Superior Developer Experience (DX)**: Leverage the full power of modern IDEs with auto-completion, refactoring tools, and instant navigation to definitions.
- **Immediate Feedback**: Say goodbye to runtime surprises. Typos and type mismatches are caught instantly by your editor, and inconsistencies are detected at compile time.
- **Expressive Logic**: Use loops, conditionals, and variables to describe massive systems concisely and intelligently.
- **Above all, it's Fun**: Replace the stress of wrestling with giant JSON files with the pure joy of "coding" your system design in Go.

## Key Improvements

- **Functional Options Pattern**: Use `DefineNode` with `WithOwner`, `WithMeta`, etc., for declarative and clean node definitions.
- **Metadata Composition**: The `Merge` helper allows combining reusable metadata parts (e.g., Tier1, DBA, ManagedService) while detecting key collisions at runtime.
- **Fluent Connection API**: Intuitive wiring via `node.ConnectTo(dest)` and structured relationship management using `LinksContainer`.
- **Type-safe Flow Construction**: Reduced reliance on string IDs by using relationship objects or their `GetID()` method to define flow steps.
- **Single Source of Truth for `owner`**: `WithOwner` sets only the top-level `owner` field and keeps metadata free of duplicated ownership to avoid drift.
- **Dynamic Configuration**: Dependencies and flows are built dynamically from node slices (like `n.Gateways`), making it easy to change component counts or IDs.

## Benefits of "Architecture as Code" for Consistency and Maintainability

### Glossary of Terms

- **SSoT (Single Source of Truth)**: A design principle that ensures every piece of data is mastered in one place, eliminating redundancy and inconsistencies.
- **DRY (Don't Repeat Yourself)**: A principle aimed at reducing repetition of information, ensuring that a change in one place propagates across the system.
- **DX (Developer Experience)**: The overall quality, comfort, and efficiency of a developer's workflow when using a specific language or tool.

Using the Go DSL provides significant advantages over manual JSON editing, making architecture models more robust and easier to evolve.

### 1. Enforced Consistency through Single Source of Truth (SSoT)

Manual JSONs often duplicate ownership across top-level fields and metadata, which makes drift easy. This DSL keeps `owner` as a single source of truth at the node level and avoids duplicating it in metadata. If a downstream tool requires `metadata["owner"]`, add it explicitly via `WithMeta` for that node.

The reference JSON in this repository (`architectures/ecommerce-platform.json`) is kept in sync with the Go DSL output to maintain this standard.

For example, if you choose to add `metadata["owner"]` for compatibility, the JSON diff looks like:

```diff
     {
       "unique-id": "customer",
       "name": "Customer",
+      "metadata": {
+        "owner": "marketing-team"
+      },
       "owner": "marketing-team"
     }
```

### 2. Dynamic Scaling and Automatic Consistency

In a raw JSON file, adding a single gateway requires dozens of manual updates across node definitions, dependency lists, relationship wiring, and flow steps—a process highly prone to inconsistency.

With the Go DSL, changing a single constant reconfigures the entire system automatically while ensuring perfect internal integrity.

### 2. Code as a "Design Document" and Decision Tracking

An architecture defined in Go is more than just a tool—it is the design document itself. A decision such as "deploying two gateways" is not just a configuration value but a critical **design decision**.

By explicitly defining these decisions as constants in code, we gain significant advantages:

- **Explicit Intent**: Writing `const numGateways = 2` makes the design intent readable directly from the source code.
- **Traceability via Git**: The evolution of the architecture (e.g., "Why did we scale from 1 to 2 gateways?") is accurately recorded in Git history alongside Pull Requests and commit messages.
- **Guaranteed Integrity**: Since flows, wiring, and dependencies are derived from these constants, the model remains internally consistent even as the design evolves.

```go
// 2025-12-27: Increased gateway count to 2 for high availability
const numGateways = 2
```

### 3. Eliminating Redundancy through Metadata Composition (DRY)

Defining complex operational metadata (Tier, On-Call, Backup, etc.) in JSON often leads to fragile copy-pasting.

In Go, common configurations are defined as reusable "parts." Merging these parts ensures that a single update propagates correctly to all nodes, with the added safety of runtime collision detection.

```go
metaOps = map[string]any{"oncall": "#oncall-ops"}
// Reusable across nodes; Go panics if any keys collide
a.DefineNode("svc", Service, ..., WithMeta(Merge(metaTier1, metaOps)))
```

### 4. Structured Relationships and Type-Safe References

JSON flow definitions are merely lists of string IDs. If a relationship ID changes, flow references silently break.

The Go DSL manages all connections via the `LinksContainer` struct. Referencing these via code variables enables IDE auto-completion and ensures that missing or invalid IDs are caught at compile time.

```go
// IDE auto-completion works, and typos are caught by the compiler
fb.Step(lc.OrderToInv.GetID(), "Checking inventory")
```

### 5. Proving "Redundancy" and "Distribution" via Flow Definitions

CALM flow definitions are "samples" carved out from the vast number of communication patterns within a system.

In our DSL implementation, the Order Flow and Inventory Flow intentionally use different gateway instances. This is not a physical constraint (i.e., flows don't require dedicated gateways) but rather an **intentional modeling technique** to convey the following architectural messages:

- **Visualizing High Availability (HA)**: Demonstrates that multiple gateways are actually running (Active-Active) rather than just being idle standbys.
- **Representing Load Balancing**: Illustrates how the Load Balancer distributes requests across different instances by tracing different paths for different business processes.
- **Validating Wiring Coverage**: Ensures that every physical connection (relationship) is utilized by at least one flow, proving there are no "dead components" in the architecture.

By separating "flow recipes" (steps) from "instances," the DSL automatically generates these sophisticated representations while maintaining perfect internal integrity.

## Comparison: Manual JSON vs. Go DSL (Architecture as Code)

We evaluate the transition from static JSON to a Go DSL through the lenses of maintainability and the joy of development.

### 1. Maintainability and Reliability

| Aspect            | Standard CALM JSON (Manual)                           | Go-based DSL (Code)                                             |
| :---------------- | :---------------------------------------------------- | :-------------------------------------------------------------- |
| **Scalability**   | Adding a component requires dozens of manual updates. | Changing a single constant triggers automatic reconfiguration.  |
| **Consistency**   | Redundant entries lead to drift and "model rot."      | **SSoT** ensures one definition syncs everywhere automatically. |
| **Deduplication** | Copy-pasting is the norm, leading to bloated files.   | **Composition** allows merging reusable metadata parts safely.  |
| **Refactoring**   | Relies on fragile string replacement (grep/sed).      | IDEs rename safely; compilers catch broken references.          |

### 2. Developer Experience (DX) and "The Joy of Coding"

| Aspect             | Standard CALM JSON (Manual)                | Go-based DSL (Code)                                            |
| :----------------- | :----------------------------------------- | :------------------------------------------------------------- |
| **Writing Flow**   | A "chore" of managing brackets and commas. | A "flow" state powered by IDE auto-completion.                 |
| **Feedback**       | Errors only appear at runtime (slow).      | Editors flag errors immediately (instant feedback).            |
| **Expressiveness** | Restricted to static declarations.         | Use loops, conditionals, and functions to build intelligently. |
| **Achievement**    | A sigh of relief: "It's finally done."     | A sense of pride: "I wrote a beautiful program."               |

### Summary: The Value of "Programming" Your Design

Moving to a Go DSL is not just about building a generator—it is about **writing "Expressive Code" (the design) with the same passion as "Functional Code" (the implementation).**

By freeing designers from the stress of manual integrity management and leveraging Go's expressive power to solve architectural puzzles, we transform system design into a creative development experience. This is the true essence of Architecture as Code.

## DSL Naming Conventions

| Prefix / Method                | Role                              | Description                                                       | Example                                      |
| :----------------------------- | :-------------------------------- | :---------------------------------------------------------------- | :------------------------------------------- |
| **`New...`**                   | **Independent part generation**   | Creates a standalone component with no parent.                    | `NewRequirement`, `NewSecurityConfig`        |
| **`Define...`**                | **Declarative creation (Modern)** | Generates a highly configured object using Functional Options.    | `arch.DefineNode()`, `arch.DefineFlow()`     |
| **`With...`**                  | **Option setting**                | Configuration functions to be passed to `Define...` methods.      | `WithOwner()`, `WithMeta()`, `WithControl()` |
| **`ConnectTo`**                | **Node-centric connection**       | Initiates a connection from the node itself, returning a Builder. | `node.ConnectTo(dest)`                       |
| **`Via` / `Is` / `Encrypted`** | **Attribute setting (Fluent)**    | Fluently configures object properties.                            | `rel.Via("src", "dst").Encrypted(true)`      |
| **`Steps` / `MetaMap`**        | **Bulk configuration**            | Sets multiple attributes at once using lists or maps.             | `fb.Steps(specs...)`, `fb.MetaMap(m)`        |
| **`Merge`**                    | **Metadata synthesis**            | Combines multiple maps into one. Panics on key collisions.        | `Merge(metaTier1, metaOps)`                  |

## Developer Tools

The DSL includes several tools to enhance your development experience.

### Live Server (Hot Reload)

Edit your Go files and watch the architecture diagram update automatically in your browser.

```bash
# Display with Mermaid diagrams
make watch

# Display with D2 diagrams (more beautiful rendering)
make watch-d2
```

Open `http://localhost:3000` in your browser to see real-time updates.

### D2 Diagram Generation

Generate [D2](https://d2lang.com/) format diagrams.

```bash
make d2
# → Generates architecture.d2 and architecture.svg
```

Install D2 CLI: `brew install d2`

### Validation DSL

Built-in rules automatically verify architecture quality.

```bash
make check
# → ✅ All validation rules passed
```

**Built-in Rules:**

- `AllNodesHaveOwner()` - All nodes have an owner set
- `AllServicesHaveHealthEndpoint()` - Services have health-endpoint in metadata
- `NoDanglingRelationships()` - No references to non-existent nodes
- `AllFlowsHaveValidTransitions()` - Flow steps reference valid relationships
- `AllDatabasesHaveBackupSchedule()` - Databases have backup-schedule in metadata
- `AllTier1NodesHaveRunbook()` - Tier-1 nodes have runbook in metadata

### Architecture Diff Tool

Compare two CALM JSON files and display semantic differences with color-coded output.

```bash
make diff-arch
```

Displays:

- 📦 Nodes: Added/Removed/Modified
- 🔗 Relationships: Added/Removed
- 🌊 Flows: Added/Removed
- 🛡️ Controls: Added/Removed

## Usage

```bash
# Format code (enforces 120 char limit)
make format

# Validate against CALM schema
make validate

# Run Go DSL validation rules
make check

# Compare against the original reference JSON (success if no diff)
make diff

# Show semantic architecture diff
make diff-arch

# Generate D2 diagram
make d2

# Live server (Mermaid)
make watch

# Live server (D2)
make watch-d2
```

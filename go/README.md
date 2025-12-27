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
- **Single Source of Truth for `owner`**: The `WithOwner` option automatically synchronizes the top-level `owner` field and `metadata["owner"]`, ensuring model consistency.
- **Dynamic Configuration**: Dependencies and flows are built dynamically from node slices (like `n.Gateways`), making it easy to change component counts or IDs.

## Benefits of "Architecture as Code" for Consistency and Maintainability

### Glossary of Terms

- **SSoT (Single Source of Truth)**: A design principle that ensures every piece of data is mastered in one place, eliminating redundancy and inconsistencies.
- **DRY (Don't Repeat Yourself)**: A principle aimed at reducing repetition of information, ensuring that a change in one place propagates across the system.
- **DX (Developer Experience)**: The overall quality, comfort, and efficiency of a developer's workflow when using a specific language or tool.

Using the Go DSL provides significant advantages over manual JSON editing, making architecture models more robust and easier to evolve.

### 1. Enforced Consistency through Single Source of Truth (SSoT)

Manual JSONs often suffer from drift between top-level fields and metadata. The Go DSL enforces integrity at the code level.

For example, the `WithOwner` option automatically synchronizes the top-level `owner` field and `metadata["owner"]`. This ensures high-quality models and leads to diffs that fix previously inconsistent entries:

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

The reference JSON in this repository (`architectures/ecommerce-platform.json`) is kept perfectly in sync with the Go DSL output to maintain this high standard.

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

### 3. Structured Relationships and Type-Safe References

JSON flow definitions are merely lists of string IDs. If a relationship ID changes, flow references silently break.

The Go DSL manages all connections via the `LinksContainer` struct. Referencing these via code variables enables IDE auto-completion and ensures that missing or invalid IDs are caught at compile time.

```go
// IDE auto-completion works, and typos are caught by the compiler
fb.Step(lc.OrderToInv.GetID(), "Checking inventory")
```

### 4. Proving "Redundancy" and "Distribution" via Flow Definitions

CALM flow definitions are "samples" carved out from the vast number of communication patterns within a system.

In our DSL implementation, the Order Flow and Inventory Flow intentionally use different gateway instances. This is not a physical constraint (i.e., flows don't require dedicated gateways) but rather an **intentional modeling technique** to convey the following architectural messages:

- **Visualizing High Availability (HA)**: Demonstrates that multiple gateways are actually running (Active-Active) rather than just being idle standbys.
- **Representing Load Balancing**: Illustrates how the Load Balancer distributes requests across different instances by tracing different paths for different business processes.
- **Validating Wiring Coverage**: Ensures that every physical connection (relationship) is utilized by at least one flow, proving there are no "dead components" in the architecture.

By separating "flow recipes" (steps) from "instances," the DSL automatically generates these sophisticated representations while maintaining perfect internal integrity.

## DSL Naming Conventions

| Prefix / Method | Role | Description | Example |
| :--- | :--- | :--- | :--- |
| **`New...`** | **Independent part generation** | Creates a standalone component with no parent. | `NewRequirement`, `NewSecurityConfig` |
| **`Define...`** | **Declarative creation (Modern)** | Generates a highly configured object using Functional Options. | `arch.DefineNode()`, `arch.DefineFlow()` |
| **`With...`** | **Option setting** | Configuration functions to be passed to `Define...` methods. | `WithOwner()`, `WithMeta()`, `WithControl()` |
| **`ConnectTo`** | **Node-centric connection** | Initiates a connection from the node itself, returning a Builder. | `node.ConnectTo(dest)` |
| **`Via` / `Is` / `Encrypted`** | **Attribute setting (Fluent)** | Fluently configures object properties. | `rel.Via("src", "dst").Encrypted(true)` |
| **`Steps` / `MetaMap`** | **Bulk configuration** | Sets multiple attributes at once using lists or maps. | `fb.Steps(specs...)`, `fb.MetaMap(m)` |
| **`Merge`** | **Metadata synthesis** | Combines multiple maps into one. Panics on key collisions. | `Merge(metaTier1, metaOps)` |

## Usage

```bash
# Format code (enforces 120 char limit)
make -C go format

# Validate against CALM schema
make -C go validate

# Compare against the original reference JSON (success if no diff)
make -C go diff
```
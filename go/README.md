# CALM Architecture Definition in Go

This directory contains a custom DSL in Go designed to define and generate CALM (Common Architecture Language Model) architectures. The goal is to move from manual JSON editing to "Architecture as Code," leveraging Go's type safety and composition patterns.

## Key Improvements

- **Functional Options Pattern**: Use `DefineNode` with `WithOwner`, `WithMeta`, etc., for declarative and clean node definitions.
- **Metadata Composition**: The `Merge` helper allows combining reusable metadata parts (e.g., Tier1, DBA, ManagedService) while detecting key collisions at runtime.
- **Fluent Connection API**: Intuitive wiring via `node.ConnectTo(dest)` and structured relationship management using `LinksContainer`.
- **Type-safe Flow Construction**: Reduced reliance on string IDs by using relationship objects or their `GetID()` method to define flow steps.
- **Single Source of Truth for `owner`**: The `WithOwner` option automatically synchronizes the top-level `owner` field and `metadata["owner"]`, ensuring model consistency.
- **Dynamic Configuration**: Dependencies and flows are built dynamically from node slices (like `n.Gateways`), making it easy to change component counts or IDs.

## Model Consistency and `make diff`

The latest DSL enforces high consistency by ensuring `owner` information is present in both top-level fields and metadata for all nodes.

Accordingly, the reference JSON in the repository (`architectures/ecommerce-platform.json`) has been updated with the output from the Go DSL. This ensures the entire model library maintains top-tier quality and consistency.

Example of a diff compared to an older, less consistent JSON:

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

## DSL Naming Conventions

| Prefix / Method | Role | Description | Example |
| :--- | :--- | :--- | :--- |
| **`New...`** | **Independent part generation** | Creates a standalone component with no parent. | `NewRequirement`, `NewSecurityConfig` |
| **`Define...`** | **Declarative creation (Modern)** | Generates a highly configured object using Functional Options. | `arch.DefineNode()`, `arch.DefineFlow()` |
| **`With...`** | **Option setting** | Configuration functions to be passed to `Define...` methods. | `WithOwner()`, `WithMeta()`, `WithControl()` |
| **`ConnectTo`** | **Node-centric connection** | Initiates a connection from the node itself, returning a Builder. | `node.ConnectTo(dest)` |
| **`Via` / `Is` / `Encrypted`** | **Attribute setting (Fluent)** | Fluently configures object properties. | `rel.Via("src", "dst").Encrypted(true)` |
| **`Merge`** | **Metadata synthesis** | Combines multiple maps into one. Panics on key collisions. | `Merge(metaTier1, metaOps)` |

## Recommended Coding Patterns

### 1. Metadata Composition

Define common settings as variables and merge them as needed.

```go
metaOps = map[string]any{"oncall": "#oncall-ops"}
a.DefineNode("svc", Service, "Name", "Desc", WithMeta(Merge(metaTier1, metaOps)))
```

### 2. Structured Relationship Management

Maintain all connections in a `LinksContainer` struct and reference them in flow definitions.

```go
lc.OrderToInv = n.OrderSvc.ConnectTo(n.InventorySvc, "Check stock").WithID("order-rel")
// Reference in a flow
fb.Step(lc.OrderToInv.GetID(), "Checking inventory")
```

## Usage

```bash
# Format code (enforces 120 char limit)
make -C go format

# Validate against CALM schema
make -C go validate

# Compare against the original reference JSON (success if no diff)
make -C go diff
```
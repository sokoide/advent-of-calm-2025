package domain

import "testing"

func TestValidationError_String(t *testing.T) {
	t.Run("with node id", func(t *testing.T) {
		err := ValidationError{Rule: "RuleA", NodeID: "node-1", Message: "bad"}
		if got := err.String(); got != "[RuleA] node-1: bad" {
			t.Fatalf("unexpected string: %s", got)
		}
	})

	t.Run("without node id", func(t *testing.T) {
		err := ValidationError{Rule: "RuleB", Message: "bad"}
		if got := err.String(); got != "[RuleB] bad" {
			t.Fatalf("unexpected string: %s", got)
		}
	})
}

func TestValidationRules(t *testing.T) {
	t.Run("Validate aggregates errors", func(t *testing.T) {
		arch := NewArchitecture("a", "A", "desc")
		arch.DefineNode("n1", Service, "svc", "desc")
		errs := arch.Validate(AllNodesHaveOwner())
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errs))
		}
	})

	t.Run("AllNodesHaveOwner", func(t *testing.T) {
		arch := NewArchitecture("a", "A", "desc")
		arch.DefineNode("n1", Service, "svc", "desc")
		errs := AllNodesHaveOwner().Validate(arch)
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errs))
		}
	})

	t.Run("AllServicesHaveHealthEndpoint", func(t *testing.T) {
		arch := NewArchitecture("a", "A", "desc")
		arch.DefineNode("svc", Service, "svc", "desc", WithOwner("team", "cc"))
		errs := AllServicesHaveHealthEndpoint().Validate(arch)
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errs))
		}
	})

	t.Run("NoDanglingRelationships", func(t *testing.T) {
		arch := NewArchitecture("a", "A", "desc")
		arch.DefineNode("n1", Service, "svc", "desc", WithOwner("team", "cc"))
		arch.Connect("r1", "desc", "n1", "missing")
		errs := NoDanglingRelationships().Validate(arch)
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errs))
		}
	})

	t.Run("AllFlowsHaveValidTransitions", func(t *testing.T) {
		arch := NewArchitecture("a", "A", "desc")
		arch.Flow("f1", "Flow", "desc").Step("missing", 1, "step", "source-to-destination")
		errs := AllFlowsHaveValidTransitions().Validate(arch)
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errs))
		}
	})

	t.Run("AllDatabasesHaveBackupSchedule", func(t *testing.T) {
		arch := NewArchitecture("a", "A", "desc")
		arch.DefineNode("db", Database, "db", "desc", WithOwner("team", "cc"))
		errs := AllDatabasesHaveBackupSchedule().Validate(arch)
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errs))
		}
	})

	t.Run("AllTier1NodesHaveRunbook", func(t *testing.T) {
		arch := NewArchitecture("a", "A", "desc")
		arch.DefineNode("svc", Service, "svc", "desc",
			WithOwner("team", "cc"),
			WithMeta(map[string]any{"tier": "tier-1"}),
		)
		errs := AllTier1NodesHaveRunbook().Validate(arch)
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errs))
		}
	})

	t.Run("NoUnusedNodes", func(t *testing.T) {
		arch := NewArchitecture("a", "A", "desc")
		arch.DefineNode("n1", Service, "svc", "desc", WithOwner("team", "cc"))
		arch.DefineNode("n2", Service, "svc2", "desc", WithOwner("team", "cc"))
		arch.Connect("r1", "desc", "n1", "n2")
		errs := NoUnusedNodes().Validate(arch)
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d", len(errs))
		}
	})
}

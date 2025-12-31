package domain

import "testing"

func TestArchitectureHelpersBasic(t *testing.T) {
	arch := NewArchitecture("a", "A", "desc")
	arch.AddMeta("k", "v")
	arch.AddControl("c1", "desc", NewRequirement("url", map[string]any{"k": "v"}))

	if arch.Metadata["k"] != "v" {
		t.Fatalf("expected metadata value")
	}
	if arch.Controls["c1"].Description != "desc" {
		t.Fatalf("expected control description")
	}
}

func TestNodeHelpers(t *testing.T) {
	arch := NewArchitecture("a", "A", "desc")
	node := arch.Node("n1", Service, "svc", "desc").Standard("cc", "owner")

	node.AddMeta("m1", "v1").AddControl("c1", "desc", NewRequirement("url", nil))
	iface := node.Interface("i1", "http").
		SetPort(80).
		SetHost("localhost").
		SetPath("/health").
		SetName("health").
		SetDesc("desc").
		SetDB("db")

	if node.Owner != "owner" || node.CostCenter != "cc" {
		t.Fatalf("expected owner and cost center")
	}
	if node.Metadata["m1"] != "v1" {
		t.Fatalf("expected node metadata")
	}
	if node.Controls["c1"].Description != "desc" {
		t.Fatalf("expected node control")
	}
	if iface.Port != 80 || iface.Host != "localhost" || iface.Path != "/health" || iface.Name != "health" ||
		iface.Description != "desc" ||
		iface.Database != "db" {
		t.Fatalf("expected interface fields to be set")
	}
}

func TestRelationshipHelpers(t *testing.T) {
	arch := NewArchitecture("a", "A", "desc")
	arch.DefineNode("src", Service, "src", "desc", WithOwner("team", "cc"))
	arch.DefineNode("dst", Service, "dst", "desc", WithOwner("team", "cc"))

	rel := arch.Connect("r1", "desc", "src", "dst").
		SrcIntf("out").
		DstIntf("in").
		Data("confidential", true).
		WithProtocol("grpc").
		AddMeta("k", "v")

	if rel.Protocol != "grpc" || rel.DataClassification != "confidential" {
		t.Fatalf("expected protocol and classification")
	}
	if rel.Metadata["k"] != "v" {
		t.Fatalf("expected metadata to be set")
	}
	if rel.RelationshipType.Connects.Source.Interfaces[0] != "out" ||
		rel.RelationshipType.Connects.Destination.Interfaces[0] != "in" {
		t.Fatalf("expected interfaces to be set")
	}
}

func TestConnectionBuilderHelpers(t *testing.T) {
	arch := NewArchitecture("a", "A", "desc")
	src := arch.DefineNode("src", Service, "src", "desc", WithOwner("team", "cc"))
	dst := arch.DefineNode("dst", Service, "dst", "desc", WithOwner("team", "cc"))

	cb := src.ConnectTo(dst, "desc").
		WithID("custom").
		Via("out", "in").
		Encrypted(true).
		Is("internal").
		Tag("k", "v").
		Protocol("http")
	if cb.GetID() != "custom" {
		t.Fatalf("expected custom id")
	}
	if cb.rel.DataClassification != "internal" || cb.rel.Protocol != "http" {
		t.Fatalf("expected classification and protocol")
	}
	if cb.rel.Metadata["k"] != "v" {
		t.Fatalf("expected metadata tag")
	}
	if cb.rel.RelationshipType.Connects.Source.Interfaces[0] != "out" {
		t.Fatalf("expected source interface")
	}
}

func TestFlowsHelpers(t *testing.T) {
	arch := NewArchitecture("a", "A", "desc")
	flow := arch.Flow("f1", "Flow", "desc").AddMeta("k", "v").Step("r1", 1, "step", "dir")
	if flow.Metadata["k"] != "v" || flow.Transitions[0].Direction != "dir" {
		t.Fatalf("expected flow metadata and direction")
	}

	builder := arch.DefineFlow("f2", "Flow2", "desc2").
		Meta("k2", "v2").
		MetaMap(map[string]any{"k3": "v3"}).
		Step("r2", "step2").
		StepEx("r3", "step3", "custom")
	if len(builder.flow.Transitions) != 2 {
		t.Fatalf("expected 2 transitions")
	}
	if builder.flow.Metadata["k2"] != "v2" || builder.flow.Metadata["k3"] != "v3" {
		t.Fatalf("expected metadata")
	}
}

func TestValidationRulesPositiveAndNegative(t *testing.T) {
	t.Run("AllServicesHaveHealthEndpoint pass", func(t *testing.T) {
		arch := NewArchitecture("a", "A", "desc")
		arch.DefineNode(
			"svc",
			Service,
			"svc",
			"desc",
			WithOwner("team", "cc"),
			WithMeta(map[string]any{"health-endpoint": "/health"}),
		)
		errs := AllServicesHaveHealthEndpoint().Validate(arch)
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d", len(errs))
		}
	})

	t.Run("AllDatabasesHaveBackupSchedule pass", func(t *testing.T) {
		arch := NewArchitecture("a", "A", "desc")
		arch.DefineNode(
			"db",
			Database,
			"db",
			"desc",
			WithOwner("team", "cc"),
			WithMeta(map[string]any{"backup-schedule": "daily"}),
		)
		errs := AllDatabasesHaveBackupSchedule().Validate(arch)
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d", len(errs))
		}
	})

	t.Run("AllTier1NodesHaveRunbook pass", func(t *testing.T) {
		arch := NewArchitecture("a", "A", "desc")
		arch.DefineNode(
			"svc",
			Service,
			"svc",
			"desc",
			WithOwner("team", "cc"),
			WithMeta(map[string]any{"tier": "tier-1", "runbook": "url"}),
		)
		errs := AllTier1NodesHaveRunbook().Validate(arch)
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d", len(errs))
		}
	})

	t.Run("NoDanglingRelationships for interacts and composed-of", func(t *testing.T) {
		arch := NewArchitecture("a", "A", "desc")
		arch.DefineNode("actor", Actor, "actor", "desc", WithOwner("team", "cc"))
		arch.DefineNode("node", Service, "svc", "desc", WithOwner("team", "cc"))
		arch.DefineNode("container", System, "sys", "desc", WithOwner("team", "cc"))
		arch.Interacts("i1", "desc", "actor", "node")
		arch.ComposedOf("c1", "desc", "container", []string{"node"})
		errs := NoDanglingRelationships().Validate(arch)
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d", len(errs))
		}
	})

	t.Run("NoUnusedNodes detects unused", func(t *testing.T) {
		arch := NewArchitecture("a", "A", "desc")
		arch.DefineNode("n1", Service, "svc", "desc", WithOwner("team", "cc"))
		arch.DefineNode("n2", Service, "svc", "desc", WithOwner("team", "cc"))
		errs := NoUnusedNodes().Validate(arch)
		if len(errs) != 2 {
			t.Fatalf("expected 2 errors, got %d", len(errs))
		}
	})

	t.Run("AllFlowsHaveValidTransitions pass", func(t *testing.T) {
		arch := NewArchitecture("a", "A", "desc")
		arch.DefineNode("n1", Service, "svc", "desc", WithOwner("team", "cc"))
		arch.DefineNode("n2", Service, "svc", "desc", WithOwner("team", "cc"))
		rel := arch.Connect("r1", "desc", "n1", "n2")
		arch.DefineFlow("f1", "Flow", "desc").Step(rel.UniqueID, "step")
		errs := AllFlowsHaveValidTransitions().Validate(arch)
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d", len(errs))
		}
	})
}

func TestRequirementHelpers(t *testing.T) {
	req := NewRequirement("url", map[string]any{"k": "v"})
	if req.RequirementURL != "url" || req.Config == nil {
		t.Fatalf("expected requirement fields")
	}
	url := NewRequirementURL("r", "cfg")
	if url.ConfigURL != "cfg" {
		t.Fatalf("expected config url")
	}

	sec := NewSecurityConfig("AES", "all")
	if sec["algorithm"] != "AES" {
		t.Fatalf("expected security config")
	}
	perf := NewPerformanceConfig(10, 5)
	if perf["p99-latency-ms"] != 10 {
		t.Fatalf("expected performance config")
	}
	avail := NewAvailabilityConfig(99.9, 30)
	if avail["uptime-percentage"] != 99.9 {
		t.Fatalf("expected availability config")
	}
	fail := NewFailoverConfig(1, 2, true)
	if fail["automatic-failover"] != true {
		t.Fatalf("expected failover config")
	}
	cb := NewCircuitBreakerConfig(50, 10, 5)
	if cb["failure-threshold-percentage"] != 50 {
		t.Fatalf("expected circuit breaker config")
	}
}

func TestMetadataHelpers(t *testing.T) {
	m := NewMetadata()
	if len(m) != 0 {
		t.Fatalf("expected empty metadata")
	}
	if v := BoolPtr(true); v == nil || !*v {
		t.Fatalf("expected bool ptr")
	}
}

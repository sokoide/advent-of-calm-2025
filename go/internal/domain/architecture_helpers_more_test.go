package domain

import "testing"

func TestWithInterfaces(t *testing.T) {
	arch := NewArchitecture("a", "A", "desc")
	node := arch.DefineNode("n1", Service, "svc", "desc",
		WithInterfaces(
			&Interface{UniqueID: "i1", Protocol: "http"},
			&Interface{UniqueID: "i2", Protocol: "grpc"},
		),
	)

	if len(node.Interfaces) != 2 {
		t.Fatalf("expected 2 interfaces, got %d", len(node.Interfaces))
	}
	if node.Interfaces[1].UniqueID != "i2" {
		t.Fatalf("expected interface i2, got %s", node.Interfaces[1].UniqueID)
	}
}

func TestFlowBuilderSteps(t *testing.T) {
	arch := NewArchitecture("a", "A", "desc")
	flow := arch.DefineFlow("f1", "Flow", "desc").
		Steps(
			StepSpec{ID: "r1", Desc: "step1"},
			StepSpec{ID: "r2", Desc: "step2", Dir: "destination-to-source"},
		)

	if len(flow.flow.Transitions) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(flow.flow.Transitions))
	}
	if flow.flow.Transitions[0].Direction != "source-to-destination" {
		t.Fatalf("expected default direction, got %s", flow.flow.Transitions[0].Direction)
	}
	if flow.flow.Transitions[1].Direction != "destination-to-source" {
		t.Fatalf("expected custom direction, got %s", flow.flow.Transitions[1].Direction)
	}
}

func TestFlowFromIds(t *testing.T) {
	arch := NewArchitecture("a", "A", "desc")
	flow := arch.FlowFromIds("f1", "Flow", "desc", "r1", "r2")

	if len(flow.Transitions) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(flow.Transitions))
	}
	if flow.Transitions[1].SequenceNumber != 2 {
		t.Fatalf("expected sequence 2, got %d", flow.Transitions[1].SequenceNumber)
	}
}

func TestAddRelationship(t *testing.T) {
	arch := NewArchitecture("a", "A", "desc")
	rel := &Relationship{UniqueID: "r1", Description: "desc"}
	arch.AddRelationship(rel)

	if len(arch.Relationships) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(arch.Relationships))
	}
	if arch.Relationships[0].UniqueID != "r1" {
		t.Fatalf("unexpected relationship id: %s", arch.Relationships[0].UniqueID)
	}
}

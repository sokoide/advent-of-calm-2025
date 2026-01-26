package ast

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// ApplyPatch applies a list of patch operations to the source.
func (s GoASTSyncer) ApplyPatch(src string, ops []domain.PatchOperation) (string, error) {
	fset, f, err := parseSource(src)
	if err != nil {
		return "", err
	}

	for _, op := range ops {
		var err error
		switch op.Type {
		case domain.PatchAddNode:
			err = s.handlePatchAddNode(f, op)
		case domain.PatchUpdateNode:
			err = s.handlePatchUpdateNode(f, fset, op)
		case domain.PatchDeleteNode:
			err = s.handlePatchDeleteNode(f, fset, op)
		case domain.PatchDeleteRelationship:
			err = s.handlePatchDeleteRelationship(f, op)
		case domain.PatchAddRelationship:
			err = s.handlePatchAddRelationship(f, fset, op)
		case domain.PatchAddInterface:
			err = s.handlePatchAddInterface(f, fset, op)
		case domain.PatchDeleteInterface:
			err = s.handlePatchDeleteInterface(f, op)
		case domain.PatchDeleteFlow:
			err = s.handlePatchDeleteFlow(f, op)
		case domain.PatchAddFlow:
			err = s.handlePatchAddFlow(f, op)
		case domain.PatchUpdateFlow:
			err = s.handlePatchUpdateFlow(f, op)
		case domain.PatchAddComposedOf:
			err = s.handlePatchAddComposedOf(f, op)
		case domain.PatchAddControl:
			err = s.handlePatchAddControl(f, op)
		case domain.PatchDeleteControl:
			err = s.handlePatchDeleteControl(f, op)
		case domain.PatchDeleteComposedOf:
			err = s.handlePatchDeleteComposedOf(f, op)
		case domain.PatchUpdateComposedOf:
			err = s.handlePatchUpdateComposedOf(f, op)
		case domain.PatchUpdateControl:
			err = s.handlePatchUpdateControl(f, op)
		case domain.PatchUpdateRelationship:
			err = s.handlePatchUpdateRelationship(f, op)
		}

		if err != nil {
			return "", err
		}
	}

	result, err := formatFile(fset, f)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (GoASTSyncer) handlePatchAddNode(f *ast.File, op domain.PatchOperation) error {
	if op.NodeID == "" || op.NodeName == "" {
		return domain.NewPatchError(
			domain.PatchAddNode,
			"missing required fields",
			map[string]string{"nodeId": op.NodeID, "nodeName": op.NodeName},
		)
	}
	nodeType := op.NodeTypeName
	if nodeType == "" {
		nodeType = "Service"
	}
	if err := AddNodeInAST(f, op.NodeID, nodeType, op.NodeName, op.NodeDesc); err != nil {
		return fmt.Errorf("failed to add node %s: %w", op.NodeID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchUpdateNode(f *ast.File, fset *token.FileSet, op domain.PatchOperation) error {
	if op.Origin != nil {
		if err := updateNodeAtLine(f, fset, op); err != nil {
			return fmt.Errorf("failed to update node at line %d: %w", op.Origin.Line, err)
		}
	} else if op.NodeID != "" {
		valStr := fmt.Sprintf("%v", op.Value)
		if err := UpdateNodePropertyInAST(f, op.NodeID, op.Property, valStr); err != nil {
			return fmt.Errorf("failed to update node %s: %w", op.NodeID, err)
		}
	} else {
		return domain.NewPatchError(
			domain.PatchUpdateNode,
			"requires origin or nodeId",
			map[string]string{"operation": fmt.Sprintf("%+v", op)},
		)
	}
	return nil
}

func (GoASTSyncer) handlePatchDeleteNode(f *ast.File, fset *token.FileSet, op domain.PatchOperation) error {
	var varName string
	if op.NodeID != "" {
		varName = findVariableNameForNodeID(f, op.NodeID)
		if varName != "" {
			log.Printf("Cascade delete: found variable name '%s' for nodeID '%s'", varName, op.NodeID)
		} else {
			log.Printf("Cascade delete: no variable name found for nodeID '%s'", op.NodeID)
		}
	}

	if op.Origin != nil {
		if op.Origin.LoopVar != "" {
			if err := updateLoopVariable(f, op.Origin.LoopVar, -1); err != nil {
				return fmt.Errorf("failed to decrement loop var %s: %w", op.Origin.LoopVar, err)
			}
		} else {
			if err := deleteNodeAtLine(f, fset, op.Origin.Line); err != nil {
				return fmt.Errorf("failed to delete node at line %d: %w", op.Origin.Line, err)
			}
		}
	} else if op.NodeID != "" {
		if err := DeleteNodeInAST(f, op.NodeID); err != nil {
			return fmt.Errorf("failed to delete node %s: %w", op.NodeID, err)
		}
	} else {
		return domain.NewPatchError(
			domain.PatchDeleteNode,
			"requires origin or nodeId",
			map[string]string{"operation": fmt.Sprintf("%+v", op)},
		)
	}

	if varName != "" || op.NodeID != "" {
		deleteRelationshipsReferencingVariable(f, varName, op.NodeID)
		deleteComposedOfReferencingVariable(f, varName, op.NodeID)
	}
	return nil
}

func (GoASTSyncer) handlePatchDeleteRelationship(f *ast.File, op domain.PatchOperation) error {
	if op.NodeID == "" {
		return domain.NewPatchError(
			domain.PatchDeleteRelationship,
			"requires nodeId (relationship ID)",
			nil,
		)
	}
	if err := deleteRelationshipByID(f, op.NodeID); err != nil {
		return fmt.Errorf("failed to delete relationship %s: %w", op.NodeID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchAddRelationship(f *ast.File, fset *token.FileSet, op domain.PatchOperation) error {
	if op.NodeID == "" || op.SourceNode == "" || op.TargetNode == "" {
		return domain.NewPatchError(
			domain.PatchAddRelationship,
			"requires nodeId, sourceNode, and targetNode",
			map[string]string{"nodeId": op.NodeID, "sourceNode": op.SourceNode, "targetNode": op.TargetNode},
		)
	}
	if err := addRelationshipToAST(f, fset, op.NodeID, op.SourceNode, op.TargetNode, op.IsInteracts); err != nil {
		return fmt.Errorf("failed to add relationship %s: %w", op.NodeID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchAddInterface(f *ast.File, fset *token.FileSet, op domain.PatchOperation) error {
	if op.NodeID == "" || op.InterfaceID == "" || op.Protocol == "" {
		return domain.NewPatchError(
			domain.PatchAddInterface,
			"requires nodeId, interfaceId, and protocol",
			map[string]string{"nodeId": op.NodeID, "interfaceId": op.InterfaceID, "protocol": op.Protocol},
		)
	}
	if err := addInterfaceToAST(f, fset, op.NodeID, op.InterfaceID, op.Protocol); err != nil {
		return fmt.Errorf("failed to add interface %s to %s: %w", op.InterfaceID, op.NodeID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchDeleteInterface(f *ast.File, op domain.PatchOperation) error {
	if op.InterfaceID == "" {
		return domain.NewPatchError(
			domain.PatchDeleteInterface,
			"requires interfaceId",
			nil,
		)
	}
	if err := deleteInterfaceFromAST(f, op.InterfaceID); err != nil {
		return fmt.Errorf("failed to delete interface %s: %w", op.InterfaceID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchDeleteFlow(f *ast.File, op domain.PatchOperation) error {
	if op.FlowID == "" {
		return domain.NewPatchError(
			domain.PatchDeleteFlow,
			"requires flowId",
			nil,
		)
	}
	if err := deleteFlowFromAST(f, op.FlowID); err != nil {
		return fmt.Errorf("failed to delete flow %s: %w", op.FlowID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchAddFlow(f *ast.File, op domain.PatchOperation) error {
	if op.FlowID == "" || op.FlowName == "" {
		return domain.NewPatchError(
			domain.PatchAddFlow,
			"requires flowId and flowName",
			map[string]string{"flowId": op.FlowID, "flowName": op.FlowName},
		)
	}
	if err := addFlowToAST(f, op.FlowID, op.FlowName, op.FlowDesc, op.FlowSteps); err != nil {
		return fmt.Errorf("failed to add flow %s: %w", op.FlowID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchUpdateFlow(f *ast.File, op domain.PatchOperation) error {
	if op.FlowID == "" {
		return domain.NewPatchError(
			domain.PatchUpdateFlow,
			"requires flowId",
			nil,
		)
	}
	if err := updateFlowInAST(f, op.FlowID, op.FlowName, op.FlowDesc, op.FlowSteps); err != nil {
		return fmt.Errorf("failed to update flow %s: %w", op.FlowID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchAddComposedOf(f *ast.File, op domain.PatchOperation) error {
	if op.ContainerID == "" || len(op.ChildNodeIDs) == 0 {
		return domain.NewPatchError(
			domain.PatchAddComposedOf,
			"requires containerId and childNodeIds",
			map[string]string{"containerId": op.ContainerID, "childCount": fmt.Sprintf("%d", len(op.ChildNodeIDs))},
		)
	}
	if err := addComposedOfToAST(f, op.NodeID, op.ContainerID, op.ChildNodeIDs); err != nil {
		return fmt.Errorf("failed to add composed-of: %w", err)
	}
	return nil
}

func (GoASTSyncer) handlePatchAddControl(f *ast.File, op domain.PatchOperation) error {
	if op.ControlID == "" {
		return domain.NewPatchError(
			domain.PatchAddControl,
			"requires controlId",
			nil,
		)
	}
	if err := addControlToAST(f, op.ControlID, op.ControlDesc); err != nil {
		return fmt.Errorf("failed to add control %s: %w", op.ControlID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchDeleteControl(f *ast.File, op domain.PatchOperation) error {
	if op.ControlID == "" {
		return domain.NewPatchError(
			domain.PatchDeleteControl,
			"requires controlId",
			nil,
		)
	}
	if err := deleteControlFromAST(f, op.ControlID); err != nil {
		return fmt.Errorf("failed to delete control %s: %w", op.ControlID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchDeleteComposedOf(f *ast.File, op domain.PatchOperation) error {
	if op.ComposedOfID == "" {
		return domain.NewPatchError(
			domain.PatchDeleteComposedOf,
			"requires composedOfId",
			nil,
		)
	}
	if err := deleteComposedOfFromAST(f, op.ComposedOfID); err != nil {
		return fmt.Errorf("failed to delete composed-of %s: %w", op.ComposedOfID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchUpdateComposedOf(f *ast.File, op domain.PatchOperation) error {
	if op.ComposedOfID == "" {
		return domain.NewPatchError(
			domain.PatchUpdateComposedOf,
			"requires composedOfId",
			nil,
		)
	}
	if err := updateComposedOfInAST(f, op.ComposedOfID, op.ComposedOfDesc, op.ChildNodeIDs); err != nil {
		return fmt.Errorf("failed to update composed-of %s: %w", op.ComposedOfID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchUpdateControl(f *ast.File, op domain.PatchOperation) error {
	if op.ControlID == "" {
		return domain.NewPatchError(
			domain.PatchUpdateControl,
			"requires controlId",
			nil,
		)
	}
	if err := updateControlInAST(f, op.ControlID, op.ControlDesc); err != nil {
		return fmt.Errorf("failed to update control %s: %w", op.ControlID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchUpdateRelationship(f *ast.File, op domain.PatchOperation) error {
	relationshipID := op.NodeID
	if relationshipID == "" {
		relationshipID = op.ComposedOfID
	}
	if relationshipID == "" {
		return domain.NewPatchError(
			domain.PatchUpdateRelationship,
			"requires nodeId or composedOfId",
			nil,
		)
	}
	if err := updateRelationshipInAST(f, relationshipID, op.Property, op.Value); err != nil {
		return fmt.Errorf("failed to update relationship %s: %w", relationshipID, err)
	}
	return nil
}

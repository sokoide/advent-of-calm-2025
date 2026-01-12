package ast

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strconv"
	"strings"

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
		log.Printf("Warning: add-node requires nodeId and nodeName")
		return nil
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
		log.Printf("Warning: update-node requires origin or nodeId: %v", op)
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
		log.Printf("Warning: delete-node requires origin or nodeId: %v", op)
		return nil
	}

	if varName != "" || op.NodeID != "" {
		deleteRelationshipsReferencingVariable(f, varName, op.NodeID)
		deleteComposedOfReferencingVariable(f, varName, op.NodeID)
	}
	return nil
}

func (GoASTSyncer) handlePatchDeleteRelationship(f *ast.File, op domain.PatchOperation) error {
	if op.NodeID == "" {
		log.Printf("Warning: delete-relationship requires nodeId (relationship ID)")
		return nil
	}
	if err := deleteRelationshipByID(f, op.NodeID); err != nil {
		return fmt.Errorf("failed to delete relationship %s: %w", op.NodeID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchAddRelationship(f *ast.File, fset *token.FileSet, op domain.PatchOperation) error {
	if op.NodeID == "" || op.SourceNode == "" || op.TargetNode == "" {
		log.Printf("Warning: add-relationship requires nodeId, sourceNode, and targetNode")
		return nil
	}
	if err := addRelationshipToAST(f, fset, op.NodeID, op.SourceNode, op.TargetNode, op.IsInteracts); err != nil {
		return fmt.Errorf("failed to add relationship %s: %w", op.NodeID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchAddInterface(f *ast.File, fset *token.FileSet, op domain.PatchOperation) error {
	if op.NodeID == "" || op.InterfaceID == "" || op.Protocol == "" {
		log.Printf("Warning: add-interface requires nodeId, interfaceId, and protocol")
		return nil
	}
	if err := addInterfaceToAST(f, fset, op.NodeID, op.InterfaceID, op.Protocol); err != nil {
		return fmt.Errorf("failed to add interface %s to %s: %w", op.InterfaceID, op.NodeID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchDeleteInterface(f *ast.File, op domain.PatchOperation) error {
	if op.InterfaceID == "" {
		log.Printf("Warning: delete-interface requires interfaceId")
		return nil
	}
	if err := deleteInterfaceFromAST(f, op.InterfaceID); err != nil {
		return fmt.Errorf("failed to delete interface %s: %w", op.InterfaceID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchDeleteFlow(f *ast.File, op domain.PatchOperation) error {
	if op.FlowID == "" {
		log.Printf("Warning: delete-flow requires flowId")
		return nil
	}
	if err := deleteFlowFromAST(f, op.FlowID); err != nil {
		return fmt.Errorf("failed to delete flow %s: %w", op.FlowID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchAddFlow(f *ast.File, op domain.PatchOperation) error {
	if op.FlowID == "" || op.FlowName == "" {
		log.Printf("Warning: add-flow requires flowId and flowName")
		return nil
	}
	if err := addFlowToAST(f, op.FlowID, op.FlowName, op.FlowDesc, op.FlowSteps); err != nil {
		return fmt.Errorf("failed to add flow %s: %w", op.FlowID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchUpdateFlow(f *ast.File, op domain.PatchOperation) error {
	if op.FlowID == "" {
		log.Printf("Warning: update-flow requires flowId")
		return nil
	}
	if err := updateFlowInAST(f, op.FlowID, op.FlowName, op.FlowDesc, op.FlowSteps); err != nil {
		return fmt.Errorf("failed to update flow %s: %w", op.FlowID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchAddComposedOf(f *ast.File, op domain.PatchOperation) error {
	if op.ContainerID == "" || len(op.ChildNodeIDs) == 0 {
		log.Printf("Warning: add-composed-of requires containerId and childNodeIds")
		return nil
	}
	if err := addComposedOfToAST(f, op.NodeID, op.ContainerID, op.ChildNodeIDs); err != nil {
		return fmt.Errorf("failed to add composed-of: %w", err)
	}
	return nil
}

func (GoASTSyncer) handlePatchAddControl(f *ast.File, op domain.PatchOperation) error {
	if op.ControlID == "" {
		log.Printf("Warning: add-control requires controlId")
		return nil
	}
	if err := addControlToAST(f, op.ControlID, op.ControlDesc); err != nil {
		return fmt.Errorf("failed to add control %s: %w", op.ControlID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchDeleteControl(f *ast.File, op domain.PatchOperation) error {
	if op.ControlID == "" {
		log.Printf("Warning: delete-control requires controlId")
		return nil
	}
	if err := deleteControlFromAST(f, op.ControlID); err != nil {
		return fmt.Errorf("failed to delete control %s: %w", op.ControlID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchDeleteComposedOf(f *ast.File, op domain.PatchOperation) error {
	if op.ComposedOfID == "" {
		log.Printf("Warning: delete-composed-of requires composedOfId")
		return nil
	}
	if err := deleteComposedOfFromAST(f, op.ComposedOfID); err != nil {
		return fmt.Errorf("failed to delete composed-of %s: %w", op.ComposedOfID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchUpdateComposedOf(f *ast.File, op domain.PatchOperation) error {
	if op.ComposedOfID == "" {
		log.Printf("Warning: update-composed-of requires composedOfId")
		return nil
	}
	if err := updateComposedOfInAST(f, op.ComposedOfID, op.ComposedOfDesc, op.ChildNodeIDs); err != nil {
		return fmt.Errorf("failed to update composed-of %s: %w", op.ComposedOfID, err)
	}
	return nil
}

func (GoASTSyncer) handlePatchUpdateControl(f *ast.File, op domain.PatchOperation) error {
	if op.ControlID == "" {
		log.Printf("Warning: update-control requires controlId")
		return nil
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
		log.Printf("Warning: update-relationship requires nodeId or composedOfId")
		return nil
	}
	if err := updateRelationshipInAST(f, relationshipID, op.Property, op.Value); err != nil {
		return fmt.Errorf("failed to update relationship %s: %w", relationshipID, err)
	}
	return nil
}

func updateNodeAtLine(f *ast.File, fset *token.FileSet, op domain.PatchOperation) error {
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}
		// We are looking for the DefineNode call
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check line number
		// Note: fset.Position(n.Pos()).Line might be what we need
		// The Origin.Line points to the start of the defined node, likely the DefineNode call or assignment
		pos := fset.Position(call.Pos())
		if pos.Line != op.Origin.Line {
			return true
		}

		// Verify it's a DefineNode call
		if !isDefineNode(call) {
			return true
		}

		// Apply update
		if err := applyPropertyUpdate(call, op.Property, op.Value); err != nil {
			log.Printf("Failed to apply property update: %v", err)
			return false
		}

		found = true
		return false
	})

	if !found {
		return fmt.Errorf("DefineNode call not found at line %d", op.Origin.Line)
	}
	return nil
}

func isDefineNode(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "DefineNode"
}

func applyPropertyUpdate(call *ast.CallExpr, property string, value interface{}) error {
	valStr := fmt.Sprintf("%v", value)

	// DefineNode args: id, type, name, desc, options...
	if len(call.Args) < 4 {
		return fmt.Errorf("DefineNode call has too few arguments")
	}

	switch property {
	case "name":
		call.Args[2] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", valStr)}
	case "description":
		call.Args[3] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", valStr)}
	case "node-type":
		// TODO: Handle type update (requires import awareness for domain.Service etc)
		// For now, simple string replacement if it was using string literals,
		// but typically it uses domain.Type constants.
		// Let's defer complex type updates or assume string based on current impl needs.
	case "owner":
		// Search for WithOwner option
		updated := false
		for _, arg := range call.Args[4:] {
			if updateOptionCall(arg, "WithOwner", 0, valStr) {
				updated = true
				break
			}
		}
		// If not found, we should append a new option?
		// For simplicity in this iteration, we only update existing options.
		if !updated {
			// append WithOwner? Implementation complex for appending
		}
	}
	return nil
}

func updateOptionCall(expr ast.Expr, funcName string, argIndex int, newValue string) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	if !isOwnerCall(call.Fun, funcName) {
		return false
	}
	if len(call.Args) <= argIndex {
		return false
	}
	call.Args[argIndex] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", newValue)}
	return true
}

func isOwnerCall(fun ast.Expr, name string) bool {
	switch v := fun.(type) {
	case *ast.Ident:
		return v.Name == name
	case *ast.SelectorExpr:
		return v.Sel.Name == name
	}
	return false
}

func deleteNodeAtLine(f *ast.File, fset *token.FileSet, line int) error {
	// We need to find the statement containing line X and remove it from its block
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		// Look for blocks
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		newOrderedList := make([]ast.Stmt, 0, len(block.List))
		for _, stmt := range block.List {
			pos := fset.Position(stmt.Pos())
			// We check if the statement STARTS at the line, or contains it.
			// Ideally the Origin line is the start line.
			if pos.Line == line {
				found = true
				continue // Skip this statement (delete)
			}
			newOrderedList = append(newOrderedList, stmt)
		}

		if found {
			block.List = newOrderedList
			return false // Stop traversing
		}
		return true
	})

	if !found {
		// Try finding it in top-level declarations (unlikely for DefineNode inside functions, but possible)
		// Actually typical DefineNode is inside Build() which is in a FuncDecl, so BlockStmt traversal should catch it.
		return fmt.Errorf("statement at line %d not found in any block", line)
	}
	return nil
}

// deleteRelationshipByID finds a Connect() call with the specified relationship ID and removes it.
func deleteRelationshipByID(f *ast.File, relationshipID string) error {
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		newOrderedList := make([]ast.Stmt, 0, len(block.List))
		for _, stmt := range block.List {
			// Check if this statement contains a Connect() or Interacts() call with matching ID
			if containsRelationshipWithID(stmt, relationshipID) {
				found = true
				continue // Skip this statement (delete)
			}
			newOrderedList = append(newOrderedList, stmt)
		}

		if found {
			block.List = newOrderedList
			return false
		}
		return true
	})

	if !found {
		return fmt.Errorf("relationship with ID %q not found", relationshipID)
	}
	return nil
}

// containsRelationshipWithID checks if a statement contains a Connect() or Interacts() call with the given ID.
// Handles frontend-generated IDs with -N suffix (e.g., "customer-interacts-lb-0" -> "customer-interacts-lb").
func containsRelationshipWithID(stmt ast.Stmt, relationshipID string) bool {
	// Strip -N suffix from relationshipID if present (frontend adds -0, -1, etc.)
	baseID := relationshipID
	if idx := strings.LastIndex(relationshipID, "-"); idx > 0 {
		suffix := relationshipID[idx+1:]
		// Check if suffix is numeric
		if _, err := strconv.Atoi(suffix); err == nil {
			baseID = relationshipID[:idx]
		}
	}

	found := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if it's a Connect(), Interacts(), or ComposedOf() call
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		funcName := sel.Sel.Name
		if funcName != "Connect" && funcName != "Interacts" && funcName != "ComposedOf" {
			return true
		}

		// Check first argument is the relationship ID
		if len(call.Args) < 1 {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}

		// Remove quotes from literal value
		idValue := lit.Value
		if len(idValue) >= 2 && idValue[0] == '"' && idValue[len(idValue)-1] == '"' {
			idValue = idValue[1 : len(idValue)-1]
		}

		// Match against both full ID and base ID
		if idValue == relationshipID || idValue == baseID {
			found = true
			return false
		}
		return true
	})
	return found
}

// addToArrayVariable finds a variable definition (varName := []string{...}) and adds a new element.
func addToArrayVariable(f *ast.File, varName, newValue string) bool {
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}
		// Look for assignment: varName := []string{...} or varName = append(...)
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		if len(assign.Lhs) == 0 || len(assign.Rhs) == 0 {
			return true
		}
		// Check if LHS is our variable
		lhsIdent, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || lhsIdent.Name != varName {
			return true
		}
		// Check if RHS is a composite literal (array)
		compLit, ok := assign.Rhs[0].(*ast.CompositeLit)
		if !ok {
			return true
		}
		// Add new element
		newElem := &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", newValue)}
		compLit.Elts = append(compLit.Elts, newElem)
		found = true
		return false
	})
	return found
}

// updateRelationshipInAST finds a Connect() or Interacts() call with the specified ID and updates its properties.
func updateRelationshipInAST(f *ast.File, relationshipID, property string, value interface{}) error {
	found := false
	valStr := fmt.Sprintf("%v", value)

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		stmt, ok := n.(ast.Stmt)
		if !ok {
			return true
		}

		if !containsRelationshipWithID(stmt, relationshipID) {
			return true
		}

		exprStmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			return true
		}

		// Traverse call chain from outermost to innermost
		// e.g., Connect(id, desc, ...).Protocol(proto)
		// Call(Protocol) -> X: Call(Connect)
		curr := exprStmt.X
		for {
			call, ok := curr.(*ast.CallExpr)
			if !ok {
				break
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				break
			}

			// 1. Check if the current call is the property method (e.g. .Protocol("..."))
			if strings.EqualFold(sel.Sel.Name, property) {
				if len(call.Args) >= 1 {
					call.Args[0] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", valStr)}
					found = true
					return false
				}
			}

			// 2. Check if the current call is the root Connect/Interacts/ComposedOf call
			if sel.Sel.Name == "Connect" || sel.Sel.Name == "Interacts" {
				if strings.EqualFold(property, "description") && len(call.Args) >= 2 {
					call.Args[1] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", valStr)}
					found = true
					return false
				}
				// ID match is already verified by containsRelationshipWithID
				// Stop here as we've hit the root
				break
			}

			// 3. Handle ComposedOf - signature: ComposedOf(id, desc, container, []string{nodes})
			if sel.Sel.Name == "ComposedOf" {
				if strings.EqualFold(property, "add-child-node") {
					// Add node to the array (4th argument)
					if len(call.Args) >= 4 {
						// Case 1: Direct array literal []string{...}
						if compLit, ok := call.Args[3].(*ast.CompositeLit); ok {
							newElem := &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", valStr)}
							compLit.Elts = append(compLit.Elts, newElem)
							found = true
							return false
						}
						// Case 2: Variable reference (e.g., sysNodes)
						if ident, ok := call.Args[3].(*ast.Ident); ok {
							varName := ident.Name
							// Find the variable definition and add to its array
							if addToArrayVariable(f, varName, valStr) {
								found = true
								return false
							}
						}
					}
				} else if strings.EqualFold(property, "description") && len(call.Args) >= 2 {
					call.Args[1] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", valStr)}
					found = true
					return false
				}
				break
			}

			curr = sel.X
		}

		return true
	})

	if !found {
		return fmt.Errorf("relationship %s property %s not found/updated", relationshipID, property)
	}
	return nil
}

// addRelationshipToAST adds a new Connect() or Interacts() call to the AST.
func addRelationshipToAST(
	f *ast.File,
	fset *token.FileSet,
	relationshipID, sourceNode, targetNode string,
	isInteracts bool,
) error {
	fn := findFunctionInAST(f, []string{"wireComponents", "build"})
	if fn == nil {
		return fmt.Errorf("wireComponents or build function not found in AST")
	}

	receiverName := getReceiverName(fn, "a")
	funcName := "Connect"
	if isInteracts {
		funcName = "Interacts"
	}

	desc := fmt.Sprintf("%s from %s to %s", funcName, sourceNode, targetNode)
	newStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(receiverName),
				Sel: ast.NewIdent(funcName),
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", relationshipID)},
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", desc)},
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", sourceNode)},
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", targetNode)},
			},
		},
	}

	insertStmtBeforeReturn(fn, newStmt)

	if isInteracts {
		log.Printf("📝 Adding Interacts: %s from %s to %s", relationshipID, sourceNode, targetNode)
	} else {
		log.Printf("📝 Adding Connect: %s from %s to %s", relationshipID, sourceNode, targetNode)
	}
	return nil
}

func updateLoopVariable(f *ast.File, varName string, delta int) error {
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		// 1. Const declaration: const numGateways = 2
		if gen, ok := n.(*ast.GenDecl); ok && gen.Tok == token.CONST {
			for _, spec := range gen.Specs {
				vspec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, name := range vspec.Names {
					if name.Name == varName {
						if len(vspec.Values) > i {
							if lit, ok := vspec.Values[i].(*ast.BasicLit); ok && lit.Kind == token.INT {
								val, err := strconv.Atoi(lit.Value)
								if err == nil {
									newVal := val + delta
									if newVal < 0 {
										newVal = 0
									}
									lit.Value = strconv.Itoa(newVal)
									found = true
									return false
								}
							}
						}
					}
				}
			}
		}

		// 2. Assignment: numGateways := 2 or var numGateways = 2
		// Handling var declaration
		if gen, ok := n.(*ast.GenDecl); ok && gen.Tok == token.VAR {
			for _, spec := range gen.Specs {
				vspec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, name := range vspec.Names {
					if name.Name == varName {
						if len(vspec.Values) > i {
							if lit, ok := vspec.Values[i].(*ast.BasicLit); ok && lit.Kind == token.INT {
								val, err := strconv.Atoi(lit.Value)
								if err == nil {
									newVal := val + delta
									if newVal < 0 {
										newVal = 0
									}
									lit.Value = strconv.Itoa(newVal)
									found = true
									return false
								}
							}
						}
					}
				}
			}
		}

		return true
	})

	if !found {
		return fmt.Errorf("variable/constant %q not found or not an integer literal", varName)
	}
	return nil
}

// generateVarNameFromID converts a node ID to a valid Go variable name.
// Example: "payment-service-copy-123" -> "paymentServiceCopy123"
func generateVarNameFromID(nodeID string) string {
	parts := strings.Split(nodeID, "-")
	var result strings.Builder
	for i, part := range parts {
		if part == "" {
			continue
		}
		if i == 0 {
			// First part: lowercase
			result.WriteString(strings.ToLower(part))
		} else {
			// Subsequent parts: capitalize first letter
			if len(part) > 0 {
				result.WriteString(strings.ToUpper(part[:1]) + part[1:])
			}
		}
	}
	// Ensure it starts with a letter
	name := result.String()
	if name == "" || !((name[0] >= 'a' && name[0] <= 'z') || (name[0] >= 'A' && name[0] <= 'Z')) {
		name = "node" + name
	}
	return name
}

// addInterfaceToAST finds the DefineNode call for the given nodeID and adds an Interface() call after it.
func addInterfaceToAST(f *ast.File, fset *token.FileSet, nodeID, interfaceID, protocol string) error {
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		// Look for the statement containing DefineNode with matching ID
		var insertIndex int = -1
		var varName string

		for i, stmt := range block.List {
			// Case 1: Assignment statement: varName := a.DefineNode(...)
			if assign, ok := stmt.(*ast.AssignStmt); ok {
				for j, expr := range assign.Rhs {
					if isDefineNodeCallWithID(expr, nodeID) {
						// Found it! Get LHS name
						if ident, ok := assign.Lhs[j].(*ast.Ident); ok {
							varName = ident.Name
							insertIndex = i + 1
							break
						}
					}
				}
			}

			// Case 2: Expression statement: a.DefineNode(...)
			// Convert to variable assignment and add Interface as separate statement
			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				if isDefineNodeCallWithID(exprStmt.X, nodeID) {
					// Generate a variable name from nodeID
					varName = generateVarNameFromID(nodeID)

					// Convert expression statement to assignment
					assignStmt := &ast.AssignStmt{
						Lhs: []ast.Expr{ast.NewIdent(varName)},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{exprStmt.X},
					}

					// Replace the expression statement with assignment
					block.List[i] = assignStmt

					// Set insert index for the interface call
					insertIndex = i + 1
					break
				}
			}

			if insertIndex != -1 {
				break
			}
		}

		// Handle variable assignment case
		if insertIndex != -1 && varName != "" {
			// Create the new statement: varName.Interface("id", "proto")
			// We can't easily construct a full AST manually with positions, so we construct a call expression
			// and rely on go/printer or formatFile to handle it.

			// Construct: varName.Interface
			fun := &ast.SelectorExpr{
				X:   &ast.Ident{Name: varName},
				Sel: &ast.Ident{Name: "Interface"},
			}

			// Args: "id", "proto"
			args := []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", interfaceID)},
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", protocol)},
			}

			call := &ast.CallExpr{Fun: fun, Args: args}
			stmt := &ast.ExprStmt{X: call}

			// Insert into block
			newOrderedList := append([]ast.Stmt{}, block.List[:insertIndex]...)
			newOrderedList = append(newOrderedList, stmt)
			newOrderedList = append(newOrderedList, block.List[insertIndex:]...)
			block.List = newOrderedList

			found = true
			return false
		}

		return true
	})

	if !found {
		return fmt.Errorf("definition for node %s not found", nodeID)
	}
	return nil
}

// isDefineNodeCallWithID checks if expr is a DefineNode call for the specific nodeID, possibly inside a chain.
func isDefineNodeCallWithID(expr ast.Expr, nodeID string) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	// Case 1: Direct DefineNode call
	if isDefineNode(call) {
		if len(call.Args) < 1 {
			return false
		}
		// Check first arg (ID)
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return false
		}
		val := strings.Trim(lit.Value, "\"`")
		return val == nodeID
	}

	// Case 2: Method chain (e.g. DefineNode(...).Interface(...))
	// We need to check the object being called (X in SelectorExpr)
	if _, ok := call.Fun.(*ast.SelectorExpr); ok {
		// Traverse down the chain
		// The structure is CallExpr -> SelectorExpr -> X (recursive)
		// We'll perform a simple unwrapping
		curr := call.Fun
		for {
			sel, ok := curr.(*ast.SelectorExpr)
			if !ok {
				break
			}
			// If X is a call, check if it's our target
			if subCall, ok := sel.X.(*ast.CallExpr); ok {
				if isDefineNodeCallWithID(subCall, nodeID) {
					return true
				}
				// If not directly DefineNode, continue unwrapping
				curr = subCall.Fun
			} else {
				break
			}
		}
	}

	return false
}

// deleteInterfaceFromAST finds an Interface("id", ...) call and removes it.
func deleteInterfaceFromAST(f *ast.File, interfaceID string) error {
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		newOrderedList := make([]ast.Stmt, 0, len(block.List))
		for _, stmt := range block.List {
			if isInterfaceCallWithID(stmt, interfaceID) {
				found = true
				continue // Delete
			}
			newOrderedList = append(newOrderedList, stmt)
		}

		if found {
			block.List = newOrderedList
			return false
		}
		return true
	})

	if !found {
		return fmt.Errorf("interface call %q not found", interfaceID)
	}
	return nil
}

func isInterfaceCallWithID(stmt ast.Stmt, interfaceID string) bool {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}

	// Traverse down the call chain to find the root Interface() call
	// Example: node.Interface("id").SetName("name")
	// AST: Call(SetName) -> X: Call(Interface) -> X: node
	curr := exprStmt.X
	for {
		call, ok := curr.(*ast.CallExpr)
		if !ok {
			return false
		}

		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "Interface" {
				if len(call.Args) >= 1 {
					if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
						val := strings.Trim(lit.Value, "\"")
						if val == interfaceID {
							return true
						}
					}
				}
			}
			// Move down to the receiver
			curr = sel.X
		} else {
			return false
		}
	}
}

// deleteFlowFromAST finds a DefineFlow("id", ...) call and removes it.
func deleteFlowFromAST(f *ast.File, flowID string) error {
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		newOrderedList := make([]ast.Stmt, 0, len(block.List))
		for _, stmt := range block.List {
			if isDefineFlowCallWithID(stmt, flowID) {
				found = true
				continue // Delete
			}
			newOrderedList = append(newOrderedList, stmt)
		}

		if found {
			block.List = newOrderedList
			return false
		}
		return true
	})

	if !found {
		return fmt.Errorf("flow definition %q not found", flowID)
	}
	return nil
}

func isDefineFlowCallWithID(stmt ast.Stmt, flowID string) bool {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}

	// Traverse down the call chain to find the root DefineFlow() call
	// Example: a.DefineFlow("id", ...).Steps(...)
	curr := exprStmt.X
	for {
		call, ok := curr.(*ast.CallExpr)
		if !ok {
			return false
		}

		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "DefineFlow" {
				if len(call.Args) >= 1 {
					if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
						val := strings.Trim(lit.Value, "\"")
						if val == flowID {
							return true
						}
					}
				}
			}
			// Move down to the receiver
			curr = sel.X
		} else {
			return false
		}
	}
}

// addFlowToAST adds a new DefineFlow call to the AST.
func addFlowToAST(f *ast.File, flowID, name, desc string, steps []string) error {
	fn := findFunctionInAST(f, []string{"defineFlows", "build"})
	if fn == nil {
		return fmt.Errorf("defineFlows or build function not found in AST")
	}

	receiverName := getReceiverName(fn, "a")

	// 1. Create DefineFlow call
	defineFlowCall := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(receiverName),
			Sel: ast.NewIdent("DefineFlow"),
		},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", flowID)},
			&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", name)},
			&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", desc)},
		},
	}

	// 2. Create Steps call
	stepSpecs := make([]ast.Expr, 0, len(steps))
	for _, stepID := range steps {
		stepSpecs = append(stepSpecs, &ast.CompositeLit{
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent("domain"),
				Sel: ast.NewIdent("StepSpec"),
			},
			Elts: []ast.Expr{
				&ast.KeyValueExpr{
					Key:   ast.NewIdent("ID"),
					Value: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", stepID)},
				},
			},
		})
	}

	stepsCall := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   defineFlowCall,
			Sel: ast.NewIdent("Steps"),
		},
		Args: stepSpecs,
	}

	insertStmtBeforeReturn(fn, &ast.ExprStmt{X: stepsCall})

	log.Printf("📝 Adding Flow: %s", flowID)
	return nil
}

// updateFlowInAST updates an existing DefineFlow call.
func updateFlowInAST(f *ast.File, flowID, name, desc string, steps []string) error {
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		// Look for statement containing DefineFlow
		stmt, ok := n.(ast.Stmt)
		if !ok {
			return true
		}

		if !isDefineFlowCallWithID(stmt, flowID) {
			return true
		}

		// Found the statement. Now we need to update arguments of DefineFlow and Steps.
		exprStmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			return true
		}

		// Traverse chain to find DefineFlow and Steps calls
		var defineFlowCall *ast.CallExpr
		var stepsCall *ast.CallExpr

		curr := exprStmt.X
		for {
			call, ok := curr.(*ast.CallExpr)
			if !ok {
				break
			}

			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "DefineFlow" {
					defineFlowCall = call
				} else if sel.Sel.Name == "Steps" {
					stepsCall = call
				}
				curr = sel.X
			} else {
				break
			}
		}

		if defineFlowCall != nil {
			// Update Name (arg 1) and Desc (arg 2)
			// Arg 0 is ID
			if len(defineFlowCall.Args) >= 3 {
				defineFlowCall.Args[1] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", name)}
				defineFlowCall.Args[2] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", desc)}
			}
		}

		if stepsCall != nil {
			// Update Steps args
			newArgs := make([]ast.Expr, 0, len(steps))
			for _, stepID := range steps {
				// domain.StepSpec{ID: "stepID"}
				val := &ast.CompositeLit{
					Type: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "domain"},
						Sel: &ast.Ident{Name: "StepSpec"},
					},
					Elts: []ast.Expr{
						&ast.KeyValueExpr{
							Key:   &ast.Ident{Name: "ID"},
							Value: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", stepID)},
						},
						// We don't preserve description for now as it's hard to track
					},
				}
				newArgs = append(newArgs, val)
			}
			stepsCall.Args = newArgs
		} else {
			// If Steps() call is missing, we need to append it.
			// Ideally we would wrap the existing expression `X` with `CallExpr{Fun: SelectorExpr{X: X, Sel: "Steps"}, Args: ...}`
			// But since we can't easily modify the structure here without re-writing the whole statement,
			// we will log a warning.
			// TODO: Implement adding Steps() if missing.
			log.Printf("Warning: Steps() call not found for flow %s, skipping steps update", flowID)
		}

		found = true
		return false
	})

	if !found {
		return fmt.Errorf("flow %q not found", flowID)
	}
	return nil
}

// addComposedOfToAST adds a new ComposedOf relationship to the AST.
func addComposedOfToAST(f *ast.File, id, containerID string, childNodeIDs []string) error {
	if id == "" {
		id = fmt.Sprintf("composed-%s", containerID)
	}

	fn := findFunctionInAST(f, []string{"wireComponents", "build"})
	if fn == nil {
		return fmt.Errorf("wireComponents or build function not found in AST")
	}

	receiverName := getReceiverName(fn, "a")

	// Build the node IDs slice: []string{"n1", "n2"}
	elts := make([]ast.Expr, len(childNodeIDs))
	for i, nid := range childNodeIDs {
		elts[i] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", nid)}
	}

	newStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(receiverName),
				Sel: ast.NewIdent("ComposedOf"),
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", id)},
				&ast.BasicLit{Kind: token.STRING, Value: "\"Container relationship\""},
				ast.NewIdent(containerID),
				&ast.CompositeLit{
					Type: &ast.ArrayType{Elt: ast.NewIdent("string")},
					Elts: elts,
				},
			},
		},
	}

	insertStmtBeforeReturn(fn, newStmt)

	log.Printf("📝 Adding ComposedOf: %s (container: %s)", id, containerID)
	return nil
}

// deleteComposedOfFromAST removes a ComposedOf call from the AST.
func deleteComposedOfFromAST(f *ast.File, composedOfID string) error {
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		newList := make([]ast.Stmt, 0, len(block.List))
		for _, stmt := range block.List {
			if isComposedOfCallWithID(stmt, composedOfID) {
				found = true
				continue // Skip this statement to delete it
			}
			newList = append(newList, stmt)
		}

		if found {
			block.List = newList
		}

		return true
	})

	if !found {
		return fmt.Errorf("composed-of %q not found", composedOfID)
	}
	return nil
}

// isComposedOfCallWithID checks if a statement is a ComposedOf call with the given ID.
func isComposedOfCallWithID(stmt ast.Stmt, composedOfID string) bool {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}

	call, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}

	// Check if it's a method call like a.ComposedOf(...)
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != "ComposedOf" {
		return false
	}

	if len(call.Args) < 1 {
		return false
	}

	// Check first argument (composed-of ID)
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}

	val := strings.Trim(lit.Value, "\"")
	return val == composedOfID
}

// addControlToAST adds a new AddControl call to the AST.
func addControlToAST(f *ast.File, controlID, desc string) error {
	fn := findFunctionInAST(f, []string{"addGlobalControls", "build", "defineNodes"})
	if fn == nil {
		return fmt.Errorf("suitable function for AddControl not found in AST")
	}

	receiverName := getReceiverName(fn, "arch")

	newStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(receiverName),
				Sel: ast.NewIdent("AddControl"),
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", controlID)},
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", desc)},
			},
		},
	}

	insertStmtBeforeReturn(fn, newStmt)

	log.Printf("📝 Adding Control: %s", controlID)
	return nil
}

// deleteControlFromAST removes an AddControl call from the AST.
func deleteControlFromAST(f *ast.File, controlID string) error {
	found := false

	// Controls are typically added via arch.AddControl("id", "desc")
	// We need to find and remove such statements
	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		newList := make([]ast.Stmt, 0, len(block.List))
		for _, stmt := range block.List {
			if isAddControlCallWithID(stmt, controlID) {
				found = true
				continue // Skip this statement to delete it
			}
			newList = append(newList, stmt)
		}

		if found {
			block.List = newList
		}

		return true
	})

	if !found {
		return fmt.Errorf("control %q not found", controlID)
	}
	return nil
}

// isAddControlCallWithID checks if a statement is an AddControl call with the given ID.
func isAddControlCallWithID(stmt ast.Stmt, controlID string) bool {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}

	call, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}

	// Check if it's a method call like arch.AddControl(...)
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != "AddControl" {
		return false
	}

	if len(call.Args) < 1 {
		return false
	}

	// Check first argument (control ID)
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}

	val := strings.Trim(lit.Value, "\"")
	return val == controlID
}

// updateComposedOfInAST updates a ComposedOf call in the AST (description and/or child nodes).
func updateComposedOfInAST(f *ast.File, composedOfID, newDesc string, newChildNodeIDs []string) error {
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if it's a ComposedOf call
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if sel.Sel.Name != "ComposedOf" {
			return true
		}

		if len(call.Args) < 2 {
			return true
		}

		// Check first argument (ID)
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}

		val := strings.Trim(lit.Value, "\"")
		if val != composedOfID {
			return true
		}

		// Update description (second argument) if provided
		if newDesc != "" && len(call.Args) >= 2 {
			call.Args[1] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", newDesc)}
		}

		// Update child nodes if provided
		if len(newChildNodeIDs) > 0 && len(call.Args) >= 4 {
			// Args[2] = container, Args[3] = nodes slice
			// Build new composite literal for nodes
			elts := make([]ast.Expr, len(newChildNodeIDs))
			for i, nodeID := range newChildNodeIDs {
				elts[i] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", nodeID)}
			}
			call.Args[3] = &ast.CompositeLit{
				Type: &ast.ArrayType{Elt: &ast.Ident{Name: "string"}},
				Elts: elts,
			}
		}

		found = true
		return false
	})

	if !found {
		return fmt.Errorf("composed-of %q not found", composedOfID)
	}
	return nil
}

// updateControlInAST updates the description of an AddControl call in the AST.
func updateControlInAST(f *ast.File, controlID, newDesc string) error {
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if it's an AddControl call
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if sel.Sel.Name != "AddControl" {
			return true
		}

		if len(call.Args) < 2 {
			return true
		}

		// Check first argument (ID)
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}

		val := strings.Trim(lit.Value, "\"")
		if val != controlID {
			return true
		}

		// Update description (second argument)
		call.Args[1] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", newDesc)}

		found = true
		return false
	})

	if !found {
		return fmt.Errorf("control %q not found", controlID)
	}
	return nil
}

// findVariableNameForNodeID finds the Go variable name that holds a node with the given unique-id
// by scanning DefineNode calls like: nc.OrderReplica = a.DefineNode("order-database-replica", ...)
func findVariableNameForNodeID(f *ast.File, nodeID string) string {
	var varName string

	ast.Inspect(f, func(n ast.Node) bool {
		if varName != "" {
			return false
		}

		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		if len(assign.Rhs) == 0 {
			return true
		}

		// Check if RHS is a DefineNode call
		call, ok := assign.Rhs[0].(*ast.CallExpr)
		if !ok {
			return true
		}

		if !isDefineNodeCallSimple(call) {
			return true
		}

		// Check if first argument is the nodeID
		if len(call.Args) > 0 {
			if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				if strings.Trim(lit.Value, "\"") == nodeID {
					// Found it! Get the variable name from LHS
					if len(assign.Lhs) > 0 {
						if sel, ok := assign.Lhs[0].(*ast.SelectorExpr); ok {
							varName = sel.Sel.Name
						} else if ident, ok := assign.Lhs[0].(*ast.Ident); ok {
							varName = ident.Name
						}
					}
				}
			}
		}
		return true
	})

	return varName
}

// isDefineNodeCallSimple checks if a call expression is a DefineNode call
func isDefineNodeCallSimple(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == "DefineNode"
}

// deleteRelationshipsReferencingVariable removes all ConnectTo calls that reference the given Go variable name or node ID
func deleteRelationshipsReferencingVariable(f *ast.File, varName string, nodeID string) {
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		newList := make([]ast.Stmt, 0, len(fn.Body.List))
		for _, stmt := range fn.Body.List {
			// Check both ExprStmt and AssignStmt
			shouldDelete := false

			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				if call, ok := exprStmt.X.(*ast.CallExpr); ok {
					if callChainReferencesVariable(call, varName, nodeID) {
						shouldDelete = true
					}
				}
			} else if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
				if len(assignStmt.Rhs) > 0 {
					if call, ok := assignStmt.Rhs[0].(*ast.CallExpr); ok {
						if callChainReferencesVariable(call, varName, nodeID) {
							shouldDelete = true
						}
					}
				}
			}

			if !shouldDelete {
				newList = append(newList, stmt)
			}
		}
		fn.Body.List = newList
	}
}

// callChainReferencesVariable checks if a call expression chain references the given Go variable name or node ID
func callChainReferencesVariable(call *ast.CallExpr, varName string, nodeID string) bool {
	// Use a stack to traverse all call expressions in the chain
	var checkExpr func(expr ast.Expr) bool
	checkExpr = func(expr ast.Expr) bool {
		switch e := expr.(type) {
		case *ast.CallExpr:
			// Special handling for Connect(), Interacts(), and Interface() calls:
			// Check if arguments reference the variable via variableReferenceMatches
			if isRelationshipOrInterfaceCall(e) {
				for _, arg := range e.Args {
					if variableReferenceMatches(arg, varName, nodeID) {
						return true
					}
				}
			}

			// Check other arguments
			for _, arg := range e.Args {
				if checkExpr(arg) {
					return true
				}
			}
			// Check the function/receiver
			return checkExpr(e.Fun)

		case *ast.SelectorExpr:
			// Check if this is n.<varName> or nc.<varName>
			if varName != "" && e.Sel.Name == varName {
				return true
			}
			// Continue checking the receiver
			return checkExpr(e.X)

		case *ast.Ident:
			return varName != "" && e.Name == varName

		case *ast.BasicLit:
			if nodeID != "" && e.Kind == token.STRING {
				val := strings.Trim(e.Value, "\"`")
				return val == nodeID
			}
			return false

		default:
			return false
		}
	}

	return checkExpr(call)
}

// deleteComposedOfReferencingVariable removes or updates ComposedOf calls that reference the given variable or node ID
func deleteComposedOfReferencingVariable(f *ast.File, varName string, nodeID string) {
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		newList := make([]ast.Stmt, 0, len(fn.Body.List))
		for _, stmt := range fn.Body.List {
			exprStmt, ok := stmt.(*ast.ExprStmt)
			if !ok {
				newList = append(newList, stmt)
				continue
			}

			call, ok := exprStmt.X.(*ast.CallExpr)
			if !ok {
				newList = append(newList, stmt)
				continue
			}

			// Check if it's a ComposedOf call
			if isComposedOfCall(call) {
				// 1. Check if the container (parent) matches varName or nodeID -> Delete entire statement
				if composedOfContainerMatches(call, varName, nodeID) {
					continue // Delete statement
				}

				// 2. Check and filter children slice
				if removeComposedOfChild(call, varName, nodeID) {
					// Modified in place, keep the statement
				}
			}

			newList = append(newList, stmt)
		}
		fn.Body.List = newList
	}
}

func isComposedOfCall(call *ast.CallExpr) bool {
	// Walk up chain to find ComposedOf
	curr := call
	for curr != nil {
		sel, ok := curr.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		if sel.Sel.Name == "ComposedOf" {
			return true
		}
		// Move up
		if innerCall, ok := sel.X.(*ast.CallExpr); ok {
			curr = innerCall
		} else {
			break
		}
	}
	return false
}

func getComposedOfBaseCall(call *ast.CallExpr) *ast.CallExpr {
	curr := call
	for curr != nil {
		sel, ok := curr.Fun.(*ast.SelectorExpr)
		if !ok {
			return nil
		}
		if sel.Sel.Name == "ComposedOf" {
			return curr
		}
		if innerCall, ok := sel.X.(*ast.CallExpr); ok {
			curr = innerCall
		} else {
			break
		}
	}
	return nil
}

func composedOfContainerMatches(call *ast.CallExpr, varName string, nodeID string) bool {
	baseCall := getComposedOfBaseCall(call)
	if baseCall == nil || len(baseCall.Args) < 3 {
		return false
	}
	// Container is 3rd argument
	return variableReferenceMatches(baseCall.Args[2], varName, nodeID)
}

// removeComposedOfChild removes the varName or nodeID from the children slice. Returns true if modified.
func removeComposedOfChild(call *ast.CallExpr, varName string, nodeID string) bool {
	baseCall := getComposedOfBaseCall(call)
	if baseCall == nil || len(baseCall.Args) < 4 {
		return false
	}

	comp, ok := baseCall.Args[3].(*ast.CompositeLit)
	if !ok {
		return false
	}

	newElts := make([]ast.Expr, 0, len(comp.Elts))
	modified := false
	for _, elt := range comp.Elts {
		if variableReferenceMatches(elt, varName, nodeID) {
			modified = true
			continue
		}
		newElts = append(newElts, elt)
	}

	if modified {
		comp.Elts = newElts
	}
	return modified
}

// isRelationshipOrInterfaceCall checks if a call expression is a.Connect(), a.Interacts(), or varName.Interface()
func isRelationshipOrInterfaceCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == "Connect" || sel.Sel.Name == "Interacts" || sel.Sel.Name == "Interface"
}

// variableReferenceMatches checks if an expression references varName (e.g. n.OrderReplica or n.OrderReplica.UniqueID) or nodeID
func variableReferenceMatches(expr ast.Expr, varName string, nodeID string) bool {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		// Check "UniqueID" pattern: n.OrderReplica.UniqueID
		if e.Sel.Name == "UniqueID" {
			// Check X
			return variableReferenceMatches(e.X, varName, nodeID)
		}
		// Check direct access: n.OrderReplica
		if varName != "" && e.Sel.Name == varName {
			return true
		}
	case *ast.Ident:
		if varName != "" && e.Name == varName {
			return true
		}
	case *ast.BasicLit:
		if nodeID != "" && e.Kind == token.STRING {
			val := strings.Trim(e.Value, "\"`")
			if val == nodeID {
				return true
			}
		}
	}
	return false
}

// --- AST Helpers ---

// findFunctionInAST searches for a function matching any of the nameParts,
// respecting the order of nameParts as priority.
func findFunctionInAST(f *ast.File, nameParts []string) *ast.FuncDecl {
	for _, part := range nameParts {
		partLower := strings.ToLower(part)
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			nameLower := strings.ToLower(fn.Name.Name)
			if strings.Contains(nameLower, partLower) {
				return fn
			}
		}
	}
	return nil
}

func getReceiverName(fn *ast.FuncDecl, defaultName string) string {
	// 1. Check method receiver: func (a *Architecture) Method()
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		for _, p := range fn.Recv.List {
			if isArchitectureType(p.Type) {
				if len(p.Names) > 0 {
					return p.Names[0].Name
				}
			}
		}
	}

	// 2. Check parameters for *domain.Architecture or *Architecture
	if fn.Type.Params != nil {
		for _, p := range fn.Type.Params.List {
			if isArchitectureType(p.Type) {
				if len(p.Names) > 0 {
					return p.Names[0].Name
				}
			}
		}
	}

	// 3. Also check for local variable assignments like arch := domain.NewArchitecture(...)
	if fn.Body != nil {
		for _, stmt := range fn.Body.List {
			assign, ok := stmt.(*ast.AssignStmt)
			if !ok {
				continue
			}
			for _, rhs := range assign.Rhs {
				if isNewArchitectureCall(rhs) {
					if len(assign.Lhs) > 0 {
						if ident, ok := assign.Lhs[0].(*ast.Ident); ok {
							return ident.Name
						}
					}
				}
			}
		}
	}

	return defaultName
}

func isArchitectureType(expr ast.Expr) bool {
	// Handles Architecture, *Architecture, domain.Architecture, *domain.Architecture
	var target ast.Expr = expr
	if star, ok := expr.(*ast.StarExpr); ok {
		target = star.X
	}

	switch v := target.(type) {
	case *ast.Ident:
		return v.Name == "Architecture"
	case *ast.SelectorExpr:
		return v.Sel.Name == "Architecture"
	}
	return false
}

func isNewArchitectureCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		// Might be a direct call if dot-imported
		if ident, ok := call.Fun.(*ast.Ident); ok {
			return ident.Name == "NewArchitecture"
		}
		return false
	}
	return sel.Sel.Name == "NewArchitecture"
}

func insertStmtBeforeReturn(fn *ast.FuncDecl, stmt ast.Stmt) {
	for i, s := range fn.Body.List {
		if _, ok := s.(*ast.ReturnStmt); ok {
			newList := make([]ast.Stmt, 0, len(fn.Body.List)+1)
			newList = append(newList, fn.Body.List[:i]...)
			newList = append(newList, stmt)
			newList = append(newList, fn.Body.List[i:]...)
			fn.Body.List = newList
			return
		}
	}
	fn.Body.List = append(fn.Body.List, stmt)
}

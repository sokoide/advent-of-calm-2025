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
func (GoASTSyncer) ApplyPatch(src string, ops []domain.PatchOperation) (string, error) {
	fset, f, err := parseSource(src)
	if err != nil {
		return "", err
	}

	for _, op := range ops {
		switch op.Type {
		case domain.PatchUpdateNode:
			if op.Origin == nil {
				log.Printf("Warning: update-node requires origin: %v", op)
				continue
			}
			if err := updateNodeAtLine(f, fset, op); err != nil {
				return "", fmt.Errorf("failed to update node at line %d: %w", op.Origin.Line, err)
			}
		case domain.PatchDeleteNode:
			if op.Origin == nil {
				log.Printf("Warning: delete-node requires origin: %v", op)
				continue
			}
			if op.Origin.LoopVar != "" {
				// Loop node deletion = Decrement loop variable
				if err := updateLoopVariable(f, op.Origin.LoopVar, -1); err != nil {
					return "", fmt.Errorf("failed to decrement loop var %s: %w", op.Origin.LoopVar, err)
				}
			} else {
				// Explicit node deletion = Remove statement
				if err := deleteNodeAtLine(f, fset, op.Origin.Line); err != nil {
					return "", fmt.Errorf("failed to delete node at line %d: %w", op.Origin.Line, err)
				}
			}
		case domain.PatchDeleteRelationship:
			// Delete Connect() call by relationship ID (no origin needed)
			if op.NodeID == "" {
				log.Printf("Warning: delete-relationship requires nodeId (relationship ID)")
				continue
			}
			if err := deleteRelationshipByID(f, op.NodeID); err != nil {
				return "", fmt.Errorf("failed to delete relationship %s: %w", op.NodeID, err)
			}
		case domain.PatchAddRelationship:
			// Add a new Connect() or Interacts() call
			if op.NodeID == "" || op.SourceNode == "" || op.TargetNode == "" {
				log.Printf("Warning: add-relationship requires nodeId, sourceNode, and targetNode")
				continue
			}
			if err := addRelationshipToAST(f, fset, op.NodeID, op.SourceNode, op.TargetNode, op.IsInteracts); err != nil {
				return "", fmt.Errorf("failed to add relationship %s: %w", op.NodeID, err)
			}
		case domain.PatchAddInterface:
			// Add a new Interface() call for a node
			if op.NodeID == "" || op.InterfaceID == "" || op.Protocol == "" {
				log.Printf("Warning: add-interface requires nodeId, interfaceId, and protocol")
				continue
			}
			if err := addInterfaceToAST(f, fset, op.NodeID, op.InterfaceID, op.Protocol); err != nil {
				return "", fmt.Errorf("failed to add interface %s to %s: %w", op.InterfaceID, op.NodeID, err)
			}
		case domain.PatchDeleteInterface:
			// Delete Interface() call
			if op.InterfaceID == "" {
				log.Printf("Warning: delete-interface requires interfaceId")
				continue
			}
			if err := deleteInterfaceFromAST(f, op.InterfaceID); err != nil {
				return "", fmt.Errorf("failed to delete interface %s: %w", op.InterfaceID, err)
			}
		case domain.PatchDeleteFlow:
			// Delete DefineFlow() call
			if op.FlowID == "" {
				log.Printf("Warning: delete-flow requires flowId")
				continue
			}
			if err := deleteFlowFromAST(f, op.FlowID); err != nil {
				return "", fmt.Errorf("failed to delete flow %s: %w", op.FlowID, err)
			}
		case domain.PatchAddFlow:
			if op.FlowID == "" || op.FlowName == "" {
				log.Printf("Warning: add-flow requires flowId and flowName")
				continue
			}
			if err := addFlowToAST(f, op.FlowID, op.FlowName, op.FlowDesc, op.FlowSteps); err != nil {
				return "", fmt.Errorf("failed to add flow %s: %w", op.FlowID, err)
			}
		case domain.PatchUpdateFlow:
			if op.FlowID == "" {
				log.Printf("Warning: update-flow requires flowId")
				continue
			}
			if err := updateFlowInAST(f, op.FlowID, op.FlowName, op.FlowDesc, op.FlowSteps); err != nil {
				return "", fmt.Errorf("failed to update flow %s: %w", op.FlowID, err)
			}
		}
	}

	result, err := formatFile(fset, f)
	if err != nil {
		return "", err
	}

	// If there's a pending relationship to add, insert it into the source
	if pendingRelationship.pending {
		result = insertRelationshipIntoSource(
			result,
			pendingRelationship.id,
			pendingRelationship.source,
			pendingRelationship.target,
			pendingRelationship.isInteracts,
		)
		pendingRelationship.pending = false
	}

	// If there's a pending flow to add, insert it into the source
	if pendingFlow.pending {
		result = insertFlowIntoSource(
			result,
			pendingFlow.id,
			pendingFlow.name,
			pendingFlow.desc,
			pendingFlow.steps,
		)
		pendingFlow.pending = false
	}

	return result, nil
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

		// Check if it's a Connect() or Interacts() call
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		funcName := sel.Sel.Name
		if funcName != "Connect" && funcName != "Interacts" {
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

// addRelationshipToAST records a new Connect() call to be added.
// Since AST insertion is complex, we use source code manipulation approach.
// This function stores the relationship info in a global variable that will be
// appended to the formatted source code at the end of ApplyPatch.
var pendingRelationship struct {
	id, source, target string
	isInteracts        bool
	pending            bool
}

func addRelationshipToAST(
	f *ast.File,
	fset *token.FileSet,
	relationshipID, sourceNode, targetNode string,
	isInteracts bool,
) error {
	// Store the relationship for later source code insertion
	pendingRelationship.id = relationshipID
	pendingRelationship.source = sourceNode
	pendingRelationship.target = targetNode
	pendingRelationship.isInteracts = isInteracts
	pendingRelationship.pending = true

	if isInteracts {
		log.Printf("📝 Adding Interacts: %s from %s to %s", relationshipID, sourceNode, targetNode)
	} else {
		log.Printf("📝 Adding Connect: %s from %s to %s", relationshipID, sourceNode, targetNode)
	}
	return nil
}

// insertRelationshipIntoSource inserts a Connect() or Interacts() call into the Go source code.
// It looks for the wireComponents function and inserts before "return lc".
func insertRelationshipIntoSource(src, relationshipID, sourceNode, targetNode string, isInteracts bool) string {
	// Generate the relationship line to insert
	var relationshipLine string
	if isInteracts {
		relationshipLine = fmt.Sprintf(`
	// GUI-generated interaction: %s
	a.Interacts("%s", "Interaction from %s to %s", "%s", "%s")
`, relationshipID, relationshipID, sourceNode, targetNode, sourceNode, targetNode)
	} else {
		relationshipLine = fmt.Sprintf(`
	// GUI-generated connection: %s
	a.Connect("%s", "Connection from %s to %s", "%s", "%s")
`, relationshipID, relationshipID, sourceNode, targetNode, sourceNode, targetNode)
	}

	// Find "return lc" in wireComponents and insert before it
	pattern := "\treturn lc\n"
	insertPoint := strings.LastIndex(src, pattern)
	if insertPoint == -1 {
		pattern = "return lc"
		insertPoint = strings.LastIndex(src, pattern)
	}

	if insertPoint == -1 {
		log.Printf("Warning: Could not find insertion point in wireComponents")
		return src
	}

	return src[:insertPoint] + relationshipLine + src[insertPoint:]
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

		// Look for the assignment statement containing DefineNode with matching ID
		var insertIndex int = -1
		var varName string

		for i, stmt := range block.List {
			// Check for assignment: varName := a.DefineNode("nodeID", ...) or varName = ...
			// 1. Short Variable Declaration: varName := ...
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
			if insertIndex != -1 {
				break
			}
		}

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
		return fmt.Errorf("definition for node %s not found (must be explicit variable assignment)", nodeID)
	}
	return nil
}

// isDefineNodeCallWithID checks if expr is a DefineNode call for the specific nodeID
func isDefineNodeCallWithID(expr ast.Expr, nodeID string) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	if !isDefineNode(call) {
		return false
	}
	if len(call.Args) < 1 {
		return false
	}
	// Check first arg (ID)
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}
	val := strings.Trim(lit.Value, "\"")
	return val == nodeID
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

// addFlowToAST adds a new DefineFlow call to the source.
// It inserts the code textually into wireComponents for simplicity.
func addFlowToAST(f *ast.File, flowID, name, desc string, steps []string) error {
	pendingFlow.id = flowID
	pendingFlow.name = name
	pendingFlow.desc = desc
	pendingFlow.steps = steps
	pendingFlow.pending = true

	log.Printf("📝 Adding Flow: %s", flowID)
	return nil
}

var pendingFlow struct {
	id, name, desc string
	steps          []string
	pending        bool
}

// insertFlowIntoSource inserts a DefineFlow call into the Go source code.
// It inserts into wireComponents before "return lc".
func insertFlowIntoSource(src string, flowID, name, desc string, steps []string) string {
	stepsCode := ""
	for _, stepID := range steps {
		// Use quotes for ID as it's a string literal in the generated call
		stepsCode += fmt.Sprintf("\t\t\tdomain.StepSpec{ID: \"%s\"},\n", stepID)
	}

	flowCode := fmt.Sprintf(`
	// GUI-generated flow: %s
	a.DefineFlow("%s", "%s", "%s").
		Steps(
%s		)
`, flowID, flowID, name, desc, stepsCode)

	// Find "return lc" in wireComponents and insert before it
	pattern := "\treturn lc\n"
	insertPoint := strings.LastIndex(src, pattern)
	if insertPoint == -1 {
		pattern = "return lc"
		insertPoint = strings.LastIndex(src, pattern)
	}

	if insertPoint == -1 {
		log.Printf("Warning: Could not find insertion point in wireComponents")
		return src
	}

	return src[:insertPoint] + flowCode + src[insertPoint:]
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

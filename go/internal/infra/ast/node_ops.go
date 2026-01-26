package ast

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// AddNodeInAST appends a new DefineNode call to the build or defineNodes function in the AST.
func AddNodeInAST(f *ast.File, nodeID, nodeType, name, desc string) error {
	fn := findFunctionInAST(f, []string{"defineNodes", "build"})
	if fn == nil {
		return fmt.Errorf("defineNodes or build function not found in AST")
	}

	receiverName := getReceiverName(fn, "arch")

	// Create type expression from nodeType parameter
	// Use domain.Service, domain.Database, etc.
	typeExpr := &ast.SelectorExpr{
		X:   ast.NewIdent("domain"),
		Sel: ast.NewIdent(nodeType),
	}

	// Create: <receiver>.DefineNode("id", <Type>, "name", "desc")
	newStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(receiverName),
				Sel: ast.NewIdent("DefineNode"),
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", nodeID)},
				typeExpr,
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", name)},
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", desc)},
			},
		},
	}

	insertStmtBeforeReturn(fn, newStmt)
	return nil
}

// UpdateNodePropertyInAST updates a specific property (name, description, owner, etc.) of a node.
func UpdateNodePropertyInAST(f *ast.File, nodeID, property, value string) error {
	found := false
	var updateErr error
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "DefineNode" {
			return true
		}

		if len(call.Args) < 1 {
			return true
		}
		idLit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || strings.Trim(idLit.Value, "\"") != nodeID {
			return true
		}

		// Handle specific properties
		switch property {
		case "node-type":
			if len(call.Args) >= 2 {
				nodeTypeName, ok := normalizeNodeType(value)
				if !ok {
					updateErr = fmt.Errorf("unsupported node-type %q", value)
					found = true
					return false
				}
				qualifier := ast.NewIdent("domain")
				if sel, ok := call.Args[1].(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok {
						qualifier = ident
					}
				}
				call.Args[1] = &ast.SelectorExpr{
					X:   qualifier,
					Sel: ast.NewIdent(nodeTypeName),
				}
				found = true
			}
		case "name":
			if len(call.Args) >= 3 {
				call.Args[2] = &ast.BasicLit{Kind: idLit.Kind, Value: fmt.Sprintf("%q", value)}
				found = true
			}
		case "description":
			if len(call.Args) >= 4 {
				call.Args[3] = &ast.BasicLit{Kind: idLit.Kind, Value: fmt.Sprintf("%q", value)}
				found = true
			}
		case "owner":
			// owner is usually in WithOwner("owner", "cc") option
			updatedOwner := false
			for _, arg := range call.Args[4:] {
				optCall, ok := arg.(*ast.CallExpr)
				if !ok {
					continue
				}
				if isWithOwnerCall(optCall.Fun) && len(optCall.Args) >= 1 {
					optCall.Args[0] = &ast.BasicLit{Kind: idLit.Kind, Value: fmt.Sprintf("%q", value)}
					updatedOwner = true
					found = true
				}
			}
			if !updatedOwner {
				// Append new WithOwner option: domain.WithOwner("value", "TBD")
				newOption := &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("domain"),
						Sel: ast.NewIdent("WithOwner"),
					},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", value)},
						&ast.BasicLit{Kind: token.STRING, Value: "\"TBD\""},
					},
				}
				call.Args = append(call.Args, newOption)
				found = true
			}
		}

		return !found
	})

	if updateErr != nil {
		return updateErr
	}
	if !found {
		return fmt.Errorf("property %q for node %q not updated in AST", property, nodeID)
	}
	return nil
}

// DeleteNodeInAST removes a DefineNode call with the given id from the AST.
func DeleteNodeInAST(f *ast.File, nodeID string) error {
	var varName string
	// 1. Find the variable name assigned to this nodeID (if any)
	ast.Inspect(f, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, rhs := range assign.Rhs {
			if isDefineNodeCall(rhs, nodeID) {
				if len(assign.Lhs) == 1 {
					varName = exprToSimpleString(assign.Lhs[0])
				}
				return false
			}
		}
		return true
	})

	found := false
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		newList, removed := deleteNodeFromStmtList(fn.Body.List, nodeID, varName)
		fn.Body.List = newList
		if removed {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("node with id %q not found in AST", nodeID)
	}
	return nil
}

func normalizeNodeType(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "service":
		return "Service", true
	case "database":
		return "Database", true
	case "actor":
		return "Actor", true
	case "system":
		return "System", true
	case "queue":
		return "Queue", true
	case "webclient":
		return "WebClient", true
	default:
		return "", false
	}
}

func findDefineNodeTypeExpr(f *ast.File) ast.Expr {
	var found ast.Expr
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "DefineNode" {
			return true
		}
		if len(call.Args) < 2 {
			return true
		}
		found = call.Args[1]
		return false
	})
	return found
}

func defaultNodeTypeExpr(nodeType string) ast.Expr {
	if strings.Contains(nodeType, ".") {
		parts := strings.SplitN(nodeType, ".", 2)
		return &ast.SelectorExpr{
			X:   ast.NewIdent(parts[0]),
			Sel: ast.NewIdent(parts[1]),
		}
	}
	return ast.NewIdent(nodeType)
}

func cloneTypeExpr(expr ast.Expr) ast.Expr {
	switch v := expr.(type) {
	case *ast.Ident:
		return ast.NewIdent(v.Name)
	case *ast.SelectorExpr:
		return &ast.SelectorExpr{
			X:   cloneTypeExpr(v.X),
			Sel: ast.NewIdent(v.Sel.Name),
		}
	default:
		return expr
	}
}

func isWithOwnerCall(fn ast.Expr) bool {
	switch v := fn.(type) {
	case *ast.Ident:
		return v.Name == "WithOwner"
	case *ast.SelectorExpr:
		return v.Sel.Name == "WithOwner"
	default:
		return false
	}
}

func deleteNodeFromStmtList(stmts []ast.Stmt, nodeID string, varName string) ([]ast.Stmt, bool) {
	found := false
	newList := make([]ast.Stmt, 0, len(stmts))

	for _, stmt := range stmts {
		// 1. Remove the DefineNode call
		if stmtDefinesNode(stmt, nodeID) {
			found = true
			continue
		}

		// 2. Remove any statement that references the variable assigned to the deleted node
		if varName != "" && stmtReferencesVariable(stmt, varName) {
			found = true
			continue
		}

		// 3. Remove any statement that references the nodeID as a string literal
		if stmtReferencesStringLiteral(stmt, nodeID) {
			found = true
			continue
		}

		if deleteNodeFromStmt(stmt, nodeID, varName) {
			found = true
		}
		newList = append(newList, stmt)
	}

	return newList, found
}

func stmtReferencesStringLiteral(stmt ast.Stmt, val string) bool {
	referenced := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if referenced {
			return false
		}
		lit, ok := n.(*ast.BasicLit)
		if ok && lit.Kind == token.STRING {
			if strings.Trim(lit.Value, "\"`") == val {
				referenced = true
			}
		}
		return !referenced
	})
	return referenced
}

func stmtReferencesVariable(stmt ast.Stmt, varName string) bool {
	if varName == "" {
		return false
	}
	// Extract base field name if it's a selector (e.g. nc.PaymentSvc -> PaymentSvc)
	parts := strings.Split(varName, ".")
	baseName := parts[len(parts)-1]

	referenced := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if referenced {
			return false
		}
		switch v := n.(type) {
		case *ast.Ident:
			if v.Name == baseName {
				referenced = true
			}
		case *ast.SelectorExpr:
			if v.Sel.Name == baseName {
				referenced = true
			}
		}
		return !referenced
	})
	return referenced
}

func exprToSimpleString(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		x := exprToSimpleString(v.X)
		if x == "" {
			return v.Sel.Name
		}
		return x + "." + v.Sel.Name
	default:
		return ""
	}
}

func deleteNodeFromStmt(stmt ast.Stmt, nodeID string, varName string) bool {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		newList, found := deleteNodeFromStmtList(s.List, nodeID, varName)
		s.List = newList
		return found
	case *ast.ForStmt:
		if s.Body == nil {
			return false
		}
		newList, found := deleteNodeFromStmtList(s.Body.List, nodeID, varName)
		s.Body.List = newList
		return found
	case *ast.RangeStmt:
		if s.Body == nil {
			return false
		}
		newList, found := deleteNodeFromStmtList(s.Body.List, nodeID, varName)
		s.Body.List = newList
		return found
	case *ast.IfStmt:
		found := false
		if s.Body != nil {
			newList, removed := deleteNodeFromStmtList(s.Body.List, nodeID, varName)
			s.Body.List = newList
			found = found || removed
		}
		if s.Else != nil {
			found = found || deleteNodeFromElse(s.Else, nodeID, varName)
		}
		return found
	case *ast.SwitchStmt:
		found := false
		for _, clauseStmt := range s.Body.List {
			clause, ok := clauseStmt.(*ast.CaseClause)
			if !ok {
				continue
			}
			newList, removed := deleteNodeFromStmtList(clause.Body, nodeID, varName)
			clause.Body = newList
			found = found || removed
		}
		return found
	case *ast.TypeSwitchStmt:
		found := false
		for _, clauseStmt := range s.Body.List {
			clause, ok := clauseStmt.(*ast.CaseClause)
			if !ok {
				continue
			}
			newList, removed := deleteNodeFromStmtList(clause.Body, nodeID, varName)
			clause.Body = newList
			found = found || removed
		}
		return found
	case *ast.SelectStmt:
		found := false
		for _, clauseStmt := range s.Body.List {
			clause, ok := clauseStmt.(*ast.CommClause)
			if !ok {
				continue
			}
			newList, removed := deleteNodeFromStmtList(clause.Body, nodeID, varName)
			clause.Body = newList
			found = found || removed
		}
		return found
	case *ast.LabeledStmt:
		return deleteNodeFromStmt(s.Stmt, nodeID, varName)
	default:
		return false
	}
}

func deleteNodeFromElse(elseStmt ast.Stmt, nodeID string, varName string) bool {
	switch stmt := elseStmt.(type) {
	case *ast.BlockStmt:
		newList, found := deleteNodeFromStmtList(stmt.List, nodeID, varName)
		stmt.List = newList
		return found
	case *ast.IfStmt:
		return deleteNodeFromStmt(stmt, nodeID, varName)
	default:
		return false
	}
}

func stmtDefinesNode(stmt ast.Stmt, nodeID string) bool {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		return isDefineNodeCall(s.X, nodeID)
	case *ast.AssignStmt:
		for _, expr := range s.Rhs {
			if isDefineNodeCall(expr, nodeID) {
				return true
			}
		}
		return false
	case *ast.DeclStmt:
		decl, ok := s.Decl.(*ast.GenDecl)
		if !ok {
			return false
		}
		for _, spec := range decl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, val := range valueSpec.Values {
				if isDefineNodeCall(val, nodeID) {
					return true
				}
			}
		}
		return false
	default:
		return false
	}
}

func isDefineNodeCall(expr ast.Expr, nodeID string) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "DefineNode" {
		return false
	}
	if len(call.Args) == 0 {
		return false
	}
	// Try to match nodeID against first argument
	// It could be a literal string or an identifier
	switch arg := call.Args[0].(type) {
	case *ast.BasicLit:
		if arg.Kind == token.STRING {
			val := strings.Trim(arg.Value, "\"`")
			return val == nodeID
		}
	case *ast.Ident:
		// TODO: If the ID is a variable (e.g. 'id' in a loop), we can't easily know its value
		// without more analysis. For now, we only match if the variable name itself
		// matches the nodeID (unlikely for loops, but possible for constants).
		return arg.Name == nodeID
	}
	return false
}

// updateNodeAtLine updates a node at a specific line number.
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

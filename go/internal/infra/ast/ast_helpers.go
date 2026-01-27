package ast

import (
	"go/ast"
	"go/token"
	"strings"
)

// findFunctionInAST searches for a function matching any of the nameParts,
// respecting the order of nameParts as priority.
//
// Uses substring matching (strings.Contains) on lowercase function names.
// For example, searching for ["build"] will match "buildComponents", "rebuild", etc.
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

// getReceiverName returns the receiver name for an Architecture method.
func getReceiverName(fn *ast.FuncDecl, defaultName string) string {
	// 1. Check method receiver: func (a *Architecture) Method()
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		for _, p := range fn.Recv.List {
			if isArchitectureType(p.Type) && len(p.Names) > 0 {
				return p.Names[0].Name
			}
		}
	}

	// 2. Check parameters for *domain.Architecture or *Architecture
	if fn.Type.Params != nil {
		for _, p := range fn.Type.Params.List {
			if isArchitectureType(p.Type) && len(p.Names) > 0 {
				return p.Names[0].Name
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
				if isNewArchitectureCall(rhs) && len(assign.Lhs) > 0 {
					if ident, ok := assign.Lhs[0].(*ast.Ident); ok {
						return ident.Name
					}
				}
			}
		}
	}

	return defaultName
}

// isArchitectureType checks if an expression is an Architecture type.
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

// isNewArchitectureCall checks if an expression is a NewArchitecture call.
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

// insertStmtBeforeReturn inserts a statement before the first return statement in a function.
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

// isDefineNode checks if a call expression is a DefineNode call.
func isDefineNode(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "DefineNode"
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

// isOwnerCall checks if a function expression is an owner-related call (WithOwner, etc).
func isOwnerCall(fun ast.Expr, name string) bool {
	switch v := fun.(type) {
	case *ast.Ident:
		return v.Name == name
	case *ast.SelectorExpr:
		return v.Sel.Name == name
	}
	return false
}

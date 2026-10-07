package coctyl

import (
	"fmt"
	"go/ast"
	"go/token"
)

// Normalizer transforms an AST into a canonical, alpha-normalized AST (Pass 1).
// It resolves variable bindings, scopes, and filters superficial declarations.
type Normalizer struct {
	includeImports bool
	includePackage bool
}

// NewNormalizer creates a new Normalizer configured with options.
func NewNormalizer(opts Options) *Normalizer {
	return &Normalizer{
		includeImports: opts.IncludeImports,
		includePackage: opts.IncludePackage,
	}
}

// Normalize rewrites file into a canonical AST ready for serialization.
func (n *Normalizer) Normalize(file *ast.File) *ast.File {
	// 1. Strip comments
	file.Comments = nil

	// 2. Normalize package declaration
	if !n.includePackage {
		file.Name = ast.NewIdent("_")
	}

	// 3. Filter imports if requested
	if !n.includeImports {
		var filteredDecls []ast.Decl
		for _, decl := range file.Decls {
			if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
				continue
			}
			filteredDecls = append(filteredDecls, decl)
		}
		file.Decls = filteredDecls
	}

	// 4. Build file-level scope for top-level symbols
	fileScope := NewRootScope()
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			fileScope.Define(d.Name.Name)
		case *ast.GenDecl:
			if d.Tok == token.VAR || d.Tok == token.CONST {
				for _, spec := range d.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range vs.Names {
							fileScope.Define(name.Name)
						}
					}
				}
			}
		}
	}

	// 5. Alpha-normalize declarations
	for _, decl := range file.Decls {
		n.normalizeDecl(decl, fileScope)
	}

	return file
}

func (n *Normalizer) normalizeDecl(decl ast.Decl, fileScope *Scope) {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		n.normalizeFuncDecl(d, fileScope)
	case *ast.GenDecl:
		n.normalizeGenDecl(d, fileScope)
	}
}

func (n *Normalizer) normalizeFuncDecl(fn *ast.FuncDecl, fileScope *Scope) {
	// Canonicalize function name in fileScope
	if fileScope != nil {
		if canonical, ok := fileScope.Lookup(fn.Name.Name); ok {
			fn.Name = ast.NewIdent(canonical)
		} else {
			fn.Name = ast.NewIdent("_")
		}
	} else {
		fn.Name = ast.NewIdent("_")
	}

	// Create function scope with independent local variable counter
	fnCounter := 0
	fnScope := &Scope{
		parent:  fileScope,
		symbols: make(map[string]string),
		nextID:  &fnCounter,
	}

	// Receiver
	if fn.Recv != nil {
		n.normalizeFieldList(fn.Recv, fnScope)
	}

	// Signature (TypeParams, Params, Results)
	if fn.Type != nil {
		if fn.Type.TypeParams != nil {
			n.normalizeFieldList(fn.Type.TypeParams, fnScope)
		}
		if fn.Type.Params != nil {
			n.normalizeFieldList(fn.Type.Params, fnScope)
		}
		if fn.Type.Results != nil {
			n.normalizeFieldList(fn.Type.Results, fnScope)
		}
	}

	// Body
	if fn.Body != nil {
		n.normalizeBlockStmt(fn.Body, fnScope)
	}
}

func (n *Normalizer) normalizeGenDecl(gen *ast.GenDecl, scope *Scope) {
	for _, spec := range gen.Specs {
		switch sp := spec.(type) {
		case *ast.ValueSpec:
			// Evaluate values first in current scope
			for i, val := range sp.Values {
				sp.Values[i] = n.normalizeExpr(val, scope)
			}
			if sp.Type != nil {
				sp.Type = n.normalizeExpr(sp.Type, scope)
			}
			// Then bind names in scope
			for i, name := range sp.Names {
				if scope != nil {
					canonical := scope.DefineOrReuse(name.Name)
					sp.Names[i] = ast.NewIdent(canonical)
				}
			}
		case *ast.TypeSpec:
			sp.Type = n.normalizeExpr(sp.Type, scope)
		}
	}
}

func (n *Normalizer) normalizeFieldList(fl *ast.FieldList, scope *Scope) {
	if fl == nil {
		return
	}
	for _, field := range fl.List {
		for i, name := range field.Names {
			canonical := scope.Define(name.Name)
			field.Names[i] = ast.NewIdent(canonical)
		}
		if field.Type != nil {
			field.Type = n.normalizeExpr(field.Type, scope)
		}
	}
}

func (n *Normalizer) normalizeBlockStmt(block *ast.BlockStmt, parentScope *Scope) {
	if block == nil {
		return
	}
	childScope := parentScope.NewChild()
	for i, stmt := range block.List {
		block.List[i] = n.normalizeStmt(stmt, childScope)
	}
}

func (n *Normalizer) normalizeStmt(stmt ast.Stmt, scope *Scope) ast.Stmt {
	if stmt == nil {
		return nil
	}

	switch st := stmt.(type) {
	case *ast.BlockStmt:
		n.normalizeBlockStmt(st, scope)
		return st

	case *ast.AssignStmt:
		if st.Tok == token.DEFINE {
			// RHS evaluated before LHS is defined
			for i, rhs := range st.Rhs {
				st.Rhs[i] = n.normalizeExpr(rhs, scope)
			}
			for i, lhs := range st.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					canonical := scope.DefineOrReuse(ident.Name)
					st.Lhs[i] = ast.NewIdent(canonical)
				} else {
					st.Lhs[i] = n.normalizeExpr(lhs, scope)
				}
			}
		} else {
			// Normal assignment (=, +=, etc.)
			for i, lhs := range st.Lhs {
				st.Lhs[i] = n.normalizeExpr(lhs, scope)
			}
			for i, rhs := range st.Rhs {
				st.Rhs[i] = n.normalizeExpr(rhs, scope)
			}
		}
		return st

	case *ast.ReturnStmt:
		for i, res := range st.Results {
			st.Results[i] = n.normalizeExpr(res, scope)
		}
		return st

	case *ast.IfStmt:
		ifScope := scope.NewChild()
		if st.Init != nil {
			st.Init = n.normalizeStmt(st.Init, ifScope)
		}
		st.Cond = n.normalizeExpr(st.Cond, ifScope)
		n.normalizeBlockStmt(st.Body, ifScope)
		if st.Else != nil {
			st.Else = n.normalizeStmt(st.Else, ifScope)
		}
		return st

	case *ast.ForStmt:
		forScope := scope.NewChild()
		if st.Init != nil {
			st.Init = n.normalizeStmt(st.Init, forScope)
		}
		if st.Cond != nil {
			st.Cond = n.normalizeExpr(st.Cond, forScope)
		}
		if st.Post != nil {
			st.Post = n.normalizeStmt(st.Post, forScope)
		}
		n.normalizeBlockStmt(st.Body, forScope)
		return st

	case *ast.RangeStmt:
		rangeScope := scope.NewChild()
		st.X = n.normalizeExpr(st.X, scope) // X evaluated in parent scope
		if st.Tok == token.DEFINE {
			if st.Key != nil {
				if ident, ok := st.Key.(*ast.Ident); ok {
					st.Key = ast.NewIdent(rangeScope.DefineOrReuse(ident.Name))
				}
			}
			if st.Value != nil {
				if ident, ok := st.Value.(*ast.Ident); ok {
					st.Value = ast.NewIdent(rangeScope.DefineOrReuse(ident.Name))
				}
			}
		} else {
			if st.Key != nil {
				st.Key = n.normalizeExpr(st.Key, rangeScope)
			}
			if st.Value != nil {
				st.Value = n.normalizeExpr(st.Value, rangeScope)
			}
		}
		n.normalizeBlockStmt(st.Body, rangeScope)
		return st

	case *ast.ExprStmt:
		st.X = n.normalizeExpr(st.X, scope)
		return st

	case *ast.DeclStmt:
		n.normalizeDecl(st.Decl, scope)
		return st

	case *ast.IncDecStmt:
		st.X = n.normalizeExpr(st.X, scope)
		return st

	case *ast.BranchStmt:
		return st

	case *ast.SwitchStmt:
		swScope := scope.NewChild()
		if st.Init != nil {
			st.Init = n.normalizeStmt(st.Init, swScope)
		}
		if st.Tag != nil {
			st.Tag = n.normalizeExpr(st.Tag, swScope)
		}
		n.normalizeBlockStmt(st.Body, swScope)
		return st

	case *ast.TypeSwitchStmt:
		tsScope := scope.NewChild()
		if st.Init != nil {
			st.Init = n.normalizeStmt(st.Init, tsScope)
		}
		st.Assign = n.normalizeStmt(st.Assign, tsScope)
		n.normalizeBlockStmt(st.Body, tsScope)
		return st

	case *ast.CaseClause:
		caseScope := scope.NewChild()
		for i, expr := range st.List {
			st.List[i] = n.normalizeExpr(expr, caseScope)
		}
		for i, bodyStmt := range st.Body {
			st.Body[i] = n.normalizeStmt(bodyStmt, caseScope)
		}
		return st

	case *ast.SelectStmt:
		n.normalizeBlockStmt(st.Body, scope)
		return st

	case *ast.CommClause:
		commScope := scope.NewChild()
		if st.Comm != nil {
			st.Comm = n.normalizeStmt(st.Comm, commScope)
		}
		for i, bodyStmt := range st.Body {
			st.Body[i] = n.normalizeStmt(bodyStmt, commScope)
		}
		return st

	case *ast.DeferStmt:
		st.Call = n.normalizeExpr(st.Call, scope).(*ast.CallExpr)
		return st

	case *ast.GoStmt:
		st.Call = n.normalizeExpr(st.Call, scope).(*ast.CallExpr)
		return st

	case *ast.SendStmt:
		st.Chan = n.normalizeExpr(st.Chan, scope)
		st.Value = n.normalizeExpr(st.Value, scope)
		return st

	case *ast.LabeledStmt:
		st.Stmt = n.normalizeStmt(st.Stmt, scope)
		return st

	case *ast.EmptyStmt:
		return st

	default:
		return stmt
	}
}

func (n *Normalizer) normalizeExpr(expr ast.Expr, scope *Scope) ast.Expr {
	if expr == nil {
		return nil
	}

	switch ex := expr.(type) {
	case *ast.Ident:
		return n.normalizeIdent(ex, scope)

	case *ast.BasicLit:
		return ex

	case *ast.BinaryExpr:
		ex.X = n.normalizeExpr(ex.X, scope)
		ex.Y = n.normalizeExpr(ex.Y, scope)
		return ex

	case *ast.UnaryExpr:
		ex.X = n.normalizeExpr(ex.X, scope)
		return ex

	case *ast.CallExpr:
		ex.Fun = n.normalizeExpr(ex.Fun, scope)
		for i, arg := range ex.Args {
			ex.Args[i] = n.normalizeExpr(arg, scope)
		}
		return ex

	case *ast.SelectorExpr:
		// X is evaluated in scope, but Sel is a struct field or package symbol - never a local var!
		ex.X = n.normalizeExpr(ex.X, scope)
		return ex

	case *ast.IndexExpr:
		ex.X = n.normalizeExpr(ex.X, scope)
		ex.Index = n.normalizeExpr(ex.Index, scope)
		return ex

	case *ast.IndexListExpr:
		ex.X = n.normalizeExpr(ex.X, scope)
		for i, idx := range ex.Indices {
			ex.Indices[i] = n.normalizeExpr(idx, scope)
		}
		return ex

	case *ast.SliceExpr:
		ex.X = n.normalizeExpr(ex.X, scope)
		if ex.Low != nil {
			ex.Low = n.normalizeExpr(ex.Low, scope)
		}
		if ex.High != nil {
			ex.High = n.normalizeExpr(ex.High, scope)
		}
		if ex.Max != nil {
			ex.Max = n.normalizeExpr(ex.Max, scope)
		}
		return ex

	case *ast.TypeAssertExpr:
		ex.X = n.normalizeExpr(ex.X, scope)
		if ex.Type != nil {
			ex.Type = n.normalizeExpr(ex.Type, scope)
		}
		return ex

	case *ast.ParenExpr:
		// Unpack parenthesis wrapper
		return n.normalizeExpr(ex.X, scope)

	case *ast.CompositeLit:
		if ex.Type != nil {
			ex.Type = n.normalizeExpr(ex.Type, scope)
		}
		isMapOrSlice := false
		if ex.Type != nil {
			switch ex.Type.(type) {
			case *ast.MapType, *ast.ArrayType:
				isMapOrSlice = true
			}
		}
		for i, elt := range ex.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				if isMapOrSlice {
					kv.Key = n.normalizeExpr(kv.Key, scope)
				} else if ident, ok := kv.Key.(*ast.Ident); ok {
					kv.Key = ast.NewIdent(fmt.Sprintf("key:%s", ident.Name))
				} else {
					kv.Key = n.normalizeExpr(kv.Key, scope)
				}
				kv.Value = n.normalizeExpr(kv.Value, scope)
			} else {
				ex.Elts[i] = n.normalizeExpr(elt, scope)
			}
		}
		return ex

	case *ast.KeyValueExpr:
		if ident, ok := ex.Key.(*ast.Ident); ok {
			ex.Key = ast.NewIdent(fmt.Sprintf("key:%s", ident.Name))
		} else {
			ex.Key = n.normalizeExpr(ex.Key, scope)
		}
		ex.Value = n.normalizeExpr(ex.Value, scope)
		return ex

	case *ast.FuncLit:
		closureCounter := 0
		closureScope := &Scope{
			parent:  scope,
			symbols: make(map[string]string),
			nextID:  &closureCounter,
		}
		if ex.Type != nil {
			if ex.Type.Params != nil {
				n.normalizeFieldList(ex.Type.Params, closureScope)
			}
			if ex.Type.Results != nil {
				n.normalizeFieldList(ex.Type.Results, closureScope)
			}
		}
		n.normalizeBlockStmt(ex.Body, closureScope)
		return ex

	case *ast.StarExpr:
		ex.X = n.normalizeExpr(ex.X, scope)
		return ex

	case *ast.ArrayType:
		if ex.Len != nil {
			ex.Len = n.normalizeExpr(ex.Len, scope)
		}
		ex.Elt = n.normalizeExpr(ex.Elt, scope)
		return ex

	case *ast.MapType:
		ex.Key = n.normalizeExpr(ex.Key, scope)
		ex.Value = n.normalizeExpr(ex.Value, scope)
		return ex

	case *ast.ChanType:
		ex.Value = n.normalizeExpr(ex.Value, scope)
		return ex

	case *ast.StructType:
		if ex.Fields != nil {
			for _, f := range ex.Fields.List {
				if f.Type != nil {
					f.Type = n.normalizeExpr(f.Type, scope)
				}
			}
		}
		return ex

	case *ast.InterfaceType:
		return ex

	case *ast.Ellipsis:
		if ex.Elt != nil {
			ex.Elt = n.normalizeExpr(ex.Elt, scope)
		}
		return ex

	default:
		return expr
	}
}

func (n *Normalizer) normalizeIdent(ident *ast.Ident, scope *Scope) *ast.Ident {
	if ident.Name == "_" {
		return ident
	}

	// 1. Check if it's a declared variable in scope (local or file-level)
	if scope != nil {
		if canonical, found := scope.Lookup(ident.Name); found {
			return ast.NewIdent(canonical)
		}
	}

	// 2. Check if it's a Go predeclared identifier (type, constant, builtin function)
	if IsPredeclared(ident.Name) {
		return ast.NewIdent(fmt.Sprintf("builtin:%s", ident.Name))
	}

	// 3. Otherwise it's a package-level external or type identifier
	return ast.NewIdent(fmt.Sprintf("ident:%s", ident.Name))
}

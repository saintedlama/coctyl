package coctyl

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
)

// Serializer converts a normalized AST into a canonical deterministic S-expression string (Pass 2).
// It is completely stateless, immutable, and re-entrant.
type Serializer struct{}

// NewSerializer creates a new Serializer instance.
func NewSerializer() *Serializer {
	return &Serializer{}
}

// Serialize serializes a normalized ast.File into a canonical string directly.
func (s *Serializer) Serialize(file *ast.File) string {
	var buf bytes.Buffer
	p := &astPrinter{buf: &buf}
	p.writeFile(file)
	return buf.String()
}

// SerializeFile is an alias for Serialize for compatibility.
func (s *Serializer) SerializeFile(file *ast.File) string {
	return s.Serialize(file)
}

// astPrinter renders AST nodes into the canonical S-expression buffer.
type astPrinter struct {
	buf *bytes.Buffer
}

func (p *astPrinter) write(str string) {
	p.buf.WriteString(str)
}

func (p *astPrinter) writeFile(file *ast.File) {
	p.write("(File ")
	p.write(fmt.Sprintf("(Pkg %s) ", file.Name.Name))
	p.write("(Decls")
	for _, decl := range file.Decls {
		p.write(" ")
		p.writeDecl(decl)
	}
	p.write("))")
}

func (p *astPrinter) writeDecl(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		p.writeFuncDecl(d)
	case *ast.GenDecl:
		p.writeGenDecl(d)
	default:
		p.write("(Decl _)")
	}
}

func (p *astPrinter) writeFuncDecl(fn *ast.FuncDecl) {
	p.write(fmt.Sprintf("(FuncDecl (Name %s) ", fn.Name.Name))

	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		p.write("(Recv ")
		for i, field := range fn.Recv.List {
			if i > 0 {
				p.write(" ")
			}
			p.writeField(field)
		}
		p.write(") ")
	}

	p.writeFuncType(fn.Type)

	if fn.Body != nil {
		p.write(" ")
		p.writeBlockStmt(fn.Body)
	}

	p.write(")")
}

func (p *astPrinter) writeGenDecl(gen *ast.GenDecl) {
	p.write(fmt.Sprintf("(GenDecl %s", gen.Tok.String()))
	for _, spec := range gen.Specs {
		p.write(" ")
		switch sp := spec.(type) {
		case *ast.ValueSpec:
			p.write("(ValueSpec")
			for _, val := range sp.Values {
				p.write(" ")
				p.writeExpr(val)
			}
			if sp.Type != nil {
				p.write(" (Type ")
				p.writeExpr(sp.Type)
				p.write(")")
			}
			p.write(" (Names")
			for _, name := range sp.Names {
				p.write(" " + name.Name)
			}
			p.write("))")
		case *ast.TypeSpec:
			p.write(fmt.Sprintf("(TypeSpec %s ", sp.Name.Name))
			p.writeExpr(sp.Type)
			p.write(")")
		case *ast.ImportSpec:
			if sp.Name != nil {
				p.write(fmt.Sprintf("(Import %s %s)", sp.Name.Name, sp.Path.Value))
			} else {
				p.write(fmt.Sprintf("(Import %s)", sp.Path.Value))
			}
		}
	}
	p.write(")")
}

func (p *astPrinter) writeFuncType(fnType *ast.FuncType) {
	p.write("(Sig")

	if fnType.TypeParams != nil && len(fnType.TypeParams.List) > 0 {
		p.write(" (TypeParams")
		for _, field := range fnType.TypeParams.List {
			p.write(" ")
			p.writeField(field)
		}
		p.write(")")
	}

	p.write(" (Params")
	if fnType.Params != nil {
		for _, field := range fnType.Params.List {
			p.write(" ")
			p.writeField(field)
		}
	}
	p.write(")")

	if fnType.Results != nil {
		p.write(" (Results")
		for _, field := range fnType.Results.List {
			p.write(" ")
			p.writeField(field)
		}
		p.write(")")
	}
	p.write(")")
}

func (p *astPrinter) writeField(field *ast.Field) {
	p.write("(Field")
	if len(field.Names) > 0 {
		p.write(" (Names")
		for _, name := range field.Names {
			p.write(" " + name.Name)
		}
		p.write(")")
	}
	if field.Type != nil {
		p.write(" (Type ")
		p.writeExpr(field.Type)
		p.write(")")
	}
	p.write(")")
}

func (p *astPrinter) writeBlockStmt(block *ast.BlockStmt) {
	p.write("(Block")
	for _, stmt := range block.List {
		p.write(" ")
		p.writeStmt(stmt)
	}
	p.write(")")
}

func (p *astPrinter) writeStmt(stmt ast.Stmt) {
	if stmt == nil {
		p.write("nil")
		return
	}

	switch st := stmt.(type) {
	case *ast.BlockStmt:
		p.writeBlockStmt(st)

	case *ast.AssignStmt:
		p.write(fmt.Sprintf("(Assign %s (Lhs", st.Tok.String()))
		for _, lhs := range st.Lhs {
			p.write(" ")
			p.writeExpr(lhs)
		}
		p.write(") (Rhs")
		for _, rhs := range st.Rhs {
			p.write(" ")
			p.writeExpr(rhs)
		}
		p.write("))")

	case *ast.ReturnStmt:
		p.write("(Return")
		for _, res := range st.Results {
			p.write(" ")
			p.writeExpr(res)
		}
		p.write(")")

	case *ast.IfStmt:
		p.write("(If ")
		if st.Init != nil {
			p.write("(Init ")
			p.writeStmt(st.Init)
			p.write(") ")
		}
		p.write("(Cond ")
		p.writeExpr(st.Cond)
		p.write(") (Body ")
		p.writeBlockStmt(st.Body)
		p.write(")")
		if st.Else != nil {
			p.write(" (Else ")
			p.writeStmt(st.Else)
			p.write(")")
		}
		p.write(")")

	case *ast.ForStmt:
		p.write("(For ")
		if st.Init != nil {
			p.write("(Init ")
			p.writeStmt(st.Init)
			p.write(") ")
		}
		if st.Cond != nil {
			p.write("(Cond ")
			p.writeExpr(st.Cond)
			p.write(") ")
		}
		if st.Post != nil {
			p.write("(Post ")
			p.writeStmt(st.Post)
			p.write(") ")
		}
		p.write("(Body ")
		p.writeBlockStmt(st.Body)
		p.write("))")

	case *ast.RangeStmt:
		p.write("(Range (X ")
		p.writeExpr(st.X)
		p.write(fmt.Sprintf(") %s ", st.Tok.String()))
		if st.Key != nil {
			p.write("(Key ")
			p.writeExpr(st.Key)
			p.write(") ")
		}
		if st.Value != nil {
			p.write("(Val ")
			p.writeExpr(st.Value)
			p.write(") ")
		}
		p.write("(Body ")
		p.writeBlockStmt(st.Body)
		p.write("))")

	case *ast.ExprStmt:
		p.write("(ExprStmt ")
		p.writeExpr(st.X)
		p.write(")")

	case *ast.DeclStmt:
		p.write("(DeclStmt ")
		p.writeDecl(st.Decl)
		p.write(")")

	case *ast.IncDecStmt:
		p.write(fmt.Sprintf("(IncDec %s ", st.Tok.String()))
		p.writeExpr(st.X)
		p.write(")")

	case *ast.BranchStmt:
		p.write(fmt.Sprintf("(Branch %s", st.Tok.String()))
		if st.Label != nil {
			p.write(fmt.Sprintf(" %s", st.Label.Name))
		}
		p.write(")")

	case *ast.SwitchStmt:
		p.write("(Switch ")
		if st.Init != nil {
			p.write("(Init ")
			p.writeStmt(st.Init)
			p.write(") ")
		}
		if st.Tag != nil {
			p.write("(Tag ")
			p.writeExpr(st.Tag)
			p.write(") ")
		}
		p.write("(Body ")
		p.writeBlockStmt(st.Body)
		p.write("))")

	case *ast.TypeSwitchStmt:
		p.write("(TypeSwitch ")
		if st.Init != nil {
			p.write("(Init ")
			p.writeStmt(st.Init)
			p.write(") ")
		}
		p.write("(Assign ")
		p.writeStmt(st.Assign)
		p.write(") (Body ")
		p.writeBlockStmt(st.Body)
		p.write("))")

	case *ast.CaseClause:
		p.write("(Case ")
		if len(st.List) > 0 {
			p.write("(List")
			for _, expr := range st.List {
				p.write(" ")
				p.writeExpr(expr)
			}
			p.write(") ")
		} else {
			p.write("(Default) ")
		}
		p.write("(Body")
		for _, bodyStmt := range st.Body {
			p.write(" ")
			p.writeStmt(bodyStmt)
		}
		p.write("))")

	case *ast.SelectStmt:
		p.write("(Select ")
		p.writeBlockStmt(st.Body)
		p.write(")")

	case *ast.CommClause:
		p.write("(Comm ")
		if st.Comm != nil {
			p.write("(Clause ")
			p.writeStmt(st.Comm)
			p.write(") ")
		} else {
			p.write("(Default) ")
		}
		p.write("(Body")
		for _, bodyStmt := range st.Body {
			p.write(" ")
			p.writeStmt(bodyStmt)
		}
		p.write("))")

	case *ast.DeferStmt:
		p.write("(Defer ")
		p.writeExpr(st.Call)
		p.write(")")

	case *ast.GoStmt:
		p.write("(Go ")
		p.writeExpr(st.Call)
		p.write(")")

	case *ast.SendStmt:
		p.write("(Send ")
		p.writeExpr(st.Chan)
		p.write(" ")
		p.writeExpr(st.Value)
		p.write(")")

	case *ast.LabeledStmt:
		p.write(fmt.Sprintf("(Labeled %s ", st.Label.Name))
		p.writeStmt(st.Stmt)
		p.write(")")

	case *ast.EmptyStmt:
		// Omit empty statement

	default:
		p.write("(Stmt _)")
	}
}

func (p *astPrinter) writeExpr(expr ast.Expr) {
	if expr == nil {
		p.write("nil")
		return
	}

	switch ex := expr.(type) {
	case *ast.Ident:
		p.write(ex.Name)

	case *ast.BasicLit:
		p.write(fmt.Sprintf("(Lit %s %s)", ex.Kind.String(), ex.Value))

	case *ast.BinaryExpr:
		p.write(fmt.Sprintf("(Binary %s ", ex.Op.String()))
		p.writeExpr(ex.X)
		p.write(" ")
		p.writeExpr(ex.Y)
		p.write(")")

	case *ast.UnaryExpr:
		p.write(fmt.Sprintf("(Unary %s ", ex.Op.String()))
		p.writeExpr(ex.X)
		p.write(")")

	case *ast.CallExpr:
		p.write("(Call ")
		p.writeExpr(ex.Fun)
		p.write(" (Args")
		for _, arg := range ex.Args {
			p.write(" ")
			p.writeExpr(arg)
		}
		if ex.Ellipsis != token.NoPos {
			p.write(" ...")
		}
		p.write("))")

	case *ast.SelectorExpr:
		p.write("(Sel ")
		p.writeExpr(ex.X)
		p.write(fmt.Sprintf(" .%s)", ex.Sel.Name))

	case *ast.IndexExpr:
		p.write("(Index ")
		p.writeExpr(ex.X)
		p.write(" ")
		p.writeExpr(ex.Index)
		p.write(")")

	case *ast.IndexListExpr:
		p.write("(IndexList ")
		p.writeExpr(ex.X)
		for _, idx := range ex.Indices {
			p.write(" ")
			p.writeExpr(idx)
		}
		p.write(")")

	case *ast.SliceExpr:
		p.write("(Slice ")
		p.writeExpr(ex.X)
		p.write(" ")
		p.writeExpr(ex.Low)
		p.write(" ")
		p.writeExpr(ex.High)
		if ex.Slice3 {
			p.write(" ")
			p.writeExpr(ex.Max)
		}
		p.write(")")

	case *ast.TypeAssertExpr:
		p.write("(TypeAssert ")
		p.writeExpr(ex.X)
		p.write(" ")
		p.writeExpr(ex.Type)
		p.write(")")

	case *ast.ParenExpr:
		p.writeExpr(ex.X)

	case *ast.CompositeLit:
		p.write("(CompLit ")
		p.writeExpr(ex.Type)
		p.write(" (Elts")
		for _, elt := range ex.Elts {
			p.write(" ")
			p.writeExpr(elt)
		}
		p.write("))")

	case *ast.KeyValueExpr:
		p.write("(KV ")
		p.writeExpr(ex.Key)
		p.write(" ")
		p.writeExpr(ex.Value)
		p.write(")")

	case *ast.FuncLit:
		p.write("(FuncLit ")
		p.writeFuncType(ex.Type)
		p.write(" ")
		p.writeBlockStmt(ex.Body)
		p.write(")")

	case *ast.StarExpr:
		p.write("(Star ")
		p.writeExpr(ex.X)
		p.write(")")

	case *ast.ArrayType:
		p.write("(Array ")
		if ex.Len != nil {
			p.writeExpr(ex.Len)
			p.write(" ")
		}
		p.writeExpr(ex.Elt)
		p.write(")")

	case *ast.MapType:
		p.write("(Map ")
		p.writeExpr(ex.Key)
		p.write(" ")
		p.writeExpr(ex.Value)
		p.write(")")

	case *ast.ChanType:
		p.write(fmt.Sprintf("(Chan %d ", ex.Dir))
		p.writeExpr(ex.Value)
		p.write(")")

	case *ast.StructType:
		p.write("(Struct")
		if ex.Fields != nil {
			for _, f := range ex.Fields.List {
				p.write(" (Field")
				for _, name := range f.Names {
					p.write(fmt.Sprintf(" %s", name.Name))
				}
				p.write(" ")
				p.writeExpr(f.Type)
				p.write(")")
			}
		}
		p.write(")")

	case *ast.InterfaceType:
		p.write("(Interface)")

	case *ast.Ellipsis:
		p.write("(Ellipsis ")
		p.writeExpr(ex.Elt)
		p.write(")")

	default:
		p.write("(Expr _)")
	}
}

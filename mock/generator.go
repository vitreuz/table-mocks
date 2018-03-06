package mock

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func GenerateFile(m *Mock, file *os.File) error {
	m.addSyncImport()

	node := m.ToFile()
	fset = token.NewFileSet()
	return format.Node(file, fset, node)
}

func (m *Mock) addSyncImport() {
	m.Imports = append(m.Imports, "sync")
}

func (m Mock) ToFile() *ast.File {
	node := &ast.File{Name: ast.NewIdent(m.Package)}

	if len(m.Imports) > 1 {
		node.Decls = m.toImports()
	}

	for _, ifce := range m.Interfaces {
		node.Decls = append(node.Decls, ifce.GenerateStructs()...)
	}

	return node
}

func (m Mock) toImports() []ast.Decl {
	node := &ast.GenDecl{
		Lparen: 2,
		Tok:    token.IMPORT,
		Specs:  []ast.Spec{},
	}

	for i := range m.Imports {
		imprtSpec := &ast.ImportSpec{
			Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", m.Imports[i])},
		}
		node.Specs = append(node.Specs, imprtSpec)
	}

	return []ast.Decl{node}
}

func (ifce Interface) GenerateStructs() []ast.Decl {
	decls := []ast.Decl{ifce.generateInterfaceStruct()}
	for _, method := range ifce.Methods {
		decls = append(decls, method.generateMethodStruct(ifce.Name))
	}

	return decls
}

func (ifce Interface) GenerateMethods() []ast.Decl {
	// generate Constructor
	decls := []ast.Decl{ifce.generateConstructor()}
	for _, method := range ifce.Methods {
		// generate Returns
		returns := method.generateReturns(ifce.Name)
		decls = append(decls, returns)
	}
	// generate GetArgs
	// generate callback
	// generate ForCall
	return decls
}

func (meth Method) generateReturns(ifceName string) *ast.FuncDecl {
	fakeMethod := ast.NewIdent("fakeMethod")
	fakeMethodField := selectorExpr(ast.NewIdent("fake"), meth.fieldName())
	fakeMethodMutex := selectorExpr(ast.NewIdent("fake"), meth.mutexName())

	recv := &ast.FieldList{
		List: []*ast.Field{{
			Names: []*ast.Ident{ast.NewIdent("fake")},
			Type:  &ast.StarExpr{X: ast.NewIdent(strings.Title(ifceName))},
		}},
	}
	returnsName := strings.Title(meth.Name) + "Returns"
	params := &ast.FieldList{List: []*ast.Field{}}
	for _, ret := range meth.Rets {
		params.List = append(params.List, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(ret.Name + "Result")},
			Type:  ret.Type,
		})
	}

	results := &ast.FieldList{List: []*ast.Field{{
		Type: &ast.StarExpr{X: ast.NewIdent(ifceName)},
	}}}
	body := &ast.BlockStmt{List: []ast.Stmt{
		&ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: selectorExpr(fakeMethodMutex, "Lock"),
			},
		}, &ast.AssignStmt{
			Lhs: expression(fakeMethod),
			Tok: token.DEFINE,
			Rhs: expression(&ast.IndexExpr{
				X:     fakeMethodField,
				Index: &ast.BasicLit{Value: "0"},
			}),
		},
	}}

	for _, ret := range meth.Rets {
		asgn := &ast.AssignStmt{
			Lhs: expression(
				selectorExpr(fakeMethod, ret.fieldName(true)),
			),
			Tok: token.ASSIGN,
			Rhs: expression(ast.NewIdent(ret.Name + "Result")),
		}

		body.List = append(body.List, asgn)
	}

	body.List = append(body.List, []ast.Stmt{
		&ast.AssignStmt{
			Lhs: expression(&ast.IndexExpr{
				X:     fakeMethodField,
				Index: &ast.BasicLit{Value: "0"},
			}),
			Tok: token.ASSIGN,
			Rhs: expression(fakeMethod),
		}, &ast.ExprStmt{
			&ast.CallExpr{
				Fun: selectorExpr(fakeMethodMutex, "Unlock"),
			},
		}, &ast.ReturnStmt{
			Results: expression(ast.NewIdent("fake")),
		},
	}...)

	return generateFuncDecl(recv, returnsName, params, results, body)
}

func (ifce Interface) generateConstructor() *ast.FuncDecl {
	constructorName := "New" + strings.Title(ifce.Name)
	params := &ast.FieldList{}
	results := &ast.FieldList{
		List: []*ast.Field{{
			Type: &ast.StarExpr{X: ast.NewIdent(strings.Title(ifce.Name))},
		}},
	}
	body := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.AssignStmt{
				Lhs: expression(ast.NewIdent("fake")),
				Tok: token.DEFINE,
				Rhs: expression(compositeLit(strings.Title(ifce.Name), true)),
			},
		},
	}

	for _, method := range ifce.Methods {
		stmt := &ast.AssignStmt{
			Lhs: expression(
				selectorExpr(ast.NewIdent("fake"), toMethodName(method.Name, "Method")),
			),
			Tok: token.ASSIGN,
			Rhs: expression(
				&ast.CallExpr{
					Fun: ast.NewIdent("make"),
					Args: expression(&ast.MapType{
						Key:   ast.NewIdent("int"),
						Value: ast.NewIdent(toMethodStructName(ifce.Name, method.Name)),
					}),
				},
			),
		}
		body.List = append(body.List, stmt)
	}

	body.List = append(body.List, &ast.ReturnStmt{
		Results: expression(ast.NewIdent("fake"))},
	)

	return generateFuncDecl(nil, constructorName, params, results, body)
}

func generateFuncDecl(recv *ast.FieldList, name string, params, results *ast.FieldList, body *ast.BlockStmt) *ast.FuncDecl {
	return &ast.FuncDecl{
		Recv: recv,
		Name: ast.NewIdent(name),
		Type: &ast.FuncType{Params: params, Results: results},
		Body: body,
	}
}

func expression(expr ...ast.Expr) []ast.Expr { return expr }
func compositeLit(name string, deref bool) *ast.UnaryExpr {
	expr := &ast.UnaryExpr{X: &ast.CompositeLit{Type: ast.NewIdent(name)}}
	if deref {
		expr.Op = token.AND
	}
	return expr
}
func selectorExpr(x ast.Expr, sel string) *ast.SelectorExpr {
	return &ast.SelectorExpr{X: x, Sel: ast.NewIdent(sel)}
}

func (ifce Interface) generateInterfaceStruct() ast.Decl {
	fieldList := []*ast.Field{}
	for _, method := range ifce.Methods {
		methField := &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(toMethodName(method.Name, "Method"))},
			Type: &ast.MapType{
				Key:   ast.NewIdent("int"),
				Value: ast.NewIdent(toMethodStructName(ifce.Name, method.Name)),
			},
		}
		methMutex := &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(toMethodName(method.Name, "Mutex"))},
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent("sync"),
				Sel: ast.NewIdent("RWMutex"),
			},
		}

		fieldList = append(fieldList, methField, methMutex)
	}

	return generateStruct(ifce.Name, fieldList)
}

func (meth Method) generateMethodStruct(ifceName string) ast.Decl {
	fieldList := []*ast.Field{}
	for i, arg := range meth.Args {
		fieldList = append(fieldList, arg.generateValue(i, false))
	}

	fieldList = append(fieldList, &ast.Field{
		Names: []*ast.Ident{ast.NewIdent("Called")},
		Type:  ast.NewIdent("bool"),
	})

	for i, res := range meth.Rets {
		fieldList = append(fieldList, res.generateValue(i, true))
	}

	return generateStruct(toMethodStructName(ifceName, meth.Name), fieldList)
}

func (value Value) generateValue(i int, isResult bool) *ast.Field {
	suffix := "Arg"
	if isResult {
		suffix = "Result"
	}
	name := strings.Title(value.Name) + suffix
	if value.Name == "" {
		panic("no name")
	}

	return &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(name)},
		Type:  value.Type,
	}
}

func (value Value) fieldName(isResult bool) string {
	suffix := "Arg"
	if isResult {
		suffix = "Result"
	}

	if value.Repeat == 0 {
		return strings.Title(value.Name) + suffix
	}
	return strings.Title(value.Name) + suffix + strconv.Itoa(value.Repeat)
}

func (value Value) argName(isResult bool) string {
	suffix := "Arg"
	if isResult {
		suffix = "Result"
	}

	if value.Repeat == 0 {
		return value.Name + suffix
	}
	return value.Name + suffix + strconv.Itoa(value.Repeat)
}

func (method Method) fieldName() string {
	return toMethodName(method.Name, "Method")
}
func (method Method) mutexName() string {
	return toMethodName(method.Name, "Mutex")
}

func toMethodName(name, suffix string) string {
	if name == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(name)
	return string(unicode.ToLower(r)) + name[n:] + suffix

}

func toMethodStructName(ifceName, methodName string) string {
	return strings.Title(ifceName) + strings.Title(methodName) + "Method"
}

func generateStruct(name string, fieldList []*ast.Field) *ast.GenDecl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(name),
				Type: &ast.StructType{Fields: &ast.FieldList{List: fieldList}},
			},
		},
	}
}

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

func GenerateFile(ifce *Interface, pkg string, file *os.File) error {
	ifce.addSyncImport()

	node := ifce.ToFile(pkg)
	fset = token.NewFileSet()
	return format.Node(file, fset, node)
}

func (m *Interface) addSyncImport() {
	m.Imports = append(m.Imports, "sync")
}

func (ifce Interface) ToFile(pkg string) *ast.File {
	node := &ast.File{Name: ast.NewIdent("fake")}

	if len(ifce.Imports) > 1 {
		node.Decls = ifce.toImports()
	}

	node.Decls = append(node.Decls, ifce.GenerateStructs()...)
	node.Decls = append(node.Decls, ifce.GenerateMethods()...)

	return node
}

func (m Interface) toImports() []ast.Decl {
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
		// generate interfaceMethod
		ifceMethod := method.generateInterfaceMethod(ifce.Name)
		// generate Returns
		returns := method.generateReturns(ifce.Name)
		// generate GetArgs
		getArgs := method.generateGetArgs(ifce.Name)
		// generate callback
		callbck := method.generateCallback(ifce.Name)
		// generate ForCall
		forCall := method.generateForCall(ifce.Name)
		decls = append(decls, ifceMethod, returns, getArgs, callbck, forCall)
	}
	return decls
}

func (meth Method) generateInterfaceMethod(ifceName string) *ast.FuncDecl {
	fake := ast.NewIdent("fake")
	fakeMethod := ast.NewIdent("fakeMethod")
	fakeMethodField := selectorExpr(fake, meth.fieldName())
	fakeMethodMutex := selectorExpr(fake, meth.mutexName())
	fakeMethodCalls := selectorExpr(fake, meth.callsName())

	body := blockStmt(
		&ast.ExprStmt{
			X: call(selectorExpr(fakeMethodMutex, "Lock")),
		},
		&ast.AssignStmt{
			Lhs: expression(fakeMethod),
			Rhs: expression(&ast.IndexExpr{X: fakeMethodField, Index: fakeMethodCalls}),
			Tok: token.DEFINE,
		},
	)

	params := fieldList()
	for _, arg := range meth.Args {
		params.List = append(params.List, arg.variable(false))
		body.List = append(body.List, arg.assignToField(fakeMethod, false))
	}

	body.List = append(body.List, []ast.Stmt{
		&ast.AssignStmt{
			Lhs: expression(&ast.IndexExpr{X: fakeMethodField, Index: fakeMethodCalls}),
			Rhs: expression(fakeMethod),
			Tok: token.ASSIGN,
		},
		&ast.IncDecStmt{
			X:   fakeMethodCalls,
			Tok: token.INC,
		},
		&ast.ExprStmt{
			X: call(selectorExpr(fakeMethodMutex, "Unlock")),
		},
	}...)

	results := fieldList()
	returns := &ast.ReturnStmt{}
	for _, ret := range meth.Rets {
		results.List = append(results.List, ret.variable(true))
		returns.Results = append(returns.Results, selectorExpr(fakeMethod, ret.fieldName(true)))
	}

	body.List = append(body.List, returns)
	recv := field(starExpr(strings.Title(ifceName)), "fake")
	funcName := meth.Name

	return funcDecl(recv, funcName, params, results, body)
}

func (ifce Interface) generateConstructor() *ast.FuncDecl {
	fake := ast.NewIdent("fake")

	body := blockStmt(
		&ast.AssignStmt{
			Lhs: expression(fake),
			Tok: token.DEFINE,
			Rhs: expression(compositeLit(strings.Title(ifce.Name), true)),
		},
	)
	for _, method := range ifce.Methods {
		field := method.fieldName()
		methodStruct := method.structName(ifce.Name)

		asgn := &ast.AssignStmt{
			Lhs: expression(selectorExpr(fake, field)),
			Tok: token.ASSIGN,
			Rhs: expression(call(ast.NewIdent("make"), simpleMap("int", methodStruct))),
		}
		body.List = append(body.List, asgn)
	}
	body.List = append(body.List, &ast.ReturnStmt{Results: expression(fake)})

	funcName := "New" + strings.Title(ifce.Name)
	params := fieldList()
	results := fieldList(field(starExpr(ifce.Name)))

	return funcDecl(nil, funcName, params, results, body)
}

func (meth Method) generateReturns(ifceName string) *ast.FuncDecl {
	// Define resued ast.Nodes
	fakeMethod := ast.NewIdent("fakeMethod")
	fakeMethodField := selectorExpr(ast.NewIdent("fake"), meth.fieldName())
	fakeMethodMutex := selectorExpr(ast.NewIdent("fake"), meth.mutexName())

	params := fieldList()
	body := blockStmt(
		&ast.ExprStmt{
			X: call(selectorExpr(fakeMethodMutex, "Lock")),
		},
		&ast.AssignStmt{
			Lhs: expression(fakeMethod),
			Rhs: expression(valueIndex(fakeMethodField, "0")),
			Tok: token.DEFINE,
		},
	)

	for _, ret := range meth.Rets {
		params.List = append(params.List, field(ret.Type, ret.argName(true)))
		body.List = append(body.List, ret.assignToField(fakeMethod, true))
	}

	body.List = append(body.List, []ast.Stmt{
		&ast.AssignStmt{
			Lhs: expression(valueIndex(fakeMethodField, "0")),
			Tok: token.ASSIGN,
			Rhs: expression(fakeMethod),
		},
		&ast.ExprStmt{
			X: call(selectorExpr(fakeMethodMutex, "Unlock")),
		},
		&ast.ReturnStmt{
			Results: expression(ast.NewIdent("fake")),
		},
	}...)

	recv := field(starExpr(strings.Title(ifceName)), "fake")
	name := strings.Title(meth.Name) + "Returns"
	results := fieldList(field(starExpr(ifceName)))

	return funcDecl(recv, name, params, results, body)
}

func (val Value) assignToField(method ast.Expr, isResult bool) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: expression(selectorExpr(method, val.fieldName(isResult))),
		Rhs: expression(ast.NewIdent(val.argName(isResult))),
		Tok: token.ASSIGN,
	}
}

func (val Value) assignToVar(method ast.Expr, isResult bool) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: expression(ast.NewIdent(val.argName(isResult))),
		Rhs: expression(selectorExpr(method, val.fieldName(isResult))),
		Tok: token.ASSIGN,
	}
}

func (meth Method) generateGetArgs(ifceName string) *ast.FuncDecl {
	fakeMethodField := selectorExpr(ast.NewIdent("fake"), meth.fieldName())
	fakeMethodMutex := selectorExpr(ast.NewIdent("fake"), meth.mutexName())

	body := blockStmt(&ast.ExprStmt{X: call(selectorExpr(fakeMethodMutex, "RLock"))})
	results := fieldList()
	returns := expression()

	for _, arg := range meth.Args {
		results.List = append(results.List, field(arg.Type, arg.argName(false)))
		returns = append(returns, ast.NewIdent(arg.argName(false)))
		body.List = append(body.List, arg.assignToVar(valueIndex(fakeMethodField, "0"), false))
	}
	body.List = append(body.List, []ast.Stmt{
		&ast.ExprStmt{X: call(selectorExpr(fakeMethodMutex, "RUnlock"))},
		&ast.ReturnStmt{Results: returns},
	}...)

	recv := field(starExpr(strings.Title(ifceName)), "fake")
	funcName := strings.Title(meth.Name) + "GetArgs"
	params := fieldList()

	return funcDecl(recv, funcName, params, results, body)
}

func (meth Method) generateCallback(ifceName string) *ast.GenDecl {
	fnIdent := ast.NewIdent(meth.structName(ifceName))

	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(meth.funcName(ifceName)),
				Type: &ast.FuncType{
					Params:  fieldList(field(fnIdent)),
					Results: fieldList(field(fnIdent)),
				},
			},
		},
	}
}

func (meth Method) assignFromMap(index string, expr ast.Expr, define bool) *ast.AssignStmt {
	fakeMethodField := selectorExpr(ast.NewIdent("fake"), meth.fieldName())
	fakeMethodMap := valueIndex(fakeMethodField, index)
	tok := token.ASSIGN
	if define {
		tok = token.DEFINE
	}

	return &ast.AssignStmt{
		Lhs: expression(expr),
		Rhs: expression(fakeMethodMap),
		Tok: tok,
	}
}

func (meth Method) assignToMap(index string, expr ast.Expr) *ast.AssignStmt {
	fakeMethodField := selectorExpr(ast.NewIdent("fake"), meth.fieldName())
	fakeMethodMap := valueIndex(fakeMethodField, index)

	return &ast.AssignStmt{
		Lhs: expression(fakeMethodMap),
		Rhs: expression(expr),
		Tok: token.ASSIGN,
	}
}

func (meth Method) generateForCall(ifceName string) *ast.FuncDecl {
	fakeMethod := ast.NewIdent("fakeMethod")
	fakeMethodMutex := selectorExpr(ast.NewIdent("fake"), meth.mutexName())

	body := blockStmt(
		&ast.ExprStmt{X: call(selectorExpr(fakeMethodMutex, "Lock"))},
		&ast.RangeStmt{
			Key: ast.NewIdent("_"), Value: ast.NewIdent("fn"),
			Tok: token.DEFINE,
			X:   ast.NewIdent("fns"),
			Body: blockStmt(
				meth.assignFromMap("call", fakeMethod, true),
				meth.assignToMap("call", call(ast.NewIdent("fn"), fakeMethod)),
			),
		},
		&ast.ExprStmt{X: call(selectorExpr(fakeMethodMutex, "Unlock"))},
		&ast.ReturnStmt{Results: expression(ast.NewIdent("fake"))},
	)
	recv := field(starExpr(strings.Title(ifceName)), "fake")
	funcName := strings.Title(meth.Name) + "ForCall"
	params := fieldList(
		field(ast.NewIdent("int"), "call"),
		field(&ast.Ellipsis{Elt: ast.NewIdent(meth.funcName(ifceName))}, "fns"),
	)
	results := fieldList(field(starExpr(ifceName)))

	return funcDecl(recv, funcName, params, results, body)
}

func (ifce Interface) generateInterfaceStruct() ast.Decl {
	fieldList := []*ast.Field{}
	for _, method := range ifce.Methods {
		methField := field(
			simpleMap("int", method.structName(ifce.Name)),
			method.fieldName(),
		)
		methMutex := field(
			selectorExpr(ast.NewIdent("sync"), "RWMutex"),
			method.mutexName(),
		)

		fieldList = append(fieldList, methField, methMutex)
	}

	return generateStruct(ifce.Name, fieldList)
}

func (meth Method) generateMethodStruct(ifceName string) ast.Decl {
	fieldList := []*ast.Field{}
	for _, arg := range meth.Args {
		fieldList = append(fieldList, arg.field(false))
	}
	fieldList = append(fieldList, field(ast.NewIdent("bool"), "Called"))
	for _, res := range meth.Rets {
		fieldList = append(fieldList, res.field(true))
	}

	return generateStruct(toMethodStructName(ifceName, meth.Name), fieldList)
}

func (value Value) field(isResult bool) *ast.Field {
	suffix := "Arg"
	if isResult {
		suffix = "Result"
	}
	name := strings.Title(value.Name) + suffix
	if value.Repeat != 0 {
		name += strconv.Itoa(value.Repeat)
	}

	return &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(name)},
		Type:  value.Type,
	}
}

func (value Value) variable(isResult bool) *ast.Field {
	suffix := "Arg"
	if isResult {
		suffix = "Result"
	}
	name := value.Name + suffix
	if value.Repeat != 0 {
		name += strconv.Itoa(value.Repeat)
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
func (method Method) callsName() string {
	return toMethodName(method.Name, "Calls")
}

func (method Method) structName(ifceName string) string {
	return strings.Title(ifceName) + strings.Title(method.Name) + "Method"
}

func (method Method) funcName(ifceName string) string {
	return strings.Title(ifceName) + strings.Title(method.Name) + "Func"
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

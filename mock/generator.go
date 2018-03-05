package mock

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
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

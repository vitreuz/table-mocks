package mock

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
)

// Mock holds an array of all of the interfaces within a file.
type Mock struct {
	Package    string
	Interfaces []Interface
}

// Interface represents a single instance of an interface.
type Interface struct {
	Name    string
	Imports []string
	Methods []Method
}

// Method represents a single interface method with all of its args and return
// values.
type Method struct {
	Name string
	Args []Value
	Rets []Value
}

// Value represents an arg or return value.
type Value struct {
	Name       string
	Repeat     int
	Type       ast.Expr
	IsVariadic bool
}

var fset *token.FileSet

// ReadFile is the primary parser for a file to get mocked. This method walks
// along the file ast to create a Mock object. Interfaces with embedded fields
// outside of this file is not currently supported.
func ReadFile(reader io.Reader) *Mock {
	fset = token.NewFileSet()
	node, err := parser.ParseFile(fset, "", reader, 0)
	if err != nil {
		panic(err)
	}

	mock := new(Mock)
	ast.Print(fset, node)

	for _, d := range genDecls(node) {
		specToks := interfaceSpecTokens(d)

		for _, specTok := range specToks {
			mock.Interfaces = append(mock.Interfaces, parseInterfaceToken(specTok))
		}
	}

	return mock
}

func genDecls(node *ast.File) []*ast.GenDecl {
	toks := []*ast.GenDecl{}

	for _, d := range node.Decls {
		if gd, ok := d.(*ast.GenDecl); ok {
			toks = append(toks, gd)
		}
	}

	return toks
}

// interfaceTokens returns ast.TypeSpec becuase the name of the interface can
// only be pulled from the TypeSpec.
func interfaceSpecTokens(node *ast.GenDecl) []*ast.TypeSpec {
	toks := []*ast.TypeSpec{}

	for _, spec := range node.Specs {
		if tspec, ok := spec.(*ast.TypeSpec); ok {
			if _, ok := tspec.Type.(*ast.InterfaceType); ok {
				toks = append(toks, tspec)
			}
		}
	}

	return toks
}

func parseInterfaceToken(tok *ast.TypeSpec) Interface {
	itfcTok := tok.Type.(*ast.InterfaceType)
	methods := []Method{}

	for _, methTok := range itfcTok.Methods.List {
		if specType, ok := embeddedInterface(methTok); ok {
			embedded := parseInterfaceToken(specType)
			methods = append(methods, embedded.Methods...)
			continue
		}
		methods = append(methods, parseMethodToken(methTok))
	}

	return Interface{Name: tok.Name.Name, Methods: methods}
}

// embeddedInterface returns any embedded interfaces or false otherwise. An
// embedded interface can be determined by checking the Names attribute. Methods
// have a Names tokens, whereas embedded interfaces do not.
func embeddedInterface(tok *ast.Field) (*ast.TypeSpec, bool) {
	if len(tok.Names) > 0 {
		return nil, false
	}

	switch tokType := tok.Type.(type) {
	case *ast.Ident:
		if specTok, ok := tokType.Obj.Decl.(*ast.TypeSpec); ok {
			_, ok := specTok.Type.(*ast.InterfaceType)
			return specTok, ok
		}
	case *ast.SelectorExpr:

		// ifaceName := tokType.Sel.Name

	}

	return nil, false
}

func parseMethodToken(tok *ast.Field) Method {
	var method Method

	for _, idenTok := range tok.Names {
		method.Name = idenTok.Name
	}

	if funcTok, ok := tok.Type.(*ast.FuncType); ok {
		method.Args, method.Rets = parseFuncToken(funcTok)
	}

	return method
}

func parseFuncToken(tok *ast.FuncType) ([]Value, []Value) {
	args := []Value{}
	for ai, arg := range tok.Params.List {
		args = append(args, parseFieldToken(arg, "arg", ai)...)
	}

	if tok.Results == nil {
		return args, nil
	}
	rets := []Value{}
	for ri, ret := range tok.Results.List {
		rets = append(rets, parseFieldToken(ret, "ret", ri)...)
	}

	return args, rets
}

func parseFieldToken(tok *ast.Field, fieldType string, i int) []Value {
	value := []Value{}

	_, isVar := parseType(tok.Type)

	if len(tok.Names) == 0 {
		return append(value, Value{
			Name: fmt.Sprintf("%s%d", fieldType, i+1),
			// Type:       typ,
			IsVariadic: isVar,
		})
	}
	for _, idenTok := range tok.Names {
		value = append(value, Value{
			Name: idenTok.Name,
			// Type:       typ,
			IsVariadic: isVar,
		})
	}

	return value
}

func parseType(tok ast.Expr) (string, bool) {
	switch typeTok := tok.(type) {
	case *ast.Ident:
		return typeTok.Name, false
	case *ast.Ellipsis:
		typ, _ := parseType(typeTok.Elt)
		return typ, true
	case *ast.SelectorExpr:
		pack, _ := parseType(typeTok.X)
		return fmt.Sprintf("%s.%s", pack, typeTok.Sel.Name), false
	case *ast.ArrayType:
		typ, _ := parseType(typeTok.Elt)
		return fmt.Sprintf("[]%s", typ), false
	}

	return "", false
}

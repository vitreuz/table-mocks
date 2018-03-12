package mock

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"strings"
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
	Name string
	Type ast.Expr
}

var fset *token.FileSet

type fileReader struct{}

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

	// ast.Print(fset, node)

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

	unnameArgs := make(map[string]repeat)
	args := []Value{}
	for ai, argTok := range tok.Params.List {
		args = append(args, parseFieldToken(argTok, "Arg", unnameArgs, ai)...)
	}

	if tok.Results == nil {
		return args, nil
	}
	var rets []Value
	unnamedRets := make(map[string]repeat)
	for ri, ret := range tok.Results.List {
		rets = append(rets, parseFieldToken(ret, "Result", unnamedRets, ri)...)
	}

	for _, arg := range unnameArgs {
		if arg.repeats {
			args[arg.first].Name += "1"
		}
	}
	for _, ret := range unnamedRets {
		if ret.repeats {
			rets[ret.first].Name += "1"
		}
	}
	return args, rets
}

type repeat struct {
	repeats bool
	count   int
	first   int
}

func parseFieldToken(tok *ast.Field, fieldType string, unnamed map[string]repeat, i int) []Value {
	value := []Value{}

	valName, valType := parseType(tok.Type)

	if len(tok.Names) == 0 {
		repeat, ok := unnamed[valName]
		if !ok {
			repeat.first = i
			repeat.count++
			unnamed[valName] = repeat
			return append(value, Value{
				Name: fmt.Sprintf("%s%s", valName, fieldType),
				Type: valType,
			})
		}
		repeat.repeats = true
		repeat.count++
		unnamed[valName] = repeat

		return append(value, Value{
			Name: fmt.Sprintf("%s%s%d", valName, fieldType, repeat.count),
			Type: valType,
		})
	}
	for _, idenTok := range tok.Names {
		value = append(value, Value{
			Name: idenTok.Name,
			Type: valType,
		})
	}

	return value
}

func parseType(tok ast.Expr) (string, ast.Expr) {
	switch typeTok := tok.(type) {
	case *ast.Ident:
		name := typeTok.Name
		if name == "error" {
			name = "err"
		}
		return name, ast.NewIdent(typeTok.Name)
	case *ast.Ellipsis:
		name, expr := parseType(typeTok.Elt)
		return name + "Var", &ast.Ellipsis{Elt: expr}
	case *ast.SelectorExpr:
		_, expr := parseType(typeTok.X)
		return lowerFirst(typeTok.Sel.Name), &ast.SelectorExpr{X: expr, Sel: ast.NewIdent(typeTok.Sel.Name)}
	case *ast.ArrayType:
		name, expr := parseType(typeTok.Elt)
		if !strings.HasSuffix(name, "Arr") {
			name += "Arr"
		}
		return name, &ast.ArrayType{Elt: expr}
	}

	return "", nil
}

package mock

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"sort"
	"strings"

	"gopkg.in/Sirupsen/logrus.v0"
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

type packageParser struct {
	pkg *ast.Ident
}

func NewPackageParser(pkg *ast.Ident) *packageParser {
	return &packageParser{pkg: pkg}
}

type fileReader struct{}

func ReadPkg(dir string) *Mock {
	logrus.WithField("dir", dir).Println("reading dir")

	fset = token.NewFileSet()
	noTests := func(f os.FileInfo) bool { return !strings.Contains(f.Name(), "_test") }
	pkgs, err := parser.ParseDir(fset, dir, noTests, parser.AllErrors)
	if err != nil {
		panic(err)
	}

	// Don't know what this would mean. Should only have one package per read.
	if len(pkgs) > 1 {
		panic("too many packages")
	}

	mock := new(Mock)
	for pkgName, pkg := range pkgs {
		logrus.WithFields(logrus.Fields{
			"pkg_name":   pkgName,
			"file_count": len(pkg.Files),
		}).Println("parsing package")

		pkg.Scope = ast.NewScope(nil)

		// ast.Print(fset, pkg.Scope)

		var files []string
		for fname, node := range pkg.Files {
			files = append(files, fname)

			for identity, obj := range node.Scope.Objects {
				pkg.Scope.Objects[identity] = obj
			}
		}
		sort.Strings(files)

		for _, fname := range files {
			logrus.WithField("file_name", fname).Println("parings file")

			node := pkg.Files[fname]
			pp := NewPackageParser(node.Name)

			// TODO: find a way to pass the scope to fix embedded interfaces
			for _, d := range genDecls(node) {
				specToks := interfaceSpecTokens(d)

				for _, specTok := range specToks {
					mock.Interfaces = append(mock.Interfaces, pp.parseInterfaceToken(specTok, pkg.Scope))
				}
			}
		}
	}

	return mock
}

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

		pkg := new(packageParser)
		for _, specTok := range specToks {
			mock.Interfaces = append(mock.Interfaces, pkg.parseInterfaceToken(specTok, nil))
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

func (pkg packageParser) parseInterfaceToken(tok *ast.TypeSpec, scope *ast.Scope) Interface {
	itfcTok := tok.Type.(*ast.InterfaceType)
	methods := []Method{}

	for _, methTok := range itfcTok.Methods.List {
		if specType, ok := embeddedInterface(methTok, scope); ok {
			embedded := pkg.parseInterfaceToken(specType, scope)
			methods = append(methods, embedded.Methods...)
			continue
		}
		methods = append(methods, pkg.parseMethodToken(methTok))
	}

	return Interface{Name: tok.Name.Name, Methods: methods}
}

// embeddedInterface returns any embedded interfaces or false otherwise. An
// embedded interface can be determined by checking the Names attribute. Methods
// have a Names tokens, whereas embedded interfaces do not.
func embeddedInterface(tok *ast.Field, scope *ast.Scope) (*ast.TypeSpec, bool) {
	if len(tok.Names) > 0 {
		return nil, false
	}

	switch tokType := tok.Type.(type) {
	case *ast.Ident:
		obj := tokType.Obj
		if tokType.Obj == nil {
			obj = scope.Objects[tokType.Name]
		}

		if specTok, ok := obj.Decl.(*ast.TypeSpec); ok {
			_, ok := specTok.Type.(*ast.InterfaceType)
			return specTok, ok
		}
	case *ast.SelectorExpr:

		// ifaceName := tokType.Sel.Name

	}

	return nil, false
}

func (pkg packageParser) parseMethodToken(tok *ast.Field) Method {
	var method Method

	for _, idenTok := range tok.Names {
		method.Name = idenTok.Name
	}

	if funcTok, ok := tok.Type.(*ast.FuncType); ok {
		method.Args, method.Rets = pkg.parseFuncToken(funcTok)
	}

	return method
}

func (pkg packageParser) parseFuncToken(tok *ast.FuncType) ([]Value, []Value) {

	unnameArgs := make(map[string]repeat)
	args := []Value{}
	for ai, argTok := range tok.Params.List {
		args = append(args, pkg.parseFieldToken(argTok, "Arg", unnameArgs, ai)...)
	}

	if tok.Results == nil {
		return args, nil
	}
	var rets []Value
	unnamedRets := make(map[string]repeat)
	for ri, ret := range tok.Results.List {
		rets = append(rets, pkg.parseFieldToken(ret, "Result", unnamedRets, ri)...)
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

func (pkg packageParser) parseFieldToken(tok *ast.Field, fieldType string, unnamed map[string]repeat, i int) []Value {
	value := []Value{}

	valName, valType := pkg.parseType(tok.Type)

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

func (pkg packageParser) parseType(tok ast.Expr) (string, ast.Expr) {
	switch typeTok := tok.(type) {
	case *ast.Ident:
		name := typeTok.Name
		if typeTok.Obj != nil {
			if typeTok.Obj.Kind == ast.Typ {
				return lowerFirst(name), &ast.SelectorExpr{X: ast.NewIdent(pkg.pkg.Name), Sel: ast.NewIdent(typeTok.Name)}
			}
		}
		if name == "error" {
			name = "err"
		}
		return name, ast.NewIdent(typeTok.Name)
	case *ast.Ellipsis:
		name, expr := pkg.parseType(typeTok.Elt)
		return name + "Var", &ast.Ellipsis{Elt: expr}
	case *ast.SelectorExpr:
		_, expr := pkg.parseType(typeTok.X)
		return lowerFirst(typeTok.Sel.Name), &ast.SelectorExpr{X: expr, Sel: ast.NewIdent(typeTok.Sel.Name)}
	case *ast.ArrayType:
		name, expr := pkg.parseType(typeTok.Elt)
		if !strings.HasSuffix(name, "Arr") {
			name += "Arr"
		}
		return name, &ast.ArrayType{Elt: expr}
	}

	return "", nil
}

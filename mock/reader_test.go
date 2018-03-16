package mock_test

import (
	"errors"
	"fmt"
	"go/ast"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	. "github.com/vitreuz/table-mocks/mock"
)

func TestReadPkg(t *testing.T) {
	type pkgMaker func(dir string, i int)
	pkg := func(ss ...pkgMaker) string {
		dir, err := ioutil.TempDir("", "read_file_dir_")
		if err != nil {
			panic(err)
		}

		for si, s := range ss {
			s(dir, si)
		}

		return dir
	}
	file := func(s string) pkgMaker {
		return func(dir string, i int) {
			p := filepath.Join(dir, fmt.Sprintf("tmp%d.go", i))
			f, err := os.Create(p)
			if err != nil {
				panic(err)
			}

			r := strings.NewReader(s)
			io.Copy(f, r)
			f.Close()
		}
	}

	type checkOut func(*Mock) []error
	check := func(fns ...checkOut) []checkOut { return fns }
	expectInterfaceCount := func(count int) checkOut {
		return func(mocks *Mock) []error {
			if len(mocks.Interfaces) != count {
				return []error{fmt.Errorf(
					"expected to get %d, interfaces but got %d",
					count, len(mocks.Interfaces),
				)}
			}
			return nil
		}
	}
	type checkOutInterface func(Interface) []error
	checkInterface := func(i int, fns ...checkOutInterface) checkOut {
		return func(m *Mock) []error {
			if len(m.Interfaces) < i+1 {
				return []error{errors.New(
					"unable to run checks on interface",
				)}
			}

			var errs []error
			for fi := 0; fi < len(fns); fi++ {
				fn := fns[fi]
				ifce := m.Interfaces[i]
				for _, err := range fn(ifce) {
					errs = append(errs, fmt.Errorf("Interface %s: %v", ifce.Name, err))
				}
			}
			return errs
		}
	}

	interfaceHasName := func(name string) checkOutInterface {
		return func(iface Interface) []error {
			if iface.Name != name {
				return []error{fmt.Errorf(
					"expected to have name %q but got %q",
					name, iface.Name,
				)}
			}
			return nil
		}
	}
	interfaceHasMethodCount := func(count int) checkOutInterface {
		return func(iface Interface) []error {
			if len(iface.Methods) != count {
				return []error{fmt.Errorf(
					"expected to have %d methods but got %d",
					count, len(iface.Methods),
				)}
			}
			return nil
		}
	}

	type checkOutMethod func(Method) []error
	checkMethod := func(i int, fns ...checkOutMethod) checkOutInterface {
		return func(ifce Interface) []error {
			if len(ifce.Methods) < i+1 {
				return []error{errors.New(
					"unable to run checks on method",
				)}
			}

			var errs []error
			for fi := 0; fi < len(fns); fi++ {
				fn := fns[fi]
				method := ifce.Methods[i]
				for _, err := range fn(method) {
					errs = append(errs, fmt.Errorf(
						"Method %s: %v",
						method.Name, err,
					))
				}
			}

			return errs
		}
	}

	methodHasName := func(name string) checkOutMethod {
		return func(method Method) []error {
			if method.Name != name {
				return []error{fmt.Errorf(
					"expected to have name %q but got %q",
					name, method.Name,
				)}
			}
			return nil
		}
	}
	methodHasArgCount := func(count int) checkOutMethod {
		return func(method Method) []error {
			if len(method.Args) != count {
				return []error{fmt.Errorf(
					"expected to have %d args but got %d",
					count, len(method.Args),
				)}
			}
			return nil
		}
	}
	methodHasRetCount := func(count int) checkOutMethod {
		return func(method Method) []error {
			if len(method.Rets) != count {
				return []error{fmt.Errorf(
					"expected have %d returns but got %d",
					count, len(method.Rets),
				)}
			}
			return nil
		}
	}

	type checkOutValue func(Value) []error
	checkArgs := func(fns ...checkOutValue) checkOutMethod {
		return func(method Method) []error {
			var errs []error
			for i := 0; i < len(fns); i++ {
				if i >= len(method.Args) {
					return append(errs, fmt.Errorf(
						"unable to check arg %d",
						i,
					))
				}

				fn := fns[i]
				for _, err := range fn(method.Args[i]) {
					errs = append(errs, fmt.Errorf(
						"Arg %d: %v",
						i, err,
					))
				}
			}
			return errs
		}
	}
	checkRets := func(fns ...checkOutValue) checkOutMethod {
		return func(method Method) []error {
			var errs []error
			for i := 0; i < len(fns); i++ {
				if i >= len(method.Rets) {
					return append(errs, fmt.Errorf(
						"unable to check arg %d",
						i,
					))
				}

				fn := fns[i]
				for _, err := range fn(method.Rets[i]) {
					errs = append(errs, fmt.Errorf(
						"Arg %d: %v",
						i, err,
					))
				}
			}
			return errs
		}
	}
	checkValue := func(name string, typ ast.Expr) checkOutValue {
		return func(actual Value) []error {
			var errs []error
			if actual.Name != name {
				errs = append(errs, fmt.Errorf(
					"expected to have name %s but got %s",
					name, actual.Name,
				))
			}
			if !reflect.DeepEqual(typ, actual.Type) {
				errs = append(errs, fmt.Errorf(
					"expected to have type %T %+v but got %T %+v",
					typ, typ, actual.Type, actual.Type,
				))
			}
			return errs
		}
	}
	stringType := ast.NewIdent("string")
	intType := ast.NewIdent("int")
	errorType := ast.NewIdent("error")
	arrayType := func(expr ast.Expr) *ast.ArrayType { return &ast.ArrayType{Elt: expr} }
	ellipseType := func(expr ast.Expr) *ast.Ellipsis { return &ast.Ellipsis{Elt: expr} }
	selectorType := func(x ast.Expr, sel string) *ast.SelectorExpr {
		return &ast.SelectorExpr{X: x, Sel: ast.NewIdent(sel)}
	}
	bytesSelector := selectorType(ast.NewIdent("bytes"), "Buffer")
	// ----------       ----------
	// ----------       ----------
	// ---------- Tests ----------
	// ----------       ----------
	// ----------       ----------
	tests := [...]struct {
		name   string
		input  string
		checks []checkOut
	}{
		{
			"Simplest interface",
			pkg(file(`
				package a

				type B interface{
					C(d string) (e string)
				}`,
			)),
			check(
				expectInterfaceCount(1),
				checkInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(1),
						checkArgs(
							checkValue("d", stringType),
						),
						methodHasRetCount(1),
						checkRets(
							checkValue("e", stringType),
						),
					),
				),
			),
		}, {
			"Unnamed args",
			pkg(file(`
				package a

				type B interface {
					C(string, int, string) (int, error)
				}`,
			)),
			check(
				expectInterfaceCount(1),
				checkInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(3),
						checkArgs(
							checkValue("stringArg1", stringType),
							checkValue("intArg", intType),
							checkValue("stringArg2", stringType),
						),
						methodHasRetCount(2),
						checkRets(
							checkValue("intResult", intType),
							checkValue("errResult", errorType),
						),
					),
				),
			),
		}, {
			"Minimal identifier",
			pkg(file(`
				package a

				type B interface {
					C(d, e string) (f int, g error)
				}`,
			)),
			check(
				expectInterfaceCount(1),
				checkInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(2),
						checkArgs(
							checkValue("d", stringType),
							checkValue("e", stringType),
						),
						methodHasRetCount(2),
						checkRets(
							checkValue("f", intType),
							checkValue("g", errorType),
						),
					),
				),
			),
		}, {
			"Embedded interface same file",
			pkg(file(`
				package a

				type B interface{
					C() string
					D() string
				}

				type E interface {
					B
					F(int) string
				}`,
			)),
			check(
				expectInterfaceCount(2),
				checkInterface(1,
					interfaceHasName("E"),
					interfaceHasMethodCount(3),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(0),
						methodHasRetCount(1),
						checkRets(
							checkValue("stringResult", stringType),
						),
					),
					checkMethod(1,
						methodHasName("D"),
						methodHasArgCount(0),
						methodHasRetCount(1),
						checkRets(
							checkValue("stringResult", stringType),
						),
					),
					checkMethod(2,
						methodHasName("F"),
						methodHasArgCount(1),
						checkArgs(
							checkValue("intArg", intType),
						),
						methodHasRetCount(1),
						checkRets(
							checkValue("stringResult", stringType),
						),
					),
				),
			),
			// },
			// // NOTE: embedded interfaces in different files not supported right now
		}, {
			"Embedded interface different file (same package)",
			pkg(file(`
		     package a

			type B interface {
				D
				C() string
			}`,
			), file(`
			package a

			type D interface{
				E()
				F()
			}`,
			)),
			check(
				expectInterfaceCount(2),
				checkInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(3),
					checkMethod(0,
						methodHasName("E"),
						methodHasArgCount(0),
					),
					checkMethod(1,
						methodHasName("F"),
						methodHasArgCount(0),
					),
					checkMethod(2,
						methodHasName("C"),
						methodHasArgCount(0),
					),
				),
			),
			// }, {
			// 	"Embedded interface different file (different package)",
			// 	pkg(file(`
			// 	 package a

			// 	 import "io"

			// 	type B interface {
			// 		io.Reader
			// 		C() string
			// 	}`,
			// 	)),
			// 	check(
			// 		expectInterfaceCount(2),
			// 		checkInterface(0,
			// 			interfaceHasName("B"),
			// 			interfaceHasMethodCount(3),
			// 			checkMethod(0,
			// 				methodHasName("Read"),
			// 				methodHasArgCount(0),
			// 			),
			// 			checkMethod(1,
			// 				methodHasName("C"),
			// 				methodHasRetCount(1),
			// 			),
			// 		),
			// 	),
		}, {
			"Variadic args",
			pkg(file(`
				package a

				type B interface {
					C(...string)
				}`,
			)),
			check(
				expectInterfaceCount(1),
				checkInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(1),
						checkArgs(
							checkValue("stringVarArg", ellipseType(stringType)),
						),
					),
				),
			),
		}, {
			"No returns method",
			pkg(file(`
				package a

				type B interface{
					C()
				}`,
			)),
			check(
				expectInterfaceCount(1),
				checkInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(0),
						methodHasRetCount(0),
					),
				),
			),
		}, {
			"Type value declared from other file",
			pkg(file(`
				package a

				type B interface{
					C(bytes.Buffer)
				}
				`,
			)),
			check(
				expectInterfaceCount(1),
				checkInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(1),
						checkArgs(
							checkValue("bufferArg", bytesSelector),
						),
						methodHasRetCount(0),
					),
				),
			),
		}, {
			"Array values",
			pkg(file(`
				package a

				type B interface{
					C([][]string, []bytes.Buffer) []error
				}
				`,
			)),
			check(
				expectInterfaceCount(1),
				checkInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(2),
						checkArgs(
							checkValue("stringArrArg", arrayType(arrayType(stringType))),
							checkValue("bufferArrArg", arrayType(bytesSelector)),
						),
						methodHasRetCount(1),
						checkRets(
							checkValue("errArrResult", arrayType(errorType)),
						),
					),
				),
			),
		}, {
			"Using a package struct type (same file)",
			pkg(file(`
				package a

				type B struct {}

				type C interface{
					D(B)
				}
				`,
			)),
			check(
				expectInterfaceCount(1),
				checkInterface(0,
					interfaceHasName("C"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("D"),
						methodHasArgCount(1),
						checkArgs(
							checkValue("bArg", selectorType(ast.NewIdent("a"), "B")),
						),
					),
				),
			),
		}, {
			"Using a package struct type (same file)",
			pkg(file(`
				package a

				type B struct {}
				`), file(`
				package a

				type C interface{
					D(B)
				}
				`,
			)),
			check(
				expectInterfaceCount(1),
				checkInterface(0,
					interfaceHasName("C"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("D"),
						methodHasArgCount(1),
						checkArgs(
							checkValue("bArg", selectorType(ast.NewIdent("a"), "B")),
						),
					),
				),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.input
			defer os.RemoveAll(dir)

			mocks := ReadPkg(dir)
			for _, check := range tt.checks {
				for _, checkErr := range check(mocks) {
					if checkErr != nil {
						t.Error(checkErr)
					}
				}
			}
		})
	}
}

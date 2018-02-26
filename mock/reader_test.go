package mock_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	. "github.com/vitreuz/table-mocks/mock"
)

func TestReadFile(t *testing.T) {

	type checkOut func(*Mock) []error
	check := func(fns ...[]checkOut) []checkOut {
		checks := []checkOut{}

		for _, fnArray := range fns {
			checks = append(checks, fnArray...)
		}

		return checks
	}
	checkMock := func(fns ...checkOut) []checkOut { return fns }

	hasInterfaceCount := func(count int) checkOut {
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
	type checkOutInterface func(string, Interface) []error
	checkInterface := func(i int, fns ...checkOutInterface) []checkOut {
		checks := []checkOut{}
		// Iterating over callbacks ends up evaluating the last callback
		// len(fns) number of times instead of actually passing each callback.
		for fi := 0; fi < len(fns); fi++ {
			fn := fns[fi]
			checks = append(checks, func(mock *Mock) []error {
				return fn(
					mock.Interfaces[i].Name,
					mock.Interfaces[i],
				)
			})
		}

		return checks
	}

	interfaceHasName := func(name string) checkOutInterface {
		return func(_ string, iface Interface) []error {
			if iface.Name != name {
				return []error{fmt.Errorf(
					"expected Interface to have name %q but got %q",
					name, iface.Name,
				)}
			}
			return nil
		}
	}
	interfaceHasMethodCount := func(count int) checkOutInterface {
		return func(name string, iface Interface) []error {
			if len(iface.Methods) != count {
				return []error{fmt.Errorf(
					"expected interface %s to have %d methods but got %d",
					name, count, len(iface.Methods),
				)}
			}
			return nil
		}
	}

	type checkOutInterfaceMethod func(string, string, Method) []error
	checkInterfaceMethod := func(ii, mi int, fns ...checkOutInterfaceMethod) []checkOut {
		checks := []checkOut{}

		for fi := 0; fi < len(fns); fi++ {
			fn := fns[fi]
			checks = append(checks, func(mock *Mock) []error {
				return fn(
					mock.Interfaces[ii].Name,
					mock.Interfaces[ii].Methods[mi].Name,
					mock.Interfaces[ii].Methods[mi],
				)
			})
		}

		return checks
	}
	interfaceMethodHasName := func(name string) checkOutInterfaceMethod {
		return func(iName, mName string, method Method) []error {
			if method.Name != name {
				return []error{fmt.Errorf(
					"expected %s Method to have name %q but got %q",
					iName, name, method.Name,
				)}
			}
			return nil
		}
	}
	interfaceMethodHasArgCount := func(count int) checkOutInterfaceMethod {
		return func(iName, mName string, method Method) []error {
			if len(method.Args) != count {
				return []error{fmt.Errorf(
					"expected %s.%s to have %d args, but got %d",
					iName, mName, count, len(method.Args),
				)}
			}
			return nil
		}
	}
	interfaceMethodHasRetCount := func(count int) checkOutInterfaceMethod {
		return func(iName, mName string, method Method) []error {
			if len(method.Rets) != count {
				return []error{fmt.Errorf(
					"expected %s.%s to have %d returns, but got %d",
					iName, mName, count, len(method.Rets),
				)}
			}
			return nil
		}
	}

	type checkOutInterfaceMethodArg func(string, string, int, Value) []error
	checkInterfaceMethodArg := func(ii, mi, ai int, fns ...checkOutInterfaceMethodArg) []checkOut {
		checks := []checkOut{}

		for fi := 0; fi < len(fns); fi++ {
			fn := fns[fi]
			checks = append(checks, func(mock *Mock) []error {
				return fn(
					mock.Interfaces[ii].Name,
					mock.Interfaces[ii].Methods[mi].Name,
					ai,
					mock.Interfaces[ii].Methods[mi].Args[ai],
				)
			})
		}

		return checks
	}
	interfaceMethodArgExpectValue := func(name, typ string) checkOutInterfaceMethodArg {
		return func(iName, mName string, ai int, arg Value) []error {
			expectValue := Value{Name: name, Type: typ}
			if !reflect.DeepEqual(arg, expectValue) {
				return []error{fmt.Errorf(
					"expected %s.%s arg %d to be %+v but got %+v",
					iName, mName, ai, arg, expectValue,
				)}
			}
			return nil
		}
	}

	type checkOutInterfaceMethodRet func(string, string, int, Value) []error
	checkInterfaceMethodRet := func(ii, mi, ri int, fns ...checkOutInterfaceMethodRet) []checkOut {
		checks := []checkOut{}

		for fi := 0; fi < len(fns); fi++ {
			fn := fns[fi]
			checks = append(checks, func(mock *Mock) []error {
				return fn(
					mock.Interfaces[ii].Name,
					mock.Interfaces[ii].Methods[mi].Name,
					ri,
					mock.Interfaces[ii].Methods[mi].Rets[ri],
				)
			})
		}

		return checks
	}
	interfaceMethodRetExpectValue := func(name, typ string) checkOutInterfaceMethodRet {
		return func(iName, mName string, ri int, ret Value) []error {
			expectValue := Value{Name: name, Type: typ}
			if !reflect.DeepEqual(ret, expectValue) {
				return []error{fmt.Errorf(
					"expected %s.%s ret %d to be %+v but got %+v",
					iName, mName, ri, ret, expectValue,
				)}
			}
			return nil
		}
	}
	// ----------       ----------
	// ----------       ----------
	// ---------- Tests ----------
	// ----------       ----------
	// ----------       ----------

	tests := [...]struct {
		name   string
		input  *strings.Reader
		checks []checkOut
	}{
		{
			"Simplest interface",
			strings.NewReader(`
				package a

				type B interface{
					C(d string) (e string)
				}`,
			),
			check(
				checkMock(
					hasInterfaceCount(1),
				),
				checkInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
				),
				checkInterfaceMethod(0, 0,
					interfaceMethodHasName("C"),
					interfaceMethodHasArgCount(1),
					interfaceMethodHasRetCount(1),
				),
				checkInterfaceMethodArg(0, 0, 0,
					interfaceMethodArgExpectValue("d", "string"),
				),
				checkInterfaceMethodRet(0, 0, 0,
					interfaceMethodRetExpectValue("e", "string"),
				),
			),
		}, {
			"Unnamed args",
			strings.NewReader(`
				package a

				type B interface {
					C(string, string) int
				}`,
			),
			check(
				checkMock(
					hasInterfaceCount(1),
				),
				checkInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
				),
				checkInterfaceMethod(0, 0,
					interfaceMethodHasName("C"),
					interfaceMethodHasArgCount(2),
					interfaceMethodHasRetCount(1),
				),
				checkInterfaceMethodArg(0, 0, 0,
					interfaceMethodArgExpectValue("arg1", "string"),
				),
				checkInterfaceMethodArg(0, 0, 1,
					interfaceMethodArgExpectValue("arg2", "string"),
				),
				checkInterfaceMethodRet(0, 0, 0,
					interfaceMethodRetExpectValue("ret1", "int"),
				),
			),
		}, {
			"Minimal identifier",
			strings.NewReader(`
				package a

				type B interface {
					C(d, e string) (f int, g error)
				}`,
			),
			check(
				checkMock(
					hasInterfaceCount(1),
				),
				checkInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
				),
				checkInterfaceMethod(0, 0,
					interfaceMethodHasName("C"),
					interfaceMethodHasArgCount(2),
					interfaceMethodHasRetCount(2),
				),
				checkInterfaceMethodArg(0, 0, 0,
					interfaceMethodArgExpectValue("d", "string"),
				),
				checkInterfaceMethodArg(0, 0, 1,
					interfaceMethodArgExpectValue("e", "string"),
				),
				checkInterfaceMethodRet(0, 0, 0,
					interfaceMethodRetExpectValue("f", "int"),
				),
				checkInterfaceMethodRet(0, 0, 1,
					interfaceMethodRetExpectValue("g", "error"),
				),
			),
		}, {
			"Embedded interface same file",
			strings.NewReader(`
				package a

				type B interface{
					C() string
					D() string
				}

				type E interface {
					B
					F(int) string
				}`,
			),
			check(
				checkMock(
					hasInterfaceCount(2),
				),
				checkInterface(1,
					interfaceHasName("E"),
					interfaceHasMethodCount(3),
				),
				checkInterfaceMethod(1, 0,
					interfaceMethodHasName("C"),
					interfaceMethodHasArgCount(0),
					interfaceMethodHasRetCount(1),
				),
				checkInterfaceMethod(1, 1,
					interfaceMethodHasName("D"),
					interfaceMethodHasArgCount(0),
					interfaceMethodHasRetCount(1),
				),
				checkInterfaceMethod(1, 2,
					interfaceMethodHasName("F"),
					interfaceMethodHasArgCount(1),
					interfaceMethodHasRetCount(1),
				),
			),
		},
		// NOTE: embedded interfaces in different files not supported right now
		// }, {
		//  "Embedded interface different file",
		//  strings.NewReader(`
		//      package a

		// 		import "io"

		// 		type B interface {
		// 			io.Reader
		// 			C() string
		// 		}`,
		// 	),
		// 	check(
		// 		checkMock(
		// 			hasInterfaceCount(1),
		// 		),
		// 		checkInterface(0,
		// 			interfaceHasName("B"),
		// 			interfaceHasMethodCount(2),
		// 		),
		// 		checkInterfaceMethod(0, 0,
		// 			interfaceMethodHasName("Reader"),
		// 		),
		// 	),
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocks := ReadFile(tt.input)
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

package mock_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	. "github.com/vitreuz/table-mocks/mock"
)

func TestReadFile(t *testing.T) {

	type checkOut func(*Mock) []error
	checks := func(fns ...checkOut) []checkOut { return fns }
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
	checksInterface := func(i int, fns ...checkOutInterface) checkOut {
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
					"expected have name %q but got %q",
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

	type checkOutInterfaceMethod func(Method) []error
	checkMethod := func(i int, fns ...checkOutInterfaceMethod) checkOutInterface {
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

	methodHasName := func(name string) checkOutInterfaceMethod {
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
	methodHasArgCount := func(count int) checkOutInterfaceMethod {
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
	methodHasRetCount := func(count int) checkOutInterfaceMethod {
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

	type checkOutInterfaceMethodArg func(Value) []error
	// checkInterfaceMethodArg := func(ii, mi, ai int, fns ...checkOutInterfaceMethodArg) []checkOut {
	// 	checks := []checkOut{}

	// 	for fi := 0; fi < len(fns); fi++ {
	// 		fn := fns[fi]
	// 		checks = append(checks, func(mock *Mock) []error {
	// 			return fn(
	// 				mock.Interfaces[ii].Name,
	// 				mock.Interfaces[ii].Methods[mi].Name,
	// 				ai,
	// 				mock.Interfaces[ii].Methods[mi].Args[ai],
	// 			)
	// 		})
	// 	}

	// 	return checks
	// }
	// interfaceMethodArgExpectValue := func(name, typ string, isVariadic bool) checkOutInterfaceMethodArg {
	// 	return func(iName, mName string, ai int, arg Value) []error {
	// 		// expectValue := Value{Name: name, Type: typ, IsVariadic: isVariadic}
	// 		expectValue := Value{}
	// 		if !reflect.DeepEqual(arg, expectValue) {
	// 			return []error{fmt.Errorf(
	// 				"expected %s.%s arg %d to be %+v but got %+v",
	// 				iName, mName, ai, expectValue, arg,
	// 			)}
	// 		}
	// 		return nil
	// 	}
	// }

	// type checkOutInterfaceMethodRet func(string, string, int, Value) []error
	// checkInterfaceMethodRet := func(ii, mi, ri int, fns ...checkOutInterfaceMethodRet) []checkOut {
	// 	checks := []checkOut{}

	// 	for fi := 0; fi < len(fns); fi++ {
	// 		fn := fns[fi]
	// 		checks = append(checks, func(mock *Mock) []error {
	// 			return fn(
	// 				mock.Interfaces[ii].Name,
	// 				mock.Interfaces[ii].Methods[mi].Name,
	// 				ri,
	// 				mock.Interfaces[ii].Methods[mi].Rets[ri],
	// 			)
	// 		})
	// 	}

	// 	return checks
	// }
	// interfaceMethodRetExpectValue := func(name, typ string) checkOutInterfaceMethodRet {
	// 	return func(iName, mName string, ri int, ret Value) []error {
	// 		// expectValue := Value{Name: name, Type: typ}
	// 		expectValue := Value{}
	// 		if !reflect.DeepEqual(ret, expectValue) {
	// 			return []error{fmt.Errorf(
	// 				"expected %s.%s ret %d to be %+v but got %+v",
	// 				iName, mName, ri, expectValue, ret,
	// 			)}
	// 		}
	// 		return nil
	// 	}
	// }
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
			checks(
				expectInterfaceCount(1),
				checksInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(1),
						methodHasRetCount(1),
					),
				),
			),
			// check(
			// 	checkMock(
			// 		hasInterfaceCount(1),
			// 	),
			// 	checkInterface(0,
			// 		interfaceHasName("B"),
			// 		interfaceHasMethodCount(1),
			// 	),
			// 	checkInterfaceMethod(0, 0,
			// 		interfaceMethodHasName("C"),
			// 		interfaceMethodHasArgCount(1),
			// 		interfaceMethodHasRetCount(1),
			// 	),
			// 	checkInterfaceMethodArg(0, 0, 0,
			// 		interfaceMethodArgExpectValue("d", "string", false),
			// 	),
			// 	checkInterfaceMethodRet(0, 0, 0,
			// 		interfaceMethodRetExpectValue("e", "string"),
			// 	),
			// ),
		}, {
			"Unnamed args",
			strings.NewReader(`
				package a

				type B interface {
					C(string, string) int
				}`,
			),
			checks(
				expectInterfaceCount(1),
				checksInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(2),
						methodHasRetCount(1),
					),
				),
			),
			// check(
			// 	checkMock(
			// 		hasInterfaceCount(1),
			// 	),
			// 	checkInterface(0,
			// 		interfaceHasName("B"),
			// 		interfaceHasMethodCount(1),
			// 	),
			// 	checkInterfaceMethod(0, 0,
			// 		interfaceMethodHasName("C"),
			// 		interfaceMethodHasArgCount(2),
			// 		interfaceMethodHasRetCount(1),
			// 	),
			// 	checkInterfaceMethodArg(0, 0, 0,
			// 		interfaceMethodArgExpectValue("arg1", "string", false),
			// 	),
			// 	checkInterfaceMethodArg(0, 0, 1,
			// 		interfaceMethodArgExpectValue("arg2", "string", false),
			// 	),
			// 	checkInterfaceMethodRet(0, 0, 0,
			// 		interfaceMethodRetExpectValue("ret1", "int"),
			// 	),
			// ),
		}, {
			"Minimal identifier",
			strings.NewReader(`
				package a

				type B interface {
					C(d, e string) (f int, g error)
				}`,
			),
			checks(
				expectInterfaceCount(1),
				checksInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(2),
						methodHasRetCount(2),
					),
				),
			),
			// check(
			// 	checkMock(
			// 		hasInterfaceCount(1),
			// 	),
			// 	checkInterface(0,
			// 		interfaceHasName("B"),
			// 		interfaceHasMethodCount(1),
			// 	),
			// 	checkInterfaceMethod(0, 0,
			// 		interfaceMethodHasName("C"),
			// 		interfaceMethodHasArgCount(2),
			// 		interfaceMethodHasRetCount(2),
			// 	),
			// 	checkInterfaceMethodArg(0, 0, 0,
			// 		interfaceMethodArgExpectValue("d", "string", false),
			// 	),
			// 	checkInterfaceMethodArg(0, 0, 1,
			// 		interfaceMethodArgExpectValue("e", "string", false),
			// 	),
			// 	checkInterfaceMethodRet(0, 0, 0,
			// 		interfaceMethodRetExpectValue("f", "int"),
			// 	),
			// 	checkInterfaceMethodRet(0, 0, 1,
			// 		interfaceMethodRetExpectValue("g", "error"),
			// 	),
			// ),
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
			checks(
				expectInterfaceCount(2),
				checksInterface(1,
					interfaceHasName("E"),
					interfaceHasMethodCount(3),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(0),
						methodHasRetCount(1),
					),
					checkMethod(1,
						methodHasName("D"),
						methodHasArgCount(0),
						methodHasRetCount(1),
					),
					checkMethod(2,
						methodHasName("F"),
						methodHasArgCount(1),
						methodHasRetCount(1),
					),
				),
			),
			// 	check(
			// 		checkMock(
			// 			hasInterfaceCount(2),
			// 		),
			// 		checkInterface(1,
			// 			interfaceHasName("E"),
			// 			interfaceHasMethodCount(3),
			// 		),
			// 		checkInterfaceMethod(1, 0,
			// 			interfaceMethodHasName("C"),
			// 			interfaceMethodHasArgCount(0),
			// 			interfaceMethodHasRetCount(1),
			// 		),
			// 		checkInterfaceMethod(1, 1,
			// 			interfaceMethodHasName("D"),
			// 			interfaceMethodHasArgCount(0),
			// 			interfaceMethodHasRetCount(1),
			// 		),
			// 		checkInterfaceMethod(1, 2,
			// 			interfaceMethodHasName("F"),
			// 			interfaceMethodHasArgCount(1),
			// 			interfaceMethodHasRetCount(1),
			// 		),
			// 	),
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
		{
			"Variadic args",
			strings.NewReader(`
				package a

				type B interface {
					C(...string) string
				}`,
			),
			checks(
				expectInterfaceCount(1),
				checksInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(1),
						methodHasRetCount(1),
					),
				),
			),
			// check(
			// 	checkMock(
			// 		hasInterfaceCount(1),
			// 	),
			// 	checkInterface(0,
			// 		interfaceHasName("B"),
			// 		interfaceHasMethodCount(1),
			// 	),
			// 	checkInterfaceMethod(0, 0,
			// 		interfaceMethodHasName("C"),
			// 		interfaceMethodHasArgCount(1),
			// 		interfaceMethodHasRetCount(1),
			// 	),
			// 	checkInterfaceMethodArg(0, 0, 0,
			// 		interfaceMethodArgExpectValue("arg1", "string", true),
			// 	),
			// 	checkInterfaceMethodRet(0, 0, 0,
			// 		interfaceMethodRetExpectValue("ret1", "string"),
			// 	),
			// ),
		}, {
			"No returns method",
			strings.NewReader(`
				package a

				type B interface{
					C()
				}`,
			),
			checks(
				expectInterfaceCount(1),
				checksInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(0),
						methodHasRetCount(0),
					),
				),
			),
			// check(
			// 	checkMock(
			// 		hasInterfaceCount(1),
			// 	),
			// 	checkInterface(0,
			// 		interfaceHasName("B"),
			// 		interfaceHasMethodCount(1),
			// 	),
			// 	checkInterfaceMethod(0, 0,
			// 		interfaceMethodHasName("C"),
			// 		interfaceMethodHasArgCount(0),
			// 		interfaceMethodHasRetCount(0),
			// 	),
			// ),
		}, {
			"Type value declared from other file",
			strings.NewReader(`
				package a

				type B interface{
					C(bytes.Buffer)
				}
				`,
			),
			checks(
				expectInterfaceCount(1),
				checksInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(1),
						methodHasRetCount(0),
					),
				),
			),
			// check(
			// 	checkMock(
			// 		hasInterfaceCount(1),
			// 	),
			// 	checkInterface(0,
			// 		interfaceHasName("B"),
			// 		interfaceHasMethodCount(1),
			// 	),
			// 	checkInterfaceMethod(0, 0,
			// 		interfaceMethodHasName("C"),
			// 		interfaceMethodHasArgCount(1),
			// 		interfaceMethodHasRetCount(0),
			// 	),
			// 	checkInterfaceMethodArg(0, 0, 0,
			// 		interfaceMethodArgExpectValue("arg1", "bytes.Buffer", false),
			// 	),
			// ),
		}, {
			"Array values",
			strings.NewReader(`
				package a

				type B interface{
					C([][]string, []bytes.Buffer) []error
				}
				`,
			),
			checks(
				expectInterfaceCount(1),
				checksInterface(0,
					interfaceHasName("B"),
					interfaceHasMethodCount(1),
					checkMethod(0,
						methodHasName("C"),
						methodHasArgCount(2),
						methodHasRetCount(1),
					),
				),
			),
			// check(
			// 	checkMock(
			// 		hasInterfaceCount(1),
			// 	),
			// 	checkInterface(0,
			// 		interfaceHasName("B"),
			// 		interfaceHasMethodCount(1),
			// 	),
			// 	checkInterfaceMethod(0, 0,
			// 		interfaceMethodHasName("C"),
			// 		interfaceMethodHasArgCount(2),
			// 		interfaceMethodHasRetCount(1),
			// 	),
			// 	checkInterfaceMethodArg(0, 0, 0,
			// 		interfaceMethodArgExpectValue("arg1", "[][]string", false),
			// 	),
			// 	checkInterfaceMethodArg(0, 0, 1,
			// 		interfaceMethodArgExpectValue("arg2", "[]bytes.Buffer", false),
			// 	),
			// 	checkInterfaceMethodRet(0, 0, 0,
			// 		interfaceMethodRetExpectValue("ret1", "[]error"),
			// 	),
			// ),
		},
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

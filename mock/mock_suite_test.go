package mock_test

import (
	"go/ast"

	. "github.com/vitreuz/table-mocks/mock"
)

type testInterface struct {
	Interface
}

func newTestInterface(name string) testInterface {
	return testInterface{Interface{Name: name}}
}

func (t testInterface) WithMethod(method testMethod) testInterface {
	t.Methods = append(t.Methods, method.Method)
	return t
}

func (t testInterface) ToInterface() Interface { return t.Interface }

// METHOD

type testMethod struct {
	Method
}

func newTestMethod(name string) testMethod {
	return testMethod{Method{Name: name}}
}

func (t testMethod) WithArg(val testValue) testMethod {
	t.Args = append(t.Args, val.Value)
	return t
}
func (t testMethod) WithRet(val testValue) testMethod {
	t.Rets = append(t.Rets, val.Value)
	return t
}

func (t testMethod) ToMethod() Method { return t.Method }

// VALUE

type testValue struct {
	Value
}

func newTestValue(name string) testValue {
	return testValue{Value{Name: name, Type: ast.NewIdent("string")}}
}

func (t testValue) asEllipse() testValue {
	t.Type = &ast.Ellipsis{Elt: ast.NewIdent("string")}
	return t
}

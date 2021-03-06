package mock_test

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"

	. "github.com/vitreuz/table-mocks/mock"
)

func prefix(pre, text string) string {
	var printer func(...interface{}) string
	switch pre {
	case "+ ":
		printer = color.New(color.FgGreen).SprintFunc()
	case "- ":
		printer = color.New(color.FgRed).SprintFunc()
	default:
		printer = fmt.Sprint
	}

	lines := strings.Split(text, "\n")
	for i := 0; i < len(lines)-1; i++ {
		lines[i] = pre + lines[i]
	}

	return printer(strings.Join(lines, "\n"))
}

type checkReader func(io.Reader) []error

func compareBuffers(expect, actual *bytes.Buffer) []error {
	dmp := diffmatchpatch.New()
	el, al, l := dmp.DiffLinesToChars(strings.TrimPrefix(expect.String(), "\n"), actual.String())
	diffs := dmp.DiffCharsToLines(dmp.DiffMain(el, al, false), l)

	if len(diffs) > 1 {
		diffText := new(strings.Builder)
		for _, diff := range diffs {
			switch diff.Type {
			case diffmatchpatch.DiffDelete:
				fmt.Fprint(diffText, prefix("- ", diff.Text))
			case diffmatchpatch.DiffInsert:
				fmt.Fprint(diffText, prefix("+ ", diff.Text))
			default:
				fmt.Fprint(diffText, prefix("  ", diff.Text))
			}
		}

		return []error{fmt.Errorf(
			"output doesn't match:\n%s", diffText.String(),
		)}
	}

	return nil
}

func expectReader(expected io.Reader) checkReader {
	return func(actual io.Reader) []error {
		e, a := new(bytes.Buffer), new(bytes.Buffer)
		e.ReadFrom(expected)
		a.ReadFrom(actual)

		return compareBuffers(e, a)
	}
}

func TestGenerateFile(t *testing.T) {
	check := func(fns ...checkReader) []checkReader { return fns }

	tests := [...]struct {
		name   string
		input  *Interface
		checks []checkReader
	}{
		{
			"Simple generate",
			newTestInterface("Runner").
				WithMethod(newTestMethod("Run").
					WithArg(newTestValue("distanceArg")).
					WithRet(newTestValue("durationResult").asDuration()),
				).
				WithImport("time").
				ToInterface(),
			check(expectReader(strings.NewReader(`
// generated by table-mocks; DO NOT EDIT

package fake

import (
	"sync"
	"time"
)

type Runner struct {
	runMethod map[int]RunnerRunMethod
	runMutex  sync.RWMutex
	RunCalls  int
}

type RunnerRunMethod struct {
	DistanceArg    string
	DurationResult time.Duration
}

func NewRunner() *Runner {
	fake := &Runner{}
	fake.runMethod = make(map[int]RunnerRunMethod)

	return fake
}

func (fake *Runner) Run(distanceArg string) (durationResult time.Duration) {
	fake.runMutex.Lock()
	fakeMethod := fake.runMethod[fake.RunCalls]
	fakeMethod.DistanceArg = distanceArg
	fake.runMethod[fake.RunCalls] = fakeMethod
	fake.RunCalls++
	fake.runMutex.Unlock()

	return fakeMethod.DurationResult
}

func (fake *Runner) RunReturns(durationResult time.Duration) *Runner {
	fake.runMutex.Lock()
	fakeMethod := fake.runMethod[0]
	fakeMethod.DurationResult = durationResult
	fake.runMethod[0] = fakeMethod
	fake.runMutex.Unlock()

	return fake
}

func (fake *Runner) RunGetArgs() (distanceArg string) {
	fake.runMutex.RLock()
	distanceArg = fake.runMethod[0].DistanceArg
	fake.runMutex.RUnlock()

	return distanceArg
}

type RunnerRunFunc func(RunnerRunMethod) RunnerRunMethod

func (fake *Runner) RunForCall(call int, fns ...RunnerRunFunc) *Runner {
	fake.runMutex.Lock()
	for _, fn := range fns {
		fakeMethod := fake.runMethod[call]
		fake.runMethod[call] = fn(fakeMethod)
	}
	fake.runMutex.Unlock()

	return fake
}
`,
			))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := ioutil.TempFile("", "generate_file_")
			if err != nil {
				t.Fatalf("error creating tempfile: %v", err)
			}
			defer os.Remove(f.Name())

			err = GenerateFile(tt.input, "simple", f)
			f.Close()
			for _, check := range tt.checks {
				f, _ = os.Open(f.Name())
				defer f.Close()
				for _, checkErr := range check(f) {
					if checkErr != nil {
						t.Error(checkErr)
					}
				}
			}
		})
	}
}

func TestGenerateStructs(t *testing.T) {
	check := func(fns ...checkReader) []checkReader { return fns }

	tests := [...]struct {
		name   string
		input  *Interface
		checks []checkReader
	}{
		{
			"Simple generate",
			&Interface{
				Name: "Runner",
				Methods: []Method{
					{
						Name: "Run",
						Args: []Value{
							{
								Name: "distanceArg",
								Type: &ast.Ident{Name: "int"},
							},
						},
						Rets: []Value{
							{
								Name: "durationResult",
								Type: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "time"},
									Sel: &ast.Ident{Name: "Duration"},
								},
							}, {
								Name: "errResult",
								Type: &ast.Ident{Name: "error"},
							},
						},
					},
				},
			},
			check(
				expectReader(strings.NewReader(`
type Runner struct {
	runMethod map[int]RunnerRunMethod
	runMutex  sync.RWMutex
	RunCalls  int
}
type RunnerRunMethod struct {
	DistanceArg    int
	DurationResult time.Duration
	ErrResult      error
}`,
				)),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := ioutil.TempFile("", "generate_file_")
			if err != nil {
				t.Fatalf("error creating tempfile: %v", err)
			}
			defer os.Remove(f.Name())

			decls := tt.input.GenerateStructs()
			if err := format.Node(f, token.NewFileSet(), decls); err != nil {
				t.Fatalf("error writing node: %v", err)
			}
			f.Close()

			for _, check := range tt.checks {
				f, _ = os.Open(f.Name())
				defer f.Close()
				for _, checkErr := range check(f) {
					if checkErr != nil {
						t.Error(checkErr)
					}
				}
			}
		})
	}
}

func TestGenerateMethods(t *testing.T) {
	check := func(fns ...checkReader) []checkReader { return fns }

	tests := [...]struct {
		name   string
		input  *Interface
		checks []checkReader
	}{
		{
			"Simple generate",
			&Interface{
				Name: "Runner",
				Methods: []Method{
					{
						Name: "Run",
						Args: []Value{
							{
								Name: "distanceArg",
								Type: &ast.Ident{Name: "int"},
							},
						},
						Rets: []Value{
							{
								Name: "durationResult",
								Type: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "time"},
									Sel: &ast.Ident{Name: "Duration"},
								},
							}, {
								Name: "errResult",
								Type: &ast.Ident{Name: "error"},
							},
						},
					},
				},
			},
			check(
				expectReader(strings.NewReader(
					`func NewRunner() *Runner {
	fake := &Runner{}
	fake.runMethod = make(map[int]RunnerRunMethod)
	return fake
}
func (fake *Runner) Run(distanceArg int) (durationResult time.Duration, errResult error) {
	fake.runMutex.Lock()
	fakeMethod := fake.runMethod[fake.RunCalls]
	fakeMethod.DistanceArg = distanceArg
	fake.runMethod[fake.RunCalls] = fakeMethod
	fake.RunCalls++
	fake.runMutex.Unlock()
	return fakeMethod.DurationResult, fakeMethod.ErrResult
}
func (fake *Runner) RunReturns(durationResult time.Duration, errResult error) *Runner {
	fake.runMutex.Lock()
	fakeMethod := fake.runMethod[0]
	fakeMethod.DurationResult = durationResult
	fakeMethod.ErrResult = errResult
	fake.runMethod[0] = fakeMethod
	fake.runMutex.Unlock()
	return fake
}
func (fake *Runner) RunGetArgs() (distanceArg int) {
	fake.runMutex.RLock()
	distanceArg = fake.runMethod[0].DistanceArg
	fake.runMutex.RUnlock()
	return distanceArg
}

type RunnerRunFunc func(RunnerRunMethod) RunnerRunMethod

func (fake *Runner) RunForCall(call int, fns ...RunnerRunFunc) *Runner {
	fake.runMutex.Lock()
	for _, fn := range fns {
		fakeMethod := fake.runMethod[call]
		fake.runMethod[call] = fn(fakeMethod)
	}
	fake.runMutex.Unlock()
	return fake
}`,
				)),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := ioutil.TempFile("", "generate_file_")
			if err != nil {
				t.Fatalf("error creating tempfile: %v", err)
			}
			defer os.Remove(f.Name())

			decls := tt.input.GenerateMethods()
			if err := format.Node(f, token.NewFileSet(), decls); err != nil {
				t.Fatalf("error writing node: %v", err)
			}
			f.Close()

			for _, check := range tt.checks {
				f, _ = os.Open(f.Name())
				defer f.Close()
				for _, checkErr := range check(f) {
					if checkErr != nil {
						t.Error(checkErr)
					}
				}
			}
		})
	}
}

func TestGenerateInterfaceStruct(t *testing.T) {
	check := func(fns ...checkReader) []checkReader { return fns }

	tests := [...]struct {
		name   string
		ifce   *Interface
		checks []checkReader
	}{
		{
			"Basic method",
			newTestInterface("Runner").ToInterface(),
			check(expectReader(strings.NewReader(`
type Runner struct {
}
`,
			))),
		}, {
			"Simple params",
			newTestInterface("Runner").
				WithMethod(newTestMethod("Run")).
				WithMethod(newTestMethod("Walk")).
				ToInterface(),
			check(expectReader(strings.NewReader(`
type Runner struct {
	runMethod map[int]RunnerRunMethod
	runMutex  sync.RWMutex
	RunCalls  int

	walkMethod map[int]RunnerWalkMethod
	walkMutex  sync.RWMutex
	WalkCalls  int
}
`,
			))),
		},
	}

	for _, tt := range tests {
		output := GenerateInterfaceStruct(tt.ifce)
		for _, check := range tt.checks {
			for _, checkErr := range check(strings.NewReader(output)) {
				if checkErr != nil {
					t.Error(checkErr)
				}
			}
		}
	}
}

func TestGenerateMethodStruct(t *testing.T) {
	check := func(fns ...checkReader) []checkReader { return fns }

	tests := [...]struct {
		name   string
		ifce   string
		meth   Method
		checks []checkReader
	}{
		{
			"Basic method",
			"Runner",
			newTestMethod("Run").ToMethod(),
			check(expectReader(strings.NewReader(`
type RunnerRunMethod struct {
}
`,
			))),
		}, {
			"Simple params",
			"Runner",
			newTestMethod("Run").
				WithArg(newTestValue("distanceArg")).
				WithRet(newTestValue("timeResult")).
				ToMethod(),
			check(expectReader(strings.NewReader(`
type RunnerRunMethod struct {
	DistanceArg string
	TimeResult  string
}
`,
			))),
		}, {
			"With variadic params",
			"Runner",
			newTestMethod("Run").
				WithArg(newTestValue("distanceArg").asEllipse()).
				WithRet(newTestValue("timeResult")).
				ToMethod(),
			check(expectReader(strings.NewReader(`
type RunnerRunMethod struct {
	DistanceArg []string
	TimeResult  string
}
`,
			))),
		},
	}

	for _, tt := range tests {
		output := GenerateMethodStruct(tt.ifce, tt.meth)
		for _, check := range tt.checks {
			for _, checkErr := range check(strings.NewReader(output)) {
				if checkErr != nil {
					t.Error(checkErr)
				}
			}
		}
	}
}

func TestGenerateMethodFunc(t *testing.T) {
	check := func(fns ...checkReader) []checkReader { return fns }

	tests := [...]struct {
		name   string
		ifce   string
		meth   Method
		checks []checkReader
	}{
		{
			"Basic method",
			"Runner",
			newTestMethod("Run").ToMethod(),
			check(expectReader(strings.NewReader(`
func (fake *Runner) Run() {
	fake.runMutex.Lock()
	fakeMethod := fake.runMethod[fake.RunCalls]
	fake.runMethod[fake.RunCalls] = fakeMethod
	fake.RunCalls++
	fake.runMutex.Unlock()

	return
}
`,
			))),
		}, {
			"Simple params",
			"Runner",
			newTestMethod("Run").
				WithArg(newTestValue("distanceArg")).
				WithRet(newTestValue("timeResult")).
				ToMethod(),
			check(expectReader(strings.NewReader(`
func (fake *Runner) Run(distanceArg string) (timeResult string) {
	fake.runMutex.Lock()
	fakeMethod := fake.runMethod[fake.RunCalls]
	fakeMethod.DistanceArg = distanceArg
	fake.runMethod[fake.RunCalls] = fakeMethod
	fake.RunCalls++
	fake.runMutex.Unlock()

	return fakeMethod.TimeResult
}
`,
			))),
		}, {
			"With variadic params",
			"Runner",
			newTestMethod("Run").
				WithArg(newTestValue("distanceArg").asEllipse()).
				WithRet(newTestValue("timeResult")).
				ToMethod(),
			check(expectReader(strings.NewReader(`
func (fake *Runner) Run(distanceArg ...string) (timeResult string) {
	fake.runMutex.Lock()
	fakeMethod := fake.runMethod[fake.RunCalls]
	fakeMethod.DistanceArg = distanceArg
	fake.runMethod[fake.RunCalls] = fakeMethod
	fake.RunCalls++
	fake.runMutex.Unlock()

	return fakeMethod.TimeResult
}
`,
			))),
		},
	}

	for _, tt := range tests {
		output := GenerateMethodFunc(tt.ifce, tt.meth)
		for _, check := range tt.checks {
			for _, checkErr := range check(strings.NewReader(output)) {
				if checkErr != nil {
					t.Error(checkErr)
				}
			}
		}
	}
}

func TestGenerateMethodGetArgs(t *testing.T) {
	check := func(fns ...checkReader) []checkReader { return fns }

	tests := [...]struct {
		name   string
		ifce   string
		meth   Method
		checks []checkReader
	}{
		{
			"Basic method",
			"Runner",
			newTestMethod("Run").ToMethod(),
			check(expectReader(strings.NewReader(`
func (fake *Runner) RunGetArgs() {
	fake.runMutex.RLock()
	fake.runMutex.RUnlock()

	return
}
`,
			))),
		}, {
			"Simple params",
			"Runner",
			newTestMethod("Run").
				WithArg(newTestValue("distanceArg")).
				WithRet(newTestValue("timeResult")).
				ToMethod(),
			check(expectReader(strings.NewReader(`
func (fake *Runner) RunGetArgs() (distanceArg string) {
	fake.runMutex.RLock()
	distanceArg = fake.runMethod[0].DistanceArg
	fake.runMutex.RUnlock()

	return distanceArg
}
`,
			))),
		}, {
			"With variadic params",
			"Runner",
			newTestMethod("Run").
				WithArg(newTestValue("distanceArg").asEllipse()).
				WithRet(newTestValue("timeResult")).
				ToMethod(),
			check(expectReader(strings.NewReader(`
func (fake *Runner) RunGetArgs() (distanceArg []string) {
	fake.runMutex.RLock()
	distanceArg = fake.runMethod[0].DistanceArg
	fake.runMutex.RUnlock()

	return distanceArg
}
`,
			))),
		},
	}

	for _, tt := range tests {
		output := GenerateMethodGetArgs(tt.ifce, tt.meth)
		for _, check := range tt.checks {
			for _, checkErr := range check(strings.NewReader(output)) {
				if checkErr != nil {
					t.Error(checkErr)
				}
			}
		}
	}
}

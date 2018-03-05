package mock_test

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"

	. "github.com/vitreuz/table-mocks/mock"
)

func prefix(pre, text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = pre + line
	}

	return strings.Join(lines, "\n")
}

func TestGenerateFile(t *testing.T) {
	type checkFile func(*os.File) []error
	check := func(fns ...checkFile) []checkFile { return fns }
	expectOutput := func(expectedFile io.Reader) checkFile {
		return func(file *os.File) []error {
			expect, actual := new(bytes.Buffer), new(bytes.Buffer)
			expect.ReadFrom(expectedFile)
			actual.ReadFrom(file)

			dmp := diffmatchpatch.New()
			expectLines, actualLines, lines := dmp.DiffLinesToChars(expect.String(), actual.String())
			diffs := dmp.DiffCharsToLines(dmp.DiffMain(expectLines, actualLines, true), lines)

			if len(diffs) > 1 {
				log.Println(len(diffs))
				diffText := new(strings.Builder)
				for _, diff := range diffs {
					switch diff.Type {
					case diffmatchpatch.DiffDelete:
						fmt.Fprintln(diffText, prefix("- ", diff.Text))
					case diffmatchpatch.DiffInsert:
						fmt.Fprintln(diffText, prefix("+ ", diff.Text))
					default:
						fmt.Fprintln(diffText, prefix("  ", diff.Text))
					}
				}
				return []error{fmt.Errorf(
					"output doesn't match:\n%s", diffText.String(),
				)}
			}

			return nil
		}
	}

	tests := [...]struct {
		name   string
		input  *Mock
		checks []checkFile
	}{
		{
			"Simple generate",
			&Mock{
				Package: "fake",
				Imports: []string{"time"},
				Interfaces: []Interface{
					{
						Name: "Runner",
					},
				},
			},
			check(
				expectOutput(strings.NewReader(
					`package fake

import (
	"sync"
	"time"
)

type Runner struct {}
`,
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

			err = GenerateFile(tt.input, f)
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
	type checkFile func(*os.File) []error
	check := func(fns ...checkFile) []checkFile { return fns }
	expectOutput := func(expectedFile io.Reader) checkFile {
		return func(file *os.File) []error {
			expect, actual := new(bytes.Buffer), new(bytes.Buffer)
			expect.ReadFrom(expectedFile)
			actual.ReadFrom(file)

			dmp := diffmatchpatch.New()
			expectLines, actualLines, lines := dmp.DiffLinesToChars(expect.String(), actual.String())
			diffs := dmp.DiffCharsToLines(dmp.DiffMain(expectLines, actualLines, false), lines)

			if len(diffs) > 1 {
				log.Println(len(diffs))
				diffText := new(strings.Builder)
				for _, diff := range diffs {
					switch diff.Type {
					case diffmatchpatch.DiffDelete:
						fmt.Fprintln(diffText, prefix("- ", diff.Text))
					case diffmatchpatch.DiffInsert:
						fmt.Fprintln(diffText, prefix("+ ", diff.Text))
					default:
						fmt.Fprintln(diffText, prefix("  ", diff.Text))
					}
				}
				return []error{fmt.Errorf(
					"output doesn't match:\n%s", diffText.String(),
				)}
			}

			return nil
		}
	}

	tests := [...]struct {
		name   string
		input  *Interface
		checks []checkFile
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
								Name: "distance",
								Type: &ast.Ident{Name: "int"},
							},
						},
						Rets: []Value{
							{
								Name: "duration",
								Type: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "time"},
									Sel: &ast.Ident{Name: "Duration"},
								},
							}, {
								Name: "err",
								Type: &ast.Ident{Name: "error"},
							},
						},
					},
				},
			},
			check(
				expectOutput(strings.NewReader(
					`type Runner struct {
	runMethod map[int]RunnerRunMethod
	runMutex  sync.RWMutex
}

type RunnerRunMethod struct {
	DistanceArg int

	Called         bool
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

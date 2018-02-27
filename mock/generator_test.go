package mock_test

import (
	"go/ast"
	"go/token"
	"testing"

	. "github.com/vitreuz/table-mocks/mock"
)

func TestGenerateFile(t *testing.T) {
	type checkOut func(error) []error
	check := func(fns ...checkOut) []checkOut { return fns }

	tests := [...]struct {
		name   string
		input  *ast.File
		checks []checkOut
	}{
		{
			"Simple generate",
			&ast.File{
				Name: &ast.Ident{
					Name: "a",
				},
				Decls: []ast.Decl{
					&ast.GenDecl{
						Tok: token.VAR,
						Specs: []ast.Spec{
							&ast.ValueSpec{
								Names: []*ast.Ident{
									{
										Name: "b",
									},
								},
								Type: &ast.Ident{
									Name: "string",
								},
								Values: []ast.Expr{
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: `"c"`,
									},
								},
							},
						},
					},
				},
			},
			check(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GenerateFile(tt.input)
			for _, check := range tt.checks {
				for _, checkErr := range check(err) {
					if checkErr != nil {
						t.Error(checkErr)
					}
				}
			}
		})
	}
}

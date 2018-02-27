package mock

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"log"
)

func GenerateFile(node *ast.File) error {
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), node); err != nil {
		log.Fatal(err)
	}

	fmt.Println(buf.String())
	return nil
}

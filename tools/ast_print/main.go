package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"

	"github.com/spf13/pflag"
)

var path = pflag.StringP("path", "p", "", "path to a file to print")
var filter = pflag.StringP("filter", "f", "", "optional string value to filter by")

func main() {
	pflag.Parse()

	log.Println(*path)
	if *path == "" {
		log.Fatal("must provide a path to a file")
	}

	file, err := os.Open(*path)
	if err != nil {
		log.Fatal(err)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", file, 0)
	if err != nil {
		log.Fatal(err)
	}

	if *filter != "" {
		filterFn := func(node string) bool { return *filter == node }
		ast.FilterFile(f, filterFn)
	}

	ast.Print(token.NewFileSet(), f)
}

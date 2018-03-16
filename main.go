package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/serenize/snaker"

	"github.com/vitreuz/table-mocks/args"
	"github.com/vitreuz/table-mocks/mock"
)

func main() {
	path, err := args.Parse()
	if err != nil {
		log.Fatal(err)
	}

	m := mock.ReadPkg(filepath.Dir(path))

	if err := os.MkdirAll(args.FakesDir, 0755); err != nil {
		log.Fatal(err)
	}

	for _, ifce := range m.Interfaces {
		fileName := filepath.Join(args.FakesDir, snaker.CamelToSnake(ifce.Name)+".go")

		iFile, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer iFile.Close()

		if err := mock.GenerateFile(&ifce, "fake", iFile); err != nil {
			log.Fatal(err)
		}
	}
}

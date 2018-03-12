package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/serenize/snaker"

	"github.com/spf13/pflag"
	"github.com/vitreuz/table-mocks/mock"
)

var (
	path    string
	fakeDir string
)

func init() {
	pflag.StringVarP(&path, "filepath", "p", "", "the path to the file to generate mocks for")
	pflag.StringVarP(&fakeDir, "fake-dir", "d", "", "the directory to create mocks in")
}

func main() {
	pflag.Parse()

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	if fakeDir == "" {
		fakeDir = filepath.Join(filepath.Dir(path), "fakes")
	}
	m := mock.ReadFile(f)

	if err := os.Mkdir(fakeDir, 0755); err != nil {
		log.Fatal(err)
	}
	log.Println(fakeDir)
	for _, ifce := range m.Interfaces {
		fileName := filepath.Join(fakeDir, snaker.CamelToSnake(ifce.Name)+".go")
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

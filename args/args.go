package args

import (
	"errors"
	"path/filepath"

	flag "github.com/spf13/pflag"
)

var (
	FakesDir string
	Select   []string
)

func init() {
	flag.StringVarP(&FakesDir, "fake-dir", "d", "", "the directory to create the mocks package in. If unset, it will default to 'path/fake")
	flag.StringArrayVarP(&Select, "select", "s", nil, "specify which interfaces to generate mocks for. Can be a comma separated list or used repeatedly.")
}

func Parse() (string, error) {
	flag.Parse()

	if flag.NArg() != 1 {
		return "", errors.New("missing path to file/dir")
	}
	path := flag.Arg(0)

	if FakesDir == "" {
		FakesDir = filepath.Join(filepath.Dir(path), "fake")
	}

	return path, nil
}

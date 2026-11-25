package shared

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type Case struct {
	Name     string
	Source   string
	Expected string
	Input    string
}

func MustCollect(root, sourceExt, expectedExt, inputExt string) []Case {
	cases, err := collect(root, sourceExt, expectedExt, inputExt)
	if err != nil {
		panic(err)
	}
	return cases
}

func collect(root, sourceExt, expectedExt, inputExt string) ([]Case, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("cannot get source file path")
	}

	dataDir := filepath.Join(filepath.Dir(filepath.Dir(file)), "data")
	rootPath := filepath.Join(dataDir, root)

	byName := map[string]string{}
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		ext := filepath.Ext(path)
		stem := strings.TrimSuffix(path, ext)
		byName[filepath.Base(stem)] = stem
		return nil
	})
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(byName))
	for n := range byName {
		names = append(names, n)
	}
	sort.Strings(names)

	cases := make([]Case, 0, len(names))
	for _, name := range names {
		stem := byName[name]
		c := Case{
			Name:     name,
			Source:   stem + sourceExt,
			Expected: stem + expectedExt,
		}
		if inputExt != "" {
			inputPath := stem + inputExt
			_, err := os.Stat(inputPath)
			if err == nil {
				c.Input = inputPath
			}
		}
		cases = append(cases, c)
	}
	return cases, nil
}

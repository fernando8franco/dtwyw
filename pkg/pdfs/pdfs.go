package pdfs

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func GetFromRoute(pdfsDirPath string) ([]string, error) {
	entries, err := os.ReadDir(pdfsDirPath)
	if err != nil {
		return nil, err
	}

	pdfs := []string{}
	for _, e := range entries {
		if !e.IsDir() && strings.ToLower(filepath.Ext(e.Name())) == ".pdf" {
			pdfs = append(pdfs, e.Name())
		}
	}

	return pdfs, err
}

func GetFromDir(dir string) ([]string, error) {
	pdfs := []string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.EqualFold(filepath.Ext(d.Name()), ".pdf") {
			pdfs = append(pdfs, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(pdfs)

	return pdfs, nil
}

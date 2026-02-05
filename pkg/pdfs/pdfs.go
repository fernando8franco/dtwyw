package pdfs

import (
	"os"
	"path/filepath"
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

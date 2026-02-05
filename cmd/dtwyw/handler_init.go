package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fernando8franco/dtwyw/pkg/pdfs"
	"github.com/fernando8franco/dtwyw/pkg/slug"
)

func HandlerInit(s *state, cmd command) error {
	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)

	help := fs.Bool(initHelpFlag, false, "Show help message")
	title := fs.String(initTitleFlag, "", "Title for the config file")
	author := fs.String(initAuthorFlag, "", "Author name for the config file")

	err := fs.Parse(cmd.Arguments)
	if err != nil {
		return nil
	}

	if *help {
		fs.Usage()
		return nil
	}

	if _, err := os.Stat(s.paths.configPDFsFile); !errors.Is(err, os.ErrNotExist) {
		// fmt.Printf("PDF's config file not found\nTry '%s -help' for more information\n", cmd.Name)
		// return nil

		fmt.Println("The config pdfs file is already created")
		fmt.Print("You want to delete it and create another one? (y/n) ")
		var answer string
		fmt.Scan(&answer)

		lowAnswer := strings.ToLower(answer)
		if lowAnswer != "y" && lowAnswer != "yes" {
			return nil
		}
	}

	err = generateConfigPdfsFile(s.paths.pdfsDir, s.paths.configPDFsFile, *title, *author)
	if err != nil {
		return fmt.Errorf("error generating config pdfs file: %v", err)
	}

	return nil
}

type PDFsConfig struct {
	Path    string `json:"path"`
	NewName string `json:"new_name"`
	Title   string `json:"title"`
	Author  string `json:"author"`
}

func generateConfigPdfsFile(pdfDir, configPDFsFilePath, title, author string) error {
	cfgPDFsFile, err := os.Create(configPDFsFilePath)
	if err != nil {
		return err
	}
	defer cfgPDFsFile.Close()

	pdfs, err := pdfs.GetFromRoute(pdfDir)
	if err != nil {
		return err
	}

	if len(pdfs) == 0 {
		fmt.Println("No pdfs found.")
		return nil
	}

	pdfsInfo := map[string]PDFsConfig{}
	for _, pdf := range pdfs {
		filenameWithoutExt := strings.TrimSuffix(pdf, pdfExt)
		newFilename := slug.Create(filenameWithoutExt) + pdfExt

		var metaTitle string
		if title == titleFilename {
			metaTitle = filenameWithoutExt
		} else {
			metaTitle = title
		}

		pdfsInfo[pdf] = PDFsConfig{
			Path:    filepath.Join(pdfDir, pdf),
			NewName: newFilename,
			Title:   metaTitle,
			Author:  author,
		}
	}

	encoder := json.NewEncoder(cfgPDFsFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(pdfsInfo); err != nil {
		return err
	}

	return nil
}

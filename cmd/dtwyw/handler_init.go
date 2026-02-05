package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fernando8franco/dtwyw/pkg/slug"
)

func HandlerInit(s *state, cmd command) error {
	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	help := fs.Bool("help", false, "show help message")
	title := fs.String("title", "", "Title for the config file")
	author := fs.String("author", "", "Author name for the config file")

	err := fs.Parse(cmd.Arguments)
	if err != nil {
		return nil
	}

	if *help {
		fs.Usage()
		return nil
	}

	if _, err := os.Stat(s.paths.configPDFsFile); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("PDF's config file not found\nTry '%s -help' for more information\n", cmd.Name)
		return nil
	}

	fmt.Println("The config pdfs file is already created")
	fmt.Print("You want to delete it and create another one? (y/n) ")
	var answer string
	fmt.Scan(&answer)

	lowAnswer := strings.ToLower(answer)
	if lowAnswer != "y" && lowAnswer != "yes" {
		return nil
	}

	generateConfigPdfsFile(s.paths.pdfsDir, s.paths.configPDFsFile, *title, *author)

	return nil
}

func getPDFs(pdfsDirPath string) (pdfs map[string]string, err error) {
	entries, err := os.ReadDir(pdfsDirPath)
	if err != nil {
		return nil, err
	}

	pdfs = map[string]string{}
	for _, e := range entries {
		if !e.IsDir() && strings.ToLower(filepath.Ext(e.Name())) == ".pdf" {
			pdfs[e.Name()] = pdfsDirPath
		}
	}

	return pdfs, err
}

func generateConfigPdfsFile(pdfDir, configPDFsFilePath, title, author string) error {
	cfgPDFsFile, err := os.Create(configPDFsFilePath)
	if err != nil {
		return err
	}
	defer cfgPDFsFile.Close()

	pdfs, err := getPDFs(pdfDir)
	if err != nil {
		return err
	}

	pdfsInfo := map[string]PDFsConfig{}
	for filename, path := range pdfs {
		ext := filepath.Ext(filename)
		filenameWithoutExt := filename[:len(filename)-len(ext)]
		newFilename := slug.GenerateSlug(filenameWithoutExt) + ext

		var metaTitle string
		if title == "filename" {
			metaTitle = filenameWithoutExt
		} else {
			metaTitle = title
		}

		pdfsInfo[filename] = PDFsConfig{
			Path:    path,
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

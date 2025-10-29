package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fernando8franco/dtwyw/pkg/api"
	"github.com/fernando8franco/dtwyw/pkg/slug"
)

const (
	dtwywDir        = "dtwyw"
	pdfsDir         = "pdfs"
	compressPdfsDir = "compress_pdfs"
	configPDFsFile  = "config.pdfs.json"
)

type PDFsConfig struct {
	Path    string `json:"path"`
	NewName string `json:"new_name"`
	Title   string `json:"title"`
	Author  string `json:"author"`
}

func HandlerCompress(s *state, cmd command) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	pdfsDirPath := filepath.Join(homeDir, dtwywDir, pdfsDir)
	configPDFsFilePath := filepath.Join(homeDir, dtwywDir, pdfsDir, configPDFsFile)

	if len(cmd.Arguments) == 5 {
		if cmd.Arguments[0] != "-f" ||
			cmd.Arguments[1] != "-title" ||
			cmd.Arguments[3] != "-author" {
			return fmt.Errorf("usage: %v -f -title \"title\" -author \"author\"", cmd.Name)
		}

		title := cmd.Arguments[2]
		author := cmd.Arguments[4]

		generateConfigPdfsFile(pdfsDirPath, configPDFsFilePath, title, author)

		return nil
	}

	if _, err := os.Stat(configPDFsFilePath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("the config pdfs file is not created please run: %v -f -title \"title\" -author \"author\"", cmd.Name)
	}

	pdfs, err := getConfigPdfsFile(configPDFsFilePath)
	if err != nil {
		return err
	}

	defer func() {
		err := os.Remove(configPDFsFilePath)
		if err != nil {
			fmt.Printf("Error al eliminar archivo: %v\n", err)
		}
	}()

	keyInfo := s.cfg.GetKeyInfo()
	key := keyInfo.Key
	token := keyInfo.Token

	for filename, info := range pdfs {
		filePath := filepath.Join(info.Path, filename)
		fmt.Println(filename, filePath)

		startResponse, err := callWithRetry(
			s,
			key,
			&token,
			func(t string) (api.StartResponse, error) {
				return api.Start(t)
			},
		)
		if err != nil {
			return err
		}

		fmt.Println(startResponse)
		server := startResponse.Server
		task := startResponse.Task

		uploadResponse, err := callWithRetry(
			s,
			key,
			&token,
			func(t string) (api.UploadResponse, error) {
				return api.Upload(server, task, filePath, t)
			},
		)
		if err != nil {
			return err
		}

		fmt.Println(uploadResponse)
		serverFilename := uploadResponse.ServerFilename

		processResponse, err := callWithRetry(
			s,
			key,
			&token,
			func(t string) (api.ProcessResponse, error) {
				return api.Process(server, task, serverFilename, filename, info.Title, info.Author, t)
			},
		)
		if err != nil {
			return err
		}

		fmt.Println(processResponse)
		compressPdfsPath := strings.Replace(info.Path, pdfsDir, compressPdfsDir, 1)
		compressPdfsPath = filepath.Join(compressPdfsPath, info.NewName)

		dowloadResponse, err := callWithRetry(
			s,
			key,
			&token,
			func(t string) (api.DowloadResponse, error) {
				return api.Dowload(server, task, compressPdfsPath, t)
			},
		)
		if err != nil {
			return err
		}

		fmt.Println(dowloadResponse)

		err = os.Remove(filePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func callWithRetry[T any](s *state, key string, token *string, apiFunc func(t string) (T, error)) (T, error) {
	response, err := apiFunc(*token)

	if errors.Is(err, api.ErrUnauthorized) {
		newToken, errToken := api.GetToken(key)
		if errToken != nil {
			var zero T
			return zero, errToken
		}

		errToken = s.cfg.SetToken(key, newToken)
		if errToken != nil {
			var zero T
			return zero, errToken
		}

		*token = newToken

		response, err = apiFunc(*token)
	}

	return response, err
}

func generateConfigPdfsFile(homeDir, configPDFsFilePath, title, author string) error {
	cfgPDFsFile, err := os.Create(configPDFsFilePath)
	if err != nil {
		return err
	}
	defer cfgPDFsFile.Close()

	pdfs, err := getPDFs(homeDir)
	if err != nil {
		return err
	}

	pdfsInfo := map[string]PDFsConfig{}
	for filename, path := range pdfs {
		ext := filepath.Ext(filename)
		filenameWithoutExt := filename[:len(filename)-len(ext)]
		newFilename := slug.GenerateSlug(filenameWithoutExt) + ext

		pdfsInfo[filename] = PDFsConfig{
			Path:    path,
			NewName: newFilename,
			Title:   title,
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

func getConfigPdfsFile(cfgPDFsFile string) (map[string]PDFsConfig, error) {
	configPdfsFile, err := os.Open(cfgPDFsFile)
	if err != nil {
		return nil, err
	}
	defer configPdfsFile.Close()

	var cfg map[string]PDFsConfig
	if err := json.NewDecoder(configPdfsFile).Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

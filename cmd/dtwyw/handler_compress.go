package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/fernando8franco/dtwyw/pkg/api"
)

const (
	dtwywDir        = "dtwyw"
	pdfsDir         = "pdfs"
	compressPdfsDir = "compress_pdfs"
)

type PDFsConfig struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	NewName string `json:"new_name"`
	Title   string `json:"title"`
	Author  string `json:"author"`
}

func HandlerCompress(s *state, cmd command) error {
	generateConfigPdfsFile()
	/* pdfs, err := getPDFs()
	if err != nil {
		return err
	}

	for filename, path := range pdfs {
		fmt.Println(filename + " " + path)

		keyInfo := s.cfg.GetKeyInfo()
		key := keyInfo.Key
		token := keyInfo.Token

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
				return api.Upload(server, task, path, t)
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
				return api.Process(server, task, serverFilename, filename, "test", "UAEH", t)
			},
		)
		if err != nil {
			return err
		}

		fmt.Println(processResponse)
		compressPdfsPath := strings.Replace(path, pdfsDir, compressPdfsDir, 1)

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
	} */

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

func generateConfigPdfsFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPdfsPath := filepath.Join(homeDir, dtwywDir, pdfsDir, "config.pdfs.json")
	configPdfsFile, err := os.Create(configPdfsPath)
	if err != nil {
		return err
	}
	defer configPdfsFile.Close()

	pdfs, err := getPDFs()
	if err != nil {
		return err
	}

	pdfsInfo := []PDFsConfig{}
	for filename, path := range pdfs {
		newPDF := PDFsConfig{
			Name: filename,
			Path: path,
		}
		pdfsInfo = append(pdfsInfo, newPDF)
	}

	encoder := json.NewEncoder(configPdfsFile)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(pdfsInfo); err != nil {
		return err
	}

	return nil
}

func getPDFs() (pdfs map[string]string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	pdfsDir := filepath.Join(homeDir, dtwywDir, pdfsDir)
	entries, err := os.ReadDir(pdfsDir)
	if err != nil {
		return nil, err
	}

	pdfs = map[string]string{}
	for _, e := range entries {
		if !e.IsDir() && strings.ToLower(filepath.Ext(e.Name())) == ".pdf" {
			pdfs[e.Name()] = filepath.Join(pdfsDir, e.Name())
		}
	}

	return pdfs, err
}

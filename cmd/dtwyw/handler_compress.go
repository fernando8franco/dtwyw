package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	ilApi "github.com/fernando8franco/i-love-api-golang"
	"golang.org/x/sync/errgroup"
)

func HandlerCompress(s *state, cmd command) error {
	if _, err := os.Stat(s.paths.configPDFsFile); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("PDF's config file not found\nTry '%s' first.\n", initCmd)
		return nil
	}

	pdfs, err := getConfigPdfsFile(s.paths.configPDFsFile)
	if err != nil {
		return err
	}

	keyInfo := s.cfg.GetKeyInfo()
	client := &http.Client{}
	iloveapi := ilApi.NewClient(client, keyInfo.Key, keyInfo.Token)

	pdfsChannel := make(chan PDFsConfig)
	var wg errgroup.Group
	for range 3 {
		wg.Go(func() error {
			for pdf := range pdfsChannel {
				fmt.Println("Compressing:", pdf.Filename)

				startResponse, err := callWithRetry(s, iloveapi, func() (ilApi.StartResponse, error) {
					return iloveapi.Start(s.cfg.GetToken(), toolCompress, region)
				})
				if err != nil {
					return err
				}

				server := startResponse.Server
				task := startResponse.Task
				pdfFile := filepath.Join(s.paths.pdfsDir, pdf.Filename)

				file, err := os.Open(pdfFile)
				if err != nil {
					return err
				}
				defer file.Close()

				uploadResponse, err := callWithRetry(s, iloveapi, func() (ilApi.UploadResponse, error) {
					return iloveapi.Upload(s.cfg.GetToken(), server, ilApi.UploadRequest{
						Task:     task,
						File:     file,
						FileName: pdf.Filename,
					})
				})
				if err != nil {
					return err
				}

				serverFilename := uploadResponse.ServerFilename
				_, err = callWithRetry(s, iloveapi, func() (ilApi.ProcessResponse, error) {
					return iloveapi.Process(s.cfg.GetToken(), server, ilApi.ProcessRequest{
						Task: task,
						Tool: toolCompress,
						Files: []ilApi.Files{
							{
								ServerFileName: serverFilename,
								FileName:       pdf.Filename,
							},
						},
						Meta: ilApi.Meta{
							Title:  pdf.Title,
							Author: pdf.Author,
						},
					})
				})
				if err != nil {
					return err
				}

				compressPdfPath := filepath.Join(s.paths.compressPdfsDir, pdf.NewName)

				out, err := os.Create(compressPdfPath)
				if err != nil {
					return err
				}
				defer out.Close()

				dowloadResponse, err := callWithRetry(s, iloveapi, func() (io.ReadCloser, error) {
					return iloveapi.Dowload(s.cfg.GetToken(), server, task, out)
				})
				if err != nil {
					return err
				}
				defer dowloadResponse.Close()

				_, err = io.Copy(out, dowloadResponse)
				if err != nil {
					return err
				}

				err = os.Remove(pdfFile)
				if err != nil {
					return err
				}
				fmt.Println(pdf.Filename, "--- Compress GOOD")
			}

			return nil
		})
	}

	go func() {
		defer close(pdfsChannel)
		for _, info := range pdfs {
			filePath := filepath.Join(s.paths.pdfsDir, info.Filename)
			if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
				continue
			}
			pdfsChannel <- info
		}
	}()

	if err := wg.Wait(); err != nil {
		return err
	} else {
		fmt.Println("All pdfs were compressed correctly")
	}

	err = os.Remove(s.paths.configPDFsFile)
	if err != nil {
		return err
	}

	return nil
}

func callWithRetry[T any](s *state, iloveAPI *ilApi.Client, apiFunc func() (T, error)) (T, error) {
	response, err := apiFunc()

	if err != nil {
		if isUnauthorized(err) {
			err = getToken(s, iloveAPI, s.cfg.GetToken())
			if err != nil {
				return response, err
			}

			response, err = apiFunc()
		}
	}

	return response, err
}

func isUnauthorized(err error) bool {
	type unauthorized interface{ IsUnauthorized() bool }
	var u unauthorized
	return errors.As(err, &u) && u.IsUnauthorized()
}

func getToken(s *state, iloveAPI *ilApi.Client, routineToken string) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	currentKey := s.cfg.GetKeyInfo().Key
	currentSavedToken := s.cfg.GetToken()

	if routineToken == currentSavedToken {
		fmt.Println("Refresing Token...")
		newToken, err := iloveAPI.GenerateToken(currentKey)
		if err != nil {
			return err
		}

		err = s.cfg.SetToken(currentKey, newToken)
		if err != nil {
			return err
		}
	}

	return nil
}

func getConfigPdfsFile(cfgPDFsFile string) ([]PDFsConfig, error) {
	configPdfsFile, err := os.Open(cfgPDFsFile)
	if err != nil {
		return nil, err
	}
	defer configPdfsFile.Close()

	var cfg []PDFsConfig
	if err := json.NewDecoder(configPdfsFile).Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

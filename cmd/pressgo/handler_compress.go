package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fernando8franco/pressgo/pkg/pdfs"
	"github.com/fernando8franco/pressgo/pkg/slug"
)

func HandlerCompress(s *state, cmd command) error {
	fs := flag.NewFlagSet(cmd.Name, flag.ExitOnError)
	var (
		help = fs.Bool(initHelpFlag, false, "Show help message")
		init = fs.Bool(initFlag, false, "Create config file -init <title> <author>\nIf title == 'base', all filenames default to the base name.")
		// noInit = fs.Bool(noInitFlag, false, "Compress files without config file -no-init")
	)
	fs.Parse(cmd.Arguments)

	if *help {
		fs.Usage()
		return nil
	}

	if *init {
		cmd.Arguments = fs.Args()
		if len(cmd.Arguments) != 2 {
			fmt.Printf("Error: -init requires exactly two arguments: title and author.\nUsage: pressgo -init <title> <author>\n")
			os.Exit(1)
		}

		initConfig(s, cmd)
	}

	// if _, err := os.Stat(s.paths.configPDFsFile); errors.Is(err, os.ErrNotExist) {
	// 	fmt.Printf("PDF's config file not found\nTry '%s' first.\n", initCmd)
	// 	return nil
	// }

	// pdfs, err := getConfigPdfsFile(s.paths.configPDFsFile)
	// if err != nil {
	// 	return err
	// }

	// keyInfo := s.cfg.GetKeyInfo()
	// client := &http.Client{}
	// iloveapi := ilApi.NewClient(client, keyInfo.Key, keyInfo.Token)

	// pdfsChannel := make(chan PDFsConfig)
	// var wg errgroup.Group
	// for range 3 {
	// 	wg.Go(func() error {
	// 		for pdf := range pdfsChannel {
	// 			fmt.Println("Compressing:", pdf.Filename)

	// 			startResponse, err := callWithRetry(s, iloveapi, func() (ilApi.StartResponse, error) {
	// 				return iloveapi.Start(s.cfg.GetKeyInfo().Token, toolCompress, region)
	// 			})
	// 			if err != nil {
	// 				return err
	// 			}

	// 			server := startResponse.Server
	// 			task := startResponse.Task
	// 			pdfFile := filepath.Join(s.paths.pdfsDir, pdf.Filename)
	// 			file, err := os.Open(pdfFile)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			defer file.Close()

	// 			uploadResponse, err := callWithRetry(s, iloveapi, func() (ilApi.UploadResponse, error) {
	// 				return iloveapi.Upload(s.cfg.GetKeyInfo().Token, server, ilApi.UploadRequest{
	// 					Task:     task,
	// 					File:     file,
	// 					FileName: pdf.Filename,
	// 				})
	// 			})
	// 			if err != nil {
	// 				return err
	// 			}

	// 			serverFilename := uploadResponse.ServerFilename
	// 			_, err = callWithRetry(s, iloveapi, func() (ilApi.ProcessResponse, error) {
	// 				return iloveapi.Process(s.cfg.GetKeyInfo().Token, server, ilApi.ProcessRequest{
	// 					Task: task,
	// 					Tool: toolCompress,
	// 					Files: []ilApi.Files{
	// 						{
	// 							ServerFileName: serverFilename,
	// 							FileName:       pdf.Filename,
	// 						},
	// 					},
	// 					Meta: ilApi.Meta{
	// 						Title:  pdf.Title,
	// 						Author: pdf.Author,
	// 					},
	// 				})
	// 			})
	// 			if err != nil {
	// 				return err
	// 			}

	// 			compressPdfPath := filepath.Join(s.paths.compressPdfsDir, pdf.NewName)
	// 			out, err := os.Create(compressPdfPath)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			defer out.Close()

	// 			dowloadResponse, err := callWithRetry(s, iloveapi, func() (io.ReadCloser, error) {
	// 				return iloveapi.Dowload(s.cfg.GetKeyInfo().Token, server, task, out)
	// 			})
	// 			if err != nil {
	// 				return err
	// 			}
	// 			defer dowloadResponse.Close()

	// 			_, err = io.Copy(out, dowloadResponse)
	// 			if err != nil {
	// 				return err
	// 			}

	// 			err = os.Remove(pdfFile)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			fmt.Println(pdf.Filename, "--- Compressed correctly")
	// 		}

	// 		return nil
	// 	})
	// }

	// go func() {
	// 	defer close(pdfsChannel)
	// 	for _, info := range pdfs {
	// 		filePath := filepath.Join(s.paths.pdfsDir, info.Filename)
	// 		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
	// 			continue
	// 		}
	// 		pdfsChannel <- info
	// 	}
	// }()

	// if err := wg.Wait(); err != nil {
	// 	return err
	// } else {
	// 	fmt.Println("All pdfs were compressed correctly")
	// }

	// err = os.Remove(s.paths.configPDFsFile)
	// if err != nil {
	// 	return err
	// }

	// return nil
	return nil
}

// func callWithRetry[T any](s *state, iloveAPI *ilApi.Client, apiFunc func() (T, error)) (T, error) {
// 	response, err := apiFunc()

// 	if err != nil {
// 		if isUnauthorized(err) {
// 			err = checkToken(s, iloveAPI, s.cfg.GetKeyInfo().Token)
// 			if err != nil {
// 				return response, err
// 			}

// 			response, err = apiFunc()
// 		}
// 	}

// 	return response, err
// }

// func isUnauthorized(err error) bool {
// 	type unauthorized interface{ IsUnauthorized() bool }
// 	var u unauthorized
// 	return errors.As(err, &u) && u.IsUnauthorized()
// }

// func checkToken(s *state, iloveAPI *ilApi.Client, routineToken string) (err error) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	currentKey := s.cfg.GetKeyInfo().Key
// 	currentSavedToken := s.cfg.GetKeyInfo().Token

// 	if routineToken == currentSavedToken {
// 		fmt.Println("Refresing Token...")
// 		newToken, err := iloveAPI.GenerateToken(currentKey)
// 		if err != nil {
// 			return err
// 		}

// 		err = s.cfg.SetToken(currentKey, newToken)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func getConfigPdfsFile(cfgPDFsFile string) ([]PDFsConfig, error) {
// 	configPdfsFile, err := os.Open(cfgPDFsFile)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer configPdfsFile.Close()

// 	var cfg []PDFsConfig
// 	if err := json.NewDecoder(configPdfsFile).Decode(&cfg); err != nil {
// 		return nil, err
// 	}

// 	return cfg, nil
// }

func initConfig(s *state, cmd command) error {
	configFile := path.Join(s.wdir, configFile)
	if _, err := os.Stat(configFile); !errors.Is(err, os.ErrNotExist) {
		fmt.Println("The config pdfs file is already created")
		fmt.Print("You want to delete it and create another one? (y/n) ")
		var answer string
		fmt.Scan(&answer)

		lowAnswer := strings.ToLower(answer)
		if lowAnswer != "y" && lowAnswer != "yes" {
			return nil
		}
	}

	title := cmd.Arguments[1]
	author := cmd.Arguments[0]

	err := generateConfigPdfsFile(s.wdir, configFile, title, author)
	if err != nil {
		return fmt.Errorf("error generating config pdfs file: %v", err)
	}

	return nil
}

type PDFsConfig struct {
	Filename string `json:"filename"`
	NewName  string `json:"new_name"`
	Title    string `json:"title"`
	Author   string `json:"author"`
}

func generateConfigPdfsFile(pdfDir, configPDFsFilePath, title, author string) error {
	cfgPDFsFile, err := os.Create(configPDFsFilePath)
	if err != nil {
		return err
	}
	defer cfgPDFsFile.Close()

	pdfs, err := pdfs.GetFromDir(pdfDir)
	if err != nil {
		return err
	}

	if len(pdfs) == 0 {
		return fmt.Errorf("No pdfs found.")
	}

	pdfsInfo := []PDFsConfig{}
	for _, pdf := range pdfs {
		base := filepath.Base(pdf)
		ext := filepath.Ext(pdf)
		filenameWithoutExt := strings.TrimSuffix(base, ext)
		newFilename := slug.Create(filenameWithoutExt) + pdfExt

		var metaTitle string
		if title == titleFilename {
			metaTitle = filenameWithoutExt
		} else {
			metaTitle = title
		}

		pdfsInfo = append(pdfsInfo, PDFsConfig{
			Filename: pdf,
			NewName:  newFilename,
			Title:    metaTitle,
			Author:   author,
		})
	}

	encoder := json.NewEncoder(cfgPDFsFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(pdfsInfo); err != nil {
		return err
	}

	return nil
}

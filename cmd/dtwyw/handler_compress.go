package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	apiTest "github.com/fernando8franco/i-love-api-golang"
	"golang.org/x/sync/errgroup"
)

func HandlerCompress(s *state, cmd command) error {
	var title string
	var author string
	var initFlag bool

	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Println("Usage: dtwyw compress [options]")
		fmt.Println("\nOptions:")
		fs.PrintDefaults()
		fmt.Println("\nExample: dtwyw compress --init --title 'My File' --author 'Franco'")
	}
	initMode := fs.Bool("init", false, "Generate a configuration file")
	title2 := fs.String("title", "", "Title for the config")
	author2 := fs.String("author", "", "Author name")

	fs.Parse(cmd.Arguments)
	// if err := fs.Parse(cmd.Arguments); err != nil {
	// 	return err
	// }

	if !*initMode && (len(*title2) >= 0 || len(*author2) >= 0) {
		fmt.Println("The --title and --author flags can only be used with the --init flag")
		return nil
	}

	fmt.Println("2")
	return nil

	// if len(cmd.Arguments) > 0 {
	// 	switch cmd.Arguments[0] {
	// 	case "-h", "--help":
	// 		fmt.Println(getUsageMessage(cmd.Name))
	// 		return nil
	// 	case "-i", "--init":
	// 		fmt.Println("init")
	// 	default:
	// 		fmt.Printf("%s: invalid option -- '%s'\nTry '%s --help' for more information\n", cmd.Name, cmd.Arguments[0], cmd.Name)
	// 		return nil
	// 	}
	// }

	if len(cmd.Arguments) > 0 && cmd.Arguments[0] == "--help" {
		fmt.Println(getUsageMessage(cmd.Name))
		return nil
	}

	if len(cmd.Arguments) > 0 && cmd.Arguments[0] == "--init" {
		initFlag = true
		for i := 0; i < len(cmd.Arguments); i++ {
			switch cmd.Arguments[i] {
			case "--title":
				if i+1 < len(cmd.Arguments) {
					title = cmd.Arguments[i+1]
					i++
				}
			case "--author":
				if i+1 < len(cmd.Arguments) {
					author = cmd.Arguments[i+1]
					i++
				}
			}
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	pdfsDirPath := filepath.Join(homeDir, dtwywDir, pdfsDir)
	configPDFsFilePath := filepath.Join(homeDir, dtwywDir, pdfsDir, configPDFsFile)

	if initFlag {
		if _, err := os.Stat(configPDFsFilePath); !errors.Is(err, os.ErrNotExist) {
			fmt.Println("The config pdfs file is already created")
			return nil
		}

		generateConfigPdfsFile(pdfsDirPath, configPDFsFilePath, title, author)

		return nil
	} else {
		if _, err := os.Stat(configPDFsFilePath); errors.Is(err, os.ErrNotExist) {
			fmt.Printf("PDF's config file not found\nTry '%s --help' for more information\n", cmd.Name)
			return nil
		}
	}

	pdfs, err := getConfigPdfsFile(configPDFsFilePath)
	if err != nil {
		return err
	}

	keyInfo := s.cfg.GetKeyInfo()
	type PDF struct {
		Filename string
		Info     PDFsConfig
	}

	pdfsChannel := make(chan PDF)

	tool := "compress"
	client := &http.Client{}
	iloveapi := apiTest.NewClient(client, keyInfo.Key, keyInfo.Token)

	var wg errgroup.Group
	for range 3 {
		wg.Go(func() error {
			for pdf := range pdfsChannel {
				fmt.Println("Compressing:", pdf.Filename)
				routineToken := iloveapi.GetToken()

				startResponse, err := callWithRetry(s, iloveapi, routineToken, func() (apiTest.StartResponse, error) {
					return iloveapi.Start(tool, "us")
				})
				if err != nil {
					return err
				}

				server := startResponse.Server
				task := startResponse.Task
				pdfFile := filepath.Join(pdf.Info.Path, pdf.Filename)

				file, err := os.Open(pdfFile)
				if err != nil {
					return err
				}
				defer file.Close()

				uploadResponse, err := callWithRetry(s, iloveapi, routineToken, func() (apiTest.UploadResponse, error) {
					return iloveapi.Upload(server, apiTest.UploadRequest{
						Task:     task,
						File:     file,
						FileName: pdf.Filename,
					})
				})
				if err != nil {
					return err
				}

				serverFilename := uploadResponse.ServerFilename
				_, err = callWithRetry(s, iloveapi, routineToken, func() (apiTest.ProcessResponse, error) {
					return iloveapi.Process(server, apiTest.ProcessRequest{
						Task: task,
						Tool: tool,
						Files: []apiTest.Files{
							{
								ServerFileName: serverFilename,
								FileName:       pdf.Filename,
							},
						},
						Meta: apiTest.Meta{
							Title:  pdf.Info.Title,
							Author: pdf.Info.Author,
						},
					})
				})
				if err != nil {
					return err
				}

				compressPdfPath := strings.Replace(pdf.Info.Path, pdfsDir, compressPdfsDir, 1)
				compressPdfPath = filepath.Join(compressPdfPath, pdf.Info.NewName)

				out, err := os.Create(compressPdfPath)
				if err != nil {
					return err
				}
				defer out.Close()

				dowloadResponse, err := callWithRetry(s, iloveapi, routineToken, func() (io.ReadCloser, error) {
					return iloveapi.Dowload(server, task, out)
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
		for filename, info := range pdfs {
			filePath := filepath.Join(info.Path, filename)
			if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
				continue
			}
			pdf := PDF{
				Filename: filename,
				Info:     info,
			}
			pdfsChannel <- pdf
		}
	}()

	if err := wg.Wait(); err != nil {
		return err
	} else {
		fmt.Println("All pdfs were compressed correctly")
	}

	err = os.Remove(configPDFsFilePath)
	if err != nil {
		return err
	}

	return nil
}

func callWithRetry[T any](s *state, iloveAPI *apiTest.Client, routineToken string, apiFunc func() (T, error)) (T, error) {
	response, err := apiFunc()

	if err != nil {
		if isUnauthorized(err) {
			err = getToken(s, iloveAPI, routineToken)
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

func getToken(s *state, iloveAPI *apiTest.Client, routineToken string) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	currentSavedToken := s.cfg.GetToken()

	if routineToken == currentSavedToken {
		fmt.Println("Refresing Token...")
		newToken, err := iloveAPI.GenerateToken()
		if err != nil {
			return err
		}
		iloveAPI.SetToken(newToken)

		err = s.cfg.SetToken(iloveAPI.GetAPIKey(), newToken)
		if err != nil {
			return err
		}

	} else {
		iloveAPI.SetToken(currentSavedToken)
	}

	return nil
}

func getUsageMessage(commandName string) (usageMessage string) {
	return fmt.Sprintf(`Usage: %s [--init [--title "TITLE"] [--author "AUTHOR"]]

Options:
  --init   Creation of the config file (optional)
		--title "title"     Title to associate with the file (optional for init)
		--author "author"   Author name to associate with the file (optional for init)

Note: If you put "filename" as title, the name of the file would be the title
`, commandName)
}

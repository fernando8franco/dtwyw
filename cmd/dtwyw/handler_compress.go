package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	api "github.com/fernando8franco/dtwyw/pkg/iloveapi"
	"github.com/fernando8franco/dtwyw/pkg/slug"
	"golang.org/x/sync/errgroup"
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

	var title string
	var author string
	var initFlag bool

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

	if initFlag {
		if _, err := os.Stat(configPDFsFilePath); !errors.Is(err, os.ErrNotExist) {
			return errors.New("the config pdfs file is already created")
		}

		generateConfigPdfsFile(pdfsDirPath, configPDFsFilePath, title, author)

		return nil
	} else {
		if _, err := os.Stat(configPDFsFilePath); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("the config pdfs file is not created\n%s", getUsageMessage(cmd.Name))
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
	iloveapi := api.ILoveAPI{
		Key:   keyInfo.Key,
		Token: keyInfo.Token,
	}

	var wg errgroup.Group
	for range 3 {
		wg.Go(func() error {
			for pdf := range pdfsChannel {
				fmt.Println("Compressing:", pdf.Filename)
				routineToken := iloveapi.Token

				startResponse, err := callWithRetry(s, &iloveapi, routineToken, func() (api.StartResponse, error) {
					return iloveapi.Start()
				})
				if err != nil {
					return err
				}

				server := startResponse.Server
				task := startResponse.Task
				pdfFile := filepath.Join(pdf.Info.Path, pdf.Filename)

				uploadResponse, err := callWithRetry(s, &iloveapi, routineToken, func() (api.UploadResponse, error) {
					return iloveapi.Upload(server, task, pdfFile)
				})
				if err != nil {
					return err
				}

				serverFilename := uploadResponse.ServerFilename
				_, err = callWithRetry(s, &iloveapi, routineToken, func() (api.ProcessResponse, error) {
					return iloveapi.Process(server, task, serverFilename, pdf.Filename, pdf.Info.Title, pdf.Info.Author)
				})
				if err != nil {
					return err
				}

				compressPdfPath := strings.Replace(pdf.Info.Path, pdfsDir, compressPdfsDir, 1)
				compressPdfPath = filepath.Join(compressPdfPath, pdf.Info.NewName)
				dowloadResponse, err := callWithRetry(s, &iloveapi, routineToken, func() (api.DowloadResponse, error) {
					return iloveapi.Dowload(server, task, compressPdfPath)
				})
				if err != nil {
					return err
				}

				err = os.Remove(pdfFile)
				if err != nil {
					return err
				}
				fmt.Println(pdf.Filename, "--- Compress", dowloadResponse.Status)
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
		fmt.Println("Error:", err)
	} else {
		fmt.Println("All pdfs were compressed correctly")
	}

	err = os.Remove(configPDFsFilePath)
	if err != nil {
		return err
	}

	return nil
}

func callWithRetry[T any](s *state, iloveAPI *api.ILoveAPI, routineToken string, apiFunc func() (T, error)) (T, error) {
	response, err := apiFunc()

	if errors.Is(err, api.ErrUnauthorized) {
		err = getToken(s, iloveAPI, routineToken)
		if err != nil {
			return response, err
		}

		response, err = apiFunc()
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

func getToken(s *state, iloveAPI *api.ILoveAPI, routineToken string) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	currentSavedToken := s.cfg.GetToken()

	if routineToken == currentSavedToken {
		fmt.Println("Refresing Token...")
		newToken, err := iloveAPI.GetToken()
		if err != nil {
			return err
		}
		iloveAPI.Token = newToken

		fmt.Println(newToken)
		err = s.cfg.SetToken(iloveAPI.Key, newToken)
		if err != nil {
			return err
		}

	} else {
		iloveAPI.Token = currentSavedToken
	}

	return nil
}

func getUsageMessage(commandName string) (usageMessage string) {
	return fmt.Sprintf(`Usage: %s [--init [--title "TITLE"] [--author "AUTHOR"]]

Options:
  --init   Creation of the config file (optional)
	--title "title"     Title to associate with the file (optional for init)
	--author "author"   Author name to associate with the file (optional for init)

Note: if you put filename as title the name of the file would be the title
`, commandName)
}

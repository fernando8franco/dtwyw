package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const (
	AuthURL    = "https://api.ilovepdf.com/v1/auth"
	StartURL   = "https://api.ilovepdf.com/v1/start/compress/us"
	UploadURL  = "https://%s/v1/upload"
	ProcessURL = "https://%s/v1/process"
	DowloadURL = "https://%s/v1/download/%s"
)

var ErrUnauthorized = errors.New("unauthorized")

type StartResponse struct {
	Server           string `json:"server"`
	Task             string `json:"task"`
	RemainingCredits int    `json:"remaining_credits"`
}

type UploadResponse struct {
	ServerFilename string `json:"server_filename"`
}

type ProcessResponse struct {
	DownloadFilename string `json:"download_filename"`
	Filesize         int    `json:"filesize"`
	OutputFilesize   int    `json:"output_filesize"`
	OutputFilenumber int    `json:"output_filenumber"`
	OutputExtensions string `json:"output_extensions"`
	Timer            string `json:"timer"`
	Status           string `json:"status"`
}

type DowloadResponse struct {
	Status string `json:"status"`
}

func GetToken(key string) (token string, err error) {
	data := struct {
		PublicKey string `json:"public_key"`
	}{
		PublicKey: key,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		"POST",
		AuthURL,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return "", checkErrorCode(res)
	}

	tokenResponse := struct {
		Token string `json:"token"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.Token, nil
}

func Start(token string) (response StartResponse, err error) {
	req, err := http.NewRequest(
		"GET",
		StartURL,
		nil,
	)
	if err != nil {
		return StartResponse{}, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return StartResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return StartResponse{}, checkErrorCode(res)
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return StartResponse{}, err
	}

	return response, nil
}

func Upload(server, task, pdfPath, token string) (response UploadResponse, err error) {
	uploadUrl := fmt.Sprintf(UploadURL, server)

	file, err := os.Open(pdfPath)
	if err != nil {
		return UploadResponse{}, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(pdfPath))
	if err != nil {
		return UploadResponse{}, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return UploadResponse{}, err
	}

	_ = writer.WriteField("task", task)

	err = writer.Close()
	if err != nil {
		return UploadResponse{}, err
	}

	req, err := http.NewRequest(
		"POST",
		uploadUrl,
		body,
	)
	if err != nil {
		return UploadResponse{}, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return UploadResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return UploadResponse{}, checkErrorCode(res)
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return UploadResponse{}, err
	}

	return response, nil
}

func Process(server, task, serverFilename, filename, title, author, token string) (response ProcessResponse, err error) {
	processUrl := fmt.Sprintf(ProcessURL, server)
	type Meta struct {
		Title  string `json:"Title"`
		Author string `json:"Author"`
	}

	type ProcessRequest struct {
		Task  string `json:"task"`
		Tool  string `json:"tool"`
		Files []struct {
			ServerFilename string `json:"server_filename"`
			Filename       string `json:"filename"`
		} `json:"files"`
		Meta Meta `json:"meta"`
	}

	data := ProcessRequest{
		Task: task,
		Tool: "compress",
		Files: []struct {
			ServerFilename string `json:"server_filename"`
			Filename       string `json:"filename"`
		}{
			{ServerFilename: serverFilename, Filename: filename},
		},
		Meta: Meta{
			Title:  title,
			Author: author,
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return ProcessResponse{}, err
	}

	req, err := http.NewRequest(
		"POST",
		processUrl,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return ProcessResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return ProcessResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return ProcessResponse{}, checkErrorCode(res)
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return ProcessResponse{}, err
	}

	return response, nil
}

func Dowload(server, task, pdfPath, token string) (response DowloadResponse, err error) {
	dowloadUrl := fmt.Sprintf(DowloadURL, server, task)

	out, err := os.Create(pdfPath)
	if err != nil {
		return DowloadResponse{}, err
	}
	defer out.Close()

	req, err := http.NewRequest(
		"GET",
		dowloadUrl,
		nil,
	)
	if err != nil {
		return DowloadResponse{}, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return DowloadResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return DowloadResponse{}, checkErrorCode(res)
	}

	_, err = io.Copy(out, res.Body)
	if err != nil {
		return DowloadResponse{}, err
	}

	return DowloadResponse{Status: "Succed"}, err
}

func checkErrorCode(res *http.Response) error {
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	return fmt.Errorf("error: received non-successful status code: %d\nresponse Body: %s", res.StatusCode, string(bodyBytes))
}

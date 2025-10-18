package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	AuthURL = "https://api.ilovepdf.com/v1/auth"
)

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

func Start() {

}

func checkErrorCode(res *http.Response) error {
	bodyBytes, _ := io.ReadAll(res.Body)
	var bodyData map[string]interface{}
	var bodyMsg string

	if err := json.Unmarshal(bodyBytes, &bodyData); err == nil {
		bodyMsg = fmt.Sprintf("%v", bodyData)
	} else {
		bodyMsg = string(bodyBytes)
	}

	return fmt.Errorf("error consulting the auth token: status %d, response: %s", res.StatusCode, bodyMsg)
}

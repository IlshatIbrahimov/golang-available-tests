package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

// авторизация в API
func AuthApi() (string, error) {
	var defaultTransport http.RoundTripper = &http.Transport{Proxy: nil}
	client := &http.Client{Transport: defaultTransport}

	// авторизация на бэке
	authData := map[string]string{"username": AlertApiUsername, "password": AlertApiPassword}
	authJson, _ := json.Marshal(authData)
	authReq, _ := http.NewRequest(http.MethodPost, AlertApiUrl+"/auth/login", bytes.NewBuffer(authJson))
	authReq.Header.Set("Content-Type", "application/json")

	authResp, err := client.Do(authReq)
	if err != nil {
		return "", err
	}
	if authResp.StatusCode != 200 {
		return "", errors.New("Not authenticated! Status: " + authResp.Status)
	}

	defer authResp.Body.Close()
	body, err := ioutil.ReadAll(authResp.Body)
	if err != nil {
		return "", err
	}

	var tokensObject Tokens
	json.Unmarshal([]byte(string(body)), &tokensObject)

	return tokensObject.AccessToken, nil
}

// отправка алерта в API
func CreateAlert(apiToken, testUrl, testStatus string, testError error, duration int64) error {
	var defaultTransport http.RoundTripper = &http.Transport{Proxy: nil}
	client := &http.Client{Transport: defaultTransport}

	sendDataBody := map[string]string{"url": testUrl, "status": testStatus, "duration": strconv.FormatInt(duration, 10), "stand": os.Getenv("HOST"), "error": testError.Error()}
	bodyJson, _ := json.Marshal(sendDataBody)

	req, err := http.NewRequest(http.MethodPost, AlertApiUrl, bytes.NewBuffer(bodyJson))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if resp.StatusCode != 201 {
		return errors.New("Alert not created! Status: " + resp.Status)
	}
	if err != nil {
		return err
	}

	return nil
}

// получение списка url'ов из api
func GetUrlsFromApi(apiToken string) ([]ApiUrl, error) {
	// клиент без прокси
	var defaultTransport http.RoundTripper = &http.Transport{Proxy: nil}
	client := &http.Client{Transport: defaultTransport}

	req, err := http.NewRequest(http.MethodGet, AlertApiUrl+"/url", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	//bytesBody := []byte(body)
	apiUrls := make([]ApiUrl, 0)
	json.Unmarshal(body, &apiUrls)

	return apiUrls, nil
}

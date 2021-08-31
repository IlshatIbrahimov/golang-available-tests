package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
	"log"
)

type Tokens struct {
	AccessToken string
	RefreshToken string
}

var alertApiUrl = os.Getenv("ALERT_API_URL")
var alertApiUsername = os.Getenv("ALERT_API_USERNAME")
var alertApiPassword = os.Getenv("ALERT_API_PASSWORD")
//var urls = strings.Split(os.Getenv("ACCESSIBILITY_TEST_URLS"), ";")
var rawUrls,_ = ioutil.ReadFile("urls.txt")
var urls = strings.Split(string(rawUrls), ";")

func TestAvailable(t *testing.T) {
	// setup: check google.com
	_, _, err := getHttp("https://google.com/")
	if err != nil {
		sendFatalEmail("SetUp failed: google.com недоступен! Ошибка:\n" + err.Error())
		t.Fatalf("SetUp failed: google.com unavailable! Ошибка:\n" + err.Error())
	}

	// tests
	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			// get запрос
			body, duration, testError := getHttp(url)
			if testError != nil {
				t.Fail()
				t.Log(testError)
			}

			// проверка наличия title
			if !assertBodyHasTitle(body) {
				t.Fail()
				t.Log("Ответ не содержит title! Ответ:\n" + body)
			}

			// отправка алертов
			// авторизация на бэке
			alertApiToken, err := authApi()
			if err != nil {
				sendFatalEmail("Не удалось авторизоваться в API! Ошибка:\n" + err.Error())
				t.Fatal("Authentication to API failed! Error:\n" + err.Error())
			}
			// создание алерта
			var status string
			if testError != nil {
				status = "FAIL"
			} else {
				status = "PASS"
				testError = errors.New("")
			}
			err = createAlert(alertApiToken, url, status, testError, int64(duration))
			if err != nil {
				sendFatalEmail("Alert не был создан! Ошибка:\n" + err.Error())
				t.Fatal("Alert not created! Error:\n" + err.Error())
			}
		})
	}
}

// get запрос по url
func getHttp(url string) (string, time.Duration, error) {
	// клиент без прокси
	var defaultTransport http.RoundTripper = &http.Transport{Proxy: nil}
	client := &http.Client{Transport: defaultTransport}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	// время отправки запроса
	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	if resp.StatusCode != 200 {
		return "", 0, errors.New("Status code is not 200! Status: " + resp.Status)
	}

	// время выполнения запроса
	done := time.Since(start)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", done, err
	}

	return string(body), done, nil
}

// проверка наличия title в body
func assertBodyHasTitle(body string) bool {
	return strings.Contains(body, "title")
}

// авторизация в API
func authApi() (string, error) {
	var defaultTransport http.RoundTripper = &http.Transport{Proxy: nil}
	client := &http.Client{Transport: defaultTransport}

	// авторизация на бэке
	authData := map[string]string{"username": alertApiUsername, "password": alertApiPassword}
	authJson, _ := json.Marshal(authData)
	authReq, _ := http.NewRequest(http.MethodPost, alertApiUrl + "/auth/login", bytes.NewBuffer(authJson))
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
func createAlert(apiToken, testUrl, testStatus string, testError error, duration int64) error {
	var defaultTransport http.RoundTripper = &http.Transport{Proxy: nil}
	client := &http.Client{Transport: defaultTransport}

	sendDataBody := map[string]string{"url": testUrl, "status": testStatus, "duration": strconv.FormatInt(duration, 10), "stand": os.Getenv("HOST"), "error": testError.Error()}
	bodyJson, _ := json.Marshal(sendDataBody)

	req, err := http.NewRequest(http.MethodPost, alertApiUrl, bytes.NewBuffer(bodyJson))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer " + apiToken)
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

// отправка алерта на почту в случае фатальных ошибок
func sendFatalEmail(messageText string) error{
	from := os.Getenv("ACCESSIBILITY_SMTP_EMAIL")
	password := os.Getenv("ACCESSIBILITY_SMTP_PASSWORD")
	smtpHost := os.Getenv("ACCESSIBILITY_SMTP_HOST")
	smtpPort := os.Getenv("ACCESSIBILITY_SMTP_PORT")

	to := strings.Split(os.Getenv("FATAL_EMAIL_RECIPIENTS"), ",")

	message := []byte("Subject: Ошибка в тестах доступности страниц!\r\n" + messageText)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost + ":" + smtpPort, auth, from, to, message)
	if err != nil {
		return err
	}
	return nil
}

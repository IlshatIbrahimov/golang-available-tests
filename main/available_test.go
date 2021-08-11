package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
	"errors"
)

var alertApiUrl = os.Getenv("ALERT_API_URL")
var alertApiUsername = os.Getenv("ALERT_API_USERNAME")
var alertApiPassword = os.Getenv("ALERT_API_PASSWORD")
var urls = strings.Split(os.Getenv("ACCESSIBILITY_TEST_URLS"), ";")

func TestAvailable(t *testing.T) {
	var hasFailedTests = false

	// setup
	_, _, err := getHttp("https://google.com/")
	if err != nil {
		sendFatalEmail("SetUp failed: google.com недоступен! Ошибка:\n" + err.Error())
		t.Fatalf("SetUp failed: google.com недоступен! Ошибка: %s", err)
	}

	// tests
	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			var currentFail = false

			body, duration, err := getHttp(url)
			if err != nil {
				currentFail = true
				hasFailedTests = true
				t.Fail()
				t.Log(err)
			}
			if !assertBodyHasTitle(body) {
				currentFail = true
				hasFailedTests = true
				t.Fail()
				t.Logf("Ответ не содержит GTM или GTM неверный! Ответ: %s", body)
			}

			// авторизация на бэке
			alertApiToken, err := authApi()
			if err != nil {
				sendFatalEmail("Не удалось авторизоваться в API! Ошибка:\n" + err.Error())
				t.Fatalf("Authentication to api failed! Error: %s", err)
			}
			// отправка данных
			var status string
			if currentFail {
				status = "FAIL"
			} else {
				status = "PASS"
			}
			err = createAlert(alertApiToken, url, status, int64(duration))
			if err != nil {
				sendFatalEmail("Alert не был создан! Ошибка:\n" + err.Error())
				t.Fatalf("Alert not created! Error: %s", err)
			}
		})
	}

	// teardown
	if hasFailedTests {

	}
}

// get запрос по url
func getHttp(url string) (string, time.Duration, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	var defaultTransport http.RoundTripper = &http.Transport{Proxy: nil}
	client := &http.Client{Transport: defaultTransport}


	start := time.Now()
	resp, err := client.Do(req)
	done := time.Since(start)

	if err != nil {
		return "", 0, err
	}

	if resp.StatusCode != 200 {
		return "", 0, errors.New("Status code is not 200. Status: " + resp.Status)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}

	return string(body), done, nil
}

func assertBodyHasTitle(body string) bool {
	return strings.Contains(body, "title")
}

// авторизация в API
func authApi() (string, error) {
	// авторизация на бэке
	authData := map[string]string{"username": alertApiUsername, "password": alertApiPassword}
	authJson, _ := json.Marshal(authData)
	authReq, _ := http.NewRequest(http.MethodPost, alertApiUrl + "/auth/login", bytes.NewBuffer(authJson))

	var defaultTransport http.RoundTripper = &http.Transport{Proxy: nil}
	client := &http.Client{Transport: defaultTransport}

	authResp, err := client.Do(authReq)
	if err != nil {
		return "", err
	}

	defer authResp.Body.Close()
	body, err := ioutil.ReadAll(authResp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// отправка алерта в API
func createAlert(apiToken, testUrl, testStatus string, duration int64) error {
	sendDataBody := map[string]string{"url": testUrl, "status": testStatus, "duration": strconv.FormatInt(duration, 10), "stand": os.Getenv("HOST")}
	bodyJson, _ := json.Marshal(sendDataBody)

	var defaultTransport http.RoundTripper = &http.Transport{Proxy: nil}
	client := &http.Client{Transport: defaultTransport}

	req, _ := http.NewRequest(http.MethodPost, alertApiUrl, bytes.NewBuffer(bodyJson))
	req.Header.Set("Authentication", "Bearer " + apiToken)

	_, err := client.Do(req)

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

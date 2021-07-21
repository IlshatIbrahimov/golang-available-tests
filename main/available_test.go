package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestAvailable(t *testing.T) {
	var hasFailedTests = false
	var envUrls = os.Getenv("AVAILABLE_TEST_URLS")
	var urls = strings.Split(envUrls, ",")

	// setup
	_, _, err := getHttp("https://google.com/")
	if err != nil {
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
			if !assertBodyHasFooter(body) {
				currentFail = true
				hasFailedTests = true
				t.Fail()
				t.Logf("Ответ не содержит footer! Ответ: %s", body)
			}

			var sendDataBody map[string]string
			// отправка данных на бэк здесь!
			if currentFail {
				sendDataBody = map[string]string{"url": url, "status": "FAIL", "duration": strconv.FormatInt(int64(duration), 10)}
			} else {
				sendDataBody = map[string]string{"url": url, "status": "PASS", "duration": strconv.FormatInt(int64(duration), 10)}
			}

			sendDataJson, _ := json.Marshal(sendDataBody)
			sendData, _ := http.NewRequest(http.MethodPost, "", sendDataJson)
		})
	}

	// teardown
	if hasFailedTests {

	}
}

func getHttp(url string) (string, time.Duration, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	client := http.DefaultClient

	start := time.Now()
	resp, err := client.Do(req)
	done := time.Since(start)

	if err != nil {
		return "", 0, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}

	return string(body), done, nil
}

func assertBodyHasFooter(body string) bool {
	return strings.Contains(body, "footer")
}

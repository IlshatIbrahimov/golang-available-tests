package main

import (
	"io/ioutil"
	"net/http"
	"os"
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
			body, duration, err := getHttp(url)
			if err != nil {
				hasFailedTests = true
				t.Fail()
				t.Log(err)
			}
			if !assertBodyHasFooter(body) {
				hasFailedTests = true
				t.Fail()
				t.Logf("Ответ не содержит footer! Ответ: %s", body)
			}

			// отправка данных на бэк здесь!
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

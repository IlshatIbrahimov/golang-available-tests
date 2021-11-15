package main

import (
	"errors"
	"log"
	"os"
	"strings"
	"testing"
)

func TestAvailable(t *testing.T) {
	// setup: logger
	f, logErr := os.OpenFile("testLogs", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if logErr != nil {
		t.Log("error opening log file")
	}
	defer f.Close()
	log.SetOutput(f)

	// check logs for emptiness and send logged alerts to api
	fi, fiErr := os.Stat("testLogs")
	if fiErr != nil {
		t.Log("error checking log file")
	}
	alertApiToken, err := AuthApi()
	if fi.Size() != 0 {
		buf, err := os.ReadFile("testLogs")
		if err != nil {
			t.Log("error reading log file")
		}
		logAlerts := string(buf)
		logAlertsSlice := strings.Split(logAlerts, "logAlertSeparator")

		for _, alert := range logAlertsSlice {
			err = CreateAlertWithBody(alertApiToken, []byte(alert))
			if err != nil {
				t.Log("Alert is not created!")
				break
			}
			// clear logs
			err = os.Remove("testLogs")
			if err != nil {
				t.Log("Cannot delete log file")
			}
		}
	}

	// setup: check google.com
	_, _, err = GetHttp("https://google.com/")
	if err != nil {
		SendFatalEmail("SetUp failed: google.com недоступен! Ошибка:\n" + err.Error())
		t.Log("SetUp failed: google.com unavailable! Ошибка:\n" + err.Error())
		t.FailNow()
	}

	// setup: get urls

	if err != nil {
		SendFatalEmail("Не удалось авторизоваться в API! Ошибка:\n" + err.Error())
		t.Log("Authentication to API failed! Error:\n" + err.Error())
		t.FailNow()
	}
	urls, err := GetUrlsFromApi(alertApiToken)
	if err != nil {
		SendFatalEmail("Не удалось получить адреса из API! Ошибка:\n" + err.Error())
		t.Log("GET /v1/alerts/url failed! Error:\n" + err.Error())
		t.FailNow()
	}

	// tests
	for _, url := range urls {
		t.Run(url.Url, func(t *testing.T) {
			// get запрос
			body, duration, testError := GetHttp(url.Url)
			if testError != nil {
				t.Log(testError)
				t.FailNow()
			}

			// проверка наличия title
			if !AssertBodyHasTitle(body) {
				t.Log("Ответ не содержит title! Ответ:\n" + body)
				t.Fail()
			}

			// отправка алертов
			// авторизация на бэке
			alertApiToken, err := AuthApi()
			if err != nil {
				SendFatalEmail("Не удалось авторизоваться в API! Ошибка:\n" + err.Error())
				t.Log("Authentication to API failed! Error:\n" + err.Error())
				t.FailNow()
			}
			// создание алерта
			var status string
			if testError != nil {
				status = "FAIL"
			} else {
				status = "PASS"
				testError = errors.New("")
			}
			err = CreateAlert(alertApiToken, url.Url, status, testError, int64(duration))
			if err != nil {
				SendFatalEmail("Alert не был создан! Ошибка:\n" + err.Error())
				t.Log("Alert not created! Error:\n" + err.Error())
				t.FailNow()
			}
		})
	}
}

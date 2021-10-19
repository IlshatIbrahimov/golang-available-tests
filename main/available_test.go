package main

import (
	"errors"
	"testing"
)

func TestAvailable(t *testing.T) {
	// setup: check google.com
	_, _, err := GetHttp("https://google.com/")
	if err != nil {
		SendFatalEmail("SetUp failed: google.com недоступен! Ошибка:\n" + err.Error())
		t.Log("SetUp failed: google.com unavailable! Ошибка:\n" + err.Error())
		t.FailNow()
	}

	// setup: get urls
	alertApiToken, err := AuthApi()
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

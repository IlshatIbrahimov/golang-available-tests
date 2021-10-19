package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// get запрос по url
func GetHttp(url string) (string, time.Duration, error) {
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
func AssertBodyHasTitle(body string) bool {
	return strings.Contains(body, "title")
}

// отправка алерта на почту в случае фатальных ошибок
func SendFatalEmail(messageText string) error {
	message := []byte("Subject: Ошибка в тестах доступности страниц! " + os.Getenv("HOST") + "\r\n" + messageText)

	auth := smtp.PlainAuth("", EmailFrom, EmailPassword, EmailSmtpHost)

	err := smtp.SendMail(EmailSmtpHost+":"+EmailSmtpPort, auth, EmailFrom, EmailTo, message)
	if err != nil {
		return err
	}
	return nil
}

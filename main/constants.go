package main

import (
	"os"
	"strings"
)

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type ApiUrl struct {
	Url string
}

var AlertApiUrl = os.Getenv("ALERT_API_URL")
var AlertApiUsername = os.Getenv("ALERT_API_USERNAME")
var AlertApiPassword = os.Getenv("ALERT_API_PASSWORD")

var EmailFrom = os.Getenv("ACCESSIBILITY_SMTP_EMAIL")
var EmailPassword = os.Getenv("ACCESSIBILITY_SMTP_PASSWORD")
var EmailSmtpHost = os.Getenv("ACCESSIBILITY_SMTP_HOST")
var EmailSmtpPort = os.Getenv("ACCESSIBILITY_SMTP_PORT")
var EmailTo = strings.Split(os.Getenv("FATAL_EMAIL_RECIPIENTS"), ",")

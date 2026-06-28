package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBPath        string
	EmailsCSVPath string
	TemplatePath  string
	SMTPHost      string
	SMTPPort      string
	SenderEmail   string
	WorkerCount   int
	SendDelay     time.Duration
}

func loadConfig() (Config, error) {
	workerCount, err := getEnvInt("MAILCHIMP_WORKER_COUNT", 5)
	if err != nil {
		return Config{}, err
	}

	delayMS, err := getEnvInt("MAILCHIMP_SEND_DELAY_MS", 50)
	if err != nil {
		return Config{}, err
	}

	return Config{
		DBPath:        getEnvString("MAILCHIMP_DB_PATH", "mailchimp.db"),
		EmailsCSVPath: getEnvString("MAILCHIMP_EMAILS_CSV", "./emails.csv"),
		TemplatePath:  getEnvString("MAILCHIMP_TEMPLATE_PATH", "email.tmpl"),
		SMTPHost:      getEnvString("MAILCHIMP_SMTP_HOST", "127.0.0.1"),
		SMTPPort:      getEnvString("MAILCHIMP_SMTP_PORT", "1025"),
		SenderEmail:   getEnvString("MAILCHIMP_SENDER_EMAIL", "shreyashukla20042005@gmail.com"),
		WorkerCount:   workerCount,
		SendDelay:     time.Duration(delayMS) * time.Millisecond,
	}, nil
}

func getEnvString(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid value for %s: %w", key, err)
	}

	return parsed, nil
}

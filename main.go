package main

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"
	"time"
)

type Recipient struct {
	Name  string
	Email string
}

type DLQ struct {
	Recipient   Recipient
	Error       string
	FailedAt    time.Time
	FailureType string // "template_error"
	Attempts    int
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	// Initialize database
	db, err := initDB(cfg.DBPath)
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}
	defer db.Close()

	recipientChannel := make(chan Recipient)
	dlqChannel := make(chan DLQ)

	go func() {
		loadRecipient(cfg.EmailsCSVPath, recipientChannel)
	}()

	var wg sync.WaitGroup

	for i := 1; i <= cfg.WorkerCount; i++ {
		wg.Add(1)
		go emailWorker(i, recipientChannel, &wg, dlqChannel, db, cfg)

	}

	go func() {
		for failed := range dlqChannel {
			_ = failed
		}
	}()

	wg.Wait()
	close(dlqChannel)

	//Dlq summary
	total, sent, failed, err := getStats(db)
	if err == nil {
		fmt.Printf("\n=== Email Campaign Summary ===\n")
		fmt.Printf("Total: %d | Sent: %d | Failed: %d\n", total, sent, failed)
	}

	// Print failed emails from DB
	if failed > 0 {
		failures, _ := getFailedEmails(db)
		fmt.Printf("\n=== Failed Emails ===\n")
		for _, f := range failures {
			fmt.Printf("%s: %s (%s)\n", f.Recipient.Email, f.Error, f.FailureType)
		}
	}
}

func executeTemplate(r Recipient, templatePath string) (string, error) {
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, r)
	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}

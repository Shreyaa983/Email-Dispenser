package main

import (
	"database/sql"
	"fmt"
	"net/smtp"
	"sync"
	"time"
)

func emailWorker(id int, ch chan Recipient, wg *sync.WaitGroup, dlq chan DLQ, db *sql.DB, cfg Config) {
	defer wg.Done()
	for recipient := range ch {
		msg, err := executeTemplate(recipient, cfg.TemplatePath)
		if err != nil {
			fmt.Printf("Woker: %d error parsing template for %s", id, recipient.Email)
			failure := DLQ{
				Recipient:   recipient,
				Error:       err.Error(),
				FailedAt:    time.Now(),
				FailureType: "template_error",
				Attempts:    1,
			}
			dlq <- failure
			logFailedEmail(db, failure)

			continue
		}

		fmt.Printf("worker %d: Sending email to %s \n", id, recipient.Email)
		err = smtp.SendMail(cfg.SMTPHost+":"+cfg.SMTPPort, nil, cfg.SenderEmail, []string{recipient.Email}, []byte(msg))
		if err != nil {
			failure := DLQ{
				Recipient:   recipient,
				Error:       err.Error(),
				FailedAt:    time.Now(),
				FailureType: "smtp_error",
				Attempts:    1,
			}
			dlq <- failure
			logFailedEmail(db, failure)

			continue
		}
		time.Sleep(cfg.SendDelay)

		if err == nil {
			logSuccessfulEmail(db, recipient)
			fmt.Printf("worker %d: Sent email to %s \n", id, recipient.Email)
		}
	}
}

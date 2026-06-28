package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func initDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Create email_logs table
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS email_logs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT NOT NULL,
        name TEXT,
        status TEXT DEFAULT 'pending',
        failure_type TEXT,
        error_msg TEXT,
        attempts INTEGER DEFAULT 0,
        sent_at DATETIME,
        failed_at DATETIME,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    `

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	fmt.Println("Database initialized successfully")
	return db, nil
}

// Log successful email
func logSuccessfulEmail(db *sql.DB, recipient Recipient) error {
	query := `
    INSERT INTO email_logs (email, name, status, sent_at)
    VALUES (?, ?, 'sent', CURRENT_TIMESTAMP)
    `
	_, err := db.Exec(query, recipient.Email, recipient.Name)
	return err
}

// Log failed email
func logFailedEmail(db *sql.DB, dlq DLQ) error {
	query := `
    INSERT INTO email_logs (email, name, status, failure_type, error_msg, attempts, failed_at)
    VALUES (?, ?, 'failed', ?, ?, ?, CURRENT_TIMESTAMP)
    `
	_, err := db.Exec(query, dlq.Recipient.Email, dlq.Recipient.Name,
		dlq.FailureType, dlq.Error, dlq.Attempts)
	return err
}

// Get all failed emails
func getFailedEmails(db *sql.DB) ([]DLQ, error) {
	query := `SELECT email, name, failure_type, error_msg, attempts FROM email_logs WHERE status = 'failed'`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var failures []DLQ
	for rows.Next() {
		var dlq DLQ
		err := rows.Scan(&dlq.Recipient.Email, &dlq.Recipient.Name,
			&dlq.FailureType, &dlq.Error, &dlq.Attempts)
		if err != nil {
			continue
		}
		failures = append(failures, dlq)
	}
	return failures, nil
}

// Get statistics
func getStats(db *sql.DB) (total, sent, failed int, err error) {
	err = db.QueryRow("SELECT COUNT(*) FROM email_logs").Scan(&total)
	if err != nil {
		return
	}
	err = db.QueryRow("SELECT COUNT(*) FROM email_logs WHERE status = 'sent'").Scan(&sent)
	if err != nil {
		return
	}
	err = db.QueryRow("SELECT COUNT(*) FROM email_logs WHERE status = 'failed'").Scan(&failed)
	return
}

# Mailchimp Email Campaign Sender

A small Go application that reads recipients from a CSV file, renders an HTML/text email template for each recipient, sends the messages through SMTP, and logs delivery results to a local SQLite database.

## What it does

- Loads recipients from `emails.csv`
- Renders `email.tmpl` for each recipient using Go templates
- Sends email through an SMTP server
- Stores send/failure history in SQLite
- Prints a final campaign summary with totals and failed deliveries

## Project layout

- `main.go` - application entry point and worker orchestration
- `config.go` - environment-driven configuration loading
- `consumer.go` - CSV reader that streams recipients into a channel
- `producer.go` - worker logic that renders templates and sends email
- `database.go` - SQLite initialization and logging helpers
- `email.tmpl` - email template used for each recipient
- `emails.csv` - sample recipient list
- `info.md` - extra notes for the project

## How it works

1. The app reads runtime settings from environment variables.
2. A SQLite database is created if it does not already exist.
3. Recipients are loaded from the CSV file and sent to workers through a channel.
4. Each worker renders the template with recipient data.
5. The email is sent with `net/smtp`.
6. Successes and failures are written to the `email_logs` table.
7. When all workers finish, the app prints delivery statistics and any failed emails.

## Requirements

- Go 1.25 or newer
- An SMTP server for local testing or real delivery
- SQLite is embedded through the Go dependency, so no separate database server is needed

## Recommended local SMTP setup

For local testing, Mailpit is a convenient SMTP sink that accepts messages and exposes a web UI.

```bash
docker run -d --restart unless-stopped --name mailpit -p 8025:8025 -p 1025:1025 axllent/mailpit
```

If `localhost` resolves to IPv6 on your machine, use `127.0.0.1` for the SMTP host.

Mailpit UI: http://localhost:8025

## Configuration

The app is configured entirely through environment variables. Every value has a default.

| Variable | Default | Purpose |
| --- | --- | --- |
| `MAILCHIMP_DB_PATH` | `mailchimp.db` | Path to the SQLite database file |
| `MAILCHIMP_EMAILS_CSV` | `./emails.csv` | Path to the recipient CSV file |
| `MAILCHIMP_TEMPLATE_PATH` | `email.tmpl` | Path to the email template file |
| `MAILCHIMP_SMTP_HOST` | `127.0.0.1` | SMTP server host |
| `MAILCHIMP_SMTP_PORT` | `1025` | SMTP server port |
| `MAILCHIMP_SENDER_EMAIL` | `shreyashukla20042005@gmail.com` | Sender address used in SMTP delivery |
| `MAILCHIMP_WORKER_COUNT` | `5` | Number of concurrent email workers |
| `MAILCHIMP_SEND_DELAY_MS` | `50` | Delay between sends per worker in milliseconds |

## Example configuration

### PowerShell

```powershell
$env:MAILCHIMP_SMTP_HOST = "127.0.0.1"
$env:MAILCHIMP_SMTP_PORT = "1025"
$env:MAILCHIMP_SENDER_EMAIL = "sender@example.com"
$env:MAILCHIMP_WORKER_COUNT = "5"
$env:MAILCHIMP_SEND_DELAY_MS = "50"
```

### Bash

```bash
export MAILCHIMP_SMTP_HOST=127.0.0.1
export MAILCHIMP_SMTP_PORT=1025
export MAILCHIMP_SENDER_EMAIL=sender@example.com
export MAILCHIMP_WORKER_COUNT=5
export MAILCHIMP_SEND_DELAY_MS=50
```

## Running the app

From the project root:

```bash
go run .
```

If you want to point at a different CSV, template, or database file, set the environment variables before running.

## Input files

### CSV format

The CSV file must include a header row with at least these columns:

- `name`
- `email`

Example:

```csv
name,email
Alice Johnson,alice.johnson@example.com
Bob Smith,bob.smith@example.com
```

### Template format

The template is rendered with a `Recipient` object that exposes `Name` and `Email`.

Example:

```text
Subject: Hello, {{.Name}}

Hi {{.Name}}

Thanks,
Shreya
```

## Database schema

The application creates an `email_logs` table automatically.

Columns:

- `email` - recipient email address
- `name` - recipient name
- `status` - `sent` or `failed`
- `failure_type` - failure category such as `template_error` or `smtp_error`
- `error_msg` - error message from the send attempt
- `attempts` - retry count recorded by the app
- `sent_at` - timestamp for successful sends
- `failed_at` - timestamp for failed sends
- `created_at` - row creation timestamp

## Output

At the end of a run, the app prints:

- Total rows logged
- Count of sent emails
- Count of failed emails
- A list of failed recipients when failures occurred

## Troubleshooting

- If nothing is sent, confirm the SMTP host and port are reachable.
- If template parsing fails, check that `email.tmpl` exists and contains valid Go template syntax.
- If the CSV fails to load, make sure the first row is a header and the data rows have at least two columns.
- If the database file cannot be created, verify that the working directory is writable.
- If you are testing with Mailpit, open the UI at http://localhost:8025 to inspect received messages.

## Notes

- The sender address is configurable, but the current default in code is a sample personal address. Override it for real use.
- Failed sends are logged immediately and included in the final summary.
- The worker count and per-send delay make it easy to tune throughput while testing.

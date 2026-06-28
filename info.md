docker run -d --restart unless-stopped --name mailpit -p 8025:8025 -p 1025:1025 axllent/mailpit

Set MAILCHIMP_SMTP_HOST=127.0.0.1 if localhost resolves to IPv6 on your machine.
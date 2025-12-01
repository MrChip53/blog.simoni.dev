package mail

import "net/smtp"

type MailClient struct {
	smtpHost string
	smtpPort string
	from     string
	auth     smtp.Auth
}

func NewMailClient(smtpHost, smtpPort, username, password string) *MailClient {
	auth := smtp.PlainAuth("", username, password, smtpHost)
	return &MailClient{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		from:     username,
		auth:     auth,
	}
}

func (mc *MailClient) SendMail(to []string, subject, body string) error {
	msg := []byte("Subject: " + subject + "\n\n" + body)
	return smtp.SendMail(mc.smtpHost+":"+mc.smtpPort, mc.auth, mc.from, to, msg)
}

package cmd

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)

// Send via SMTP
func smtpSend(from string, to []string, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}

	defer c.Close()

	if config.STARTTLS {
		conf := &tls.Config{ServerName: config.SMTPHost, MinVersion: tls.VersionTLS12}

		conf.InsecureSkipVerify = config.AllowInsecure

		if err = c.StartTLS(conf); err != nil {
			return err
		}
	}

	var a smtp.Auth

	if config.Auth == "plain" {
		a = smtp.PlainAuth("", config.Username, config.Password, config.SMTPHost)
	}

	if config.Auth == "login" {
		a = loginAuth(config.Username, config.Password)
	}

	if config.Auth == "cram-md5" {
		a = smtp.CRAMMD5Auth(config.Username, config.Password)
	}

	if a != nil {
		if err = c.Auth(a); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	if _, err := w.Write(msg); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return c.Quit()

}

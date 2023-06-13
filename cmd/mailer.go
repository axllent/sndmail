package cmd

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// SMTP wrapper will send and optionally log the transaction
func smtpWrapper(from string, to []string, msg []byte) error {
	code, response, err := smtpSend(from, to, msg)

	if config.LogFile != "" {
		now := time.Now().Local()
		ts := now.Format("Jan 02 15:04:05")
		tls := "off"
		if config.STARTTLS {
			tls = "on"
		}
		recipients := strings.Join(to, ",")
		logMsg := fmt.Sprintf("%s host=%s tls=%s from=%s, recipients=%s mailsize=%d, smtpstatus=%d smtpmsg='%s'\n", ts, config.SMTPHost, tls, from, recipients, len(msg), code, response)

		file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666) // #nosec
		// silently fail if the file cannot be opened
		if err == nil {
			defer file.Close()
			_, _ = file.Write([]byte(logMsg))
		}
	}

	return err
}

// Send via SMTP
func smtpSend(from string, to []string, msg []byte) (int, string, error) {
	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	c, err := smtp.Dial(addr)
	if err != nil {
		return 0, "", err
	}

	defer c.Close()

	if config.STARTTLS {
		conf := &tls.Config{ServerName: config.SMTPHost, MinVersion: tls.VersionTLS12}

		conf.InsecureSkipVerify = config.AllowInsecure

		if err = c.StartTLS(conf); err != nil {
			return 0, "", err
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
			return 0, "", err
		}
	}
	if err = c.Mail(from); err != nil {
		return 0, "", err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return 0, "", err
		}
	}

	w, err := c.Data()
	if err != nil {
		return 0, "", err
	}

	if _, err := w.Write(msg); err != nil {
		return 0, "", err
	}

	defer c.Quit()

	code, message, err := closeData(c)

	return code, message, err
}

// CloseData wil ensure the SMTP server response is returned
// @see https://stackoverflow.com/a/70925659
func closeData(client *smtp.Client) (int, string, error) {
	d := &dataCloser{
		c:           client,
		WriteCloser: client.Text.DotWriter(),
	}

	return d.Close()
}

type dataCloser struct {
	c *smtp.Client
	io.WriteCloser
}

func (d *dataCloser) Close() (int, string, error) {
	d.WriteCloser.Close() // #nosec make sure WriterCloser gets closed
	code, message, err := d.c.Text.ReadResponse(250)

	return code, message, err
}

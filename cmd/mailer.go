package cmd

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lithammer/shortuuid"
)

// SMTP wrapper will send and optionally log the transaction
func smtpWrapper(from string, to []string, message []byte) error {
	msg, err := injectMissingHeaders(message, from)
	if err != nil {
		return err
	}

	code, response, err := smtpSend(from, to, msg)

	if config.LogFile != "" {
		now := time.Now().Local()
		ts := now.Format("Jan 02 15:04:05")
		tls := "off"
		if config.STARTTLS {
			tls = "on"
		}
		if err != nil {
			code, response = smtpErrParser(err.Error())
		}

		recipients := strings.Join(to, ",")
		logMsg := fmt.Sprintf("%s host=%s tls=%s from=%s, recipients=%s mailsize=%d, smtpstatus=%d smtpmsg='%s'\n", ts, config.SMTPHost, tls, from, recipients, len(msg), code, response)

		file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666) // #nosec
		// silently fail if the log file cannot be opened
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

	code, message, err := dataWithResponse(c, msg)
	if err != nil {
		return 0, "", err
	}

	return code, message, c.Quit()
}

// Go's net/smtp Data() only returns an error message. This custom function
// also returns the server's response message so it can be logged.
//
// @see https://github.com/axllent/sndmail/issues/10#issuecomment-2214807859
func dataWithResponse(c *smtp.Client, msg []byte) (int, string, error) {
	id, err := c.Text.Cmd("DATA")
	if err != nil {
		return 0, "", err
	}

	c.Text.StartResponse(id)

	code, message, err := c.Text.ReadResponse(354)
	if err != nil {
		return code, message, err
	}

	c.Text.EndResponse(id)

	w := c.Text.DotWriter()

	if _, err := w.Write(msg); err != nil {
		return 0, "", err
	}

	if err := w.Close(); err != nil {
		return 0, "", err
	}

	return c.Text.ReadResponse(250)
}

// Inject Message-Id and Date if missing. The From address is also
// optionally injected if missing.
func injectMissingHeaders(body []byte, from string) ([]byte, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(body))
	if err != nil {
		// create blank message so lookups don't fail
		msg = &mail.Message{}

		// inject a new blank line before body
		body = append([]byte(fmt.Sprintf("\r\n")), body...)
	}

	// add message ID if missing
	if msg.Header.Get("Message-Id") == "" {
		messageID := shortuuid.New() + "@sndmail"
		body = append([]byte("Message-Id: <"+messageID+">\r\n"), body...)
	}

	// add date if missing
	if msg.Header.Get("Date") == "" {
		now := time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700 (MST)")
		body = append([]byte("Date: "+now+"\r\n"), body...)
	}

	// set From header is missing
	if msg.Header.Get("From") == "" {
		body = append([]byte("From: <"+from+">\r\n"), body...)
	}

	return body, nil
}

// error parser for SMTP response messages
func smtpErrParser(s string) (int, string) {
	var re = regexp.MustCompile(`(\d\d\d) (.*)`)
	if re.MatchString(s) {
		matches := re.FindAllStringSubmatch(s, -1)
		for _, m := range matches {
			i, _ := strconv.Atoi(m[1])

			return i, m[2]
		}
	}

	return 0, s
}

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

var (
	recipientsFromMessage bool

	fromAddress string
)

// Standard sendmail via the CLI
func sendmail() {
	body, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading stdin")
		os.Exit(11)
	}

	msg, err := mail.ReadMessage(bytes.NewReader(body))
	if err != nil {
		// potentially forgot to add headers, inject a new blank line above
		body = append([]byte(fmt.Sprintf("\n")), body...)
		msg, err = mail.ReadMessage(bytes.NewReader(body))
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("error parsing message body: %s", err))
			os.Exit(11)
		}
	}

	if fromAddress == "" {
		fromAddress = msg.Header.Get("From")
		if fromAddress == "" {
			fromAddress = config.From
		}
	}

	recipients := pflag.Args()

	if recipientsFromMessage {
		// get all recipients in To, Cc and Bcc
		if to, err := msg.Header.AddressList("To"); err == nil {
			for _, a := range to {
				recipients = append(recipients, a.Address)
			}
		}
		if cc, err := msg.Header.AddressList("Cc"); err == nil {
			for _, a := range cc {
				recipients = append(recipients, a.Address)
			}
		}
		if bcc, err := msg.Header.AddressList("Bcc"); err == nil {
			for _, a := range bcc {
				recipients = append(recipients, a.Address)
			}
		}
	}

	if len(recipients) == 0 {
		fmt.Fprintln(os.Stderr, "no recipients found")
		os.Exit(11)
	}

	// get unique recipients, also sets aliases
	recipients = uniqueRecipients(recipients)

	if err := smtpSend(fromAddress, recipients, body); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(11)
	}
}

// Returns unique recipients in a slice. If an address matches an alias, it is updated.
func uniqueRecipients(slice []string) []string {
	u := make(map[string]bool, len(slice))
	for _, v := range slice {
		val := strings.ToLower(v)
		// check aliases
		alias, ok := config.Aliases[val]
		if ok {
			val = alias
		}
		u[val] = true
	}

	n := make([]string, 0, len(u))
	for k := range u {
		n = append(n, k)
	}

	return n
}

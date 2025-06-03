package cmd

import (
	"bufio"
	"fmt"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"time"
)

// SMTP transaction struct
type smtpTransaction struct {
	From       string
	To         []string
	DataWrite  bool
	Message    []string
	LastActive time.Time
}

var transaction = smtpTransaction{}

// Cancel the "connection" after 1 minute of inactivity
func exitAfterTimeout() {
	interval := time.Second

	tk := time.NewTicker(interval)

	for range tk.C {
		if transaction.LastActive.Add(time.Minute).Before(time.Now()) {
			writeSMTP(421, fmt.Sprintf("4.4.2 %s Error: timeout exceeded", config.Hostname))
			os.Exit(11)
		}
	}
}

// Runs a SMTP session in the foreground
func stdinSMTPD() {
	transaction.LastActive = time.Now() // set

	go exitAfterTimeout()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("220 sndmail-smtpd ready")
	errors := []string{}
	for {
		// Scans a line from Stdin(Console)
		scanner.Scan()
		// Holds the string that scanned
		text := scanner.Text()

		transaction.LastActive = time.Now() // update

		if transaction.DataWrite {
			if text == "." {
				if err := smtpWrapper(transaction.From, transaction.To, []byte(strings.Join(transaction.Message, "\r\n"))); err != nil {
					writeSMTP(503, err.Error())
				} else {
					writeSMTP(250, "2.0.0 Ok: message queued")
				}
				// reset
				transaction = smtpTransaction{LastActive: time.Now()}
				continue
			}

			// SMTP standard input will append an extra dot to any line which starts with a dot, so remove it again
			if strings.HasPrefix(text, "..") {
				text = text[1:]
			}

			transaction.Message = append(transaction.Message, text)
			continue
		}

		verb, args := readSMTP(text)

		switch verb {

		case "HELO":
			if args == "" {
				writeSMTP(501, "Syntax: HELO hostname")
				continue
			}
			// https://tools.ietf.org/html/rfc2821#section-4.1.1.1
			writeSMTP(250, config.Hostname)

		case "EHLO":
			if args == "" {
				writeSMTP(501, "Syntax: EHLO hostname")
				continue
			}
			// see: https://tools.ietf.org/html/rfc2821#section-4.1.4
			errors = []string{}

			writeEHLO(250, config.Hostname)
			// writeEHLO(250, fmt.Sprintf("SIZE %v", 20240000))

		case "MAIL":
			// The MAIL command starts off a new mail transaction
			// see: https://tools.ietf.org/html/rfc2821#section-4.1.1.2
			// This doesn't implement the RFC4594 addition of an AUTH param to the MAIL command
			// see: http://tools.ietf.org/html/rfc4954#section-3 for details
			if from, err := getAddressArg("FROM", args); err == nil {
				transaction.From = from.Address
				writeSMTP(250, "2.1.0 Ok")
			} else {
				writeSMTP(503, err.Error())
			}

		case "RCPT":
			// https://tools.ietf.org/html/rfc2821#section-4.1.1.3
			if transaction.From == "" {
				writeSMTP(503, "5.5.1 Error: need MAIL command")
				continue
			}
			// TODO: bubble these up to the message,
			if to, err := getAddressArg("TO", args); err == nil {
				transaction.To = append(transaction.To, to.Address)
				writeSMTP(250, "2.1.5 Ok")
			} else {
				writeSMTP(503, err.Error())
			}

		case "DATA":
			if len(transaction.To) == 0 {
				writeSMTP(503, "5.5.1 Error: need RCPT command")
				continue
			}
			writeSMTP(354, "End data with <CR><LF>.<CR><LF>")
			transaction.DataWrite = true

		case "RSET":
			// Reset the connection
			// see: https://tools.ietf.org/html/rfc2821#section-4.1.1.5
			transaction.DataWrite = false
			transaction.From = ""
			transaction.To = []string{}
			writeSMTP(250, "2.1.5 Ok")

		case "VRFY":
			// Since this is a commonly abused SPAM aid, it's better to just
			// default to 252 (apparent validity / could not verify).
			// see: https://tools.ietf.org/html/rfc2821#section-4.1.1.6
			writeSMTP(252, "But it was worth a shot, right?")

		// see: https://tools.ietf.org/html/rfc2821#section-4.1.1.7
		case "EXPN":
			writeSMTP(252, "Maybe, maybe not")

		case "NOOP":
			// NOOP doesn't do anything
			// see: https://tools.ietf.org/html/rfc2821#section-4.1.1.9
			writeSMTP(250, "2.1.5 Ok")

		case "QUIT":
			writeSMTP(221, "2.0.0 Good bye")
			os.Exit(0)

		default:
			writeSMTP(500, "Syntax error, command unrecognised")
			errors = append(errors, fmt.Sprintf("bad input: %v %v", verb, args))
			if len(errors) > 3 {
				writeSMTP(500, "Too many unrecognized commands")
				os.Exit(1)
			}

		}
	}
}

func writeSMTP(code int, msg string) {
	fmt.Printf("%d %s\n", code, msg)
}

func writeEHLO(code int, msg string) {
	fmt.Printf("%d-%s\n", code, msg)
}

func readSMTP(line string) (string, string) {
	var args string
	command := strings.SplitN(line, " ", 2)

	verb := strings.ToUpper(command[0])
	if len(command) > 1 {
		args = command[1]
	}

	return verb, args
}

var pathRegex = regexp.MustCompile(`<([^@>]+@[^@>]+)>`)

// GetAddressArg extracts the address value from a supplied SMTP argument
// for handling MAIL FROM:address@example.com and RCPT TO:address@example.com
func getAddressArg(argName string, args string) (*mail.Address, error) {
	argSplit := strings.SplitN(args, ":", 2)
	if len(argSplit) == 2 && strings.ToUpper(argSplit[0]) == argName {

		path := pathRegex.FindString(argSplit[1])
		if path == "" {
			return nil, fmt.Errorf("couldn't find valid FROM path in %v", argSplit[1])
		}

		return mail.ParseAddress(path)
	}

	return nil, fmt.Errorf("Bad arguments")
}

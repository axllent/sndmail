// Package cmd is the application logic
package cmd

import (
	"fmt"
	"os"
)

var (
	smtpViaInput bool

	// Version of the build
	Version = "dev"
)

// Exec is the main application wrapper
func Exec() {
	// init config
	config = Config{}

	initArgs()

	err := initConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if smtpViaInput {
		stdinSMTPD()
	} else {
		sendmail()
	}
}

// func main() {
// 	Cmd()
// }

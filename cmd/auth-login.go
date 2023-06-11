package cmd

import (
	"errors"
	"net/smtp"
)

// Custom implementation of LOGIN SMTP authentication
// @see https://gist.github.com/andelf/5118732
type loginAuthS struct {
	username, password string
}

// LoginAuth authentication
func loginAuth(username, password string) smtp.Auth {
	return &loginAuthS{username, password}
}

func (a *loginAuthS) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuthS) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unknown fromServer")
		}
	}

	return nil, nil
}

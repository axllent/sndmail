package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	config Config
)

// Config struct
type Config struct {
	Hostname      string
	From          string
	SMTPHost      string
	SMTPPort      int
	STARTTLS      bool
	AllowInsecure bool
	Auth          string
	Username      string
	Password      string
	Aliases       map[string]string
}

// Generate the config
func initConfig() error {
	// load from sndmail.conf if exists
	v := loadFromConfigFile()

	config.SMTPHost = v["smtp-host"]

	i, err := strconv.Atoi(v["smtp-port"])
	if err != nil {
		return err
	}

	config.SMTPPort = i
	config.STARTTLS = strToBool(v["starttls"])
	config.AllowInsecure = strToBool(v["allow-insecure"])
	config.Auth = strings.ToLower(v["auth-type"])
	config.Username = v["auth-user"]
	config.Password = v["auth-pass"]
	config.Aliases = loadAliasMap()
	if config.From == "" {
		config.From = v["from"]
	}

	return nil
}

// Load the configuration file (if found)
func loadFromConfigFile() map[string]string {
	config.Hostname = "localhost"
	hostname, err := os.Hostname()
	if err == nil {
		config.Hostname = hostname
	}

	userName := "unknown"
	curUser, err := user.Current()
	if err == nil {
		userName = curUser.Username
	}

	from := userName + "@" + config.Hostname

	// default values
	c := map[string]string{
		"smtp-host":      "localhost",
		"smtp-port":      "25",
		"starttls":       "false",
		"allow-insecure": "false",
		"auth-type":      "none",
		"auth-user":      "",
		"auth-pass":      "",
		"from":           from,
	}

	confFile, err := findConfig("sndmail.conf")
	if err != nil {
		return c
	}

	file, err := os.Open(confFile) // #nosec
	if err != nil {
		return c
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				_, validKey := c[key]
				if !validKey {
					continue
				}

				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				c[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return c
	}

	return c
}

// Load an aliases map if aliases file is found
func loadAliasMap() map[string]string {
	c := map[string]string{}

	aliasFile, err := findConfig("aliases")
	if err != nil {
		return c
	}

	file, err := os.Open(aliasFile) // #nosec
	if err != nil {
		return c
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, ":"); equal >= 0 {
			if key := strings.ToLower(strings.TrimSpace(line[:equal])); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.ToLower(strings.TrimSpace(line[equal+1:]))
				}
				c[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return c
	}

	return c
}

// Convert a string to a boolean value
func strToBool(v string) bool {
	v = strings.ToLower(v)
	return v == "true" || v == "yes" || v == "y" || v == "1"
}

// FindConfig will attempt to locate a file within default paths
func findConfig(name string) (string, error) {
	paths := []string{".", "/etc", "/usr/local/etc"}

	for _, p := range paths {
		test := filepath.Join(p, name)
		if isFile(test) {
			return test, nil
		}
	}

	return "", fmt.Errorf("config %s not found", name)
}

// Return if a path is a file
func isFile(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) || !info.Mode().IsRegular() {
		return false
	}

	return true
}

package cmd

import (
	"fmt"
	"os"

	"github.com/axllent/ghru/v2"
	"github.com/spf13/pflag"
)

var (
	displayVersion bool
	doUpdate       bool

	ghruConf = ghru.Config{
		Repo:           "axllent/sndmail",
		ArchiveName:    "sndmail-{{.OS}}-{{.Arch}}",
		BinaryName:     "sndmail",
		CurrentVersion: Version,
	}
)

func initArgs() {
	// Artificially generate the help screen to simplify the formatting.
	pflag.Usage = func() {
		fmt.Printf("sndmail %s: sendmail emulator (https://github.com/axllent/sndmail)\n\n", Version)
		fmt.Printf("Usage: %s [flags] [recipients] < message\n\n", os.Args[0])
		fmt.Println(`Options:
  -B     	Ignored
  -bm    	Read mail from standard input (default)
  -bs    	Handle SMTP commands on standard input
  -C     	Ignored
  -d     	Ignored
  -F     	Ignored
  -f     	Set the envelope sender address
  -i     	Ignored
  -L     	Ignored
  -m     	Ignored
  -N     	Ignored
  -n     	Ignored
  -o     	Ignored
  -em    	Ignored
  -ep    	Ignored
  -eq    	Ignored
  -p     	Ignored
  -q     	Ignored
  -R     	Ignored
  -r     	Ignored
  -t     	Read message for recipients
  -U     	Ignored
  -V     	Ignored
  -v     	Ignored
  -X     	Ignored
  --version	Display version and update information
  --update	Update sndmail to the latest version`)
	}

	// Given limitation in Go's default flag package (cannot handle single dash with
	// multiple characters), we use pflag in order to artificially handle `-bs`.
	pflag.BoolP("long-B", "B", false, "Ignored")
	// handles -bm & -bs
	pflag.StringP("long-b", "b", "", "Handle SMTP commands on standard input")
	pflag.BoolP("long-C", "C", false, "Ignored")
	pflag.BoolP("long-d", "d", false, "Ignored")
	pflag.StringP("long-from", "F", "", "Ignored")
	pflag.StringVarP(&fromAddress, "from", "f", "", "Set the envelope sender address")
	pflag.BoolP("long-i", "i", false, "Ignored")
	pflag.BoolP("long-L", "L", false, "Ignored")
	pflag.BoolP("long-m", "m", false, "Ignored")
	pflag.BoolP("long-N", "N", false, "Ignored")
	pflag.BoolP("long-n", "n", false, "Ignored")
	pflag.BoolP("long-o", "o", false, "Ignored")
	pflag.StringP("long-e", "e", "", "Ignored")
	pflag.BoolP("long-p", "p", false, "Ignored")
	pflag.BoolP("long-q", "q", false, "Ignored")
	pflag.BoolP("long-R", "R", false, "Ignored")
	pflag.BoolP("long-r", "r", false, "Ignored")
	pflag.BoolVarP(&recipientsFromMessage, "long-t", "t", false, "Read message for recipients")
	pflag.BoolP("long-U", "U", false, "Ignored")
	pflag.BoolP("long-V", "V", false, "Ignored")
	pflag.BoolP("long-v", "v", false, "Ignored")
	pflag.BoolP("long-X", "X", false, "Ignored")
	pflag.BoolP("help", "h", false, "")
	pflag.BoolVar(&displayVersion, "version", false, "Display version information")
	pflag.BoolVar(&doUpdate, "update", false, "Update sndmail to the latest version")

	pflag.Parse()

	if showHelp, _ := pflag.CommandLine.GetBool("help"); showHelp {
		pflag.Usage()
		os.Exit(0)
	}

	if displayVersion {
		fmt.Printf("Version: %s\n", Version)

		release, err := ghruConf.Latest()
		if err != nil {
			fmt.Printf("Error checking for latest release: %s\n", err)
			os.Exit(1)
		}

		// The latest version is the same version
		if release.Tag == Version {
			os.Exit(0)
		}

		// A newer release is available
		fmt.Printf(
			"Update available: %s\nRun `%s --update` to update (requires read/write access to install directory).\n",
			release.Tag,
			os.Args[0],
		)
		os.Exit(0)
	}

	if doUpdate {
		// Update the application
		rel, err := ghruConf.SelfUpdate()
		if err != nil {
			fmt.Printf("Error updating: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Updated %s to version %s\n", os.Args[0], rel.Tag)
		os.Exit(0)
	}

	v, err := checkParam("long-b", "b", []string{"s", "m"})
	if err == nil {
		smtpViaInput = v == "s"
	}

	_, _ = checkParam("long-e", "e", []string{"m", "p", "q"})
}

// Simple function to limit the short flags to valid options
func checkParam(long, short string, options []string) (string, error) {
	if v, err := pflag.CommandLine.GetString(long); err == nil {
		if v != "" {
			for _, o := range options {
				if v == o {
					return o, nil
				}
			}

			errorMsg := fmt.Sprintf("unknown shorthand flag: '-%s%s", short, v)
			fmt.Println(errorMsg)
			pflag.Usage()
			fmt.Println(errorMsg)
			os.Exit(1)
		}
	}

	return "", fmt.Errorf("%s not set", short)
}

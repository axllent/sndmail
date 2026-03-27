# Sndmail - a sendmail emulator

![Tests status](https://github.com/axllent/sndmail/actions/workflows/tests.yml/badge.svg)
![Build status](https://github.com/axllent/sndmail/actions/workflows/release-build.yml/badge.svg)
![CodeQL](https://github.com/axllent/sndmail/actions/workflows/codeql-analysis.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/axllent/sndmail)](https://goreportcard.com/report/github.com/axllent/sndmail)

Sndmail is a multi-platform sendmail emulator and drop-in replacement for \*nix-like platforms.

It was created primarily for use in Docker containers. While there are many different sendmail emulators available, most lack `sendmail -bs` functionality (running SMTP on standard input) which is now the default with Symfony Mailer.

## Features

- Static drop-in replacement for sendmail
- Configurable SMTP relay server, STARTTLS with PLAIN, LOGIN and CRAM-MD5 support
- SMTP on standard input (`sendmail -bs`)
- Automatically generates missing `Message-ID`, `From` and `Date` headers
- Includes the ability to check for latest version and self-update functionality

## Installation

- Static binaries can be found on the [releases](https://github.com/axllent/sndmail/releases/latest)
- Copy or symlink the `sndmail` executable from `/usr/sbin/sendmail`
- Copy the `sndmail.conf.example` to `/etc/sndmail.conf`, making any necessary edits to adjust to your SMTP relay server

### Install via a bash script (Linux & macOS)

Linux & macOS users can install it directly by using:

```bash
sudo sh < <(curl -sL https://raw.githubusercontent.com/axllent/sndmail/develop/install.sh)
```

**Note** that this will install sndmail into `/usr/local/bin/sndmail` and symlink it to `/usr/sbin/sendmail`.
If you want to install it without the symlink, you can use the `NO_SENDMAIL` environment variable:

```bash
sudo NO_SENDMAIL=true sh < <(curl -sL https://raw.githubusercontent.com/axllent/sndmail/develop/install.sh)
```

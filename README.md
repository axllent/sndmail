# Sndmail - a sendmail emulator

![Build status](https://github.com/axllent/sndmail/actions/workflows/release-build.yml/badge.svg)
![CodeQL](https://github.com/axllent/sndmail/actions/workflows/codeql-analysis.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/axllent/sndmail)](https://goreportcard.com/report/github.com/axllent/sndmail)

Sndmail is a multi-platform sendmail emulator and drop-in replacement for *nix-like platforms.

It was created primarily for use in Docker containers. Whilst there are many different sendmail emulators available, most lack working `sendmail -bs` functionality (running SMTP on standard input) which is now the default with Symfony mail.


## Features

- Static drop-in replacement for sendmail
- Configurable SMTP relay server, STARTTLS with PLAIN, LOGIN and CRAM-MD5 support
- SMTP on standard input (`sendmail -bs`)
- Auto-generates (if missing from input) `Message-Id`, `From` & `Date` headers


## Installation

- Static binaries can be found on the [releases](https://github.com/axllent/sndmail/releases/latest)
- Copy or symlink the `sndmail` executable from `/usr/sbin/sendmail`
- Copy the `sndmail.conf.example` to `/etc/sndmail.conf` making any necessary edits to adjust to your SMTP relay server


### Install via bash script (Linux & Mac)

**Warning**: This will delete any existing /usr/sbin/sendmail!

Linux & Mac users can install it directly via:

```bash
sudo bash < <(curl -sL https://raw.githubusercontent.com/axllent/sndmail/develop/install.sh)
```

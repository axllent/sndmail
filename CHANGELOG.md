# Changelog

Notable changes to sndmail will be documented in this file.

## [v1.1.1]

### Chore
- Bump actions/stale from 10.1.1 to 10.2.0
- Update Go dependencies


## [v1.1.0]

### Chore
- Refactor install script, support installation directory & support GitHub API token authentication
- Update Go version and dependencies
- Add connection timeout for SMTP connections

### Fix
- Move HELO/EHLO hostname setting to the correct position in smtpSend function
- Handle stdin closure with appropriate error message in stdinSMTPD function
- Update isFile function to handle errors more accurately
- Prevent adding empty recipients to unique recipients map

### Test
- Add tests with Mailpit
- Add unit tests for uniqueRecipients and injectMissingHeaders functions


## [v1.0.2]

### Chore
- Bump actions/stale from 10.0.0 to 10.1.1
- Bump wangyoucao577/go-release-action from 1.53 to 1.55
- Bump golangci/golangci-lint-action from 8 to 9
- Bump actions/checkout from 5 to 6
- Bump github/codeql-action from 3 to 4
- Update Go dependencies


## [v1.0.1]

### Chore
- Set HELO/EHLO using hostname
- Update Go dependencies
- Bump actions/setup-go from 5 to 6
- Bump actions/stale from 9.1.0 to 10.0.0


## [v1.0.0]

### Feature
- Add ability to check for latest version and self-update functionality

### Chore
- Change dependabot checks to quarterly
- Add golangci-lint tests to GitHub Actions
- Normalize error messages and improve resource cleanup in SMTP-related functions
- Switch to cliff for changelog generation


## [v0.0.8]

### Fix
- Remove leading dot when `-bs` is used (SMTP standard input)


## [v0.0.7]

### Chore
- Reset LastActive after successful SMTP on standard input
- Use `Message-ID` instead of `Message-Id`


## [v0.0.6]

### Chore
- Add changelog
- Automatically generate email headers if input does not contain valid email headers ([#10](https://github.com/axllent/mailpit/issues/10))


## [v0.0.5]

### Feature
- Add Message-Id and Date headers (if missing)

### Fix
- Prevent additional "\r\n.\r\n" after DATA is over ([#10](https://github.com/axllent/mailpit/issues/10))


## [v0.0.3]

### Feature
- Add optional logging


## [v0.0.1]



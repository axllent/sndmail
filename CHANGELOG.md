# Changelog

Notable changes to sndmail will be documented in this file.

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



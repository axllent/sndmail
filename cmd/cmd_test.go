package cmd

import (
	"bytes"
	"net/mail"
	"sort"
	"strings"
	"testing"
)

// ---- uniqueRecipients -------------------------------------------------------

func TestUniqueRecipients_Deduplication(t *testing.T) {
	config.Aliases = map[string]string{}
	in := []string{"alice@example.com", "Alice@Example.COM", "bob@example.com"}
	got := uniqueRecipients(in)
	if len(got) != 2 {
		t.Fatalf("expected 2 unique recipients, got %d: %v", len(got), got)
	}
}

func TestUniqueRecipients_AliasResolved(t *testing.T) {
	config.Aliases = map[string]string{"root": "admin@example.com"}
	got := uniqueRecipients([]string{"root"})
	if len(got) != 1 || got[0] != "admin@example.com" {
		t.Fatalf("expected alias to resolve to admin@example.com, got %v", got)
	}
}

func TestUniqueRecipients_EmptyAliasDropped(t *testing.T) {
	config.Aliases = map[string]string{"nobody": ""}
	got := uniqueRecipients([]string{"nobody"})
	if len(got) != 0 {
		t.Fatalf("expected empty alias to be dropped, got %v", got)
	}
}

func TestUniqueRecipients_NoAlias(t *testing.T) {
	config.Aliases = map[string]string{}
	got := uniqueRecipients([]string{"user@example.com"})
	if len(got) != 1 || got[0] != "user@example.com" {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestUniqueRecipients_CaseNormalised(t *testing.T) {
	config.Aliases = map[string]string{}
	got := uniqueRecipients([]string{"User@Example.COM"})
	if len(got) != 1 || got[0] != "user@example.com" {
		t.Fatalf("expected lowercase address, got %v", got)
	}
}

func TestUniqueRecipients_MixedAliasAndDirect(t *testing.T) {
	config.Aliases = map[string]string{"root": "admin@example.com"}
	in := []string{"root", "admin@example.com", "other@example.com"}
	got := uniqueRecipients(in)
	sort.Strings(got)
	want := []string{"admin@example.com", "other@example.com"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

// ---- injectMissingHeaders ---------------------------------------------------

func parseHeaders(t *testing.T, body []byte) mail.Header {
	t.Helper()
	msg, err := mail.ReadMessage(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to parse injected message: %v", err)
	}
	return msg.Header
}

func TestInjectMissingHeaders_AddsAll(t *testing.T) {
	body := []byte("Subject: test\r\n\r\nHello")
	got, err := injectMissingHeaders(body, "sender@example.com")
	if err != nil {
		t.Fatal(err)
	}
	h := parseHeaders(t, got)
	if h.Get("Message-ID") == "" {
		t.Error("expected Message-ID to be injected")
	}
	if h.Get("Date") == "" {
		t.Error("expected Date to be injected")
	}
	if h.Get("From") == "" {
		t.Error("expected From to be injected")
	}
}

func TestInjectMissingHeaders_PreservesExisting(t *testing.T) {
	body := []byte("Message-ID: <existing@test>\r\nDate: Mon, 01 Jan 2024 00:00:00 +0000\r\nFrom: <original@example.com>\r\n\r\nHello")
	got, err := injectMissingHeaders(body, "other@example.com")
	if err != nil {
		t.Fatal(err)
	}
	h := parseHeaders(t, got)
	if h.Get("Message-ID") != "<existing@test>" {
		t.Errorf("Message-ID should not be overwritten, got %q", h.Get("Message-ID"))
	}
	if h.Get("From") != "<original@example.com>" {
		t.Errorf("From should not be overwritten, got %q", h.Get("From"))
	}
}

func TestInjectMissingHeaders_MessageIDFormat(t *testing.T) {
	body := []byte("Subject: test\r\n\r\nHello")
	got, err := injectMissingHeaders(body, "sender@example.com")
	if err != nil {
		t.Fatal(err)
	}
	h := parseHeaders(t, got)
	mid := h.Get("Message-ID")
	if !strings.HasSuffix(mid, "@sndmail>") {
		t.Errorf("expected Message-ID to end with @sndmail>, got %q", mid)
	}
}

func TestInjectMissingHeaders_BodylessInput(t *testing.T) {
	// No headers at all — should not error and should inject all three headers.
	body := []byte("just plain text with no headers")
	got, err := injectMissingHeaders(body, "sender@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(got, []byte("Message-ID:")) {
		t.Error("expected Message-ID in output")
	}
}

// ---- smtpErrParser ----------------------------------------------------------

func TestSmtpErrParser_ValidCode(t *testing.T) {
	code, msg := smtpErrParser("550 5.1.1 User unknown")
	if code != 550 {
		t.Errorf("expected code 550, got %d", code)
	}
	if msg != "5.1.1 User unknown" {
		t.Errorf("unexpected message: %q", msg)
	}
}

func TestSmtpErrParser_NoMatch(t *testing.T) {
	raw := "something went wrong"
	code, msg := smtpErrParser(raw)
	if code != 0 {
		t.Errorf("expected code 0, got %d", code)
	}
	if msg != raw {
		t.Errorf("expected original string back, got %q", msg)
	}
}

func TestSmtpErrParser_421(t *testing.T) {
	code, msg := smtpErrParser("421 Service temporarily unavailable")
	if code != 421 {
		t.Errorf("expected 421, got %d", code)
	}
	if !strings.Contains(msg, "temporarily") {
		t.Errorf("unexpected message: %q", msg)
	}
}

// ---- getAddressArg ----------------------------------------------------------

func TestGetAddressArg_ValidFrom(t *testing.T) {
	addr, err := getAddressArg("FROM", "FROM:<sender@example.com>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr.Address != "sender@example.com" {
		t.Errorf("expected sender@example.com, got %q", addr.Address)
	}
}

func TestGetAddressArg_ValidRcpt(t *testing.T) {
	addr, err := getAddressArg("TO", "TO:<recipient@example.com>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr.Address != "recipient@example.com" {
		t.Errorf("expected recipient@example.com, got %q", addr.Address)
	}
}

func TestGetAddressArg_WrongVerb(t *testing.T) {
	_, err := getAddressArg("FROM", "TO:<recipient@example.com>")
	if err == nil {
		t.Error("expected error for mismatched verb")
	}
}

func TestGetAddressArg_MissingAngleBrackets(t *testing.T) {
	_, err := getAddressArg("FROM", "FROM:sender@example.com")
	if err == nil {
		t.Error("expected error when angle brackets are missing")
	}
}

func TestGetAddressArg_BadArguments(t *testing.T) {
	_, err := getAddressArg("FROM", "garbage")
	if err == nil {
		t.Error("expected error for garbage input")
	}
}

// ---- readSMTP ---------------------------------------------------------------

func TestReadSMTP_VerbOnly(t *testing.T) {
	verb, args := readSMTP("QUIT")
	if verb != "QUIT" || args != "" {
		t.Errorf("expected QUIT/'', got %q/%q", verb, args)
	}
}

func TestReadSMTP_VerbWithArgs(t *testing.T) {
	verb, args := readSMTP("EHLO mail.example.com")
	if verb != "EHLO" {
		t.Errorf("expected EHLO, got %q", verb)
	}
	if args != "mail.example.com" {
		t.Errorf("expected mail.example.com, got %q", args)
	}
}

func TestReadSMTP_Lowercase(t *testing.T) {
	verb, _ := readSMTP("helo localhost")
	if verb != "HELO" {
		t.Errorf("expected uppercase HELO, got %q", verb)
	}
}

func TestReadSMTP_ArgsWithSpaces(t *testing.T) {
	// Args part should not be split further.
	verb, args := readSMTP("MAIL FROM:<sender@example.com> SIZE=1234")
	if verb != "MAIL" {
		t.Errorf("expected MAIL, got %q", verb)
	}
	if args != "FROM:<sender@example.com> SIZE=1234" {
		t.Errorf("unexpected args: %q", args)
	}
}

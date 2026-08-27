package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T, name string) *os.File {
	t.Helper()
	file, err := os.Open(filepath.Join("..", "..", "fixtures", name))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { file.Close() })
	return file
}

func TestParseHAR(t *testing.T) {
	result, err := ParseHAR(fixture(t, "auth-401.har"))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Requests) != 1 || result.Statuses[0] != 401 || result.SourceSHA256 == "" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
func TestParseHARUnicode(t *testing.T) {
	result, err := ParseHAR(fixture(t, "unicode.har"))
	if err != nil {
		t.Fatal(err)
	}
	body, err := result.Requests[0].BodyBytes()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "Привет") || !strings.Contains(string(body), "你好") {
		t.Fatalf("unicode was not preserved: %s", body)
	}
}
func TestParseMalformedHAR(t *testing.T) {
	if _, err := ParseHAR(fixture(t, "malformed.har")); err == nil || !strings.Contains(err.Error(), "HAR_MALFORMED") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestParseHAREmpty(t *testing.T) {
	if _, err := ParseHAR(strings.NewReader(`{"log":{"entries":[]}}`)); err == nil {
		t.Fatal("expected empty HAR error")
	}
}

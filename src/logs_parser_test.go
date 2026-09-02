package src

import (
	"testing"
	"time"
)

func TestParseLogLine(t *testing.T) {
	line := `- - [13/Jul/2026:00:00:02 +0000] "GET /layerindex/recipe/137737/ HTTP/1.1" 200 16923 "-" "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.117 Safari/537.36"`

	got, err := ParseLogLine(line)
	if err != nil {
		t.Fatalf("ParseLogLine returned unexpected error: %v", err)
	}

	wantTimestamp := time.Date(2026, 7, 13, 0, 0, 2, 0, time.UTC)
	if !got.Timestamp.Equal(wantTimestamp) {
		t.Errorf("Timestamp = %v, want %v", got.Timestamp, wantTimestamp)
	}
	if got.Method != "GET" {
		t.Errorf("Method = %q, want %q", got.Method, "GET")
	}
	if got.Path != "/layerindex/recipe/137737/" {
		t.Errorf("Path = %q, want %q", got.Path, "/layerindex/recipe/137737/")
	}
	if got.Protocol != "HTTP/1.1" {
		t.Errorf("Protocol = %q, want %q", got.Protocol, "HTTP/1.1")
	}
	if got.ResponseCode != 200 {
		t.Errorf("ResponseCode = %d, want %d", got.ResponseCode, 200)
	}
	if got.BodyBytesSent != 16923 {
		t.Errorf("BodyBytesSent = %d, want %d", got.BodyBytesSent, 16923)
	}
	if got.HTTPReferrer != "-" {
		t.Errorf("HTTPReferrer = %q, want %q", got.HTTPReferrer, "-")
	}
	wantUserAgent := "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.117 Safari/537.36"
	if got.HTTPUserAgent != wantUserAgent {
		t.Errorf("HTTPUserAgent = %q, want %q", got.HTTPUserAgent, wantUserAgent)
	}
}

func TestParseLogLineDashBodyBytes(t *testing.T) {
	line := `- - [13/Jul/2026:00:00:02 +0000] "GET / HTTP/1.1" 304 - "-" "curl/8.0.1"`

	got, err := ParseLogLine(line)
	if err != nil {
		t.Fatalf("ParseLogLine returned unexpected error: %v", err)
	}
	if got.ResponseCode != 304 {
		t.Errorf("ResponseCode = %d, want %d", got.ResponseCode, 304)
	}
	if got.BodyBytesSent != 0 {
		t.Errorf("BodyBytesSent = %d, want %d", got.BodyBytesSent, 0)
	}
}

func TestParseLogLineInvalid(t *testing.T) {
	if _, err := ParseLogLine("not a valid log line"); err == nil {
		t.Fatal("expected an error for an invalid log line, got nil")
	}
}

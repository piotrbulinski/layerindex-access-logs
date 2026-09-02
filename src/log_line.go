package src

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// LogLine represents a single log entry with relevant fields.
// Example: - - [13/Jul/2026:00:00:02 +0000] "GET /layerindex/recipe/137737/ HTTP/1.1" 200 16923 "-" "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.117 Safari/537.36"
type LogLine struct {
	// Timestamp is parsed from the following format: 13/Jul/2026:00:00:02 +0000
	Timestamp time.Time

	// Method is the HTTP method used in the request (e.g., GET, POST).
	Method string

	// Path is the requested path in the HTTP request (e.g., /layerindex/recipe/137737/).
	Path string

	// Route is the matched route pattern for the request path (e.g., /layerindex/recipe/{id}/).
	Route string

	// Protocol is the HTTP protocol version used in the request (e.g., HTTP/1.1).
	Protocol string

	// ResponseCode is the HTTP response code returned for the request (e.g., 200, 404).
	ResponseCode int

	// BodyBytesSent is the number of bytes sent in the response body (e.g. 16923)
	BodyBytesSent int

	// HTTPReferrer is the HTTP referrer header value from the request.
	// Example: "https://layers.openembedded.org/"
	HTTPReferrer string

	// HTTPUserAgent is the HTTP user agent header value from the request.
	// Example: Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.117 Safari/537.36
	HTTPUserAgent string
}

// timeLayout matches the timestamp format used in the access logs.
// Example: 13/Jul/2026:00:00:02 +0000
const timeLayout = "02/Jan/2006:15:04:05 -0700"

// logLineRegex captures the timestamp, request line, response code,
// body bytes sent, referrer, and user agent from a combined-format log line.
var logLineRegex = regexp.MustCompile(`\[([^\]]+)\] "([^"]*)" (\d{3}) (\S+) "([^"]*)" "(.*)"$`)

// ParseLogLine parses a single log line and returns a LogLine struct.
func ParseLogLine(line string, router chi.Router) (*LogLine, error) {
	matches := logLineRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("log line does not match expected format: %q", line)
	}

	timestamp, err := time.Parse(timeLayout, matches[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp %q: %w", matches[1], err)
	}

	requestFields := strings.Fields(matches[2])
	if len(requestFields) != 3 {
		return nil, fmt.Errorf("unexpected request line format: %q", matches[2])
	}

	responseCode, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("failed to parse response code %q: %w", matches[3], err)
	}

	if responseCode >= 300 {
		return nil, fmt.Errorf("skipping requests with 3xx, 4xx and 5xx response codes")
	}

	bodyBytesSent := 0
	if matches[4] != "-" {
		bodyBytesSent, err = strconv.Atoi(matches[4])
		if err != nil {
			return nil, fmt.Errorf("failed to parse body bytes sent %q: %w", matches[4], err)
		}
	}

	httpMethod := requestFields[0]
	if httpMethod == "CONNECT" || httpMethod == "PRI" {
		// Skip CONNECT requests as they are not relevant for our analysis.
		return nil, fmt.Errorf("skipping CONNECT request log line")
	}

	parsedURL, err := url.Parse(strings.ReplaceAll(requestFields[1], "//", "/"))
	if err != nil {
		log.Println("failed to parse line:", line)
		log.Fatalln("failed to parse", requestFields[1], ":", err)
	}

	if parsedURL.Path == "/" {
		// Skip root path requests as they are not relevant for our analysis.
		return nil, fmt.Errorf("skipping root path log line")
	}

	if strings.HasPrefix(parsedURL.Path, "/static/") || parsedURL.Path == "/favicon.ico" || parsedURL.Path == "/robots.txt" || parsedURL.Path == "/yp-compat.png" {
		// Skip static files as they are not relevant for our analysis.
		return nil, fmt.Errorf("skipping static file log line")
	}

	if strings.HasPrefix(parsedURL.Path, "/logs-clean/") {
		return nil, fmt.Errorf("skipping logs-clean file log line")
	}

	if strings.HasPrefix(parsedURL.Path, "/static/") {
		// Skip static files as they are not relevant for our analysis.
		return nil, fmt.Errorf("skipping static file log line")
	}

	if strings.HasPrefix(parsedURL.Path, "/layersweb_access_logs") {
		// Skip access logs as they are not relevant for our analysis.
		return nil, fmt.Errorf("skipping access log line")
	}

	route := router.Find(chi.NewRouteContext(), http.MethodGet, parsedURL.Path)

	return &LogLine{
		Timestamp:     timestamp,
		Method:        httpMethod,
		Path:          requestFields[1],
		Protocol:      requestFields[2],
		Route:         route,
		ResponseCode:  responseCode,
		BodyBytesSent: bodyBytesSent,
		HTTPReferrer:  matches[5],
		HTTPUserAgent: matches[6],
	}, nil
}

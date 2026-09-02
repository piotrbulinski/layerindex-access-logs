package src

import (
	"bufio"
	"context"
	"fmt"
	"iter"
	"log"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/klauspost/compress/zstd"
)

// ParseAllLogs parses every available access log and imports the resulting
// LogLines into the PostgreSQL database. The target database and table are
// created when missing and the table is truncated before the data is loaded.
func ParseAllLogs(router chi.Router) {
	ctx := context.Background()

	if err := EnsureDatabase(ctx); err != nil {
		log.Fatalf("failed to prepare database: %v", err)
	}

	count, err := ImportLogLines(ctx, ParseAllLogLines(router))
	if err != nil {
		log.Fatalf("failed to import log lines: %v", err)
	}

	fmt.Printf("imported %d log lines\n", count)
}

// ParseAllLogLines returns an iterator over every parsed access-log entry,
// spanning all available daily log files.
func ParseAllLogLines(router chi.Router) iter.Seq[LogLine] {
	return func(yield func(LogLine) bool) {
		startDate := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)

		for date := startDate; date.Before(time.Now()); date = date.AddDate(0, 0, 1) {
			for logLine := range ParseLogs(date, router) {
				if !yield(logLine) {
					return
				}
			}
		}
	}
}

// ParseLogs returns an iterator over the parsed access-log entries for the
// given date. It can be used directly in a range expression. Lines that cannot
// be parsed are logged and skipped.
func ParseLogs(date time.Time, router chi.Router) iter.Seq[LogLine] {
	return func(yield func(LogLine) bool) {
		logFile, err := GetLogsFile(date)
		if err != nil {
			log.Fatalf("failed to obtain logs for date %s: %v", date.Format("2006-01-02"), err)
		}
		defer logFile.Close()

		reader, err := zstd.NewReader(logFile)
		if err != nil {
			log.Fatalf("failed to create zstd reader: %v", err)
		}
		defer reader.Close()

		scanner := bufio.NewScanner(reader)

		for scanner.Scan() {
			logLine, err := ParseLogLine(scanner.Text(), router)
			if err != nil {
				// log.Printf("skipping unparseable log line: %s\n%v", scanner.Text(), err)
				continue
			}

			if !yield(*logLine) {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatalf("error reading from zstd reader: %v", err)
		}
	}
}

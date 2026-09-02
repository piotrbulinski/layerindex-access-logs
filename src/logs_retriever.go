package src

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	logsBaseURLTemplate  = "https://layers.openembedded.org/logs-clean/"
	logsFileNameTemplate = "layersweb_%s_log.zst"
)

// GetLogsFile retrieves the logs file for the given date.
// It first attempts to open a local file with the expected name.
// If the local file does not exist, it fetches the logs from the remote URL and saves them locally for future use.
// The function returns an io.ReadCloser for reading the logs.
func GetLogsFile(date time.Time) (io.ReadCloser, error) {
	filePath := fmt.Sprintf(logsFileNameTemplate, date.Format("20060102"))

	// Attempt to open the local file first
	if file, err := os.Open(filePath); err == nil {
		return file, nil
	}

	logsURL := logsBaseURLTemplate + filePath

	log.Printf("Fetching logs for date: %s - %s", date.Format("2006-01-02"), logsURL)

	httpResponse, err := http.Get(logsURL)
	if err != nil {
		log.Fatalf("failed to obtains logs: %v", err)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		log.Fatalf("failed to obtain logs: %s", httpResponse.Status)
	}

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("failed to create local log file: %v", err)
	}

	_, err = io.Copy(file, httpResponse.Body)
	if err != nil {
		log.Fatalf("failed to save logs to local file: %v", err)
	}

	file.Seek(0, 0)

	return file, nil
}

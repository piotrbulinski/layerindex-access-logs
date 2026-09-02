package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/klauspost/compress/zstd"
	"piotr.bulinski.no/layerindex-access-logs/src"
)

func main() {
	router := src.NewRouter("urls.json")

	logsFileNameTemplate := "layersweb_%s_log.txt"

	startDate := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)
	for date := startDate; date.Before(time.Now()); date = date.AddDate(0, 0, 1) {
		logFile, err := src.GetLogsFile(date)
		if err != nil {
			log.Fatalf("failed to obtain logs for date %s: %v", date.Format("2006-01-02"), err)
		}
		defer logFile.Close()

		reader, err := zstd.NewReader(logFile)
		if err != nil {
			log.Fatalf("failed to create zstd reader: %v", err)
		}
		defer reader.Close()

		filePath := fmt.Sprintf(logsFileNameTemplate, date.Format("20060102"))

		outFile, err := os.Create(filePath)
		if err != nil {
			log.Fatalf("failed to create output file: %v", err)
		}
		defer outFile.Close()

		_, err = reader.WriteTo(outFile)
		if err != nil {
			log.Fatalf("failed to write decompressed data to output file: %v", err)
		}

		log.Printf("Decompressed logs for date: %s - saved to %s", date.Format("2006-01-02"), filePath)

	}
	src.ParseAllLogs(router)
}

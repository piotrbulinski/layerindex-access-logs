package src

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
)

// KnownURL represents a known layerindex URL with its associated module and name.
type KnownURL struct {
	URL    string
	Module string
	Name   string
}

// ListKnownURLs reads a file containing known layerindex URLs and returns a slice of KnownURL structs.
func ListKnownURLs(filePath string) []KnownURL {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	var knownURLs []KnownURL

	if err = json.NewDecoder(file).Decode(&knownURLs); err != nil {
		log.Fatalf("Failed to decode JSON: %v", err)
	}

	return knownURLs
}

func noop(w http.ResponseWriter, r *http.Request) {}

var placeholderReplacer1 = regexp.MustCompile("<([^:>]+)>")
var placeholderReplacer2 = regexp.MustCompile("<[^:]+:?([^>]+)>")

func NewRouter(filePath string) chi.Router {
	knownURLs := ListKnownURLs(filePath)

	router := chi.NewRouter()

	for _, knownURL := range knownURLs {
		// Replace "history/<var>/" with "history/<var2>/" to avoid conflicts in route matching
		url := strings.ReplaceAll(knownURL.URL, "history/<var>/", "history/<var2>/")

		// Replace Django style placeholders with chi-style placeholders
		url = placeholderReplacer1.ReplaceAllString(url, "{$1}")
		url = placeholderReplacer2.ReplaceAllString(url, "{$1}")

		// replace escaped dots with actual dots to match the route correctly
		url = strings.ReplaceAll(url, "\\.", ".")

		// Register the route with a no-op handler
		router.Get(url, noop)
	}

	return router
}

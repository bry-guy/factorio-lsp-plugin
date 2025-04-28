package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log" // Import the log package
	"net/http"
)

// DownloadAndParseAPI downloads JSON from the given URL and unmarshals it into the provided interface.
func DownloadAndParseAPI(url string, v interface{}) error {
	log.Printf("Attempting to download API from: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to download API from %s: %v", url, err)
		return fmt.Errorf("failed to download API from %s: %w", url, err)
	}
	defer resp.Body.Close()
	log.Printf("Download successful from %s, status code: %d", url, resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download API from %s: received status code %d", url, resp.StatusCode)
	}

	log.Printf("Reading response body from %s", url)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body from %s: %v", url, err)
		return fmt.Errorf("failed to read response body from %s: %w", url, err)
	}
	log.Printf("Successfully read %d bytes from %s", len(body), url)

	log.Printf("Attempting to parse JSON from %s", url)
	err = json.Unmarshal(body, v)
	if err != nil {
		log.Printf("Failed to parse JSON from %s: %v", url, err)
		return fmt.Errorf("failed to parse JSON from %s: %w", url, err)
	}
	log.Printf("Successfully parsed JSON from %s", url)

	return nil
}

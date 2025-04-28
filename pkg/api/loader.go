package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DownloadAndParseAPI downloads JSON from the given URL and unmarshals it into the provided interface.
func DownloadAndParseAPI(url string, v interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download API from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download API from %s: received status code %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		return fmt.Errorf("failed to parse JSON from %s: %w", url, err)
	}

	return nil
}

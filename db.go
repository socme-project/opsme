package opsme

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type ArtemisRecord struct {
	ID       string `json:"id"`
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
}

func GetArtemisMachines() ([]ArtemisRecord, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(os.Getenv("POCKETBASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PocketBase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"PocketBase returned status code %d: %s",
			resp.StatusCode,
			resp.Status,
		)
	}

	var result struct {
		Items []ArtemisRecord `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode PocketBase response: %w", err)
	}

	return result.Items, nil
}

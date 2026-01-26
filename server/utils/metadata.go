package utils

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type VideoMetadata struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	URL       string    `json:"url"`
}

var (
	metadataFile = "./videos/metadata.json"
	mu           sync.Mutex
)

func SaveMetadata(v VideoMetadata) error {
	mu.Lock()
	defer mu.Unlock()

	var videos []VideoMetadata
	data, err := os.ReadFile(metadataFile)
	if err == nil {
		json.Unmarshal(data, &videos)
	}

	// Update or Append
	found := false
	for i, video := range videos {
		if video.ID == v.ID {
			videos[i] = v
			found = true
			break
		}
	}
	if !found {
		videos = append(videos, v)
	}

	newData, _ := json.MarshalIndent(videos, "", "  ")
	return os.WriteFile(metadataFile, newData, 0644)
}

func GetAllMetadata() ([]VideoMetadata, error) {
	mu.Lock()
	defer mu.Unlock()

	var videos []VideoMetadata
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []VideoMetadata{}, nil
		}
		return nil, err
	}
	err = json.Unmarshal(data, &videos)
	return videos, err
}

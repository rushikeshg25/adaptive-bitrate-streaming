package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
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

var metadataFile = "./videos/metadata.json"
var mu sync.Mutex

func readMetadata() ([]VideoMetadata, error) {
	b, e := os.ReadFile(metadataFile)
	if os.IsNotExist(e) {
		return []VideoMetadata{}, nil
	}
	if e != nil {
		return nil, e
	}
	var v []VideoMetadata
	e = json.Unmarshal(b, &v)
	return v, e
}
func SaveMetadata(v VideoMetadata) error {
	mu.Lock()
	defer mu.Unlock()
	videos, e := readMetadata()
	if e != nil {
		return e
	}
	found := false
	for i := range videos {
		if videos[i].ID == v.ID {
			videos[i] = v
			found = true
			break
		}
	}
	if !found {
		videos = append(videos, v)
	}
	b, e := json.MarshalIndent(videos, "", "  ")
	if e != nil {
		return e
	}
	dir := filepath.Dir(metadataFile)
	if e = os.MkdirAll(dir, 0700); e != nil {
		return e
	}
	f, e := os.CreateTemp(dir, ".metadata-*")
	if e != nil {
		return e
	}
	name := f.Name()
	defer os.Remove(name)
	if _, e = f.Write(b); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	if e = os.Rename(name, metadataFile); e != nil {
		return e
	}
	d, e := os.Open(dir)
	if e != nil {
		return e
	}
	defer d.Close()
	return d.Sync()
}
func GetAllMetadata() ([]VideoMetadata, error) { mu.Lock(); defer mu.Unlock(); return readMetadata() }
func RecoverProcessing() error {
	videos, e := GetAllMetadata()
	if e != nil {
		return e
	}
	for _, v := range videos {
		if v.Status == "Processing" {
			v.Status = "Failed"
			if e = SaveMetadata(v); e != nil {
				return e
			}
		}
	}
	return nil
}

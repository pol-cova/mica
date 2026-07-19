// Package catalog loads explicit local service context without requiring a
// service-catalog vendor or exposing credentials to clients.
package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mica-dev/mica/internal/incidents"
)

type File struct {
	Services []incidents.Service `json:"services"`
}

func Load(path string) ([]incidents.Service, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file File
	if err := json.Unmarshal(contents, &file); err != nil {
		return nil, fmt.Errorf("parse catalog: %w", err)
	}
	if len(file.Services) == 0 {
		return nil, fmt.Errorf("catalog has no services")
	}
	seen := map[string]bool{}
	for index := range file.Services {
		service := &file.Services[index]
		if service.ID == "" || service.Name == "" || service.Environment == "" {
			return nil, fmt.Errorf("catalog service requires id, name, and environment")
		}
		if seen[service.ID] {
			return nil, fmt.Errorf("duplicate service ID %q", service.ID)
		}
		seen[service.ID] = true
		if service.Source.Kind == "" {
			service.Source.Kind = "configured catalog file"
		}
		if service.Source.RefreshedAt.IsZero() {
			service.Source.RefreshedAt = time.Now().UTC()
		}
	}
	return file.Services, nil
}

package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnrichesCatalogSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "services.json")
	if err := os.WriteFile(path, []byte(`{"services":[{"id":"checkout","name":"checkout","environment":"demo"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	services, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 1 || services[0].Source.Kind != "configured catalog file" || services[0].Source.RefreshedAt.IsZero() {
		t.Fatalf("services = %#v", services)
	}
}

func TestLoadRejectsDuplicateServiceIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "services.json")
	if err := os.WriteFile(path, []byte(`{"services":[{"id":"checkout","name":"a","environment":"demo"},{"id":"checkout","name":"b","environment":"demo"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected duplicate ID error")
	}
}

package immich

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUserKeys(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "keys.json")
		os.WriteFile(path, []byte(`{"users":{"user-1":"key-1","user-2":"key-2"}}`), 0644)

		keys, err := LoadUserKeys(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(keys.Users) != 2 {
			t.Fatalf("expected 2 users, got %d", len(keys.Users))
		}
		if keys.Users["user-1"] != "key-1" {
			t.Errorf("expected key-1, got %s", keys.Users["user-1"])
		}
	})

	t.Run("empty users map", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "keys.json")
		os.WriteFile(path, []byte(`{}`), 0644)

		keys, err := LoadUserKeys(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if keys.Users == nil {
			t.Fatal("expected non-nil users map")
		}
		if len(keys.Users) != 0 {
			t.Fatalf("expected 0 users, got %d", len(keys.Users))
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := LoadUserKeys("/nonexistent/keys.json")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "keys.json")
		os.WriteFile(path, []byte(`not json`), 0644)

		_, err := LoadUserKeys(path)
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestUserKeysGetAPIKey(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		keys := &UserKeys{Users: map[string]string{"owner-1": "api-key-1"}}
		key, ok := keys.GetAPIKey("owner-1")
		if !ok {
			t.Fatal("expected key to be found")
		}
		if key != "api-key-1" {
			t.Errorf("expected api-key-1, got %s", key)
		}
	})

	t.Run("not found", func(t *testing.T) {
		keys := &UserKeys{Users: map[string]string{"owner-1": "api-key-1"}}
		_, ok := keys.GetAPIKey("owner-2")
		if ok {
			t.Fatal("expected key to not be found")
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var keys *UserKeys
		_, ok := keys.GetAPIKey("owner-1")
		if ok {
			t.Fatal("expected key to not be found on nil")
		}
	})
}

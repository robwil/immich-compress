package immich

import (
	"encoding/json"
	"fmt"
	"os"
)

type UserKeys struct {
	Users map[string]string `json:"users"`
}

func LoadUserKeys(path string) (*UserKeys, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read user keys file: %w", err)
	}

	var keys UserKeys
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, fmt.Errorf("failed to parse user keys file: %w", err)
	}

	if keys.Users == nil {
		keys.Users = make(map[string]string)
	}

	return &keys, nil
}

func (k *UserKeys) GetAPIKey(ownerID string) (string, bool) {
	if k == nil {
		return "", false
	}
	key, ok := k.Users[ownerID]
	return key, ok
}

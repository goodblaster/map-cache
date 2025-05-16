package caches

import (
	"context"
	"encoding/json"
	"os"

	"github.com/goodblaster/errors"
	"github.com/google/uuid"
)

type BackupContainer struct {
	Name           string           `json:"name"`
	Data           map[string]any   `json:"data"`
	KeyExpirations map[string]int64 `json:"key_expirations"`
	Expiration     *int64           `json:"expiration,omitempty"`
}

// Backup creates a backup of the specified cache and saves it to the given file.
func Backup(ctx context.Context, cacheName string, outFile string) error {
	if cacheName == "" {
		cacheName = DefaultName
	}

	cache, err := FetchCache(cacheName)
	if err != nil {
		return errors.Wrapf(err, "error fetching cache %q", cacheName)
	}

	id := cacheName + "-" + uuid.New().String()
	cache.Acquire(id)
	defer cache.Release(id)

	f, err := os.Create(outFile)
	if err != nil {
		return errors.Wrapf(err, "error creating backup file %q", outFile)
	}
	defer f.Close()

	keysTTLs := make(map[string]int64)
	for k, v := range cache.keyExps {
		if v != nil {
			keysTTLs[k] = v.Expiration
		}
	}

	backup := BackupContainer{
		Name:           cacheName,
		Data:           cache.Map.Data().(map[string]any),
		KeyExpirations: keysTTLs,
	}

	if cache.exp != nil {
		backup.Expiration = &cache.exp.Expiration
	}

	err = json.NewEncoder(f).Encode(backup)
	if err != nil {
		return errors.Wrapf(err, "error encoding cache %q to file %q", cacheName, outFile)
	}

	return nil
}

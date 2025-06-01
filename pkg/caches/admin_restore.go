package caches

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/goodblaster/errors"
)

func Restore(ctx context.Context, cacheName string, inFile string) error {
	if cacheName == "" {
		cacheName = DefaultName
	}

	f, err := os.Open(inFile)
	if err != nil {
		return errors.Wrapf(err, "error opening backup file %q", inFile)
	}
	defer f.Close()

	backup := BackupContainer{
		Data:           map[string]any{},
		KeyExpirations: map[string]int64{},
	}

	err = json.NewDecoder(f).Decode(&backup)
	if err != nil {
		return errors.Wrapf(err, "error decoding backup file %q", inFile)
	}

	cache := New()
	if err = cache.cmap.Set(ctx, backup.Data); err != nil {
		return errors.Wrapf(err, "error setting data in cache %q", cacheName)
	}

	// Set the cache expiration
	if cacheName != DefaultName && backup.Expiration != nil {
		exp := time.Unix(*backup.Expiration, 0)
		cache.exp = FutureFunc(int64(exp.Sub(time.Now()).Seconds()), func() {
			caches.Delete(cacheName)
		})
	}

	// Set the key expirations
	for key, ttl := range backup.KeyExpirations {
		exp := time.Unix(ttl, 0)
		cache.keyExps[key] = FutureFunc(int64(exp.Sub(time.Now()).Seconds()), func() {
			_ = cache.Delete(ctx, key)
		})
	}

	// Delete the existing cache if it exists, and its expirations.
	_ = DeleteCache(cacheName)

	caches.Store(cacheName, cache)
	return nil
}

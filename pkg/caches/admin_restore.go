package caches

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/log"
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

	backup := RestoreContainer{
		Data:           map[string]any{},
		KeyExpirations: map[string]int64{},
		Triggers:       map[string][]RawTrigger{},
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
		ms := time.Until(exp).Milliseconds()
		cache.exp = FutureFunc(ms, func() {
			removeCache(cacheName, cache)
		})
		cache.expMillis = &ms
	}

	// Set the key expirations
	for key, ttl := range backup.KeyExpirations {
		exp := time.Unix(ttl, 0)
		duration := int64(exp.Sub(time.Now()).Milliseconds())

		// Skip expired keys - they're already expired, no need to set a timer
		if duration <= 0 {
			log.With("key", key).With("expired_ms_ago", -duration).Warn("skipping expired key during restore")
			continue
		}

		// Use the normal TTL path so the expiration worker (which holds the
		// cache lock) performs the delete.
		if err := cache.SetKeyTTL(context.Background(), key, duration); err != nil {
			log.WithError(err).With("key", key).Warn("failed to restore key expiration")
		}
	}

	// Set the triggers
	cache.triggers = make(map[string][]Trigger)
	for key, rawTriggers := range backup.Triggers {
		for _, rawTrigger := range rawTriggers {
			trigger := Trigger{
				Id:      rawTrigger.Id,
				Key:     key,
				Command: rawTrigger.Command.Command,
			}
			cache.triggers[key] = append(cache.triggers[key], trigger)
		}
	}

	// Swap in the restored cache, then shut down the one it replaced (if any).
	if old, loaded := caches.Swap(cacheName, cache); loaded {
		old.(*Cache).Close()
	}
	return nil
}

package caches

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/logos"
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
		cache.exp = FutureFunc(int64(exp.Sub(time.Now()).Milliseconds()), func() {
			caches.Delete(cacheName)
		})
	}

	// Set the key expirations
	for key, ttl := range backup.KeyExpirations {
		key := key // Capture loop variable to avoid closure bug
		exp := time.Unix(ttl, 0)
		duration := int64(exp.Sub(time.Now()).Milliseconds())

		// Skip expired keys - they're already expired, no need to set a timer
		if duration <= 0 {
			logos.Warnf("skipping expired key %s during restore (expired %d ms ago)", key, -duration)
			continue
		}

		cache.keyExps[key] = FutureFunc(duration, func() {
			// Delete the key when TTL expires
			// Log errors but don't fail - TTL cleanup is best-effort
			if err := cache.Delete(ctx, key); err != nil {
				logos.WithError(err).Warnf("failed to delete expired key %s during restore", key)
			}
		})
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

	// Delete the existing cache if it exists, and its expirations.
	// Log errors but don't fail - deletion is best-effort before restore
	if err := DeleteCache(cacheName); err != nil {
		logos.WithError(err).Warnf("failed to delete existing cache %s before restore", cacheName)
	}

	caches.Store(cacheName, cache)
	return nil
}

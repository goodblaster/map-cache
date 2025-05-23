package caches

import (
	"context"
	"strings"

	"github.com/goodblaster/errors"
)

// Trigger - A trigger is a command that is executed when a specified key is modified.
// For now, this command is only called on-change. Future versions may support
// additional commands like on-delete.
type Trigger struct {
	Id      string  `json:"id"`
	Key     string  `json:"key"`
	Command Command `json:"command"`
}

// OnChange gets called whenever there was a successful data replacement.
// All trigger keys are checked, and for each match, the trigger command is called.
// Initially, this will be immediately recursive. Future versions may complete a full
// command sequence before delayed recursion. Future version should also seek to
// prevent infinite recursion.
func (cache *Cache) OnChange(ctx context.Context, key string, oldValue any, newValue any) error {
	for triggerKey, triggers := range cache.triggers {
		matchingKeys := cache.KeysMatch(ctx, triggerKey, key)

		// Full list might be used later. For now, accept any match.
		if len(matchingKeys) > 0 {
			for _, trigger := range triggers {
				if res := trigger.Command.Do(ctx, cache); res.Error != nil {
					return errors.Wrap(res.Error, "trigger failed")
				}
			}
		}
	}

	return nil
}

func (cache *Cache) KeysMatch(ctx context.Context, triggerKey, dataKey string) []string {
	if !strings.Contains(triggerKey, "*") {
		if triggerKey == dataKey {
			return []string{dataKey}
		}
		return nil
	}

	var keys []string
	wildKeys := cache.cmap.WildKeys(ctx, triggerKey)
	for _, wildKey := range wildKeys {
		if wildKey == dataKey {
			keys = append(keys, wildKey)
		}
	}

	return keys
}

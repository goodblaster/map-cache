package caches

import (
	"context"
	"strings"

	"github.com/goodblaster/errors"
)

// Trigger recursion limits
const (
	// MaxTriggerDepth is the maximum number of nested trigger executions allowed.
	// This prevents infinite loops where trigger A fires trigger B which fires A again.
	MaxTriggerDepth = 10
)

// Context key types for type-safe context value access
type triggerDepthKey struct{}
type triggerVarsKey struct{}
type triggerOldValueKey struct{}
type triggerNewValueKey struct{}

var (
	triggerDepthContextKey    = triggerDepthKey{}
	triggerVarsContextKey     = triggerVarsKey{}
	triggerOldValueContextKey = triggerOldValueKey{}
	triggerNewValueContextKey = triggerNewValueKey{}
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
//
// INFINITE LOOP PROTECTION:
// Triggers can recursively fire other triggers. To prevent infinite loops,
// we track the recursion depth and limit it to MaxTriggerDepth (10 levels).
// If the depth limit is exceeded, an error is returned.
func (cache *Cache) OnChange(ctx context.Context, key string, oldValue any, newValue any) error {
	// Check current trigger depth
	depth := getTriggerDepth(ctx)
	if depth > MaxTriggerDepth {
		return errors.Newf("trigger recursion depth limit exceeded (max: %d) - possible infinite loop detected", MaxTriggerDepth)
	}

	// Increment depth for nested trigger executions
	ctx = context.WithValue(ctx, triggerDepthContextKey, depth+1)
	for triggerKey, triggers := range cache.triggers {
		matchingKeys := cache.KeysMatch(ctx, triggerKey, key)

		// Full list might be used later. For now, accept any match.
		if len(matchingKeys) > 0 {
			vars, err := ExtractWildcardMatches(key, triggerKey)
			if err != nil {
				return errors.Wrapf(err, "failed to extract wildcard matches for key %s", key)
			}

			for _, trigger := range triggers {
				cmdCtx := context.WithValue(ctx, triggerVarsContextKey, vars)
				cmdCtx = context.WithValue(cmdCtx, triggerOldValueContextKey, oldValue)
				cmdCtx = context.WithValue(cmdCtx, triggerNewValueContextKey, newValue)
				if res := trigger.Command.Do(cmdCtx, cache); res.Error != nil {
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

// ExtractWildcardMatches returns values that match wildcards in triggerKey.
func ExtractWildcardMatches(key, triggerKey string) ([]string, error) {
	// Normalize by trimming any leading/trailing slashes
	key = strings.Trim(key, "/")
	triggerKey = strings.Trim(triggerKey, "/")

	keyParts := strings.Split(key, "/")
	triggerParts := strings.Split(triggerKey, "/")

	if len(keyParts) != len(triggerParts) {
		return nil, errors.Newf("mismatched path lengths: %v vs %v", keyParts, triggerParts)
	}

	var matches []string
	for i := range keyParts {
		if triggerParts[i] == "*" {
			if keyParts[i] == "" {
				return nil, errors.Newf("wildcard at index %d matched empty segment", i)
			}
			matches = append(matches, keyParts[i])
		} else if triggerParts[i] != keyParts[i] {
			return nil, errors.Newf("segment mismatch at index %d: %s != %s", i, triggerParts[i], keyParts[i])
		}
	}
	return matches, nil
}

// getTriggerDepth retrieves the current trigger recursion depth from context.
// Returns 0 if not set (first trigger execution).
func getTriggerDepth(ctx context.Context) int {
	if depth, ok := ctx.Value(triggerDepthContextKey).(int); ok {
		return depth
	}
	return 0
}

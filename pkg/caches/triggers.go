package caches

import (
	"context"
	"strings"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/config"
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
		return ErrTriggerRecursionLimit.Format(MaxTriggerDepth)
	}

	// Most writes have no triggers to consider.
	if len(cache.triggers) == 0 {
		return nil
	}

	depthSet := false
	for triggerKey, triggers := range cache.triggers {
		// Match the changed key against the trigger pattern segment by segment.
		// This avoids walking the cache data for every write.
		vars, ok := matchPattern(key, triggerKey)
		if !ok {
			continue
		}

		// Increment depth for nested trigger executions
		if !depthSet {
			ctx = context.WithValue(ctx, triggerDepthContextKey, depth+1)
			depthSet = true
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

	return nil
}

// KeysMatch returns []string{dataKey} if dataKey matches the trigger pattern
// triggerKey, or nil otherwise.
func (cache *Cache) KeysMatch(ctx context.Context, triggerKey, dataKey string) []string {
	if _, err := ExtractWildcardMatches(dataKey, triggerKey); err != nil {
		return nil
	}
	return []string{dataKey}
}

// ExtractWildcardMatches returns values that match wildcards in triggerKey.
func ExtractWildcardMatches(key, triggerKey string) ([]string, error) {
	// Normalize by trimming any leading/trailing delimiters
	key = trimDelimiters(key)
	triggerKey = trimDelimiters(triggerKey)

	keyParts := strings.Split(key, config.KeyDelimiter)
	triggerParts := strings.Split(triggerKey, config.KeyDelimiter)

	if len(keyParts) != len(triggerParts) {
		return nil, ErrMismatchedPathLengths.Format(keyParts, triggerParts)
	}

	var matches []string
	for i := range keyParts {
		if triggerParts[i] == "*" {
			if keyParts[i] == "" {
				return nil, ErrWildcardEmptySegment.Format(i)
			}
			matches = append(matches, keyParts[i])
		} else if triggerParts[i] != keyParts[i] {
			return nil, ErrSegmentMismatch.Format(i, triggerParts[i], keyParts[i])
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

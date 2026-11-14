package containers

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/goodblaster/map-cache/internal/config"
)

// GabsMap wraps a Gabs container for JSON-like nested data structures.
// INVARIANT: The root container is always a map[string]any, never an array or scalar.
// This invariant is established by gabs.New() which creates a JSON object,
// and must be maintained by all Set operations.
type GabsMap struct {
	container *gabs.Container
}

// NewGabsMap creates a new GabsMap with an empty JSON object as the root.
// The root is guaranteed to be map[string]any.
func NewGabsMap() *GabsMap {
	return &GabsMap{
		container: gabs.New(), // Creates JSON object: map[string]any
	}
}

func (gMap *GabsMap) Get(ctx context.Context, hierarchy ...string) (Data, error) {
	c := gMap.container.Search(hierarchy...)
	if c == nil {
		return nil, ErrNotFound
	}
	return c.Data(), nil
}

func (gMap *GabsMap) Delete(ctx context.Context, path ...string) error {
	return gMap.container.Delete(path...)
}

func (gMap *GabsMap) ArrayRemove(ctx context.Context, index int, path ...string) error {
	return gMap.container.ArrayRemove(index, path...)
}

func (gMap *GabsMap) ArrayAppend(ctx context.Context, value any, path ...string) error {
	return gMap.container.ArrayAppend(value, path...)
}

func (gMap *GabsMap) ArrayResize(ctx context.Context, newSize int, path ...string) error {
	if newSize < 0 {
		return fmt.Errorf("newSize cannot be negative: %d", newSize)
	}

	c := gMap.container.Search(path...)
	if c == nil {
		return ErrNotFound
	}

	data := c.Data()
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return fmt.Errorf("value at path is not a slice or array, got %T", data)
	}

	currLen := v.Len()

	switch {
	case currLen == newSize:
		return nil
	case currLen < newSize:
		// Append nils until newSize is reached
		for i := currLen; i < newSize; i++ {
			if err := gMap.container.ArrayAppend(nil, path...); err != nil {
				return fmt.Errorf("error appending nil at index %d: %w", i, err)
			}
		}
	case currLen > newSize:
		// Remove from the end to avoid index shifting
		for i := currLen - 1; i >= newSize; i-- {
			if err := gMap.container.ArrayRemove(i, path...); err != nil {
				return fmt.Errorf("error removing index %d: %w", i, err)
			}
		}
	}

	return nil
}

func (gMap *GabsMap) Exists(ctx context.Context, path ...string) bool {
	return gMap.container.Exists(path...)
}

func (gMap *GabsMap) Set(ctx context.Context, value any, path ...string) error {
	_, err := gMap.container.Set(value, path...)
	return err
}

// Data returns the underlying map data.
// SAFETY: This type assertion is safe because NewGabsMap() ensures the root
// is always a map[string]any (created by gabs.New()). The container is never
// replaced with a non-map value - only nested values within it are modified.
func (gMap *GabsMap) Data(ctx context.Context) map[string]any {
	return gMap.container.Data().(map[string]any)
}

// WildKeys returns every existing key that matches path, where a "*" segment
// matches any single object key or array index. Null values are not matched.
func (gMap *GabsMap) WildKeys(ctx context.Context, path string) []string {
	delim := config.KeyDelimiter
	tokens := strings.Split(path, delim)
	var results []string

	// Walk the raw data directly. Going through gabs' Path() per segment would
	// re-split every key on "." (breaking keys that contain dots) and allocate
	// a container per visited node.
	var walk func(node any, idx int, currentPath []string)
	walk = func(node any, idx int, currentPath []string) {
		if node == nil {
			return
		}

		if idx >= len(tokens) {
			results = append(results, strings.Join(currentPath, delim))
			return
		}

		token := tokens[idx]

		switch data := node.(type) {
		case map[string]any:
			if token == "*" {
				for key, child := range data {
					walk(child, idx+1, append(currentPath, key))
				}
				return
			}
			if child, ok := data[token]; ok {
				walk(child, idx+1, append(currentPath, token))
			}
		case []any:
			if token == "*" {
				for i, child := range data {
					walk(child, idx+1, append(currentPath, strconv.Itoa(i)))
				}
				return
			}
			if i, err := strconv.Atoi(token); err == nil && i >= 0 && i < len(data) {
				walk(data[i], idx+1, append(currentPath, token))
			}
		}
	}

	walk(gMap.container.Data(), 0, make([]string, 0, len(tokens)))
	return results
}

package containers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs/v2"
)

type GabsMap struct {
	container *gabs.Container
}

func NewGabsMap() *GabsMap {
	return &GabsMap{
		container: gabs.New(),
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

func (gMap *GabsMap) Exists(ctx context.Context, path ...string) bool {
	return gMap.container.Exists(path...)
}

func (gMap *GabsMap) Set(ctx context.Context, value any, path ...string) error {
	_, err := gMap.container.Set(value, path...)
	return err
}

func (gMap *GabsMap) Data(ctx context.Context) map[string]any {
	return gMap.container.Data().(map[string]any)
}

func (gMap *GabsMap) WildKeys(ctx context.Context, path string) []string {
	var results []string
	tokens := strings.Split(path, "/")

	var walk func(node *gabs.Container, idx int, currentPath []string)
	walk = func(node *gabs.Container, idx int, currentPath []string) {
		if node == nil || node.Data() == nil {
			return
		}

		if idx >= len(tokens) {
			results = append(results, strings.Join(currentPath, "/"))
			return
		}

		token := tokens[idx]

		if token == "*" {
			switch data := node.Data().(type) {
			case map[string]interface{}:
				for key := range data {
					child := node.Path(key)
					walk(child, idx+1, append(currentPath, key))
				}
			case []interface{}:
				for i := range data {
					child := node.Index(i)
					walk(child, idx+1, append(currentPath, fmt.Sprintf("%d", i)))
				}
			}
		} else {
			// Try map lookup
			child := node.Path(token)
			if child != nil && child.Data() != nil {
				walk(child, idx+1, append(currentPath, token))
				return
			}

			// Try array index
			if i, err := strconv.Atoi(token); err == nil {
				child := node.Index(i)
				if child != nil && child.Data() != nil {
					walk(child, idx+1, append(currentPath, token))
				}
			}
		}
	}

	walk(gMap.container, 0, []string{})
	return results
}

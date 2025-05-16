package containers

import (
	"context"

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

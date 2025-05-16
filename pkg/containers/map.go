package containers

import "context"

type Data any
type Map interface {
	Get(ctx context.Context, hierarchy ...string) (Data, error)
	Set(ctx context.Context, value any, hierarchy ...string) error
	Delete(ctx context.Context, hierarchy ...string) error
	ArrayRemove(ctx context.Context, index int, path ...string) error
	Exists(ctx context.Context, hierarchy ...string) bool
	Data(ctx context.Context) map[string]any
}

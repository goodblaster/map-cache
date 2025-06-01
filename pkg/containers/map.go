package containers

import "context"

type Data any
type Map interface {
	Get(ctx context.Context, hierarchy ...string) (Data, error)
	Set(ctx context.Context, value any, hierarchy ...string) error
	Delete(ctx context.Context, hierarchy ...string) error
	ArrayAppend(ctx context.Context, value any, path ...string) error
	ArrayRemove(ctx context.Context, index int, path ...string) error
	ArrayResize(ctx context.Context, newSize int, path ...string) error
	Exists(ctx context.Context, hierarchy ...string) bool
	Data(ctx context.Context) map[string]any
	WildKeys(ctx context.Context, path string) []string
}

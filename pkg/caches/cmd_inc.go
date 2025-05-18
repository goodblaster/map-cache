package caches

import (
	"context"
	"reflect"

	"github.com/goodblaster/errors"
)

type CommandInc struct {
	key string
	val float64
}

func INC(key string, value float64) Command {
	return CommandInc{key: key, val: value}
}

func (p CommandInc) Do(ctx context.Context, cache *Cache) CmdResult {
	v, err := cache.Get(ctx, p.key)
	if err != nil {
		return CmdResult{Error: ErrKeyNotFound.Format(p.key)}
	}

	f64, ok := ToFloat64(v)
	if !ok {
		return CmdResult{Error: errors.New("not a number")}
	}

	f64 += p.val
	return CmdResult{Error: cache.Replace(ctx, p.key, f64)}
}

func ToFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(n).Int()), true
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(n).Uint()), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}

package caches

import (
	"context"
	"reflect"

	"github.com/goodblaster/errors"
)

type CommandInc struct {
	Key   string  `json:"key,required"`
	Value float64 `json:"value,required"`
}

func (CommandInc) Type() string {
	return "INC"
}

func INC(key string, value float64) Command {
	return CommandInc{Key: key, Value: value}
}

func (p CommandInc) Do(ctx context.Context, cache *Cache) CmdResult {
	v, err := cache.Get(ctx, p.Key)
	if err != nil {
		return CmdResult{Error: ErrKeyNotFound.Format(p.Key)}
	}

	f64, ok := ToFloat64(v)
	if !ok {
		return CmdResult{Error: errors.New("not a number")}
	}

	f64 += p.Value
	return CmdResult{Error: cache.Replace(ctx, p.Key, f64)}
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

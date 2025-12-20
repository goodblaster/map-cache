package caches

import (
	"context"
	"math"
	"reflect"
)

type CommandInc struct {
	Key   string  `json:"key,required"`
	Value float64 `json:"value,required"`
}

func (CommandInc) Type() CommandType {
	return CommandTypeInc
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
		return CmdResult{Error: ErrNotANumber}
	}

	f64 += p.Value
	if err := cache.Replace(ctx, p.Key, f64); err != nil {
		return CmdResult{Error: err}
	}
	return CmdResult{Value: f64}
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

func ToInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case uint:
		if uint64(n) <= math.MaxInt64 {
			return int64(n), true
		}
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		if n <= math.MaxInt64 {
			return int64(n), true
		}
	case float32:
		f := float64(n)
		if f == math.Trunc(f) && f >= math.MinInt64 && f <= math.MaxInt64 {
			return int64(f), true
		}
	case float64:
		if n == math.Trunc(n) && n >= math.MinInt64 && n <= math.MaxInt64 {
			return int64(n), true
		}
	}
	return 0, false
}

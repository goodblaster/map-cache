package triggers

import "reflect"

func Same(a, b any) bool {
	ta := reflect.TypeOf(a)
	tb := reflect.TypeOf(b)

	// If types are different, return false
	if ta != tb {
		return false
	}

	// Handle nil explicitly
	if a == nil || b == nil {
		return a == b
	}

	// Handle map[string]any
	if ta.Kind() == reflect.Map && ta.Key().Kind() == reflect.String {
		mapA, okA := a.(map[string]any)
		mapB, okB := b.(map[string]any)
		if !okA || !okB {
			return false
		}
		if len(mapA) != len(mapB) {
			return false
		}
		for k, va := range mapA {
			vb, exists := mapB[k]
			if !exists || !Same(va, vb) {
				return false
			}
		}
		return true
	}

	// For all other comparable types
	va := reflect.ValueOf(a)
	if !va.Type().Comparable() {
		return false
	}

	return a == b
}

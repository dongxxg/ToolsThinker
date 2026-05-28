package _map

import (
	"tools-thinker/support/collection"
)

type _map = map[string]any

func GetNumberPtr[N collection.Number](m _map, key string) *N {
	return collection.ParseNumberPtr[N](m[key])
}

func GetDirectPtr[D collection.Direct](m _map, key string) *D {
	return collection.ParseDirectPtr[D](m[key])
}

func GetNumber[N collection.Number](m _map, key string) N {
	return collection.ParseNumber[N](m[key])
}

func GetDirect[D collection.Direct](m _map, key string) D {
	return collection.ParseDirect[D](m[key])
}

func GetIntPtr(m _map, key string) *int {
	return GetNumberPtr[int](m, key)
}

func GetInt64Ptr(m _map, key string) *int64 {
	return GetNumberPtr[int64](m, key)
}

func GetStrPtr(m _map, key string) *string {
	return GetDirectPtr[string](m, key)
}

func GetSliceStrPtr(m _map, key string) *[]string {
	ss := collection.ParseSliceDirect[string](m[key])
	if ss == nil {
		return nil
	} else {
		return &ss
	}
}

func GetSliceMap(m _map, key string) []map[string]any {
	return collection.ParseSliceDirect[map[string]any](m[key])
}

func GetFloat32Ptr(m _map, key string) *float32 {
	return GetNumberPtr[float32](m, key)
}

func GetFloat64Ptr(m _map, key string) *float64 {
	return GetNumberPtr[float64](m, key)
}

func GetBoolPtr(m _map, key string) *bool {
	return GetDirectPtr[bool](m, key)
}

func GetInt(m _map, key string) int {
	return GetNumber[int](m, key)
}

func GetInt64(m _map, key string) int64 {
	return GetNumber[int64](m, key)
}

func GetString(m _map, key string) string {
	return GetDirect[string](m, key)
}

func GetBool(m _map, key string) bool {
	return GetDirect[bool](m, key)
}

func GetMap(m _map, key string) map[string]any {
	return GetDirect[map[string]any](m, key)
}

func SetPtr[T any](m _map, key string, value *T) {
	if value == nil {
		m[key] = nil
	} else {
		m[key] = *value
	}
}

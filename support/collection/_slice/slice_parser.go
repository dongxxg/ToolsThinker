package _slice

import "tools-thinker/support/collection"

func GetIntPtr(s []any, idx int) *int {
	return collection.ParseNumberPtr[int](s[idx])
}

func GetSliceStr(s []any, idx int) []string {
	return collection.ParseSliceDirect[string](s[idx])
}

func GetSliceMap(s []any, idx int) []map[string]any {
	return collection.ParseSliceDirect[map[string]any](s[idx])
}

func GetInt(s []any, idx int) int {
	return collection.ParseNumber[int](s[idx])
}

func GetInt64(s []any, idx int) int64 {
	return collection.ParseNumber[int64](s[idx])
}

func GetString(s []any, idx int) string {
	return collection.ParseDirect[string](s[idx])
}

func GetBool(s []any, idx int) bool {
	return collection.ParseDirect[bool](s[idx])
}

func GetMap(s []any, idx int) map[string]any {
	return collection.ParseDirect[map[string]any](s[idx])
}

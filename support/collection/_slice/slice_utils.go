package _slice

import "tools-thinker/support/collection"

// 相等
func Equals[T collection.Equal](a []T, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func HasPrefix[T collection.Equal](a []T, b []T) bool {
	if len(a) < len(b) {
		return false
	}
	for i, vb := range b {
		if a[i] != vb {
			return false
		}
	}
	return true
}

func IndexOf[T collection.Equal](s []T, e T) int {
	for i, str := range s {
		if str == e {
			return i
		}
	}
	return -1
}

// 包含
func Contains[T collection.Equal](s []T, e T) bool {
	return IndexOf(s, e) > -1
}

// 将字符slice按每段多少进行分段
func Partition[T any](original []T, pieceSize int) (res [][]T) {
	if pieceSize <= 0 {
		res = append(res, original)
		return res
	}
	counts := (len(original) + pieceSize - 1) / pieceSize
	for i := 0; i < counts; i++ {
		if i == counts-1 {
			res = append(res, original[i*pieceSize:])
		} else {
			res = append(res, original[i*pieceSize:(i+1)*pieceSize])
		}
	}
	return res
}

// 集合删除
// @set 元素唯一
func SetDelete[T collection.Equal](set []T, e T) []T {
	if i := IndexOf(set, e); i >= 0 {
		set = append(set[:i], set[i+1:]...)
	}
	return set
}

// 集合添加
// @set 元素唯一
func SetAdd[T collection.Equal](set []T, e T) []T {
	if i := IndexOf(set, e); i < 0 {
		set = append(set, e)
	}
	return set
}

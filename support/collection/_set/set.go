package _set

import (
	"fmt"
	"tools-thinker/support/collection"
)

var (
	keyExists = struct{}{}
)

const maxInt = int(^uint(0) >> 1)

// 字符串set
type Set[T collection.Equal] struct {
	m map[T]struct{}
}

// 创建set
func New[T collection.Equal](ts ...T) *Set[T] {
	s := NewWithSize[T](len(ts))
	s.Add(ts...)
	return s
}

// 使用slice创建set
func NewBySlice[T collection.Equal](ts []T) *Set[T] {
	s := NewWithSize[T](len(ts))
	for _, s2 := range ts {
		s.Add(s2)
	}
	return s
}

func NewWithSize[T collection.Equal](size int) *Set[T] {
	return &Set[T]{make(map[T]struct{}, size)}
}

func (s *Set[T]) InitWithSize(size int) {
	s.m = make(map[T]struct{}, size)
}

func (s *Set[T]) Add(items ...T) {
	for _, item := range items {
		s.m[item] = keyExists
	}
}

func (s *Set[T]) Remove(items ...T) {
	for _, item := range items {
		delete(s.m, item)
	}
}

func (s *Set[T]) Has(items ...T) bool {
	has := false
	for _, item := range items {
		if _, has = s.m[item]; !has {
			break
		}
	}
	return has
}

func (s *Set[T]) HasAny(items ...T) bool {
	has := false
	for _, item := range items {
		if _, has = s.m[item]; has {
			break
		}
	}
	return has
}

func (s *Set[T]) Size() int {
	return len(s.m)
}

func (s *Set[T]) Clear() {
	s.m = make(map[T]struct{})
}

func (s *Set[T]) IsEmpty() bool {
	return s.Size() == 0
}

func (s *Set[T]) IsEqual(t *Set[T]) bool {
	// return false if they are no the same size
	if s.Size() != t.Size() {
		return false
	}

	equal := true
	t.Each(func(item T) bool {
		_, equal = s.m[item]
		return equal // if false, Each() will end
	})

	return equal
}

func (s *Set[T]) IsSubset(t *Set[T]) bool {
	if s.Size() < t.Size() {
		return false
	}

	subset := true

	t.Each(func(item T) bool {
		_, subset = s.m[item]
		return subset
	})

	return subset
}

func (s *Set[T]) IsSuperset(t *Set[T]) bool {
	return t.IsSubset(s)
}

func (s *Set[T]) Each(f func(item T) bool) {
	for item := range s.m {
		if !f(item) {
			break
		}
	}
}

// 任意获取一个
func (s *Set[T]) GetOne() (t T, ok bool) {
	for item := range s.m {
		return item, true
	}
	var zero T
	return zero, false
}

func (s *Set[T]) Copy() *Set[T] {
	u := NewWithSize[T](s.Size())
	for item := range s.m {
		u.m[item] = keyExists
	}
	return u
}

// String returns a string representation of s
func (s *Set[T]) String() string {
	sl := s.Slice()
	return fmt.Sprintf("%v", sl)
}

func (s *Set[T]) Slice() []T {
	v := make([]T, 0, s.Size())
	for item := range s.m {
		v = append(v, item)
	}
	return v
}

func (s *Set[T]) Merge(t *Set[T]) {
	for item := range t.m {
		s.m[item] = keyExists
	}
}

// Separate removes the Set items containing in t from Set s. Please aware that
// it's not the opposite of Merge.
func (s *Set[T]) Separate(t *Set[T]) {
	for item := range t.m {
		delete(s.m, item)
	}
}

func Union[T collection.Equal](sets ...*Set[T]) *Set[T] {
	maxPos := -1
	maxSize := 0
	for i, set := range sets {
		if l := set.Size(); l > maxSize {
			maxSize = l
			maxPos = i
		}
	}
	if maxSize == 0 {
		return New[T]()
	}

	u := sets[maxPos].Copy()
	for i, set := range sets {
		if i == maxPos {
			continue
		}
		for item := range set.m {
			u.m[item] = keyExists
		}
	}
	return u
}

func Difference[T collection.Equal](set1 *Set[T], sets ...*Set[T]) *Set[T] {
	s := set1.Copy()
	for _, set := range sets {
		s.Separate(set)
	}
	return s
}

func Intersection[T collection.Equal](sets ...*Set[T]) *Set[T] {
	minPos := -1
	minSize := maxInt
	for i, set := range sets {
		if l := set.Size(); l < minSize {
			minSize = l
			minPos = i
		}
	}
	if minSize == maxInt || minSize == 0 {
		return New[T]()
	}

	t := sets[minPos].Copy()
	for i, set := range sets {
		if i == minPos {
			continue
		}
		for item := range t.m {
			if _, has := set.m[item]; !has {
				delete(t.m, item)
			}
		}
	}
	return t
}

func SymmetricDifference[T collection.Equal](s *Set[T], t *Set[T]) *Set[T] {
	u := Difference[T](s, t)
	v := Difference[T](t, s)
	return Union[T](u, v)
}

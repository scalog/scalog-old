package golib

/**
Copied from https://gist.github.com/bgadrian/cb8b9344d9c66571ef331a14eb7a2e80.
*/

type Set32 struct {
	list map[int32]struct{} //empty structs occupy 0 memory
}

func (s *Set32) Contains(v int32) bool {
	_, ok := s.list[v]
	return ok
}

func (s *Set32) Add(v int32) {
	s.list[v] = struct{}{}
}

func (s *Set32) Remove(v int32) {
	delete(s.list, v)
}

func (s *Set32) Clear() {
	s.list = make(map[int32]struct{})
}

func (s *Set32) Size() int {
	return len(s.list)
}

//for iteration
func (s *Set32) Iterable() map[int32]struct{} {
	return s.list
}

func NewSet32(list ...int32) *Set32 {
	s := &Set32{}
	s.list = make(map[int32]struct{})
	for _, v := range list {
		s.Add(v)
	}
	return s
}

//optional functionalities

type FilterFunc32 func(v int32) bool

// Filter returns a subset, that contains only the values that satisfies the given predicate P
func (s *Set32) Filter(P FilterFunc32) *Set32 {
	res := NewSet32()
	for v := range s.list {
		if !P(v) {
			continue
		}
		res.Add(v)
	}
	return res
}

func (s *Set32) Union(s2 *Set32) *Set32 {
	res := NewSet32()
	for v := range s.list {
		res.Add(v)
	}

	for v := range s2.list {
		res.Add(v)
	}
	return res
}

func (s *Set32) Intersect(s2 *Set32) *Set32 {
	res := NewSet32()
	for v := range s.list {
		if !s2.Contains(v) {
			continue
		}
		res.Add(v)
	}
	return res
}

// Difference returns the subset from s, that doesn't exists in s2 (param)
func (s *Set32) Difference(s2 *Set32) *Set32 {
	res := NewSet32()
	for v := range s.list {
		if s2.Contains(v) {
			continue
		}
		res.Add(v)
	}
	return res
}

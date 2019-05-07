package set64

/**
Copied from https://gist.github.com/bgadrian/cb8b9344d9c66571ef331a14eb7a2e80.
*/

type Set64 struct {
	list map[int64]struct{} //empty structs occupy 0 memory
}

func (s *Set64) Contains(v int64) bool {
	_, ok := s.list[v]
	return ok
}

func (s *Set64) Add(v int64) {
	s.list[v] = struct{}{}
}

func (s *Set64) Remove(v int64) {
	delete(s.list, v)
}

func (s *Set64) Clear() {
	s.list = make(map[int64]struct{})
}

func (s *Set64) Size() int {
	return len(s.list)
}

// Iterable iterates the set.
func (s *Set64) Iterable() map[int64]struct{} {
	return s.list
}

func NewSet64(list ...int64) *Set64 {
	s := &Set64{}
	s.list = make(map[int64]struct{})
	for _, v := range list {
		s.Add(v)
	}
	return s
}

//
// optional functionalities
//

type FilterFunc64 func(v int64) bool

// Filter returns a subset, that contains only the values that satisfies the given predicate P
func (s *Set64) Filter(P FilterFunc64) *Set64 {
	res := NewSet64()
	for v := range s.list {
		if !P(v) {
			continue
		}
		res.Add(v)
	}
	return res
}

func (s *Set64) Union(s2 *Set64) *Set64 {
	res := NewSet64()
	for v := range s.list {
		res.Add(v)
	}

	for v := range s2.list {
		res.Add(v)
	}
	return res
}

func (s *Set64) Intersect(s2 *Set64) *Set64 {
	res := NewSet64()
	for v := range s.list {
		if !s2.Contains(v) {
			continue
		}
		res.Add(v)
	}
	return res
}

// Difference returns the subset from s, that doesn't exists in s2 (param)
func (s *Set64) Difference(s2 *Set64) *Set64 {
	res := NewSet64()
	for v := range s.list {
		if s2.Contains(v) {
			continue
		}
		res.Add(v)
	}
	return res
}

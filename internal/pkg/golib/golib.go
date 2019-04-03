package golib

/*
SliceEq is true if [a] and [b] contain the exact same elements in
the exact same order
*/
func SliceEq(a, b []int32) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

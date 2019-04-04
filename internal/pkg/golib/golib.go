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
func Max(x int, y int) int {
	if x < y {
		return y
	}
	return x
}
func Min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
func IntSliceToInt32Slice(intSlice []int) []int32 {
	int32Slice := make([]int32, len(intSlice), len(intSlice))
	for idx, element := range intSlice {
		int32Slice[idx] = int32(element)
	}
	return int32Slice
}

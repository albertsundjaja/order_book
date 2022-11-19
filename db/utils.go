package db

import (
	"sort"
)

// SortedContainsInt32 use binary search to search for a value in the slice
// returns the index of the item or -1 if not found
func SortedContainsInt32(ascending bool, slice []int32, a int32) int {
	// edge case when slice is empty
	if len(slice) == 0 {
		return -1
	}

	// https://pkg.go.dev/sort#Search
	var i int
	if ascending {
		i = sort.Search(len(slice), func(i int) bool { return slice[i] >= a })
	} else {
		i = sort.Search(len(slice), func(i int) bool { return slice[i] <= a })
	}
	if i < len(slice) && slice[i] == a {
		return i
	}
	return -1
}

// min returns the minimum of two numbers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package inmem_db

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

// InsertionSort sort the slice using insertion sort algorithm
// this is useful as the complexity is roughly O(n) for almost sorted array
func InsertiontSortInt32(slice []int32, ascending bool) {
	var x = len(slice)
	for n := 1; n < x; n++ {
		v := n
		for v > 0 {
			if ascending {
				if slice[v-1] > slice[v] {
					slice[v-1], slice[v] = slice[v], slice[v-1]
				}
			} else {
				if slice[v-1] < slice[v] {
					slice[v-1], slice[v] = slice[v], slice[v-1]
				}
			}
			v = v - 1
		}
	}
}

// min returns the minimum of two numbers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

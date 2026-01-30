package util

import "math"

// SampleIndices returns deterministic, evenly spaced indices across [0,n-1].
func SampleIndices(n, sample int) []int {
	if n <= 0 || sample <= 0 {
		return []int{}
	}
	if sample >= n {
		idx := make([]int, n)
		for i := 0; i < n; i++ {
			idx[i] = i
		}
		return idx
	}
	if sample == 1 {
		return []int{0}
	}
	indices := make([]int, 0, sample)
	step := float64(n-1) / float64(sample-1)
	last := -1
	for i := 0; i < sample; i++ {
		idx := int(math.Round(float64(i) * step))
		if idx <= last {
			idx = last + 1
		}
		if idx >= n {
			idx = n - 1
		}
		indices = append(indices, idx)
		last = idx
	}
	return indices
}

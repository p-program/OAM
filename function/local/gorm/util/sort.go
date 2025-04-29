package util

import "sort"

// SortInts 对整数切片进行排序，并返回一个新的排序后的切片
// 原始切片不会被修改
// 参数:
//   - nums: 要排序的整数切片
//   - ascending: 排序方向，true 表示升序，false 表示降序
//
// 返回:
//   - 排序后的新切片
//
// 升序排序
// ascendingOrder := SortInts(numbers, true)
// 降序排序
// descendingOrder := SortInts(numbers, false)
func SortInts(nums []int, ascending bool) []int {
	// 创建新切片以避免修改原始切片
	result := make([]int, len(nums))
	copy(result, nums)

	if ascending {
		// 升序排序
		sort.Slice(result, func(i, j int) bool {
			return result[i] < result[j]
		})
	} else {
		// 降序排序
		sort.Slice(result, func(i, j int) bool {
			return result[i] > result[j]
		})
	}

	return result
}

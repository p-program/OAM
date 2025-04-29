package set

import (
	"fmt"
	"sort"
	"strings"
)

// 泛型类型 SortedMap，用于支持任何类型的键和值
type SortedMap[K comparable, V any] struct {
	Original   map[K]V
	SortedKeys []K
}

// 可排序的类型接口
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

// NewSortedMap 是一个泛型函数，创建一个排序后的字典
func NewSortedMap[K comparable, V any](original map[K]V) *SortedMap[K, V] {
	// 提取 map 的所有键
	keys := make([]K, 0, len(original))
	for key := range original {
		keys = append(keys, key)
	}

	// 如果map为空，直接返回
	if len(keys) == 0 {
		return &SortedMap[K, V]{
			Original:   original,
			SortedKeys: keys,
		}
	}

	// 对键进行排序，根据其具体类型选择不同的排序方法
	sortKeys(keys)

	// 返回新的 SortedMap
	return &SortedMap[K, V]{
		Original:   original,
		SortedKeys: keys,
	}
}

// 根据不同类型执行排序
func sortKeys[K comparable](keys []K) {
	// 尝试使用类型参数的约束来识别具体类型
	// 采用类型转换的方式，而不是类型断言
	var emptyKey K

	// 检查是否可以转换为string
	if _, ok := any(emptyKey).(string); ok {
		sort.Slice(keys, func(i, j int) bool {
			return any(keys[i]).(string) < any(keys[j]).(string)
		})
		return
	}

	// 检查是否可以转换为int
	if _, ok := any(emptyKey).(int); ok {
		sort.Slice(keys, func(i, j int) bool {
			return any(keys[i]).(int) < any(keys[j]).(int)
		})
		return
	}

	// 检查是否可以转换为float64
	if _, ok := any(emptyKey).(float64); ok {
		sort.Slice(keys, func(i, j int) bool {
			return any(keys[i]).(float64) < any(keys[j]).(float64)
		})
		return
	}

	fmt.Println("未支持的类型，按默认顺序排序")
}

// 更通用的解决方案，使用泛型约束
func NewOrderedSortedMap[K Ordered, V any](original map[K]V) *SortedMap[K, V] {
	keys := make([]K, 0, len(original))
	for key := range original {
		keys = append(keys, key)
	}

	// 直接使用类型约束 Ordered 进行比较
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return &SortedMap[K, V]{
		Original:   original,
		SortedKeys: keys,
	}
}

// Print 按顺序输出 map
func (sm *SortedMap[K, V]) Print() {
	for _, key := range sm.SortedKeys {
		fmt.Printf("%v: %v\n", key, sm.Original[key])
	}
}

// String 返回排序后map的字符串表示
func (sm *SortedMap[K, V]) String() string {
	var sb strings.Builder
	sb.WriteString("{\n")
	for _, key := range sm.SortedKeys {
		sb.WriteString(fmt.Sprintf("  %v: %v\n", key, sm.Original[key]))
	}
	sb.WriteString("}")
	return sb.String()
}

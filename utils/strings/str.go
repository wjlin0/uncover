package strings

import "reflect"

// ContainsAll 检查切片 a 是否包含切片 b 中的所有元素
func ContainsAll(a, b interface{}) bool {
	// 使用反射获取切片 a 的值
	valA := reflect.ValueOf(a)
	// 使用反射获取切片 b 的值
	valB := reflect.ValueOf(b)

	// 如果不是切片类型，则直接返回 false
	if valA.Kind() != reflect.Slice || valB.Kind() != reflect.Slice {
		return false
	}

	// 创建一个 map 用于记录切片 a 中的元素
	elementsMap := make(map[interface{}]bool)

	// 将切片 a 中的元素添加到 map 中
	for i := 0; i < valA.Len(); i++ {
		elementsMap[valA.Index(i).Interface()] = true
	}

	// 遍历切片 b 中的元素，检查是否都在 map 中存在
	for i := 0; i < valB.Len(); i++ {
		if !elementsMap[valB.Index(i).Interface()] {
			return false
		}
	}

	// 如果所有元素都在 map 中找到，则返回 true
	return true
}

// Contains 判断切片中是否包含指定字符串
func Contains(slice []string, target string) bool {
	for i := 0; i < len(slice); i++ {
		if slice[i] == target {
			return true
		}
	}
	return false
}

// AllStringsInMap checks if all strings in the array are present in the given string set.
func AllStringsInMap(strings []string, stringSet map[string]struct{}) bool {
	for _, s := range strings {
		if _, exists := stringSet[s]; !exists {
			return false
		}
	}
	return true
}

// FindCommonStrings 找到两个字符串切片中的公共字符串
func FindCommonStrings(arr1, arr2 []string) []string {
	stringMap := make(map[string]struct{})
	var commonStrings []string

	// 将第一个数组中的字符串添加到 map 中
	for _, str := range arr1 {
		stringMap[str] = struct{}{}
	}

	// 检查第二个数组中的字符串是否在 map 中
	for _, str := range arr2 {
		if _, ok := stringMap[str]; ok {
			commonStrings = append(commonStrings, str)
		}
	}

	return commonStrings
}

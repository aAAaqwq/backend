package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// 获取当前时间
func GetCurrentTime() time.Time {
	return time.Now()
}
// IsEmpty 检查值是否为空
func IsEmpty(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return true
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.String:
		return rv.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() == 0
	default:
		return false
	}
}

// IsNil 检查值是否为 nil
func IsNil(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		return rv.IsNil()
	default:
		return false
	}
}

// IsZero 检查值是否为零值
func IsZero(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return true
		}
		rv = rv.Elem()
	}

	return rv.IsZero()
}

// TypeOf 获取值的类型名称
func TypeOf(v interface{}) string {
	if v == nil {
		return "nil"
	}
	return reflect.TypeOf(v).String()
}

// KindOf 获取值的 Kind
func KindOf(v interface{}) reflect.Kind {
	if v == nil {
		return reflect.Invalid
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return reflect.Invalid
		}
		return rv.Elem().Kind()
	}
	return rv.Kind()
}

// DeepEqual 深度比较两个值是否相等
func DeepEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// Clone 深度克隆一个值（使用 JSON 序列化/反序列化）
func Clone(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	// 使用 JSON 序列化/反序列化进行深度克隆
	jsonBytes, err := MarshalJSON(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	// 创建相同类型的新实例
	rt := reflect.TypeOf(v)

	// 如果是指针，获取指向的类型
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		newPtr := reflect.New(rt)
		if err := json.Unmarshal(jsonBytes, newPtr.Interface()); err != nil {
			return nil, fmt.Errorf("failed to unmarshal: %w", err)
		}
		return newPtr.Interface(), nil
	}

	// 非指针类型
	newVal := reflect.New(rt)
	if err := json.Unmarshal(jsonBytes, newVal.Interface()); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}
	return newVal.Elem().Interface(), nil
}

// TruncateString 截断字符串
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 0 {
		return ""
	}
	return s[:maxLen]
}

// TruncateStringWithEllipsis 截断字符串并添加省略号
func TruncateStringWithEllipsis(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}

// PadString 填充字符串（左侧或右侧）
func PadString(s string, length int, padChar rune, padLeft bool) string {
	if len(s) >= length {
		return s
	}
	padding := strings.Repeat(string(padChar), length-len(s))
	if padLeft {
		return padding + s
	}
	return s + padding
}

// PadLeft 左侧填充字符串
func PadLeft(s string, length int, padChar rune) string {
	return PadString(s, length, padChar, true)
}

// PadRight 右侧填充字符串
func PadRight(s string, length int, padChar rune) string {
	return PadString(s, length, padChar, false)
}

// ReverseString 反转字符串
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ContainsString 检查字符串切片是否包含指定字符串
func ContainsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// ContainsInt 检查整数切片是否包含指定整数
func ContainsInt(slice []int, v int) bool {
	for _, val := range slice {
		if val == v {
			return true
		}
	}
	return false
}

// ContainsInt64 检查 int64 切片是否包含指定值
func ContainsInt64(slice []int64, v int64) bool {
	for _, val := range slice {
		if val == v {
			return true
		}
	}
	return false
}

// UniqueString 去除字符串切片中的重复项
func UniqueString(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// UniqueInt 去除整数切片中的重复项
func UniqueInt(slice []int) []int {
	seen := make(map[int]bool)
	result := make([]int, 0, len(slice))
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// UniqueInt64 去除 int64 切片中的重复项
func UniqueInt64(slice []int64) []int64 {
	seen := make(map[int64]bool)
	result := make([]int64, 0, len(slice))
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// FilterString 过滤字符串切片
func FilterString(slice []string, fn func(string) bool) []string {
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// FilterInt 过滤整数切片
func FilterInt(slice []int, fn func(int) bool) []int {
	result := make([]int, 0, len(slice))
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// MapString 映射字符串切片
func MapString(slice []string, fn func(string) string) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// MapInt 映射整数切片
func MapInt(slice []int, fn func(int) int) []int {
	result := make([]int, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// SumInt 计算整数切片的总和
func SumInt(slice []int) int {
	sum := 0
	for _, v := range slice {
		sum += v
	}
	return sum
}

// SumInt64 计算 int64 切片的总和
func SumInt64(slice []int64) int64 {
	var sum int64
	for _, v := range slice {
		sum += v
	}
	return sum
}

// SumFloat64 计算 float64 切片的总和
func SumFloat64(slice []float64) float64 {
	var sum float64
	for _, v := range slice {
		sum += v
	}
	return sum
}

// MaxInt 返回整数切片中的最大值
func MaxInt(slice []int) (int, error) {
	if len(slice) == 0 {
		return 0, fmt.Errorf("slice is empty")
	}
	max := slice[0]
	for _, v := range slice[1:] {
		if v > max {
			max = v
		}
	}
	return max, nil
}

// MinInt 返回整数切片中的最小值
func MinInt(slice []int) (int, error) {
	if len(slice) == 0 {
		return 0, fmt.Errorf("slice is empty")
	}
	min := slice[0]
	for _, v := range slice[1:] {
		if v < min {
			min = v
		}
	}
	return min, nil
}

// MaxInt64 返回 int64 切片中的最大值
func MaxInt64(slice []int64) (int64, error) {
	if len(slice) == 0 {
		return 0, fmt.Errorf("slice is empty")
	}
	max := slice[0]
	for _, v := range slice[1:] {
		if v > max {
			max = v
		}
	}
	return max, nil
}

// MinInt64 返回 int64 切片中的最小值
func MinInt64(slice []int64) (int64, error) {
	if len(slice) == 0 {
		return 0, fmt.Errorf("slice is empty")
	}
	min := slice[0]
	for _, v := range slice[1:] {
		if v < min {
			min = v
		}
	}
	return min, nil
}

package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// ParseString 将任意类型解析为字符串
func ParseString(v interface{}) string {
	return ConvertToString(v)
}

// ParseInt 将任意类型解析为 int
func ParseInt(v interface{}) (int, error) {
	if v == nil {
		return 0, fmt.Errorf("value is nil")
	}

	switch val := v.(type) {
	case int:
		return val, nil
	case int8:
		return int(val), nil
	case int16:
		return int(val), nil
	case int32:
		return int(val), nil
	case int64:
		return int(val), nil
	case uint:
		return int(val), nil
	case uint8:
		return int(val), nil
	case uint16:
		return int(val), nil
	case uint32:
		return int(val), nil
	case uint64:
		return int(val), nil
	case float32:
		return int(val), nil
	case float64:
		return int(val), nil
	case string:
		return StringToInt(val)
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot parse %T to int", v)
	}
}

// ParseInt64 将任意类型解析为 int64
func ParseInt64(v interface{}) (int64, error) {
	return ConvertToInt64(v)
}

// ParseFloat64 将任意类型解析为 float64
func ParseFloat64(v interface{}) (float64, error) {
	return ConvertToFloat64(v)
}

// ParseBool 将任意类型解析为 bool
func ParseBool(v interface{}) (bool, error) {
	return ConvertToBool(v)
}

// ParseTime 将任意类型解析为 time.Time
func ParseTime(v interface{}) (time.Time, error) {
	return ConvertToTime(v)
}

// ParseDuration 将任意类型解析为 time.Duration
func ParseDuration(v interface{}) (time.Duration, error) {
	if v == nil {
		return 0, fmt.Errorf("value is nil")
	}

	switch val := v.(type) {
	case time.Duration:
		return val, nil
	case string:
		return time.ParseDuration(val)
	case int64:
		return time.Duration(val), nil
	case int:
		return time.Duration(val), nil
	case float64:
		return time.Duration(val), nil
	default:
		return 0, fmt.Errorf("cannot parse %T to time.Duration", v)
	}
}

// ParseJSON 将 JSON 字符串解析为指定类型
func ParseJSON(jsonStr string, v interface{}) error {
	if jsonStr == "" {
		return fmt.Errorf("json string is empty")
	}
	return json.Unmarshal([]byte(jsonStr), v)
}

// ParseJSONBytes 将 JSON 字节解析为指定类型
func ParseJSONBytes(jsonBytes []byte, v interface{}) error {
	if len(jsonBytes) == 0 {
		return fmt.Errorf("json bytes is empty")
	}
	return json.Unmarshal(jsonBytes, v)
}

// ParseStringSlice 将字符串切片解析为指定类型的切片
func ParseStringSlice(strSlice []string, target interface{}) error {
	// 将字符串切片转换为 JSON，然后解析
	jsonBytes, err := json.Marshal(strSlice)
	if err != nil {
		return fmt.Errorf("failed to marshal string slice: %w", err)
	}
	return json.Unmarshal(jsonBytes, target)
}

// ParseIntSlice 将字符串切片解析为 int 切片
func ParseIntSlice(strSlice []string) ([]int, error) {
	result := make([]int, 0, len(strSlice))
	for _, s := range strSlice {
		val, err := StringToInt(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s to int: %w", s, err)
		}
		result = append(result, val)
	}
	return result, nil
}

// ParseInt64Slice 将字符串切片解析为 int64 切片
func ParseInt64Slice(strSlice []string) ([]int64, error) {
	result := make([]int64, 0, len(strSlice))
	for _, s := range strSlice {
		val, err := StringToInt64(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s to int64: %w", s, err)
		}
		result = append(result, val)
	}
	return result, nil
}

// ParseFloat64Slice 将字符串切片解析为 float64 切片
func ParseFloat64Slice(strSlice []string) ([]float64, error) {
	result := make([]float64, 0, len(strSlice))
	for _, s := range strSlice {
		val, err := StringToFloat64(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s to float64: %w", s, err)
		}
		result = append(result, val)
	}
	return result, nil
}

// ParseBoolSlice 将字符串切片解析为 bool 切片
func ParseBoolSlice(strSlice []string) ([]bool, error) {
	result := make([]bool, 0, len(strSlice))
	for _, s := range strSlice {
		val, err := StringToBool(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s to bool: %w", s, err)
		}
		result = append(result, val)
	}
	return result, nil
}

// StringToIntSlice 将逗号分隔的字符串解析为 int 切片
func StringToIntSlice(s, sep string) ([]int, error) {
	if s == "" {
		return []int{}, nil
	}
	strSlice := SplitString(s, sep)
	return ParseIntSlice(strSlice)
}

// StringToInt64Slice 将逗号分隔的字符串解析为 int64 切片
func StringToInt64Slice(s, sep string) ([]int64, error) {
	if s == "" {
		return []int64{}, nil
	}
	strSlice := SplitString(s, sep)
	return ParseInt64Slice(strSlice)
}

// SplitString 分割字符串（支持多个分隔符）
func SplitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	if sep == "" {
		return []string{s}
	}
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			if start < i {
				result = append(result, s[start:i])
			}
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

// DefaultString 如果字符串为空，返回默认值
func DefaultString(s, defaultValue string) string {
	if s == "" {
		return defaultValue
	}
	return s
}

// DefaultInt 如果值为 0，返回默认值
func DefaultInt(v, defaultValue int) int {
	if v == 0 {
		return defaultValue
	}
	return v
}

// DefaultInt64 如果值为 0，返回默认值
func DefaultInt64(v, defaultValue int64) int64 {
	if v == 0 {
		return defaultValue
	}
	return v
}

// DefaultFloat64 如果值为 0，返回默认值
func DefaultFloat64(v, defaultValue float64) float64 {
	if v == 0 {
		return defaultValue
	}
	return v
}

// DefaultBool 如果值为 false，返回默认值
func DefaultBool(v, defaultValue bool) bool {
	if !v {
		return defaultValue
	}
	return v
}

// Coalesce 返回第一个非零值
func Coalesce(values ...interface{}) interface{} {
	for _, v := range values {
		if v != nil {
			// 检查是否为字符串且非空
			if str, ok := v.(string); ok {
				if str != "" {
					return str
				}
				continue
			}
			// 检查是否为数字且非零
			if num, ok := v.(int); ok {
				if num != 0 {
					return num
				}
				continue
			}
			if num, ok := v.(int64); ok {
				if num != 0 {
					return num
				}
				continue
			}
			if num, ok := v.(float64); ok {
				if num != 0 {
					return num
				}
				continue
			}
			// 其他类型，非 nil 即返回
			return v
		}
	}
	return nil
}

// CoalesceString 返回第一个非空字符串
func CoalesceString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// CoalesceInt 返回第一个非零整数
func CoalesceInt(values ...int) int {
	for _, v := range values {
		if v != 0 {
			return v
		}
	}
	return 0
}

// CoalesceInt64 返回第一个非零 int64
func CoalesceInt64(values ...int64) int64 {
	for _, v := range values {
		if v != 0 {
			return v
		}
	}
	return 0
}

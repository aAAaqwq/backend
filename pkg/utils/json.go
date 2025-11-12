package utils

import (
	"encoding/json"
	"fmt"
)

// MarshalJSON 将任意类型转换为 JSON 字节
func MarshalJSON(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	return json.Marshal(v)
}

// MarshalJSONIndent 将任意类型转换为格式化的 JSON 字节
func MarshalJSONIndent(v interface{}, prefix, indent string) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	return json.MarshalIndent(v, prefix, indent)
}

// UnmarshalJSON 将 JSON 字节解析为指定类型
func UnmarshalJSON(data []byte, v interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("data is empty")
	}
	return json.Unmarshal(data, v)
}

// MapToJSON 将 map 转换为 JSON 字符串
func MapToJSON(m map[string]interface{}) (string, error) {
	if m == nil {
		return "null", nil
	}
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal map to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// MapToJSONBytes 将 map 转换为 JSON 字节
func MapToJSONBytes(m map[string]interface{}) ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return json.Marshal(m)
}

// JSONToMap 将 JSON 字符串解析为 map
func JSONToMap(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" {
		return nil, fmt.Errorf("json string is empty")
	}
	return JSONBytesToMap([]byte(jsonStr))
}

// JSONBytesToMap 将 JSON 字节解析为 map
func JSONBytesToMap(jsonBytes []byte) (map[string]interface{}, error) {
	if len(jsonBytes) == 0 {
		return nil, fmt.Errorf("json bytes is empty")
	}
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}
	return result, nil
}

// JSONToSlice 将 JSON 字符串解析为切片
func JSONToSlice(jsonStr string) ([]interface{}, error) {
	if jsonStr == "" {
		return nil, fmt.Errorf("json string is empty")
	}
	return JSONBytesToSlice([]byte(jsonStr))
}

// JSONBytesToSlice 将 JSON 字节解析为切片
func JSONBytesToSlice(jsonBytes []byte) ([]interface{}, error) {
	if len(jsonBytes) == 0 {
		return nil, fmt.Errorf("json bytes is empty")
	}
	var result []interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to slice: %w", err)
	}
	return result, nil
}

// PrettyJSON 将任意类型转换为格式化的 JSON 字符串
func PrettyJSON(v interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to pretty JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// CompactJSON 将 JSON 字符串压缩（移除空格和换行）
func CompactJSON(jsonStr string) (string, error) {
	if jsonStr == "" {
		return "", nil
	}
	var v interface{}
	if err := json.Unmarshal([]byte(jsonStr), &v); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// IsValidJSON 检查字符串是否为有效的 JSON
func IsValidJSON(jsonStr string) bool {
	if jsonStr == "" {
		return false
	}
	var v interface{}
	return json.Unmarshal([]byte(jsonStr), &v) == nil
}

// IsValidJSONBytes 检查字节是否为有效的 JSON
func IsValidJSONBytes(jsonBytes []byte) bool {
	if len(jsonBytes) == 0 {
		return false
	}
	var v interface{}
	return json.Unmarshal(jsonBytes, &v) == nil
}

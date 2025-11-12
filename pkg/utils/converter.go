package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// ConvertToMap 将任意类型转换为 map[string]interface{}
func ConvertToMap(v interface{}) (map[string]interface{}, error) {
	if v == nil {
		return nil, fmt.Errorf("value is nil")
	}

	// 如果已经是 map 类型
	if m, ok := v.(map[string]interface{}); ok {
		return m, nil
	}

	// 使用反射获取类型
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)

	// 如果是指针，获取指向的值
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, fmt.Errorf("pointer is nil")
		}
		rv = rv.Elem()
		rt = rt.Elem()
	}

	// 如果是 map 类型，直接转换
	if rv.Kind() == reflect.Map {
		result := make(map[string]interface{})
		for _, key := range rv.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			result[keyStr] = rv.MapIndex(key).Interface()
		}
		return result, nil
	}

	// 如果是结构体，转换为 map
	if rv.Kind() == reflect.Struct {
		result := make(map[string]interface{})
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			fieldValue := rv.Field(i)

			// 跳过未导出的字段
			if !fieldValue.CanInterface() {
				continue
			}

			// 获取字段名（优先使用 json tag）
			fieldName := field.Name
			if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
				// 处理 json tag，如 "name,omitempty"
				if idx := len(jsonTag); idx > 0 {
					for i, c := range jsonTag {
						if c == ',' {
							idx = i
							break
						}
					}
					fieldName = jsonTag[:idx]
				}
			}

			// 处理指针字段
			if fieldValue.Kind() == reflect.Ptr {
				if fieldValue.IsNil() {
					result[fieldName] = nil
				} else {
					result[fieldName] = fieldValue.Elem().Interface()
				}
			} else {
				result[fieldName] = fieldValue.Interface()
			}
		}
		return result, nil
	}

	// 其他类型，使用 JSON 序列化/反序列化
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal from JSON: %w", err)
	}

	return result, nil
}

// ConvertToString 将任意类型转换为字符串
func ConvertToString(v interface{}) string {
	if v == nil {
		return ""
	}

	// 如果已经是字符串
	if str, ok := v.(string); ok {
		return str
	}

	// 使用反射获取类型
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.String:
		return rv.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	case reflect.Slice, reflect.Array:
		// 如果是字节切片，转换为字符串
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return string(rv.Bytes())
		}
		// 其他切片/数组，使用 JSON 序列化
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(jsonBytes)
	case reflect.Map, reflect.Struct:
		// 使用 JSON 序列化
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(jsonBytes)
	case reflect.Interface:
		return ConvertToString(rv.Interface())
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Int64ToString 将 int64 转换为字符串
func Int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}

// StringToInt64 将字符串转换为 int64
func StringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// StringToInt 将字符串转换为 int
func StringToInt(s string) (int, error) {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(val), nil
}

// StringToFloat64 将字符串转换为 float64
func StringToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// StringToBool 将字符串转换为 bool
func StringToBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

// ConvertToInt64 将任意类型转换为 int64
func ConvertToInt64(v interface{}) (int64, error) {
	if v == nil {
		return 0, fmt.Errorf("value is nil")
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return 0, fmt.Errorf("pointer is nil")
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return int64(rv.Float()), nil
	case reflect.String:
		return StringToInt64(rv.String())
	case reflect.Bool:
		if rv.Bool() {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}

// ConvertToFloat64 将任意类型转换为 float64
func ConvertToFloat64(v interface{}) (float64, error) {
	if v == nil {
		return 0, fmt.Errorf("value is nil")
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return 0, fmt.Errorf("pointer is nil")
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	case reflect.String:
		return StringToFloat64(rv.String())
	case reflect.Bool:
		if rv.Bool() {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// ConvertToBool 将任意类型转换为 bool
func ConvertToBool(v interface{}) (bool, error) {
	if v == nil {
		return false, fmt.Errorf("value is nil")
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return false, fmt.Errorf("pointer is nil")
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() != 0, nil
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0, nil
	case reflect.String:
		return StringToBool(rv.String())
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}

// ConvertToTime 将任意类型转换为 time.Time
func ConvertToTime(v interface{}) (time.Time, error) {
	if v == nil {
		return time.Time{}, fmt.Errorf("value is nil")
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return time.Time{}, fmt.Errorf("pointer is nil")
		}
		rv = rv.Elem()
	}

	switch val := v.(type) {
	case time.Time:
		return val, nil
	case *time.Time:
		if val == nil {
			return time.Time{}, fmt.Errorf("time pointer is nil")
		}
		return *val, nil
	case string:
		// 尝试多种时间格式
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02",
			time.RFC1123,
			time.RFC1123Z,
		}
		for _, format := range formats {
			if t, err := time.Parse(format, val); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("cannot parse time string: %s", val)
	case int64:
		// Unix 时间戳（秒）
		return time.Unix(val, 0), nil
	case int:
		// Unix 时间戳（秒）
		return time.Unix(int64(val), 0), nil
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to time.Time", v)
	}
}

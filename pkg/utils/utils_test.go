package utils

import (
	"encoding/json"
	"testing"
)

func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()

	if id1 == id2 {
		t.Errorf("Generated IDs should be different: %d == %d", id1, id2)
	}

	if id1 <= 0 {
		t.Errorf("Generated ID should be positive: %d", id1)
	}

	idStr := GenerateIDString()
	if idStr == "" {
		t.Error("Generated ID string should not be empty")
	}
}

func TestConvertToMap(t *testing.T) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	user := User{ID: 1, Name: "Alice", Age: 30}
	m, err := ConvertToMap(user)
	if err != nil {
		t.Fatalf("ConvertToMap failed: %v", err)
	}

	if m["id"] != 1 {
		t.Errorf("Expected id=1, got %v", m["id"])
	}
	if m["name"] != "Alice" {
		t.Errorf("Expected name=Alice, got %v", m["name"])
	}
	if m["age"] != 30 {
		t.Errorf("Expected age=30, got %v", m["age"])
	}
}

func TestConvertToString(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{123, "123"},
		{3.14, "3.14"},
		{true, "true"},
		{false, "false"},
		{"hello", "hello"},
		{nil, ""},
	}

	for _, tt := range tests {
		result := ConvertToString(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertToString(%v) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestMapToJSON(t *testing.T) {
	m := map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}

	jsonStr, err := MapToJSON(m)
	if err != nil {
		t.Fatalf("MapToJSON failed: %v", err)
	}

	if jsonStr == "" {
		t.Error("JSON string should not be empty")
	}

	// 验证 JSON 是否有效
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Errorf("Generated JSON is invalid: %v", err)
	}
}

func TestJSONToMap(t *testing.T) {
	jsonStr := `{"name":"Alice","age":30}`

	m, err := JSONToMap(jsonStr)
	if err != nil {
		t.Fatalf("JSONToMap failed: %v", err)
	}

	if m["name"] != "Alice" {
		t.Errorf("Expected name=Alice, got %v", m["name"])
	}

	if age, ok := m["age"].(float64); !ok || age != 30 {
		t.Errorf("Expected age=30, got %v", m["age"])
	}
}

func TestJSONBytesToMap(t *testing.T) {
	jsonBytes := []byte(`{"name":"Alice","age":30}`)

	m, err := JSONBytesToMap(jsonBytes)
	if err != nil {
		t.Fatalf("JSONBytesToMap failed: %v", err)
	}

	if m["name"] != "Alice" {
		t.Errorf("Expected name=Alice, got %v", m["name"])
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected int
		hasError bool
	}{
		{"123", 123, false},
		{123, 123, false},
		{123.0, 123, false},
		{true, 1, false},
		{false, 0, false},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		result, err := ParseInt(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("ParseInt(%v) should return error", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseInt(%v) failed: %v", tt.input, err)
			} else if result != tt.expected {
				t.Errorf("ParseInt(%v) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	}
}

func TestParseInt64(t *testing.T) {
	result, err := ParseInt64("123")
	if err != nil {
		t.Fatalf("ParseInt64 failed: %v", err)
	}
	if result != 123 {
		t.Errorf("ParseInt64(\"123\") = %d, want 123", result)
	}
}

func TestParseFloat64(t *testing.T) {
	result, err := ParseFloat64("3.14")
	if err != nil {
		t.Fatalf("ParseFloat64 failed: %v", err)
	}
	if result != 3.14 {
		t.Errorf("ParseFloat64(\"3.14\") = %f, want 3.14", result)
	}
}

func TestParseBool(t *testing.T) {
	result, err := ParseBool("true")
	if err != nil {
		t.Fatalf("ParseBool failed: %v", err)
	}
	if !result {
		t.Error("ParseBool(\"true\") = false, want true")
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		input    interface{}
		hasError bool
	}{
		{"2021-01-01 00:00:00", false},
		{"2021-01-01T00:00:00Z", false},
		{1609459200, false}, // Unix timestamp
		{"invalid", true},
	}

	for _, tt := range tests {
		_, err := ParseTime(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("ParseTime(%v) should return error", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseTime(%v) failed: %v", tt.input, err)
			}
		}
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected bool
	}{
		{"", true},
		{0, true},
		{nil, true},
		{"hello", false},
		{123, false},
		{[]int{}, true},
		{[]int{1}, false},
	}

	for _, tt := range tests {
		result := IsEmpty(tt.input)
		if result != tt.expected {
			t.Errorf("IsEmpty(%v) = %t, want %t", tt.input, result, tt.expected)
		}
	}
}

func TestIsNil(t *testing.T) {
	var nilPtr *int
	tests := []struct {
		input    interface{}
		expected bool
	}{
		{nil, true},
		{nilPtr, true},
		{123, false},
		{"hello", false},
	}

	for _, tt := range tests {
		result := IsNil(tt.input)
		if result != tt.expected {
			t.Errorf("IsNil(%v) = %t, want %t", tt.input, result, tt.expected)
		}
	}
}

func TestUniqueString(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b"}
	expected := []string{"a", "b", "c"}
	result := UniqueString(input)

	if len(result) != len(expected) {
		t.Errorf("UniqueString length = %d, want %d", len(result), len(expected))
	}

	for i, v := range expected {
		if i >= len(result) || result[i] != v {
			t.Errorf("UniqueString[%d] = %s, want %s", i, result[i], v)
		}
	}
}

func TestSumInt(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	expected := 15
	result := SumInt(input)

	if result != expected {
		t.Errorf("SumInt = %d, want %d", result, expected)
	}
}

func TestMaxInt(t *testing.T) {
	input := []int{1, 5, 3, 2, 4}
	expected := 5
	result, err := MaxInt(input)
	if err != nil {
		t.Fatalf("MaxInt failed: %v", err)
	}
	if result != expected {
		t.Errorf("MaxInt = %d, want %d", result, expected)
	}
}

func TestMinInt(t *testing.T) {
	input := []int{5, 1, 3, 2, 4}
	expected := 1
	result, err := MinInt(input)
	if err != nil {
		t.Fatalf("MinInt failed: %v", err)
	}
	if result != expected {
		t.Errorf("MinInt = %d, want %d", result, expected)
	}
}

func TestTruncateString(t *testing.T) {
	result := TruncateString("hello", 3)
	if result != "hel" {
		t.Errorf("TruncateString = %s, want hel", result)
	}
}

func TestReverseString(t *testing.T) {
	result := ReverseString("hello")
	if result != "olleh" {
		t.Errorf("ReverseString = %s, want olleh", result)
	}
}

func TestContainsString(t *testing.T) {
	slice := []string{"a", "b", "c"}
	if !ContainsString(slice, "b") {
		t.Error("ContainsString should return true for 'b'")
	}
	if ContainsString(slice, "d") {
		t.Error("ContainsString should return false for 'd'")
	}
}

func TestIsValidJSON(t *testing.T) {
	if !IsValidJSON(`{"name":"Alice"}`) {
		t.Error("IsValidJSON should return true for valid JSON")
	}
	if IsValidJSON(`{"name":"Alice"`) {
		t.Error("IsValidJSON should return false for invalid JSON")
	}
}

func TestPrettyJSON(t *testing.T) {
	m := map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}

	pretty, err := PrettyJSON(m)
	if err != nil {
		t.Fatalf("PrettyJSON failed: %v", err)
	}

	if pretty == "" {
		t.Error("PrettyJSON should not return empty string")
	}

	// 验证格式化后的 JSON 是否有效
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(pretty), &result); err != nil {
		t.Errorf("PrettyJSON generated invalid JSON: %v", err)
	}
}

func TestDefaultString(t *testing.T) {
	result := DefaultString("", "default")
	if result != "default" {
		t.Errorf("DefaultString = %s, want default", result)
	}

	result = DefaultString("value", "default")
	if result != "value" {
		t.Errorf("DefaultString = %s, want value", result)
	}
}

func TestCoalesceString(t *testing.T) {
	result := CoalesceString("", "default", "fallback")
	if result != "default" {
		t.Errorf("CoalesceString = %s, want default", result)
	}

	result = CoalesceString("value", "default", "fallback")
	if result != "value" {
		t.Errorf("CoalesceString = %s, want value", result)
	}
}

func TestParseIntSlice(t *testing.T) {
	input := []string{"1", "2", "3"}
	expected := []int{1, 2, 3}
	result, err := ParseIntSlice(input)
	if err != nil {
		t.Fatalf("ParseIntSlice failed: %v", err)
	}

	if len(result) != len(expected) {
		t.Errorf("ParseIntSlice length = %d, want %d", len(result), len(expected))
	}

	for i, v := range expected {
		if i >= len(result) || result[i] != v {
			t.Errorf("ParseIntSlice[%d] = %d, want %d", i, result[i], v)
		}
	}
}

func TestStringToIntSlice(t *testing.T) {
	result, err := StringToIntSlice("1,2,3", ",")
	if err != nil {
		t.Fatalf("StringToIntSlice failed: %v", err)
	}

	expected := []int{1, 2, 3}
	if len(result) != len(expected) {
		t.Errorf("StringToIntSlice length = %d, want %d", len(result), len(expected))
	}

	for i, v := range expected {
		if i >= len(result) || result[i] != v {
			t.Errorf("StringToIntSlice[%d] = %d, want %d", i, result[i], v)
		}
	}
}

func TestMarshalJSON(t *testing.T) {
	m := map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}

	jsonBytes, err := MarshalJSON(m)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Error("MarshalJSON should not return empty bytes")
	}

	// 验证 JSON 是否有效
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Errorf("MarshalJSON generated invalid JSON: %v", err)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	jsonBytes := []byte(`{"name":"Alice","age":30}`)
	var result map[string]interface{}

	err := UnmarshalJSON(jsonBytes, &result)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if result["name"] != "Alice" {
		t.Errorf("Expected name=Alice, got %v", result["name"])
	}
}

func TestNewSnowflakeID(t *testing.T) {
	snowflake := NewSnowflakeID(1, 1)
	id1 := snowflake.Generate()
	id2 := snowflake.Generate()

	if id1 == id2 {
		t.Errorf("Generated IDs should be different: %d == %d", id1, id2)
	}

	if id1 <= 0 {
		t.Errorf("Generated ID should be positive: %d", id1)
	}
}

func TestConvertToInt64(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected int64
		hasError bool
	}{
		{"123", 123, false},
		{123, 123, false},
		{123.0, 123, false},
		{true, 1, false},
		{false, 0, false},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		result, err := ConvertToInt64(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("ConvertToInt64(%v) should return error", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ConvertToInt64(%v) failed: %v", tt.input, err)
			} else if result != tt.expected {
				t.Errorf("ConvertToInt64(%v) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	}
}

func TestConvertToFloat64(t *testing.T) {
	result, err := ConvertToFloat64("3.14")
	if err != nil {
		t.Fatalf("ConvertToFloat64 failed: %v", err)
	}
	if result != 3.14 {
		t.Errorf("ConvertToFloat64(\"3.14\") = %f, want 3.14", result)
	}
}

func TestConvertToBool(t *testing.T) {
	result, err := ConvertToBool("true")
	if err != nil {
		t.Fatalf("ConvertToBool failed: %v", err)
	}
	if !result {
		t.Error("ConvertToBool(\"true\") = false, want true")
	}
}

func TestPadLeft(t *testing.T) {
	result := PadLeft("123", 5, '0')
	if result != "00123" {
		t.Errorf("PadLeft = %s, want 00123", result)
	}
}

func TestPadRight(t *testing.T) {
	result := PadRight("123", 5, '0')
	if result != "12300" {
		t.Errorf("PadRight = %s, want 12300", result)
	}
}

func TestFilterString(t *testing.T) {
	input := []string{"a", "b", "c", "d"}
	result := FilterString(input, func(s string) bool {
		return s != "b"
	})

	expected := []string{"a", "c", "d"}
	if len(result) != len(expected) {
		t.Errorf("FilterString length = %d, want %d", len(result), len(expected))
	}
}

func TestMapString(t *testing.T) {
	input := []string{"a", "b", "c"}
	result := MapString(input, func(s string) string {
		return s + "x"
	})

	expected := []string{"ax", "bx", "cx"}
	if len(result) != len(expected) {
		t.Errorf("MapString length = %d, want %d", len(result), len(expected))
	}

	for i, v := range expected {
		if i >= len(result) || result[i] != v {
			t.Errorf("MapString[%d] = %s, want %s", i, result[i], v)
		}
	}
}

package utils

import (
	"fmt"
)

// ExampleGenerateID 演示雪花算法 ID 生成
func ExampleGenerateID() {
	// 使用默认生成器
	id := GenerateID()
	fmt.Printf("Generated ID: %d\n", id)

	// 生成字符串 ID
	idStr := GenerateIDString()
	fmt.Printf("Generated ID String: %s\n", idStr)

	// 创建自定义生成器
	snowflake := NewSnowflakeID(1, 1)
	customID := snowflake.Generate()
	fmt.Printf("Custom ID: %d\n", customID)
}

// ExampleConvertToMap 演示类型转换为 Map
func ExampleConvertToMap() {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	user := User{ID: 1, Name: "Alice", Age: 30}
	m, err := ConvertToMap(user)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Map: %v\n", m)
	// Output: Map: map[age:30 id:1 name:Alice]
}

// ExampleConvertToString 演示类型转换为字符串
func ExampleConvertToString() {
	fmt.Println(ConvertToString(123))         // "123"
	fmt.Println(ConvertToString(3.14))        // "3.14"
	fmt.Println(ConvertToString(true))        // "true"
	fmt.Println(ConvertToString([]int{1, 2})) // "[1,2]"
}

// ExampleMapToJSON 演示 JSON 处理函数
func ExampleMapToJSON() {
	// Map 转 JSON
	m := map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}
	jsonStr, _ := MapToJSON(m)
	fmt.Printf("JSON: %s\n", jsonStr)

	// JSON 转 Map
	m2, _ := JSONToMap(jsonStr)
	fmt.Printf("Map: %v\n", m2)

	// 格式化 JSON
	pretty, _ := PrettyJSON(m)
	fmt.Printf("Pretty JSON:\n%s\n", pretty)
}

// ExampleParseInt 演示解析函数
func ExampleParseInt() {
	// 解析为 int
	i, _ := ParseInt("123")
	fmt.Printf("Int: %d\n", i)

	// 解析为 int64
	i64, _ := ParseInt64("123")
	fmt.Printf("Int64: %d\n", i64)

	// 解析为 float64
	f, _ := ParseFloat64("3.14")
	fmt.Printf("Float64: %f\n", f)

	// 解析为 bool
	b, _ := ParseBool("true")
	fmt.Printf("Bool: %t\n", b)

	// 解析为 time.Time
	t, _ := ParseTime("2021-01-01 00:00:00")
	fmt.Printf("Time: %v\n", t)

	// 解析为 time.Duration
	d, _ := ParseDuration("1h30m")
	fmt.Printf("Duration: %v\n", d)
}

// ExampleIsEmpty 演示通用工具函数
func ExampleIsEmpty() {
	// 检查是否为空
	fmt.Printf("IsEmpty: %t\n", IsEmpty("")) // true
	fmt.Printf("IsEmpty: %t\n", IsEmpty(0))  // true

	// 检查是否为 nil
	fmt.Printf("IsNil: %t\n", IsNil(nil)) // true

	// 字符串操作
	fmt.Printf("Truncate: %s\n", TruncateString("hello", 3)) // "hel"
	fmt.Printf("Reverse: %s\n", ReverseString("hello"))      // "olleh"

	// 切片操作
	slice := []int{1, 2, 3, 2, 1}
	fmt.Printf("Unique: %v\n", UniqueInt(slice)) // [1, 2, 3]
	fmt.Printf("Sum: %d\n", SumInt(slice))       // 9
	max, _ := MaxInt(slice)
	fmt.Printf("Max: %d\n", max) // 3
}

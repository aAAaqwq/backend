# Utils 工具函数库

本包提供了一系列常用的工具函数，包括ID生成、类型转换、JSON处理等。

## 功能模块

### 1. ID 生成 (`id.go`)

#### 雪花算法 ID 生成器

```go
import "backend/pkg/utils"

// 使用默认生成器（单例模式）
// 默认生成器会优先从环境变量读取配置：
// - SNOWFLAKE_MACHINE_ID (0-1023)
// - SNOWFLAKE_DATACENTER_ID (0-31)
// 如果未设置环境变量，则基于主机名自动生成唯一值
id := utils.GenerateID()              // 生成 int64 ID
idStr := utils.GenerateIDString()     // 生成字符串 ID

// 创建自定义生成器（推荐用于多实例部署）
snowflake := utils.NewSnowflakeID(1, 1) // machineID=1, datacenterID=1
id := snowflake.Generate()
```

**配置方式：**

1. **环境变量配置（推荐）**：
```bash
export SNOWFLAKE_MACHINE_ID=10
export SNOWFLAKE_DATACENTER_ID=5
```

2. **基于主机名自动生成（默认）**：
如果未设置环境变量，系统会自动基于主机名生成唯一值，确保不同机器上的实例不会冲突。

**示例：**
```go
// 生成 ID（自动使用环境变量或主机名）
id1 := utils.GenerateID()        // 1234567890123456789
id2 := utils.GenerateIDString()  // "1234567890123456789"

// 自定义生成器（显式指定，用于多实例部署）
snowflake := utils.NewSnowflakeID(10, 5)
id := snowflake.Generate()
```

**注意事项：**
- 多实例部署时，建议通过环境变量为每个实例配置唯一的 `machineID` 和 `datacenterID`
- `machineID` 范围：0-1023（支持最多 1024 个机器）
- `datacenterID` 范围：0-31（支持最多 32 个数据中心）

### 2. 类型转换 (`converter.go`)

#### 转换为 Map

```go
// 将任意类型转换为 map[string]interface{}
m, err := utils.ConvertToMap(user)
```

**示例：**
```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

user := User{ID: 1, Name: "Alice"}
m, err := utils.ConvertToMap(user)
// m = map[string]interface{}{"id": 1, "name": "Alice"}
```

#### 转换为字符串

```go
// 将任意类型转换为字符串
str := utils.ConvertToString(123)        // "123"
str := utils.ConvertToString(3.14)      // "3.14"
str := utils.ConvertToString(true)      // "true"
str := utils.ConvertToString([]int{1,2}) // "[1,2]"
```

#### 类型转换函数

```go
// 转换为 int64
id, err := utils.ConvertToInt64("123")

// 转换为 float64
f, err := utils.ConvertToFloat64("3.14")

// 转换为 bool
b, err := utils.ConvertToBool("true")

// 转换为 time.Time
t, err := utils.ConvertToTime("2021-01-01 00:00:00")
t, err := utils.ConvertToTime(1609459200) // Unix 时间戳
```

### 3. JSON 处理 (`json.go`)

#### JSON 序列化

```go
// 转换为 JSON 字节
jsonBytes, err := utils.MarshalJSON(data)

// 格式化的 JSON
jsonBytes, err := utils.MarshalJSONIndent(data, "", "  ")

// Map 转 JSON 字符串
jsonStr, err := utils.MapToJSON(m)

// Map 转 JSON 字节
jsonBytes, err := utils.MapToJSONBytes(m)

// 格式化的 JSON 字符串
prettyJSON, err := utils.PrettyJSON(data)
```

#### JSON 反序列化

```go
// JSON 字节解析为指定类型
err := utils.UnmarshalJSON(jsonBytes, &result)

// JSON 字符串解析为 map
m, err := utils.JSONToMap(jsonStr)

// JSON 字节解析为 map
m, err := utils.JSONBytesToMap(jsonBytes)

// JSON 字符串解析为切片
slice, err := utils.JSONToSlice(jsonStr)
```

#### JSON 验证

```go
// 检查是否为有效的 JSON
isValid := utils.IsValidJSON(jsonStr)
isValid := utils.IsValidJSONBytes(jsonBytes)
```

**示例：**
```go
// 序列化
data := map[string]interface{}{"name": "Alice", "age": 30}
jsonBytes, _ := utils.MarshalJSON(data)
jsonStr, _ := utils.MapToJSON(data)

// 反序列化
var result map[string]interface{}
utils.UnmarshalJSON(jsonBytes, &result)

m, _ := utils.JSONToMap(`{"name":"Alice","age":30}`)

// 格式化
pretty, _ := utils.PrettyJSON(data)
// {
//   "age": 30,
//   "name": "Alice"
// }
```

### 4. 解析函数 (`parse.go`)

#### 类型解析

```go
// 解析为字符串
str := utils.ParseString(123)

// 解析为 int
i, err := utils.ParseInt("123")

// 解析为 int64
i64, err := utils.ParseInt64("123")

// 解析为 float64
f, err := utils.ParseFloat64("3.14")

// 解析为 bool
b, err := utils.ParseBool("true")

// 解析为 time.Time
t, err := utils.ParseTime("2021-01-01 00:00:00")

// 解析为 time.Duration
d, err := utils.ParseDuration("1h30m")
```

#### JSON 解析

```go
// JSON 字符串解析
err := utils.ParseJSON(jsonStr, &result)

// JSON 字节解析
err := utils.ParseJSONBytes(jsonBytes, &result)
```

#### 切片解析

```go
// 字符串切片解析为 int 切片
ints, err := utils.ParseIntSlice([]string{"1", "2", "3"})

// 字符串切片解析为 int64 切片
int64s, err := utils.ParseInt64Slice([]string{"1", "2", "3"})

// 字符串切片解析为 float64 切片
floats, err := utils.ParseFloat64Slice([]string{"1.1", "2.2", "3.3"})

// 字符串切片解析为 bool 切片
bools, err := utils.ParseBoolSlice([]string{"true", "false", "true"})

// 逗号分隔的字符串解析为 int 切片
ints, err := utils.StringToIntSlice("1,2,3", ",")
```

#### 默认值函数

```go
// 如果值为空，返回默认值
str := utils.DefaultString("", "default")  // "default"
i := utils.DefaultInt(0, 100)            // 100
i64 := utils.DefaultInt64(0, 100)        // 100
f := utils.DefaultFloat64(0, 3.14)       // 3.14
b := utils.DefaultBool(false, true)      // true

// 返回第一个非零值
val := utils.Coalesce("", "default", "fallback")  // "default"
val := utils.CoalesceString("", "default", "fallback")  // "default"
val := utils.CoalesceInt(0, 100, 200)  // 100
```

### 5. 通用工具 (`common.go`)

#### 值检查

```go
// 检查是否为空
isEmpty := utils.IsEmpty("")      // true
isEmpty := utils.IsEmpty(0)        // true
isEmpty := utils.IsEmpty(nil)      // true

// 检查是否为 nil
isNil := utils.IsNil(nil)          // true
isNil := utils.IsNil((*int)(nil))  // true

// 检查是否为零值
isZero := utils.IsZero(0)         // true
isZero := utils.IsZero("")        // true
```

#### 类型信息

```go
// 获取类型名称
typeName := utils.TypeOf(123)     // "int"
typeName := utils.TypeOf("abc")   // "string"

// 获取 Kind
kind := utils.KindOf(123)         // reflect.Int
kind := utils.KindOf("abc")       // reflect.String
```

#### 字符串操作

```go
// 截断字符串
str := utils.TruncateString("hello", 3)              // "hel"
str := utils.TruncateStringWithEllipsis("hello", 3) // "..."

// 填充字符串
str := utils.PadLeft("123", 5, '0')   // "00123"
str := utils.PadRight("123", 5, '0')  // "12300"

// 反转字符串
str := utils.ReverseString("hello")   // "olleh"
```

#### 切片操作

```go
// 检查是否包含
contains := utils.ContainsString([]string{"a", "b"}, "a")  // true
contains := utils.ContainsInt([]int{1, 2}, 1)             // true

// 去重
unique := utils.UniqueString([]string{"a", "b", "a"})     // ["a", "b"]
unique := utils.UniqueInt([]int{1, 2, 1})                 // [1, 2]

// 过滤
filtered := utils.FilterString([]string{"a", "b", "c"}, func(s string) bool {
    return s != "b"
})  // ["a", "c"]

// 映射
mapped := utils.MapString([]string{"a", "b"}, strings.ToUpper)  // ["A", "B"]

// 求和
sum := utils.SumInt([]int{1, 2, 3})        // 6
sum := utils.SumInt64([]int64{1, 2, 3})    // 6
sum := utils.SumFloat64([]float64{1.1, 2.2})  // 3.3

// 最大值/最小值
max, _ := utils.MaxInt([]int{1, 3, 2})     // 3
min, _ := utils.MinInt([]int{1, 3, 2})     // 1
```

## 完整示例

```go
package main

import (
	"backend/pkg/utils"
	"fmt"
)

func main() {
	// 1. 生成雪花算法 ID
	id := utils.GenerateID()
	fmt.Printf("ID: %d\n", id)

	// 2. 类型转换为 Map
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	user := User{ID: 1, Name: "Alice"}
	m, _ := utils.ConvertToMap(user)
	fmt.Printf("Map: %v\n", m)

	// 3. Map 转 JSON
	jsonStr, _ := utils.MapToJSON(m)
	fmt.Printf("JSON: %s\n", jsonStr)

	// 4. JSON 转 Map
	m2, _ := utils.JSONToMap(jsonStr)
	fmt.Printf("Map from JSON: %v\n", m2)

	// 5. 类型转换
	str := utils.ConvertToString(123)
	fmt.Printf("String: %s\n", str)

	// 6. 解析
	i, _ := utils.ParseInt("123")
	fmt.Printf("Int: %d\n", i)
}
```

## 注意事项

1. **雪花算法 ID**：
   - 机器ID范围：0-1023
   - 数据中心ID范围：0-31
   - 同一毫秒内最多生成 4096 个 ID

2. **类型转换**：
   - 转换失败会返回错误
   - 指针类型会自动解引用
   - 结构体转换会使用 JSON tag

3. **JSON 处理**：
   - 所有 JSON 函数都处理 nil 值
   - 格式化 JSON 使用 2 空格缩进

4. **性能考虑**：
   - 类型转换使用反射，性能略低
   - JSON 序列化/反序列化有性能开销
   - 建议在高性能场景下谨慎使用


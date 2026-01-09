package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// DeviceID 设备ID类型
// JSON序列化为string，数据库存储为int64
type DeviceID int64

// StringToID 将字符串转换为 DeviceID
func StringToID(s string) (DeviceID, error) {
	if s == "" {
		return 0, errors.New("设备ID不能为空")
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("设备ID格式无效: %s", s)
	}
	return DeviceID(id), nil
}

// MustStringToID 将字符串转换为 DeviceID，忽略错误（用于已知有效输入）
func MustStringToID(s string) DeviceID {
	id, _ := StringToID(s)
	return id
}

// Int64ToID 将 int64 转换为 DeviceID
func Int64ToID(i int64) DeviceID {
	return DeviceID(i)
}

// MarshalJSON 实现 json.Marshaler 接口，将 DeviceID 序列化为 string
func (d DeviceID) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(d), 10))
}

// UnmarshalJSON 实现 json.Unmarshaler 接口，从 string 反序列化为 DeviceID
func (d *DeviceID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		// 尝试解析为数字（兼容旧格式）
		var i int64
		if err2 := json.Unmarshal(b, &i); err2 != nil {
			return errors.New("设备ID必须是字符串或数字格式")
		}
		*d = DeviceID(i)
		return nil
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("设备ID格式无效: %s", s)
	}
	*d = DeviceID(id)
	return nil
}

// Value 实现 driver.Valuer 接口，用于数据库存储
func (d DeviceID) Value() (driver.Value, error) {
	return int64(d), nil
}

// Scan 实现 sql.Scanner 接口，用于从数据库读取
func (d *DeviceID) Scan(value interface{}) error {
	if value == nil {
		*d = 0
		return nil
	}
	switch v := value.(type) {
	case int64:
		*d = DeviceID(v)
	case []byte:
		i, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return err
		}
		*d = DeviceID(i)
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		*d = DeviceID(i)
	default:
		return errors.New("无法扫描设备ID")
	}
	return nil
}

// String 返回 DeviceID 的字符串表示
func (d DeviceID) String() string {
	return strconv.FormatInt(int64(d), 10)
}

// Int64 返回 DeviceID 的 int64 表示
func (d DeviceID) Int64() int64 {
	return int64(d)
}

// IsZero 检查 DeviceID 是否为零值
func (d DeviceID) IsZero() bool {
	return d == 0
}

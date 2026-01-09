package model

import "time"

const (
	DataTypeSeries   = "time_series"
	DataTypeFileData = "file_data"
)

// Metadata 传感器元数据
type Metadata struct {
	DataID       int64                  `json:"data_id" db:"data_id"`                       // 数据ID（自增）
	DevID        DeviceID               `json:"dev_id" db:"dev_id" binding:"required"`      // 设备ID
	DataType     string                 `json:"data_type" db:"data_type" binding:"required"` // 数据类型: time_series/file_data
	QualityScore string                 `json:"quality_score" db:"quality_score"`           // 数据质量评分 0-100
	ExtraData    map[string]interface{} `json:"extra_data" db:"extra_data"`                 // 额外数据（可承载文件bucket_key、data_count等）
	Timestamp    time.Time              `json:"timestamp" db:"timestamp"`                   // 时间戳
}

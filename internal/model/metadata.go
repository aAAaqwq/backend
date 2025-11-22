package model

const (
	DataTypeSeries   = "time_series"
	DataTypeFileData = "file_data"
)

type Metadata struct {
	DataID       int64                  `json:"data_id" binding:"required"`
	DevID        int64                  `json:"dev_id" binding:"required"`
	UID          int64                  `json:"uid" binding:"required"`       // 上传时必填，但时序数据tags中不包含
	DataType     string                 `json:"data_type" binding:"required"` // time_series/file_data
	QualityScore float64                `json:"quality_score" binding:"required"`
	ExtraData    map[string]interface{} `json:"extra_data" binding:"required"`
	Timestamp    string                 `json:"timestamp" binding:"required"` // Unix时间戳字符串（秒）
}

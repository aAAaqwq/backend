package model

type Metadata struct {
	DataID       int64                  `json:"data_id"`
	DevID        int64                  `json:"dev_id" binding:"required"`
	UID          int64                  `json:"uid" binding:"required"`
	DataType     string                 `json:"data_type" binding:"required"` // time_series/file_data
	QualityScore float64                `json:"quality_score"`
	ExtraData    map[string]interface{} `json:"extra_data"`
	Timestamp    int64                  `json:"timestamp"` // Unix时间戳（秒)
}


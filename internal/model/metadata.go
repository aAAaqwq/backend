package model

type Metadata struct {
	DataID       int64                  `json:"data_id"`
	DevID        int64                  `json:"dev_id"`
	UID          int64                  `json:"uid"`
	DataType     string                 `json:"data_type"`     // 时序数据、视频、音频、图片
	StorageRoute string                 `json:"storage_route"` // 存储路径
	QualityScore float64                `json:"quality_score"`
	ExtraData    map[string]interface{} `json:"extra_data"`
	Timestamp    int64                  `json:"timestamp"` // Unix时间戳（秒）
	FilePath     string                 `json:"file_path"` // 文件路径（用于非结构化数据）
}

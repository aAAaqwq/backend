package model

type UploadSensorDataRequest struct {
	Metadata Metadata `json:"metadata"`
	SeriesData SeriesData `json:"series_data" binding:"omitempty"`
	FileData FileData `json:"file_data" binding:"omitempty"`
}

type UpdateUserInfoRequest struct {
	UID int64 `json:"uid" binding:"omitempty"`
	Username string `json:"username" binding:"omitempty"`
	Email string `json:"email" binding:"omitempty"`
}

type GetSeriesDataRequest struct {
	DevID              int64             `json:"dev_id" binding:"required"`
	StartTime          int64             `json:"start_time" binding:"required"`
	EndTime            int64             `json:"end_time" binding:"required"`
	Measurement        string            `json:"measurement" binding:"required"`
	Tags               map[string]string `json:"tags"`
	Fields             map[string]interface{} `json:"fileds"` // 注意拼写与API文档一致
	Aggregate          string            `json:"aggregate"`
	DownSampleInterval string            `json:"down_sample_interval"`
	LimitPoints        int               `json:"limit_points"`
}

type DeleteSeriesDataRequest struct {
	DevID     int64  `json:"dev_id" binding:"required"`      // 设备ID
	StartTime string `json:"start_time" binding:"omitempty"` // 开始时间（可选）
	EndTime   string `json:"end_time" binding:"omitempty"`   // 结束时间（可选）
}

type DeleteFileDataRequest struct {
	BucketName string `json:"bucket_name" binding:"required"`
	BucketKey  string `json:"bucket_key" binding:"required"`
}

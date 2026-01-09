package model

// UploadSensorDataRequest 上传传感器数据请求
type UploadSensorDataRequest struct {
	Metadata   Metadata   `json:"metadata"`
	SeriesData SeriesData `json:"series_data" binding:"omitempty"`
	FileData   FileData   `json:"file_data" binding:"omitempty"`
}

// UpdateUserInfoRequest 更新用户信息请求
type UpdateUserInfoRequest struct {
	UID      int64  `json:"uid" binding:"omitempty"`
	Username string `json:"username" binding:"omitempty"`
	Email    string `json:"email" binding:"omitempty"`
}

// GetSeriesDataRequest 查询时序数据请求
type GetSeriesDataRequest struct {
	DevID              DeviceID           `json:"dev_id" binding:"required"`
	StartTime          int64              `json:"start_time" binding:"required"`
	EndTime            int64              `json:"end_time" binding:"required"`
	Measurement        string             `json:"measurement" binding:"required"`
	Tags               map[string]string  `json:"tags"`
	Fields             map[string]any     `json:"fileds"` // 注意拼写与API文档一致
	Aggregate          string             `json:"aggregate"`
	DownSampleInterval string             `json:"down_sample_interval"`
	LimitPoints        int                `json:"limit_points"`
}

// DeleteSeriesDataRequest 删除时序数据请求
type DeleteSeriesDataRequest struct {
	DevID     DeviceID `json:"dev_id" binding:"required"`
	StartTime string    `json:"start_time" binding:"omitempty"`
	EndTime   string    `json:"end_time" binding:"omitempty"`
}

// DeleteFileDataRequest 删除文件数据请求
type DeleteFileDataRequest struct {
	BucketName string `json:"bucket_name" binding:"required"`
	BucketKey  string `json:"bucket_key" binding:"required"`
}

// GetPresignedPutURLReq 获取预签名PUT URL请求
type GetPresignedPutURLReq struct {
	DevID       string `json:"dev_id" binding:"required"`
	Filename    string `json:"filename" binding:"required"`
	BucketName  string `json:"bucket_name"`
	ContentType string `json:"content_type"`
}

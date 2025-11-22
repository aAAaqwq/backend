package model

type SeriesData struct {
	Points []Point `json:"points"`
}

// Point 时序数据点
// 上传时：只包含timestamp和fields（fields是任意的k-v）
// 查询时：包含完整的measurement、tags、fields、timestamp
type Point struct {
	Measurement string                 `json:"measurement,omitempty"` // 查询时返回，上传时从extra_data获取
	Tags        map[string]string      `json:"tags,omitempty"`        // tags内容在业务层从metadata添加
	Fields      map[string]interface{} `json:"fileds"`                // API文档中拼写为fileds，任意k-v字段
	Timestamp   int64                  `json:"timestamp"`             // Unix时间戳（秒）
}

type FileData struct {
	FilePath   string `json:"file_path" binding:"required"`
	BucketName string `json:"bucket_name"` // 可空，为空时根据data_type或文件路径推断
	BucketKey  string `json:"bucket_key"`  // 可空，为空时创建默认的dev_id/filename为key
}

type FileList struct {
	BucketKey    string `json:"bucket_key"`
	PreviewUrl   string `json:"preview_url"`
	Name         string `json:"name"`
	ContentType  string `json:"content_type"`
	LastModified string `json:"last_modified"`
	Size         int64  `json:"size"` // 字节数
}

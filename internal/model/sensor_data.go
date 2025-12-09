package model

type SeriesData struct {
	Points []Point `json:"points"`
}

// Point 时序数据点
type Point struct {
	Measurement string            `json:"measurement,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Fields      map[string]any    `json:"fields"`
	Timestamp   int64             `json:"timestamp"`
}

// FileData 文件数据
type FileData struct {
	UploadID   string `json:"upload_id" binding:"required"`
	BucketName string `json:"bucket_name" `
	BucketKey  string `json:"bucket_key"`
}

type FileList struct {
	BucketKey    string `json:"bucket_key"`
	PreviewUrl   string `json:"preview_url"`
	Name         string `json:"name"`
	ContentType  string `json:"content_type"`
	LastModified string `json:"last_modified"`
	Size         int64  `json:"size"`
}

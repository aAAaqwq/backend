package model

import "time"

const (
	DataTypeVideo = "video"
	DataTypeAudio = "audio"
	DataTypeImage = "image"
	DataTypeSeries = "series"
)

type SeriesData struct{
    Points []Point `json:"points"`
}

type Point struct{
    Measurement string `json:"measurement"`
    Tags map[string]string `json:"tags"`
    Fields map[string]interface{} `json:"fields"`
    Timestamp time.Time `json:"timestamp"`
}

type FileData struct{
	FilePath string `json:"file_path"`
    BucketName string `json:"bucket_name"`
    BucketKey string `json:"bucket_key"`
}

type FileList struct{
    BucketKey string `json:"bucket_key"`
    PreviewUrl string `json:"preview_url"`
	Name string `json:"file_name"`
	ContentType string `json:"content_type"`
	LastModified time.Time `json:"last_modified"`
	Size int64 `json:"size"`
}


package model

import "time"

const (
	DataTypeVideo = "video"
	DataTypeAudio = "audio"
	DataTypeImage = "image"
	DataTypeSeries = "series"
)

type SensorData struct{
	DataID int64 `json:"data_id"`
    SeriesData SeriesData `json:"series_data"`
    FileData FileData `json:"file_data"`
}


type SeriesData struct{
    Timestamp time.Time `json:"timestamp"`
    Value float64 `json:"value"`
}

type FileData struct{
	FileID int64 `json:"file_id"`
	FilePath string `json:"file_path"`
    FileName string `json:"file_name"`
    FileType string `json:"file_type"`  // video, audio, image
    FileSize int64 `json:"file_size"`
    FileData string `json:"file_data"`
}
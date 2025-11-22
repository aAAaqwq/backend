package model

type UploadSensorDataRequest struct {
	Metadata Metadata `json:"metadata"`
	SeriesData SeriesData `json:"series_data"`
	FileData FileData `json:"file_data"`
}
package model

type UploadSensorDataRequest struct {
	Metadata Metadata `json:"metadata"`
	SeriesData SeriesData `json:"series_data"`
	FileData FileData `json:"file_data"`
}

type UpdateUserInfoRequest struct {
	UID int64 `json:"uid" binding:"omitempty"`
	Username string `json:"username" binding:"omitempty"`
	Email string `json:"email" binding:"omitempty"`
}
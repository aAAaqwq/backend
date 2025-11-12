package model

import "time"



type Metadata struct{
	DataID int64 `json:"data_id"`
	DevID int64 `json:"dev_id"`
	UID int64 `json:"uid"`
	DataType string `json:"data_type"`
	StorageRoute string `json:"storage_route"`
	QualityScore float64 `json:"quality_score"`
	ExtraData string `json:"extra_data"`
	Timestamp time.Time `json:"timestamp"`
}
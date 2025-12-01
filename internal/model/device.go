package model

import "time"

const (
	DevStatusOffline  = 0 // 离线
	DevStatusOnline   = 1 // 在线
	DevStatusAbnormal = 2 // 异常
)

type Device struct {
	DevID            int64                  `json:"dev_id" db:"dev_id"`
	DevName          string                 `json:"dev_name" db:"dev_name" binding:"required"`
	DevStatus        int                    `json:"dev_status" db:"dev_status" oneof:"0 1 2"`
	DevType          string                 `json:"dev_type" db:"dev_type" binding:"required"`
	DevPower         int                    `json:"dev_power" db:"dev_power"`
	Model            string                 `json:"model" db:"model"`                         // 硬件型号（对应SQL中的model字段）
	Version          string                 `json:"version" db:"version"`                     // 硬件版本（对应SQL中的version字段）
	SamplingRate     int                    `json:"sampling_rate" db:"sampling_rate"`         // 采样频率（对应SQL中的sampling_rate字段）
	OfflineThreshold int                    `json:"offline_threshold" db:"offline_threshold"` // 离线判断阈值
	UploadInterval   int                    `json:"upload_interval" db:"upload_interval"`     // 数据上报间隔（对应SQL中的upload_interval字段）
	ExtendedConfig   map[string]interface{} `json:"extended_config" db:"extended_config"`
	CreateAt         time.Time              `json:"create_at" db:"create_at"`
	UpdateAt         time.Time              `json:"update_at" db:"update_at"`
}

type DeviceConfig struct {
	Version          string `json:"version"`           // 硬件版本
	Model            string `json:"model"`             // 设备型号
	SamplingRate     int    `json:"sampling_rate"`     // 采样频率
	UploadInterval   int    `json:"upload_interval"`   // 数据上报间隔
	OfflineThreshold int    `json:"offline_threshold"` // 离线判断阈值
	ExtendedConfig   string `json:"extended_config"`
}

package model

import "time"

const (
	DevStatusOnline   = 0 // 在线
	DevStatusOffline  = 1 // 离线
	DevStatusAbnormal = 2 // 异常
)

type Device struct {
	DevID              int64                  `json:"dev_id"`
	DevName            string                 `json:"dev_name" binding:"required"`
	DevType            string                 `json:"dev_type" binding:"required"`
	DevModel           string                 `json:"dev_model"`
	DevPower           int                    `json:"dev_power" binding:"required"`
	DevStatus          int                    `json:"dev_status" binding:"required"`
	FirmwareVersion    string                 `json:"firmware_version"`
	SamplingFrequency  int                    `json:"sampling_frequency"`
	DataUploadInterval int                    `json:"data_upload_interval"`
	OfflineThreshold   int                    `json:"offline_threshold"`
	ExtendedConfig     map[string]interface{} `json:"extended_config"`
	CreateAt           time.Time              `json:"create_at"`
	UpdateAt           time.Time              `json:"update_at"`
}

type DeviceConfig struct {
	FirmwareVersion      string `json:"firmware_version"`
	DeviceModel          string `json:"device_model"`
	SamplingFrequency    int    `json:"sampling_frequency"`
	TransmissionInterval int    `json:"transmission_interval"`
	OfflineThreshold     int    `json:"offline_threshold"`
	ExtendedConfig       string `json:"extended_config"`
}

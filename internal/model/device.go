package model

import "time"


type Device struct{
	DevID int64    `json:"dev_id"`
	DevName string `json:"dev_name"`
	DevType string `json:"dev_type"`
	DevPower int `json:"dev_power"`
	DevStatus int `json:"dev_status"`
	DeviceConfig DeviceConfig `json:"device_config"`
	CreateAt time.Time `json:"create_at"`
	UpdateAt time.Time `json:"update_at"`
}

type DeviceConfig struct{
	FirmwareVersion string `json:"firmware_version"`
	DeviceModel string `json:"device_model"`
	SamplingFrequency int `json:"sampling_frequency"`
	TransmissionInterval int `json:"transmission_interval"`
	OfflineThreshold int `json:"offline_threshold"`
	ExtendedConfig string `json:"extended_config"`
}




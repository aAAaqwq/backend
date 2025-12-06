package model

import "time"

type WarningInfo struct {
	AlertID      int64      `json:"alert_id"`
	DataID       int64      `json:"data_id,omitempty"`       // 出现警告的数据
	DevID        int64      `json:"dev_id,omitempty"`        // 出现警告的设备
	AlertType    string     `json:"alert_type"`              // dev/data
	AlertStatus  string     `json:"alert_status"`            // active/resolved/ignored
	AlertMessage string     `json:"alert_message,omitempty"`
	TriggeredAt  time.Time  `json:"triggered_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`   // 使用指针类型，支持 NULL
}

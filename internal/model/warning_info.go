package model

import "time"


type WarningInfo struct{
	AlertID int64 `json:"alert_id"`
	AlertType string `json:"alert_type"`
	AlertMessage string `json:"alert_message"`
	AlertStatus string `json:"alert_status"`
	TriggeredAt time.Time `json:"triggered_at"`
	ResolvedAt time.Time `json:"resolved_at"`
}
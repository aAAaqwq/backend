package model

import "time"

type Log struct {
	LogID     int64     `json:"log_id"`
	Type      string    `json:"type"`              // 日志类型
	Level     int       `json:"level"`             // 日志级别
	Message   string    `json:"message"`           // 日志内容
	UserAgent string    `json:"user_agent"`        // 用户代理
	CreateAt  time.Time `json:"create_at"`         // 创建时间
}

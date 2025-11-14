package model

import "time"

const (
	PermissionLevelRead      = "r"  // 只读
	PermissionLevelWrite     = "w"  // 只写
	PermissionLevelReadWrite = "rw" // 读写
)

type DeviceUser struct {
	UID             int64     `json:"uid"`
	DevID           int64     `json:"dev_id"`
	PermissionLevel string    `json:"permission_level"` // r, w, rw
	IsActive        bool      `json:"is_active"`
	BoundAt         time.Time `json:"bound_at"`
	UpdateAt        time.Time `json:"update_at"`
}

// DeviceUserBindingReq 设备用户绑定请求
type DeviceUserBindingReq struct {
	UID             int64  `json:"uid" binding:"required"`
	PermissionLevel string `json:"permission_level" binding:"required"` // 注意：API定义中拼写为permission_levell，这里使用正确的拼写
}

// DeviceUserUpdateReq 更新设备用户绑定请求
type DeviceUserUpdateReq struct {
	PermissionLevel string `json:"permission_level" binding:"required"`
	IsActive        *bool  `json:"is_active"`
}

// DeviceUserWithInfo 带用户信息的设备用户绑定
type DeviceUserWithInfo struct {
	UID             int64     `json:"uid"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	PermissionLevel string    `json:"permission_level"`
	IsActive        bool      `json:"is_active"`
	BoundAt         time.Time `json:"bound_at"`
}

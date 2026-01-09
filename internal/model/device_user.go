package model

import "time"

const (
	PermissionLevelRead      = "r"  // 只读
	PermissionLevelWrite     = "w"  // 只写
	PermissionLevelReadWrite = "rw" // 读写
)

type DeviceUser struct {
	UID             int64     `json:"uid" db:"uid"`
	DevID           DeviceID  `json:"dev_id" db:"dev_id"`
	PermissionLevel string    `json:"permission_level" db:"permission_level"` // r, w, rw
	BindAt          time.Time `json:"bind_at" db:"bind_at"`                    // 绑定时间（对应SQL中的bind_at）
}

// DeviceUserBindingReq 设备用户绑定请求
type DeviceUserBindingReq struct {
	UID             int64    `json:"uid"`
	DevID           DeviceID `json:"dev_id" binding:"required"`
	PermissionLevel string   `json:"permission_level" binding:"oneof=r w rw"`
}

// DeviceUserUpdateReq 更新设备用户绑定请求
type DeviceUserUpdateReq struct {
	PermissionLevel string `json:"permission_level" binding:"required"`
}

// DeviceUserWithInfo 带用户信息的设备用户绑定
type DeviceUserWithInfo struct {
	UID             int64     `json:"uid"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	PermissionLevel string    `json:"permission_level"`
	BindAt          time.Time `json:"bind_at"` // 绑定时间
}

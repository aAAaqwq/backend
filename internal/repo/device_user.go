package repo

import (
	"database/sql"
	"encoding/json"
	"errors"

	"backend/internal/db/mysql"
	"backend/internal/model"
	"backend/pkg/utils"
)

type DeviceUserRepository struct{}

func NewDeviceUserRepository() *DeviceUserRepository {
	return &DeviceUserRepository{}
}

// BindDeviceUser 绑定用户到设备
func (r *DeviceUserRepository) BindDeviceUser(deviceUser *model.DeviceUser) error {
	// 检查是否已绑定
	var count int
	err := mysql.MysqlCli.Client.QueryRow(
		"SELECT COUNT(*) FROM device_user WHERE dev_id = ? AND uid = ?",
		deviceUser.DevID, deviceUser.UID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("用户已绑定到该设备")
	}

	// 检查设备绑定的用户数量（最多3个）
	var userCount int
	err = mysql.MysqlCli.Client.QueryRow(
		"SELECT COUNT(*) FROM device_user WHERE dev_id = ?",
		deviceUser.DevID).Scan(&userCount)
	if err != nil {
		return err
	}
	if userCount >= 3 {
		return errors.New("设备最多只能绑定3个用户")
	}

	// 插入绑定关系
	query := `INSERT INTO device_user (uid, dev_id, permission_level, is_active, bound_at, update_at) 
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err = mysql.MysqlCli.Client.Exec(query,
		deviceUser.UID, deviceUser.DevID, deviceUser.PermissionLevel,
		deviceUser.IsActive, deviceUser.BoundAt, deviceUser.UpdateAt)
	return err
}

// GetDeviceUsers 获取设备的绑定用户列表
func (r *DeviceUserRepository) GetDeviceUsers(devID int64) ([]*model.DeviceUserWithInfo, error) {
	query := `SELECT du.uid, u.username, u.email, du.permission_level, du.is_active, du.bound_at
		FROM device_user du
		LEFT JOIN user u ON du.uid = u.uid
		WHERE du.dev_id = ?
		ORDER BY du.bound_at DESC`

	rows, err := mysql.MysqlCli.Client.Query(query, devID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.DeviceUserWithInfo
	for rows.Next() {
		user := &model.DeviceUserWithInfo{}
		err := rows.Scan(
			&user.UID, &user.Username, &user.Email,
			&user.PermissionLevel, &user.IsActive, &user.BoundAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// GetDeviceUser 获取设备用户绑定关系
func (r *DeviceUserRepository) GetDeviceUser(devID, uid int64) (*model.DeviceUser, error) {
	deviceUser := &model.DeviceUser{}
	query := `SELECT uid, dev_id, permission_level, is_active, bound_at, update_at
		FROM device_user WHERE dev_id = ? AND uid = ?`

	err := mysql.MysqlCli.Client.QueryRow(query, devID, uid).Scan(
		&deviceUser.UID, &deviceUser.DevID, &deviceUser.PermissionLevel,
		&deviceUser.IsActive, &deviceUser.BoundAt, &deviceUser.UpdateAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("设备用户绑定关系不存在")
		}
		return nil, err
	}

	return deviceUser, nil
}

// UpdateDeviceUser 更新设备用户绑定关系
func (r *DeviceUserRepository) UpdateDeviceUser(deviceUser *model.DeviceUser) error {
	query := `UPDATE device_user SET permission_level = ?, is_active = ?, update_at = ?
		WHERE dev_id = ? AND uid = ?`

	_, err := mysql.MysqlCli.Client.Exec(query,
		deviceUser.PermissionLevel, deviceUser.IsActive, deviceUser.UpdateAt,
		deviceUser.DevID, deviceUser.UID)
	return err
}

// UnbindDeviceUser 解绑用户设备
func (r *DeviceUserRepository) UnbindDeviceUser(devID, uid int64) error {
	_, err := mysql.MysqlCli.Client.Exec(
		"DELETE FROM device_user WHERE dev_id = ? AND uid = ?", devID, uid)
	return err
}

// GetUserDevices 获取用户绑定的设备列表
func (r *DeviceUserRepository) GetUserDevices(uid int64, page, pageSize int, devType string, devStatus *int, permissionLevel string, isActive *bool) ([]*model.Device, int64, error) {
	whereClause := "WHERE du.uid = ?"
	args := []interface{}{uid}

	if !utils.IsEmpty(devType) {
		whereClause += " AND d.dev_type = ?"
		args = append(args, devType)
	}
	if devStatus != nil {
		whereClause += " AND d.dev_status = ?"
		args = append(args, *devStatus)
	}
	if !utils.IsEmpty(permissionLevel) {
		whereClause += " AND du.permission_level = ?"
		args = append(args, permissionLevel)
	}
	if isActive != nil {
		whereClause += " AND du.is_active = ?"
		args = append(args, *isActive)
	}

	// 查询总数
	var total int64
	countQuery := `SELECT COUNT(*) FROM device_user du
		INNER JOIN device d ON du.dev_id = d.dev_id ` + whereClause
	err := mysql.MysqlCli.Client.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	offset := (page - 1) * pageSize
	query := `SELECT d.dev_id, d.dev_name, d.dev_type, d.dev_model, d.dev_power, d.dev_status,
		d.firmware_version, d.sampling_frequency, d.data_upload_interval, d.offline_threshold,
		d.extended_config, d.create_at, d.update_at
		FROM device_user du
		INNER JOIN device d ON du.dev_id = d.dev_id
		` + whereClause + `
		ORDER BY du.bound_at DESC
		LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := mysql.MysqlCli.Client.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var devices []*model.Device
	for rows.Next() {
		device := &model.Device{}
		var extendedConfigJSON sql.NullString
		err := rows.Scan(
			&device.DevID, &device.DevName, &device.DevType, &device.DevModel,
			&device.DevPower, &device.DevStatus, &device.FirmwareVersion,
			&device.SamplingFrequency, &device.DataUploadInterval, &device.OfflineThreshold,
			&extendedConfigJSON, &device.CreateAt, &device.UpdateAt)
		if err != nil {
			return nil, 0, err
		}

		if extendedConfigJSON.Valid {
			// 解析JSON
			json.Unmarshal([]byte(extendedConfigJSON.String), &device.ExtendedConfig)
		}

		devices = append(devices, device)
	}

	return devices, total, nil
}

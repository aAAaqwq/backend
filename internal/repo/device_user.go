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
		"SELECT COUNT(*) FROM user_dev WHERE dev_id = ? AND uid = ?",
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
		"SELECT COUNT(*) FROM user_dev WHERE dev_id = ?",
		deviceUser.DevID).Scan(&userCount)
	if err != nil {
		return err
	}
	if userCount >= 3 {
		return errors.New("设备最多只能绑定3个用户")
	}

	// 插入绑定关系
	query := `INSERT INTO user_dev (uid, dev_id, permission_level, bind_at)
		VALUES (?, ?, ?, ?)`
	_, err = mysql.MysqlCli.Client.Exec(query,
		deviceUser.UID, deviceUser.DevID, deviceUser.PermissionLevel, deviceUser.BindAt)
	return err
}

// GetDeviceUsers 获取设备的绑定用户列表
func (r *DeviceUserRepository) GetDeviceUsers(devID int64) ([]*model.DeviceUserWithInfo, error) {
	query := `SELECT ud.uid, u.username, u.email, ud.permission_level, ud.bind_at
		FROM user_dev ud
		LEFT JOIN user u ON ud.uid = u.uid
		WHERE ud.dev_id = ?
		ORDER BY ud.bind_at DESC`

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
			&user.PermissionLevel, &user.BindAt)
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
	query := `SELECT uid, dev_id, permission_level, bind_at
		FROM user_dev WHERE dev_id = ? AND uid = ?`

	err := mysql.MysqlCli.Client.QueryRow(query, devID, uid).Scan(
		&deviceUser.UID, &deviceUser.DevID, &deviceUser.PermissionLevel, &deviceUser.BindAt)
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
	query := `UPDATE user_dev SET permission_level = ?
		WHERE dev_id = ? AND uid = ?`

	_, err := mysql.MysqlCli.Client.Exec(query,
		deviceUser.PermissionLevel, deviceUser.DevID, deviceUser.UID)
	return err
}

// UnbindDeviceUser 解绑用户设备
func (r *DeviceUserRepository) UnbindDeviceUser(devID, uid int64) error {
	_, err := mysql.MysqlCli.Client.Exec(
		"DELETE FROM user_dev WHERE dev_id = ? AND uid = ?", devID, uid)
	return err
}

// GetUserDevices 获取用户绑定的设备列表
func (r *DeviceUserRepository) GetUserDevices(uid int64, page, pageSize int, devType string, devStatus *int, permissionLevel string, isActive *bool) ([]*model.Device, int64, error) {
	whereClause := "WHERE ud.uid = ?"
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
		whereClause += " AND ud.permission_level = ?"
		args = append(args, permissionLevel)
	}
	// 注意：isActive参数已废弃，因为SQL表中没有该字段

	// 查询总数
	var total int64
	countQuery := `SELECT COUNT(*) FROM user_dev ud
		INNER JOIN device d ON ud.dev_id = d.dev_id ` + whereClause
	err := mysql.MysqlCli.Client.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	offset := (page - 1) * pageSize
	query := `SELECT d.dev_id, d.dev_name, d.dev_status, d.dev_type, d.dev_power,
		d.model, d.version, d.sampling_rate, d.offline_threshold, d.upload_interval,
		d.extended_config, d.create_at, d.update_at
		FROM user_dev ud
		INNER JOIN device d ON ud.dev_id = d.dev_id
		` + whereClause + `
		ORDER BY ud.bind_at DESC
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
			&device.DevID, &device.DevName, &device.DevStatus, &device.DevType,
			&device.DevPower, &device.Model, &device.Version,
			&device.SamplingRate, &device.OfflineThreshold, &device.UploadInterval,
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

// GetUserDeviceIDs 获取用户有权限的设备ID列表
func (r *DeviceUserRepository) GetUserDeviceIDs(uid int64) ([]int64, error) {
	query := `SELECT dev_id FROM user_dev WHERE uid = ?`
	rows, err := mysql.MysqlCli.Client.Query(query, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devIDs []int64
	for rows.Next() {
		var devID int64
		if err := rows.Scan(&devID); err != nil {
			return nil, err
		}
		devIDs = append(devIDs, devID)
	}

	return devIDs, nil
}

// GetUserDeviceStatistics 获取用户设备统计信息
func (r *DeviceUserRepository) GetUserDeviceStatistics(uid int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 查询用户绑定的设备总数
	var total int64
	err := mysql.MysqlCli.Client.QueryRow(`SELECT COUNT(*) FROM user_dev WHERE uid = ?`, uid).Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// 查询在线设备数量
	var online int64
	err = mysql.MysqlCli.Client.QueryRow(`SELECT COUNT(*) FROM user_dev ud
		INNER JOIN device d ON ud.dev_id = d.dev_id
		WHERE ud.uid = ? AND d.dev_status = ?`, uid, model.DevStatusOnline).Scan(&online)
	if err != nil {
		return nil, err
	}
	stats["online"] = online

	// 查询离线设备数量
	var offline int64
	err = mysql.MysqlCli.Client.QueryRow(`SELECT COUNT(*) FROM user_dev ud
		INNER JOIN device d ON ud.dev_id = d.dev_id
		WHERE ud.uid = ? AND d.dev_status = ?`, uid, model.DevStatusOffline).Scan(&offline)
	if err != nil {
		return nil, err
	}
	stats["offline"] = offline

	// 查询异常设备数量
	var abnormal int64
	err = mysql.MysqlCli.Client.QueryRow(`SELECT COUNT(*) FROM user_dev ud
		INNER JOIN device d ON ud.dev_id = d.dev_id
		WHERE ud.uid = ? AND d.dev_status = ?`, uid, model.DevStatusAbnormal).Scan(&abnormal)
	if err != nil {
		return nil, err
	}
	stats["abnormal"] = abnormal

	// 按类型统计
	typeStats := make(map[string]int64)
	rows, err := mysql.MysqlCli.Client.Query(`SELECT d.dev_type, COUNT(*) FROM user_dev ud
		INNER JOIN device d ON ud.dev_id = d.dev_id
		WHERE ud.uid = ?
		GROUP BY d.dev_type`, uid)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var devType string
			var count int64
			if err := rows.Scan(&devType, &count); err == nil {
				typeStats[devType] = count
			}
		}
	}
	stats["by_type"] = typeStats

	// 设备种类数
	stats["type"] = len(typeStats)

	return stats, nil
}

package repo

import (
	"backend/internal/db/mysql"
	"backend/internal/model"
	"backend/pkg/utils"
	"database/sql"
	"encoding/json"
)

type DeviceRepository struct{}

func NewDeviceRepository() *DeviceRepository {
	return &DeviceRepository{}
}

// CreateDevice 创建设备
func (r *DeviceRepository) CreateDevice(device *model.Device) error {
	extendedConfigJSON, _ := json.Marshal(device.ExtendedConfig)
	query := `INSERT INTO device (dev_id, dev_name, dev_type, dev_model, dev_power, dev_status, 
		firmware_version, sampling_frequency, data_upload_interval, offline_threshold, extended_config, create_at, update_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := mysql.MysqlCli.Client.Exec(query,
		device.DevID, device.DevName, device.DevType, device.DevModel, device.DevPower, device.DevStatus,
		device.FirmwareVersion, device.SamplingFrequency, device.DataUploadInterval, device.OfflineThreshold,
		string(extendedConfigJSON), device.CreateAt, device.UpdateAt)
	return err
}

// GetDevice 获取指定设备
func (r *DeviceRepository) GetDevice(devID int64) (*model.Device, error) {
	device := &model.Device{}
	var extendedConfigJSON sql.NullString

	query := `SELECT dev_id, dev_name, dev_type, dev_model, dev_power, dev_status, 
		firmware_version, sampling_frequency, data_upload_interval, offline_threshold, extended_config, create_at, update_at 
		FROM device WHERE dev_id = ?`

	err := mysql.MysqlCli.Client.QueryRow(query, devID).Scan(
		&device.DevID, &device.DevName, &device.DevType, &device.DevModel, &device.DevPower, &device.DevStatus,
		&device.FirmwareVersion, &device.SamplingFrequency, &device.DataUploadInterval, &device.OfflineThreshold,
		&extendedConfigJSON, &device.CreateAt, &device.UpdateAt)

	if err != nil {
		return nil, err
	}

	if extendedConfigJSON.Valid {
		json.Unmarshal([]byte(extendedConfigJSON.String), &device.ExtendedConfig)
	}

	return device, nil
}

// GetDevices 获取设备列表（分页）
func (r *DeviceRepository) GetDevices(page, pageSize int, devType string, devStatus *int, keyword, sortBy, sortOrder string) ([]*model.Device, int64, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if !utils.IsEmpty(devType) {
		whereClause += " AND dev_type = ?"
		args = append(args, devType)
	}
	if devStatus != nil {
		whereClause += " AND dev_status = ?"
		args = append(args, *devStatus)
	}
	if !utils.IsEmpty(keyword) {
		whereClause += " AND (dev_name LIKE ? OR dev_model LIKE ?)"
		keywordPattern := "%" + keyword + "%"
		args = append(args, keywordPattern, keywordPattern)
	}

	if utils.IsEmpty(sortBy) {
		sortBy = "create_at"
	}
	if utils.IsEmpty(sortOrder) {
		sortOrder = "DESC"
	}
	orderClause := "ORDER BY " + sortBy + " " + sortOrder

	offset := (page - 1) * pageSize
	limitClause := "LIMIT ? OFFSET ?"
	countArgs := args
	args = append(args, pageSize, offset)

	// 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM device " + whereClause
	err := mysql.MysqlCli.Client.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := `SELECT dev_id, dev_name, dev_type, dev_model, dev_power, dev_status, 
		firmware_version, sampling_frequency, data_upload_interval, offline_threshold, extended_config, create_at, update_at 
		FROM device ` + whereClause + " " + orderClause + " " + limitClause

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
			&device.DevID, &device.DevName, &device.DevType, &device.DevModel, &device.DevPower, &device.DevStatus,
			&device.FirmwareVersion, &device.SamplingFrequency, &device.DataUploadInterval, &device.OfflineThreshold,
			&extendedConfigJSON, &device.CreateAt, &device.UpdateAt)
		if err != nil {
			return nil, 0, err
		}

		if extendedConfigJSON.Valid {
			json.Unmarshal([]byte(extendedConfigJSON.String), &device.ExtendedConfig)
		}

		devices = append(devices, device)
	}

	return devices, total, nil
}

// UpdateDevice 更新设备
func (r *DeviceRepository) UpdateDevice(device *model.Device) error {
	extendedConfigJSON, _ := json.Marshal(device.ExtendedConfig)
	query := `UPDATE device SET dev_name = ?, dev_type = ?, dev_model = ?, dev_power = ?, dev_status = ?, 
		firmware_version = ?, sampling_frequency = ?, data_upload_interval = ?, offline_threshold = ?, 
		extended_config = ?, update_at = ? WHERE dev_id = ?`

	_, err := mysql.MysqlCli.Client.Exec(query,
		device.DevName, device.DevType, device.DevModel, device.DevPower, device.DevStatus,
		device.FirmwareVersion, device.SamplingFrequency, device.DataUploadInterval, device.OfflineThreshold,
		string(extendedConfigJSON), device.UpdateAt, device.DevID)
	return err
}

// DeleteDevice 删除设备
func (r *DeviceRepository) DeleteDevice(devID int64) error {
	_, err := mysql.MysqlCli.Client.Exec("DELETE FROM device WHERE dev_id = ?", devID)
	return err
}

// GetDeviceStatistics 获取设备统计信息
func (r *DeviceRepository) GetDeviceStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 查询总数
	var total int64
	err := mysql.MysqlCli.Client.QueryRow("SELECT COUNT(*) FROM device").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// 查询在线数量
	var online int64
	err = mysql.MysqlCli.Client.QueryRow("SELECT COUNT(*) FROM device WHERE dev_status = ?", model.DevStatusOnline).Scan(&online)
	if err != nil {
		return nil, err
	}
	stats["online"] = online

	// 查询离线数量
	var offline int64
	err = mysql.MysqlCli.Client.QueryRow("SELECT COUNT(*) FROM device WHERE dev_status = ?", model.DevStatusOffline).Scan(&offline)
	if err != nil {
		return nil, err
	}
	stats["offline"] = offline

	// 查询异常数量
	var abnormal int64
	err = mysql.MysqlCli.Client.QueryRow("SELECT COUNT(*) FROM device WHERE dev_status = ?", model.DevStatusAbnormal).Scan(&abnormal)
	if err != nil {
		return nil, err
	}
	stats["abnormal"] = abnormal

	// 按类型统计
	typeStats := make(map[string]int64)
	rows, err := mysql.MysqlCli.Client.Query("SELECT dev_type, COUNT(*) FROM device GROUP BY dev_type")
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

	return stats, nil
}

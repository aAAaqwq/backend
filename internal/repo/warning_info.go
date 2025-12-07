package repo

import (
	"backend/internal/db/mysql"
	"backend/internal/model"
	"backend/pkg/utils"
	"database/sql"
	"time"
)

type WarningInfoRepository struct{}

func NewWarningInfoRepository() *WarningInfoRepository {
	return &WarningInfoRepository{}
}

// CreateWarningInfo 创建告警信息
func (r *WarningInfoRepository) CreateWarningInfo(warning *model.WarningInfo) error {
	query := `INSERT INTO alert_event (alert_id, data_id, dev_id, alert_type, alert_message, alert_status, triggered_at, resolved_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := mysql.MysqlCli.Client.Exec(query,
		warning.AlertID, warning.DataID, warning.DevID, warning.AlertType, warning.AlertMessage, warning.AlertStatus,
		warning.TriggeredAt, warning.ResolvedAt)
	return err
}

// GetWarningInfo 获取告警信息
func (r *WarningInfoRepository) GetWarningInfo(alertID int64) (*model.WarningInfo, error) {
	warning := &model.WarningInfo{}
	var resolvedAt sql.NullTime

	query := `SELECT alert_id, data_id, dev_id, alert_type, alert_message, alert_status, triggered_at, resolved_at
		FROM alert_event WHERE alert_id = ?`

	err := mysql.MysqlCli.Client.QueryRow(query, alertID).Scan(
		&warning.AlertID, &warning.DataID, &warning.DevID, &warning.AlertType, &warning.AlertMessage, &warning.AlertStatus,
		&warning.TriggeredAt, &resolvedAt)

	if err != nil {
		return nil, err
	}

	// 处理可空的 resolved_at
	if resolvedAt.Valid {
		warning.ResolvedAt = &resolvedAt.Time
	}

	return warning, nil
}

// GetWarningInfoList 获取告警信息列表
func (r *WarningInfoRepository) GetWarningInfoList(page, pageSize int, alertType, alertStatus string, devID, dataID *int64, devIDs []int64) ([]*model.WarningInfo, int64, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if !utils.IsEmpty(alertType) {
		whereClause += " AND alert_type = ?"
		args = append(args, alertType)
	}
	if !utils.IsEmpty(alertStatus) {
		whereClause += " AND alert_status = ?"
		args = append(args, alertStatus)
	}
	if devID != nil {
		whereClause += " AND dev_id = ?"
		args = append(args, *devID)
	}
	if dataID != nil {
		whereClause += " AND data_id = ?"
		args = append(args, *dataID)
	}
	// 添加设备ID列表过滤（用于普通用户只能查看有权限的设备）
	if len(devIDs) > 0 {
		placeholders := ""
		for i := range devIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, devIDs[i])
		}
		whereClause += " AND dev_id IN (" + placeholders + ")"
	}

	offset := (page - 1) * pageSize
	limitClause := "LIMIT ? OFFSET ?"
	countArgs := args
	args = append(args, pageSize, offset)

	// 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM alert_event " + whereClause
	err := mysql.MysqlCli.Client.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := `SELECT alert_id, data_id, dev_id, alert_type, alert_message, alert_status, triggered_at, resolved_at
		FROM alert_event ` + whereClause + " ORDER BY triggered_at DESC " + limitClause

	rows, err := mysql.MysqlCli.Client.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var warnings []*model.WarningInfo
	for rows.Next() {
		warning := &model.WarningInfo{}
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&warning.AlertID, &warning.DataID, &warning.DevID, &warning.AlertType, &warning.AlertMessage, &warning.AlertStatus,
			&warning.TriggeredAt, &resolvedAt)
		if err != nil {
			return nil, 0, err
		}

		// 处理可空的 resolved_at
		if resolvedAt.Valid {
			warning.ResolvedAt = &resolvedAt.Time
		}

		warnings = append(warnings, warning)
	}

	return warnings, total, nil
}

// UpdateWarningInfo 更新告警信息
func (r *WarningInfoRepository) UpdateWarningInfo(warning *model.WarningInfo) error {
	query := `UPDATE alert_event SET data_id = ?, dev_id = ?, alert_type = ?, alert_message = ?, alert_status = ?, resolved_at = ?
		WHERE alert_id = ?`

	_, err := mysql.MysqlCli.Client.Exec(query,
		warning.DataID, warning.DevID, warning.AlertType, warning.AlertMessage, warning.AlertStatus, warning.ResolvedAt, warning.AlertID)
	return err
}

// UpdateWarningStatus 更新告警状态
func (r *WarningInfoRepository) UpdateWarningStatus(alertID int64, alertStatus string, resolvedAt *time.Time) error {
	query := `UPDATE alert_event SET alert_status = ?, resolved_at = ? WHERE alert_id = ?`
	_, err := mysql.MysqlCli.Client.Exec(query, alertStatus, resolvedAt, alertID)
	return err
}

// DeleteWarningInfo 删除告警信息
func (r *WarningInfoRepository) DeleteWarningInfo(alertID int64) error {
	_, err := mysql.MysqlCli.Client.Exec("DELETE FROM alert_event WHERE alert_id = ?", alertID)
	return err
}

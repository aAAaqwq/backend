package repo

import (
	"backend/internal/db/mysql"
	"backend/internal/model"
)

type LogRepository struct{}

func NewLogRepository() *LogRepository {
	return &LogRepository{}
}

// CreateLog 创建日志
func (r *LogRepository) CreateLog(log *model.Log) error {
	query := `INSERT INTO system_log (log_id, type, level, message, user_agent, create_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err := mysql.MysqlCli.Client.Exec(query,
		log.LogID, log.Type, log.Level, log.Message, log.UserAgent, log.CreateAt)
	return err
}

// GetLog 根据ID获取日志
func (r *LogRepository) GetLog(logID int64) (*model.Log, error) {
	log := &model.Log{}
	query := `SELECT log_id, type, level, message, user_agent, create_at
		FROM system_log WHERE log_id = ?`

	err := mysql.MysqlCli.Client.QueryRow(query, logID).Scan(
		&log.LogID, &log.Type, &log.Level, &log.Message, &log.UserAgent, &log.CreateAt)
	if err != nil {
		return nil, err
	}

	return log, nil
}

// GetLogs 查询日志列表
func (r *LogRepository) GetLogs(logType string, level *int, startTime, endTime string) ([]*model.Log, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if logType != "" {
		whereClause += " AND type = ?"
		args = append(args, logType)
	}
	if level != nil {
		whereClause += " AND level = ?"
		args = append(args, *level)
	}
	if startTime != "" {
		whereClause += " AND create_at >= ?"
		args = append(args, startTime)
	}
	if endTime != "" {
		whereClause += " AND create_at <= ?"
		args = append(args, endTime)
	}

	query := `SELECT log_id, type, level, message, user_agent, create_at
		FROM system_log ` + whereClause + " ORDER BY create_at DESC LIMIT 100"

	rows, err := mysql.MysqlCli.Client.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*model.Log
	for rows.Next() {
		log := &model.Log{}
		err := rows.Scan(
			&log.LogID, &log.Type, &log.Level, &log.Message, &log.UserAgent, &log.CreateAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// DeleteLog 删除日志
func (r *LogRepository) DeleteLog(logID int64) error {
	_, err := mysql.MysqlCli.Client.Exec("DELETE FROM system_log WHERE log_id = ?", logID)
	return err
}

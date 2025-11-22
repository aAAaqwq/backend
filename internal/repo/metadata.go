package repo

import (
	"backend/internal/db/mysql"
	"backend/internal/model"
	"backend/pkg/utils"
	"database/sql"
	"encoding/json"
	"fmt"
)

type MetadataRepository struct{}

func NewMetadataRepository() *MetadataRepository {
	return &MetadataRepository{}
}

// CreateMetadata 创建元数据
func (r *MetadataRepository) CreateMetadata(metadata *model.Metadata) error {
	extraDataJSON, _ := json.Marshal(metadata.ExtraData)

	// 将timestamp字符串转换为int64
	timestamp, _ := utils.ConvertToInt64(metadata.Timestamp)
	if timestamp == 0 {
		timestamp = 0 // 如果转换失败，使用0，由service层处理
	}

	query := `INSERT INTO metadata (data_id, dev_id, uid, data_type, quality_score, extra_data, timestamp) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := mysql.MysqlCli.Client.Exec(query,
		metadata.DataID, metadata.DevID, metadata.UID, metadata.DataType, metadata.QualityScore,
		string(extraDataJSON), timestamp)
	return err
}

// GetMetadata 获取元数据
func (r *MetadataRepository) GetMetadata(dataID int64) (*model.Metadata, error) {
	metadata := &model.Metadata{}
	var extraDataJSON sql.NullString
	var timestampInt int64

	query := `SELECT data_id, dev_id, uid, data_type, quality_score, extra_data, timestamp 
		FROM metadata WHERE data_id = ?`

	err := mysql.MysqlCli.Client.QueryRow(query, dataID).Scan(
		&metadata.DataID, &metadata.DevID, &metadata.UID, &metadata.DataType,
		&metadata.QualityScore, &extraDataJSON, &timestampInt)

	if err != nil {
		return nil, err
	}

	// 将int64转换为字符串
	metadata.Timestamp = fmt.Sprintf("%d", timestampInt)

	if extraDataJSON.Valid {
		json.Unmarshal([]byte(extraDataJSON.String), &metadata.ExtraData)
	}

	return metadata, nil
}

// GetMetadataList 获取元数据列表（分页）
func (r *MetadataRepository) GetMetadataList(page, pageSize int, dataType string, startTime, endTime *int64,
	minQuality, maxQuality *float64, keyword, sortBy, sortOrder string, devID, uid int64) ([]*model.Metadata, int64, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if !utils.IsEmpty(dataType) {
		whereClause += " AND data_type = ?"
		args = append(args, dataType)
	}
	if devID > 0 {
		whereClause += " AND dev_id = ?"
		args = append(args, devID)
	}
	if uid > 0 {
		whereClause += " AND uid = ?"
		args = append(args, uid)
	}
	if startTime != nil {
		whereClause += " AND timestamp >= ?"
		args = append(args, *startTime)
	}
	if endTime != nil {
		whereClause += " AND timestamp <= ?"
		args = append(args, *endTime)
	}
	if minQuality != nil {
		whereClause += " AND quality_score >= ?"
		args = append(args, *minQuality)
	}
	if maxQuality != nil {
		whereClause += " AND quality_score <= ?"
		args = append(args, *maxQuality)
	}
	if !utils.IsEmpty(keyword) {
		whereClause += " AND (file_name LIKE ? OR file_type LIKE ?)"
		keywordPattern := "%" + keyword + "%"
		args = append(args, keywordPattern, keywordPattern)
	}

	if utils.IsEmpty(sortBy) {
		sortBy = "timestamp"
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
	countQuery := "SELECT COUNT(*) FROM metadata " + whereClause
	err := mysql.MysqlCli.Client.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := `SELECT data_id, dev_id, uid, data_type, quality_score, extra_data, timestamp 
		FROM metadata ` + whereClause + " " + orderClause + " " + limitClause

	rows, err := mysql.MysqlCli.Client.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var metadataList []*model.Metadata
	for rows.Next() {
		metadata := &model.Metadata{}
		var extraDataJSON sql.NullString
		var timestampInt int64

		err := rows.Scan(
			&metadata.DataID, &metadata.DevID, &metadata.UID, &metadata.DataType,
			&metadata.QualityScore, &extraDataJSON, &timestampInt)
		if err != nil {
			return nil, 0, err
		}

		// 将int64转换为字符串
		metadata.Timestamp = fmt.Sprintf("%d", timestampInt)

		if extraDataJSON.Valid {
			json.Unmarshal([]byte(extraDataJSON.String), &metadata.ExtraData)
		}

		metadataList = append(metadataList, metadata)
	}

	return metadataList, total, nil
}

// DeleteMetadata 删除元数据
func (r *MetadataRepository) DeleteMetadata(dataID int64) error {
	_, err := mysql.MysqlCli.Client.Exec("DELETE FROM metadata WHERE data_id = ?", dataID)
	return err
}

// GetMetadataStatistics 获取元数据统计信息
func (r *MetadataRepository) GetMetadataStatistics(devID, uid int64, dataType string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 构建WHERE条件
	whereClause := "WHERE dev_id = ?"
	args := []interface{}{devID}

	if uid > 0 {
		whereClause += " AND uid = ?"
		args = append(args, uid)
	}

	if !utils.IsEmpty(dataType) {
		whereClause += " AND data_type = ?"
		args = append(args, dataType)
	}

	// 查询总数
	var total int64
	query := "SELECT COUNT(*) FROM metadata " + whereClause
	err := mysql.MysqlCli.Client.QueryRow(query, args...).Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// 查询异常数据数量（quality_score < 30）
	var abnormal int64
	query = "SELECT COUNT(*) FROM metadata " + whereClause + " AND quality_score < 30"
	err = mysql.MysqlCli.Client.QueryRow(query, args...).Scan(&abnormal)
	if err != nil {
		return nil, err
	}
	stats["abnormal"] = abnormal

	return stats, nil
}

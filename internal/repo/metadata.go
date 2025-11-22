package repo

import (
	"backend/internal/db/mysql"
	"backend/internal/model"
	"backend/pkg/utils"
	"database/sql"
	"encoding/json"
)

type MetadataRepository struct{}

func NewMetadataRepository() *MetadataRepository {
	return &MetadataRepository{}
}

// CreateMetadata 创建元数据
func (r *MetadataRepository) CreateMetadata(metadata *model.Metadata) error {
	extraDataJSON, _ := json.Marshal(metadata.ExtraData)
	query := `INSERT INTO metadata (data_id, dev_id, uid, data_type, extra_data, timestamp) 
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err := mysql.MysqlCli.Client.Exec(query,
		metadata.DataID, metadata.DevID, metadata.UID, metadata.DataType,
		string(extraDataJSON), metadata.Timestamp)
	return err
}

// GetMetadata 获取元数据
func (r *MetadataRepository) GetMetadata(dataID int64) (*model.Metadata, error) {
	metadata := &model.Metadata{}
	var extraDataJSON sql.NullString

	query := `SELECT data_id, dev_id, uid, data_type, extra_data, timestamp 
		FROM metadata WHERE data_id = ?`

	err := mysql.MysqlCli.Client.QueryRow(query, dataID).Scan(
		&metadata.DataID, &metadata.DevID, &metadata.UID, &metadata.DataType, &extraDataJSON, &metadata.Timestamp)

	if err != nil {
		return nil, err
	}

	if extraDataJSON.Valid {
		json.Unmarshal([]byte(extraDataJSON.String), &metadata.ExtraData)
	}

	return metadata, nil
}

// GetMetadataList 获取元数据列表（分页）
func (r *MetadataRepository) GetMetadataList(page, pageSize int, dataType string, startTime, endTime *int64,
	minQuality, maxQuality *float64, keyword, sortBy, sortOrder string) ([]*model.Metadata, int64, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if !utils.IsEmpty(dataType) {
		whereClause += " AND data_type = ?"
		args = append(args, dataType)
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
	query := `SELECT data_id, dev_id, uid, data_type, storage_route, quality_score, extra_data, timestamp, file_path 
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

		err := rows.Scan(
			&metadata.DataID, &metadata.DevID, &metadata.UID, &metadata.DataType,
			&extraDataJSON, &metadata.Timestamp)
		if err != nil {
			return nil, 0, err
		}

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
func (r *MetadataRepository) GetMetadataStatistics(devID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 查询总数
	var total int64
	query := "SELECT COUNT(*) FROM metadata WHERE dev_id = ?"
	err := mysql.MysqlCli.Client.QueryRow(query, devID).Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// 查询异常数据数量（quality_score < 30）
	var abnormal int64
	query = "SELECT COUNT(*) FROM metadata WHERE dev_id = ? AND quality_score < 30"
	err = mysql.MysqlCli.Client.QueryRow(query, devID).Scan(&abnormal)
	if err != nil {
		return nil, err
	}
	stats["abnormal"] = abnormal

	return stats, nil
}

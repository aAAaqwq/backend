package repo

import (
	"backend/internal/db/mysql"
	"backend/internal/model"
	"backend/pkg/utils"
	"database/sql"
	"encoding/json"
	"strconv"
)

type MetadataRepository struct{}

func NewMetadataRepository() *MetadataRepository {
	return &MetadataRepository{}
}

// CreateMetadata 创建元数据
func (r *MetadataRepository) CreateMetadata(metadata *model.Metadata) error {
	extraDataJSON, _ := json.Marshal(metadata.ExtraData)

	query := `INSERT INTO metadata (data_id, dev_id, data_type, quality_score, extra_data, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err := mysql.MysqlCli.Client.Exec(query,
		metadata.DataID, metadata.DevID, metadata.DataType, metadata.QualityScore,
		string(extraDataJSON), metadata.Timestamp)
	return err
}

// GetMetadata 获取元数据
func (r *MetadataRepository) GetMetadata(dataID int64) (*model.Metadata, error) {
	metadata := &model.Metadata{}
	var extraDataJSON sql.NullString

	query := `SELECT data_id, dev_id, data_type, quality_score, extra_data, timestamp
		FROM metadata WHERE data_id = ?`

	err := mysql.MysqlCli.Client.QueryRow(query, dataID).Scan(
		&metadata.DataID, &metadata.DevID, &metadata.DataType,
		&metadata.QualityScore, &extraDataJSON, &metadata.Timestamp)

	if err != nil {
		return nil, err
	}

	if extraDataJSON.Valid {
		json.Unmarshal([]byte(extraDataJSON.String), &metadata.ExtraData)
	}

	return metadata, nil
}

// GetMetadataList 获取元数据列表（分页）
// 注意：uid参数已废弃，但保留以兼容旧代码
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
	// uid 参数已废弃（SQL表中没有uid字段）

	if startTime != nil {
		whereClause += " AND timestamp >= FROM_UNIXTIME(?)"
		args = append(args, *startTime)
	}
	if endTime != nil {
		whereClause += " AND timestamp <= FROM_UNIXTIME(?)"
		args = append(args, *endTime)
	}
	// quality_score 现在是 varchar 类型，需要转换为数字进行比较
	if minQuality != nil {
		whereClause += " AND CAST(quality_score AS DECIMAL(5,2)) >= ?"
		args = append(args, *minQuality)
	}
	if maxQuality != nil {
		whereClause += " AND CAST(quality_score AS DECIMAL(5,2)) <= ?"
		args = append(args, *maxQuality)
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
	query := `SELECT data_id, dev_id, data_type, quality_score, extra_data, timestamp
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
			&metadata.DataID, &metadata.DevID, &metadata.DataType,
			&metadata.QualityScore, &extraDataJSON, &metadata.Timestamp)
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

// DeleteMetadataByDataType 删除指定 data_type 的所有元数据
func (r *MetadataRepository) DeleteMetadataByDataType(dataType string) error {
	_, err := mysql.MysqlCli.Client.Exec("DELETE FROM metadata WHERE data_type = ?", dataType)
	return err
}

// DeleteMetadataByDevIDAndTimeRange 根据设备ID和时间范围删除元数据
func (r *MetadataRepository) DeleteMetadataByDevIDAndTimeRange(devID int64, startTime, endTime *int64) error {
	query := "DELETE FROM metadata WHERE dev_id = ?"
	args := []interface{}{devID}

	if startTime != nil {
		query += " AND timestamp >= FROM_UNIXTIME(?)"
		args = append(args, *startTime)
	}
	if endTime != nil {
		query += " AND timestamp <= FROM_UNIXTIME(?)"
		args = append(args, *endTime)
	}

	_, err := mysql.MysqlCli.Client.Exec(query, args...)
	return err
}

// GetMetadataStatistics 获取元数据统计信息
// 注意：uid参数已废弃，但保留以兼容旧代码
func (r *MetadataRepository) GetMetadataStatistics(devID, uid int64, dataType string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 构建WHERE条件
	whereClause := "WHERE dev_id = ?"
	args := []interface{}{devID}

	// uid 参数已废弃（SQL表中没有uid字段）

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
	// quality_score 现在是 varchar 类型，需要转换
	var abnormal int64
	query = "SELECT COUNT(*) FROM metadata " + whereClause + " AND CAST(quality_score AS DECIMAL(5,2)) < 30"
	err = mysql.MysqlCli.Client.QueryRow(query, args...).Scan(&abnormal)
	if err != nil {
		// 如果转换失败，使用字符串比较（不太准确但作为fallback）
		query = "SELECT COUNT(*) FROM metadata " + whereClause + " AND quality_score < '30'"
		mysql.MysqlCli.Client.QueryRow(query, args...).Scan(&abnormal)
	}
	stats["abnormal"] = abnormal

	return stats, nil
}

// ParseQualityScore 解析quality_score字符串为float64
func ParseQualityScore(qualityScoreStr string) float64 {
	score, err := strconv.ParseFloat(qualityScoreStr, 64)
	if err != nil {
		return 0
	}
	return score
}

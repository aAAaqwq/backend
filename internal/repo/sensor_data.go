package repo

import (
	"backend/internal/db/influxdb"
	"backend/internal/db/minio"
	"backend/internal/model"
	"context"
	"fmt"
	"time"
)

const (
	PresignedURLExpiry = 30 * time.Minute
)

type SensorDataRepository struct{}

func NewSensorDataRepository() *SensorDataRepository {
	return &SensorDataRepository{}
}

// FPutFile 上传文件到MinIO
func (r *SensorDataRepository) FPutFile(bucketName, objectName, filePath, contentType string) error {
	if minio.MinIOCli == nil {
		return nil // MinIO未初始化时返回nil
	}
	err := minio.MinIOCli.UploadFile(bucketName, objectName, filePath, contentType)
	return err
}

// DownloadFile 从MinIO下载文件到本地
func (r *SensorDataRepository) DownloadFile(bucketName, objectName, filePath string) (string, error) {
	if minio.MinIOCli == nil {
		return "", fmt.Errorf("MinIO客户端未初始化")
	}
	presignedURL, err := minio.MinIOCli.DownloadFile(bucketName, objectName, filePath)
	return presignedURL, err
}

// GetFileLists 获取文件列表
func (r *SensorDataRepository) GetFileLists(bucketName string) ([]model.FileList, error) {
	if minio.MinIOCli == nil {
		return nil, nil // MinIO未初始化时返回nil
	}
	fileLists := make([]model.FileList, 0)
	objectsChan := minio.MinIOCli.ListObjects(bucketName, "", true)
	for object := range objectsChan {
		url, err := minio.MinIOCli.Client.PresignedGetObject(context.Background(),
			bucketName, object.Key, PresignedURLExpiry, nil)
		if err != nil {
			return nil, err
		}
		// 转换时间格式为字符串
		lastModifiedStr := object.LastModified.Format(time.RFC3339)

		fileLists = append(fileLists, model.FileList{
			BucketKey:    object.Key,
			Name:         object.Key,
			PreviewUrl:   url.String(),
			ContentType:  object.ContentType,
			LastModified: lastModifiedStr,
			Size:         object.Size, // int64类型
		})
	}
	return fileLists, nil
}

// CreateSeriesData 创建时间序列数据
func (r *SensorDataRepository) CreateSeriesData(seriesData *model.SeriesData) error {
	if influxdb.InfluxDBCli == nil {
		return fmt.Errorf("InfluxDB客户端未初始化")
	}

	// 直接使用seriesData.Points
	return influxdb.InfluxDBCli.WritePoints(seriesData.Points)
}

// DeleteObject 从MinIO删除文件
func (r *SensorDataRepository) DeleteObject(bucketName, objectName string) error {
	if minio.MinIOCli == nil {
		return fmt.Errorf("MinIO客户端未初始化")
	}
	return minio.MinIOCli.DeleteObject(bucketName, objectName)
}

// QuerySeriesData 查询时序数据
func (r *SensorDataRepository) QuerySeriesData(measurement string, devID int64, startTime, endTime int64,
	tags map[string]string, fields map[string]interface{}, downSampleInterval string, aggregate string, limitPoints int) ([]model.Point, error) {
	if influxdb.InfluxDBCli == nil {
		return nil, fmt.Errorf("InfluxDB客户端未初始化")
	}

	// 使用传入的tags（应该已经包含dev_id）
	queryTags := tags
	if queryTags == nil {
		queryTags = make(map[string]string)
	}
	// 确保包含dev_id（如果传入的tags中没有，则添加）
	if _, exists := queryTags["dev_id"]; !exists {
		queryTags["dev_id"] = fmt.Sprintf("%d", devID)
	}

	// 构建时间范围
	timeRange := influxdb.TimeRange{
		Start: time.Unix(startTime, 0),
		End:   time.Unix(endTime, 0),
	}

	// 解析下采样间隔
	var downSampleDuration time.Duration
	if downSampleInterval != "" {
		var err error
		downSampleDuration, err = time.ParseDuration(downSampleInterval)
		if err != nil {
			return nil, fmt.Errorf("下采样间隔格式错误: %v", err)
		}
	}

	// 将fields转换为[]string（如果不为nil）
	var fieldNames []string
	if fields != nil {
		for fieldName := range fields {
			fieldNames = append(fieldNames, fieldName)
		}
	}

	// 构建查询选项
	opts := influxdb.QueryOptions{
		Measurement:     measurement,
		Tags:            queryTags,
		Fields:          fieldNames,
		TimeRange:       timeRange,
		DownsampleEvery: downSampleDuration,
		Aggregate:       aggregate,
		LimitPoints:     limitPoints,
	}

	// 查询数据
	result, err := influxdb.InfluxDBCli.Query(opts)
	if err != nil {
		return nil, err
	}

	// 类型转换
	points, ok := result.([]model.Point)
	if !ok {
		return nil, fmt.Errorf("查询结果类型错误")
	}

	return points, nil
}

// GetSeriesDataStatistics 获取时序数据统计信息
func (r *SensorDataRepository) GetSeriesDataStatistics(measurement string, devID int64) (map[string]interface{}, error) {
	if influxdb.InfluxDBCli == nil {
		return nil, fmt.Errorf("InfluxDB客户端未初始化")
	}
	return influxdb.InfluxDBCli.GetSeriesDataStatistics(measurement, devID)
}

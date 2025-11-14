package service

import (
	"backend/internal/db/influxdb"
	"backend/internal/db/minio"
	"backend/internal/model"
	"backend/internal/repo"
	"backend/pkg/logger"
	"backend/pkg/utils"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SensorDataService struct {
	metadataRepo *repo.MetadataRepository
}

func NewSensorDataService() *SensorDataService {
	return &SensorDataService{metadataRepo: repo.NewMetadataRepository()}
}

// CreateSensorData 创建传感器数据
func (s *SensorDataService) CreateSensorData(metadata *model.Metadata) (*model.Metadata, error) {
	// 生成数据ID
	metadata.DataID = utils.GetDefaultSnowflake().Generate()

	// 根据数据类型选择存储方式
	switch metadata.DataType {
	case model.DataTypeSeries:
		// 存储到InfluxDB
		if influxdb.InfluxDBCli == nil {
			return nil, errors.New("InfluxDB客户端未初始化")
		}

		// 创建InfluxDB客户端包装
		influxClient := &influxdb.InfluxDBClient{Client: influxdb.InfluxDBCli}

		// 准备InfluxDB数据
		tags := map[string]string{
			"dev_id": fmt.Sprintf("%d", metadata.DevID),
			"uid":    fmt.Sprintf("%d", metadata.UID),
		}

		fields := map[string]interface{}{
			"quality_score": metadata.QualityScore,
			"data_id":       metadata.DataID,
		}

		// 添加额外数据到fields
		if metadata.ExtraData != nil {
			for k, v := range metadata.ExtraData {
				fields[k] = v
			}
		}

		// 转换时间戳为纳秒
		timestamp := metadata.Timestamp * 1000000000

		// 写入时序数据到InfluxDB
		err := influxClient.WritePoint("sensor_data", tags, fields, timestamp)
		if err != nil {
			logger.L().Error("写入InfluxDB失败", logger.WithError(err))
			return nil, fmt.Errorf("写入InfluxDB失败: %v", err)
		}

		metadata.StorageRoute = "influxdb"

	case model.DataTypeVideo, model.DataTypeAudio, model.DataTypeImage:
		// 存储到MinIO
		if minio.MinIOCli == nil {
			return nil, errors.New("MinIO客户端未初始化")
		}

		if metadata.FilePath == "" {
			return nil, errors.New("非结构化数据需要提供文件路径")
		}

		// 检查文件是否存在
		if _, err := os.Stat(metadata.FilePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("文件不存在: %s", metadata.FilePath)
		}

		// 根据数据类型选择bucket
		bucketName := minio.GetBucketName(metadata.DataType)

		// 生成对象名称（使用数据ID和文件名）
		fileName := filepath.Base(metadata.FilePath)
		objectName := fmt.Sprintf("%d/%d/%s", metadata.DevID, metadata.UID, fileName)
		// 如果文件路径已包含时间戳，可以添加时间戳前缀
		if metadata.Timestamp > 0 {
			ext := filepath.Ext(fileName)
			nameWithoutExt := fileName[:len(fileName)-len(ext)]
			objectName = fmt.Sprintf("%d/%d/%s_%d%s", metadata.DevID, metadata.UID, nameWithoutExt, metadata.Timestamp, ext)
		}

		// 上传文件到MinIO
		err := minio.MinIOCli.PutObject(bucketName, objectName, metadata.FilePath)
		if err != nil {
			logger.L().Error("上传文件到MinIO失败", logger.WithError(err))
			return nil, fmt.Errorf("上传文件到MinIO失败: %v", err)
		}

		// 更新文件路径为MinIO中的路径
		metadata.FilePath = fmt.Sprintf("%s/%s", bucketName, objectName)
		metadata.StorageRoute = "minio"

	default:
		return nil, errors.New("不支持的数据类型")
	}

	// 保存元数据到MySQL
	err := s.metadataRepo.CreateMetadata(metadata)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// GetSensorData 获取传感器数据列表
func (s *SensorDataService) GetSensorData(page, pageSize int, dataType string, startTime, endTime *int64,
	minQuality, maxQuality *float64, keyword, sortBy, sortOrder string) ([]*model.Metadata, int64, error) {
	return s.metadataRepo.GetMetadataList(page, pageSize, dataType, startTime, endTime, minQuality, maxQuality, keyword, sortBy, sortOrder)
}

// GetSensorDataStatistic 获取传感器数据统计
func (s *SensorDataService) GetSensorDataStatistic(devID int64) (map[string]interface{}, error) {
	return s.metadataRepo.GetMetadataStatistics(devID)
}

// DeleteSensorData 删除传感器数据
func (s *SensorDataService) DeleteSensorData(dataID int64, dataType string) error {
	// 获取元数据
	metadata, err := s.metadataRepo.GetMetadata(dataID)
	if err != nil {
		return err
	}

	// 根据存储类型删除数据
	switch metadata.StorageRoute {
	case "influxdb":
		if influxdb.InfluxDBCli == nil {
			return errors.New("InfluxDB客户端未初始化")
		}

		// 创建InfluxDB客户端包装
		influxClient := &influxdb.InfluxDBClient{Client: influxdb.InfluxDBCli}

		// 从InfluxDB删除数据
		err = influxClient.DeleteSensorDataByDataID(metadata.DevID, metadata.Timestamp)
		if err != nil {
			logger.L().Error("从InfluxDB删除数据失败", logger.WithError(err))
			return fmt.Errorf("从InfluxDB删除数据失败: %v", err)
		}

	case "minio":
		if minio.MinIOCli == nil {
			return errors.New("MinIO客户端未初始化")
		}

		// 从MinIO删除文件
		// 解析文件路径：bucket/object_name
		parts := strings.SplitN(metadata.FilePath, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("无效的文件路径格式: %s", metadata.FilePath)
		}

		bucketName := parts[0]
		objectName := parts[1]

		err = minio.MinIOCli.DeleteObject(bucketName, objectName)
		if err != nil {
			logger.L().Error("从MinIO删除文件失败", logger.WithError(err))
			return fmt.Errorf("从MinIO删除文件失败: %v", err)
		}
	}

	// 删除元数据
	return s.metadataRepo.DeleteMetadata(dataID)
}

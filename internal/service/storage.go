package service

import (
	"backend/internal/db/influxdb"
	"backend/internal/db/minio"
	"backend/pkg/logger"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StorageService struct{}

func NewStorageService() *StorageService {
	return &StorageService{}
}

// GetSensorDataFromInfluxDB 从InfluxDB读取传感器时序数据
func (s *StorageService) GetSensorDataFromInfluxDB(devID int64, startTime, endTime *int64, limit int) ([]map[string]interface{}, error) {
	if influxdb.InfluxDBCli == nil {
		return nil, errors.New("InfluxDB客户端未初始化")
	}

	influxClient := &influxdb.InfluxDBClient{Client: influxdb.InfluxDBCli}

	result, err := influxClient.QuerySensorData(devID, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("查询InfluxDB数据失败: %v", err)
	}

	// 解析查询结果
	// InfluxDB3的QueryResult需要根据实际API进行解析
	// 这里返回空列表，实际使用时需要根据influxdb3-go的API文档进行实现
	var dataList []map[string]interface{}

	// TODO: 根据influxdb3-go的实际API解析result
	// 示例代码（需要根据实际API调整）:
	// if result != nil {
	//     // 解析result中的数据
	// }

	logger.L().Info("查询InfluxDB数据", logger.WithAny("result", result))

	return dataList, nil
}

// DownloadFileFromMinIO 从MinIO下载文件到本地
func (s *StorageService) DownloadFileFromMinIO(bucketName, objectName, localPath string) error {
	if minio.MinIOCli == nil {
		return errors.New("MinIO客户端未初始化")
	}

	// 确保本地目录存在
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	return minio.MinIOCli.GetObject(bucketName, objectName, localPath)
}

// GetFileFromMinIO 从MinIO获取文件作为Reader
func (s *StorageService) GetFileFromMinIO(bucketName, objectName string) (io.ReadCloser, error) {
	if minio.MinIOCli == nil {
		return nil, errors.New("MinIO客户端未初始化")
	}

	return minio.MinIOCli.GetObjectAsReader(bucketName, objectName)
}

// GetFileInfoFromMinIO 从MinIO获取文件信息
func (s *StorageService) GetFileInfoFromMinIO(bucketName, objectName string) (map[string]interface{}, error) {
	if minio.MinIOCli == nil {
		return nil, errors.New("MinIO客户端未初始化")
	}

	info, err := minio.MinIOCli.GetObjectInfo(bucketName, objectName)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"size":          info.Size,
		"content_type":  info.ContentType,
		"last_modified": info.LastModified,
		"etag":          info.ETag,
	}, nil
}

// ListFilesFromMinIO 列出MinIO bucket中的文件
func (s *StorageService) ListFilesFromMinIO(bucketName, prefix string, recursive bool) ([]map[string]interface{}, error) {
	if minio.MinIOCli == nil {
		return nil, errors.New("MinIO客户端未初始化")
	}

	var files []map[string]interface{}
	for objInfo := range minio.MinIOCli.ListObjects(bucketName, prefix, recursive) {
		if objInfo.Err != nil {
			logger.L().Error("列出文件时出错", logger.WithError(objInfo.Err))
			continue
		}

		files = append(files, map[string]interface{}{
			"name":          objInfo.Key,
			"size":          objInfo.Size,
			"content_type":  objInfo.ContentType,
			"last_modified": objInfo.LastModified,
			"etag":          objInfo.ETag,
		})
	}

	return files, nil
}

// GeneratePresignedURL 生成MinIO文件的预签名URL（用于临时访问）
func (s *StorageService) GeneratePresignedURL(bucketName, objectName string, expiry time.Duration) (string, error) {
	if minio.MinIOCli == nil {
		return "", errors.New("MinIO客户端未初始化")
	}

	ctx := context.Background()
	url, err := minio.MinIOCli.Client.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("生成预签名URL失败: %v", err)
	}

	return url.String(), nil
}

// UploadFileToMinIO 上传文件到MinIO（从本地文件路径）
func (s *StorageService) UploadFileToMinIO(bucketName, objectName, filePath string) error {
	if minio.MinIOCli == nil {
		return errors.New("MinIO客户端未初始化")
	}

	return minio.MinIOCli.PutObject(bucketName, objectName, filePath)
}

// UploadFileFromReaderToMinIO 从Reader上传文件到MinIO
func (s *StorageService) UploadFileFromReaderToMinIO(bucketName, objectName string, reader io.Reader, size int64, contentType string) error {
	if minio.MinIOCli == nil {
		return errors.New("MinIO客户端未初始化")
	}

	return minio.MinIOCli.PutObjectFromReader(bucketName, objectName, reader, size, contentType)
}

// DeleteFileFromMinIO 从MinIO删除文件
func (s *StorageService) DeleteFileFromMinIO(bucketName, objectName string) error {
	if minio.MinIOCli == nil {
		return errors.New("MinIO客户端未初始化")
	}

	return minio.MinIOCli.DeleteObject(bucketName, objectName)
}

// ParseFilePath 解析文件路径（bucket/object_name格式）
func (s *StorageService) ParseFilePath(filePath string) (bucketName, objectName string, err error) {
	parts := strings.SplitN(filePath, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("无效的文件路径格式: %s，应为 bucket/object_name", filePath)
	}
	return parts[0], parts[1], nil
}

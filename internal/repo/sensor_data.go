package repo

import (
	"backend/internal/db/influxdb"
	"backend/internal/db/minio"
	"backend/internal/model"
	"context"
	"fmt"
	"time"
)
const(
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
	err := minio.MinIOCli.UploadFile(bucketName, objectName, filePath,contentType)
	return err
}

// DownloadFile 从MinIO下载文件到本地
func (r *SensorDataRepository) DownloadFile(bucketName, objectName, filePath string) (string, error) {
	if minio.MinIOCli == nil {
		return "", fmt.Errorf("MinIO客户端未初始化")
	}
	presignedURL,err := minio.MinIOCli.DownloadFile(bucketName, objectName, filePath)
	return presignedURL,err
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
		fileLists = append(fileLists, model.FileList{
			BucketKey: object.Key,
			Name: object.Key,
			PreviewUrl: url.String(),
			ContentType: object.ContentType,
			LastModified: object.LastModified,
			Size: object.Size,
		})
	}
	return fileLists, nil
}

// CreateSeriesData 创建时间序列数据
func (r *SensorDataRepository) CreateSeriesData(seriesData *model.SeriesData) error {
	ctx := context.Background()
	err := influxdb.InfluxDBCli.WritePoints()
	return err
}



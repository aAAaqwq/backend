package repo

import (
	"backend/internal/db/minio"
	"context"

	minioClient "github.com/minio/minio-go/v7"
)

type SensorDataRepository struct{}

func NewSensorDataRepository() *SensorDataRepository {
	return &SensorDataRepository{}
}

func (r *SensorDataRepository) FPutImage(bucketName, objectName, filePath string) error {
	if minio.MinIOCli == nil {
		return nil // MinIO未初始化时返回nil
	}
	_, err := minio.MinIOCli.Client.FPutObject(context.Background(), bucketName, objectName, filePath, minioClient.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	return err
}

func (r *SensorDataRepository) FPutVideo(bucketName, objectName, filePath string) error {
	if minio.MinIOCli == nil {
		return nil // MinIO未初始化时返回nil
	}
	_, err := minio.MinIOCli.Client.FPutObject(context.Background(), bucketName, objectName, filePath, minioClient.PutObjectOptions{
		ContentType: "video/mp4",
	})
	return err
}

func (r *SensorDataRepository) FPutAudio(bucketName, objectName, filePath string) error {
	if minio.MinIOCli == nil {
		return nil // MinIO未初始化时返回nil
	}
	_, err := minio.MinIOCli.Client.FPutObject(context.Background(), bucketName, objectName, filePath, minioClient.PutObjectOptions{
		ContentType: "audio/mpeg",
	})
	return err
}

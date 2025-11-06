package minio

import (
	"backend/config"
	"backend/pkg/logger"
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient MinIO客户端结构
type MinIOClient struct {
	Client *minio.Client
}

// GetMinIOClient 获取MinIO客户端
// 使用 config.MinIOConfig 作为参数类型
func GetMinIOClient(cfg config.MinIOConfig) (*MinIOClient, error) {
	client, err := InitMinIOClient(cfg)
	if err != nil {
		return nil, err
	}
	return &MinIOClient{Client: client}, nil
}

// InitMinIOClient 初始化MinIO客户端
func InitMinIOClient(cfg config.MinIOConfig) (*minio.Client, error) {
	// 创建MinIO客户端
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("创建MinIO客户端失败: %v", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("MinIO连接测试失败: %v", err)
	}

	logger.L().Info("MinIO客户端初始化成功")
	return client, nil
}

package minio

import (
	"backend/config"
	"backend/pkg/logger"
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var MinIOCli *MinIOClient

// MinIOClient MinIO客户端结构
type MinIOClient struct {
	Client *minio.Client
}

// GetMinIOClient 获取MinIO客户端
// 使用 config.MinIOConfig 作为参数类型
func GetMinIOClient(cfg config.MinIOConfig) (*MinIOClient, error) {
	if MinIOCli != nil {
		return MinIOCli, nil
	}
	client, err := InitMinIOClient(cfg)
	if err != nil {
		return nil, err
	}
	MinIOCli = &MinIOClient{Client: client}
	return MinIOCli, nil
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

	// 初始化bucket（确保bucket存在）
	if err := initBuckets(client); err != nil {
		logger.L().Warn("初始化bucket失败", logger.WithError(err))
	}

	return client, nil
}

// initBuckets 初始化MinIO bucket（确保bucket存在）
func initBuckets(client *minio.Client) error {
	ctx := context.Background()
	buckets := []string{"image", "video", "audio"}

	for _, bucketName := range buckets {
		exists, err := client.BucketExists(ctx, bucketName)
		if err != nil {
			return fmt.Errorf("检查bucket %s 失败: %v", bucketName, err)
		}

		if !exists {
			err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
				Region: "us-east-1",
			})
			if err != nil {
				return fmt.Errorf("创建bucket %s 失败: %v", bucketName, err)
			}
			logger.L().Info("创建bucket成功", logger.WithString("bucket", bucketName))
		}
	}

	return nil
}

// GetBucketName 根据数据类型获取bucket名称
func GetBucketName(dataType string) string {
	switch dataType {
	case "image":
		return "image"
	case "video":
		return "video"
	case "audio":
		return "audio"
	default:
		return "image" // 默认使用image bucket
	}
}

// UploadFile 上传文件到MinIO
func (c *MinIOClient) UploadFile(bucketName, objectName string, filePath string, contentType string) error {
	ctx := context.Background()

	// 检查bucket是否存在
	exists, err := c.Client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("检查bucket失败: %v", err)
	}
	if !exists {
		err = c.CreateBucket(bucketName)
		if err != nil {
			return fmt.Errorf("Bucket %s 不存在,创建bucket失败: %v", bucketName, err)
		}
	}

	// 上传文件
	_, err = c.Client.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("上传文件失败: %v", err)
	}

	return nil
}

// PutObjectFromReader 从Reader上传文件到MinIO
func (c *MinIOClient) PutObjectFromReader(bucketName, objectName string, reader io.Reader, size int64, contentType string) error {
	ctx := context.Background()

	// 检查bucket是否存在
	exists, err := c.Client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("检查bucket失败: %v", err)
	}
	if !exists {
		err = c.CreateBucket(bucketName)
		if err != nil {
			return fmt.Errorf("Bucket %s 不存在,创建bucket失败: %v", bucketName, err)
		}
	}

	// 上传文件
	_, err = c.Client.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("上传文件失败: %v", err)
	}

	return nil
}

// DownloadFile 从MinIO下载文件到本地
func (c *MinIOClient) DownloadFile(bucketName, objectName, filePath string) (string, error) {
	ctx := context.Background()
	
	// 强制浏览器下载
	params := url.Values{}
	params.Add("response-content-disposition", "attachment; filename="+objectName)

	presignedURL, err := c.Client.PresignedGetObject(ctx, bucketName, objectName, 5*time.Minute, params)
	if err != nil {
		return "", fmt.Errorf("获取下载文件URL失败: %v", err)
	}

	return presignedURL.String(), nil
}

// GetObjectAsReader 从MinIO获取文件作为Reader
func (c *MinIOClient) GetObjectAsReader(bucketName, objectName string) (*minio.Object, error) {
	ctx := context.Background()

	obj, err := c.Client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取文件失败: %v", err)
	}

	return obj, nil
}

// DeleteObject 从MinIO删除文件
func (c *MinIOClient) DeleteObject(bucketName, objectName string) error {
	ctx := context.Background()

	err := c.Client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	return nil
}

// GetObjectInfo 获取文件信息
func (c *MinIOClient) GetObjectInfo(bucketName, objectName string) (minio.ObjectInfo, error) {
	ctx := context.Background()

	info, err := c.Client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return info, fmt.Errorf("获取文件信息失败: %v", err)
	}

	return info, nil
}

// ListObjects 列出bucket中的文件
func (c *MinIOClient) ListObjects(bucketName string, prefix string, recursive bool) <-chan minio.ObjectInfo {
	ctx := context.Background()
	return c.Client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})
}

// CreateBucket 创建bucket
func (c *MinIOClient) CreateBucket(bucketName string) error {
	ctx := context.Background()
	err := c.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		Region: "us-east-1",
	})
	if err != nil {
		return fmt.Errorf("创建bucket失败: %v", err)
	}
	return nil
}

// PresignedPutObject 生成预签名PUT URL（用于客户端直接上传）
// 返回一个预签名的URL，客户端可以直接使用该URL上传文件到MinIO
func (c *MinIOClient) PresignedPutObject(bucketName, objectName string, expiry time.Duration) (string, error) {
	ctx := context.Background()

	// 确保bucket存在
	exists, err := c.Client.BucketExists(ctx, bucketName)
	if err != nil {
		return "", fmt.Errorf("检查bucket失败: %v", err)
	}
	if !exists {
		err = c.CreateBucket(bucketName)
		if err != nil {
			return "", fmt.Errorf("Bucket %s 不存在,创建bucket失败: %v", bucketName, err)
		}
	}

	// 生成预签名PUT URL
	presignedURL, err := c.Client.PresignedPutObject(ctx, bucketName, objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("生成预签名PUT URL失败: %v", err)
	}

	return presignedURL.String(), nil
}

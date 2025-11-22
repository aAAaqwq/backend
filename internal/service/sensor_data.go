package service

import (
	"backend/internal/model"
	"backend/internal/repo"
	"backend/pkg/utils"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type SensorDataService struct {
	metadataRepo   *repo.MetadataRepository
	sensorDataRepo *repo.SensorDataRepository
	deviceUserRepo *repo.DeviceUserRepository
}

func NewSensorDataService() *SensorDataService {
	return &SensorDataService{
		metadataRepo:   repo.NewMetadataRepository(),
		sensorDataRepo: repo.NewSensorDataRepository(),
		deviceUserRepo: repo.NewDeviceUserRepository(),
	}
}

// UploadSensorData 上传传感器数据（统一接口）
func (s *SensorDataService) UploadSensorData(req *model.UploadSensorDataRequest) (int64, error) {
	if req.Metadata.DataType == model.DataTypeSeries {
		return s.uploadSeriesData(req)
	} else if req.Metadata.DataType == model.DataTypeFileData {
		metadata, err := s.uploadFileData(req)
		if err != nil {
			return 0, err
		}
		return metadata.DataID, nil
	}
	return 0, errors.New("不支持的data_type")
}

// uploadSeriesData 上传时序数据
func (s *SensorDataService) uploadSeriesData(req *model.UploadSensorDataRequest) (int64, error) {
	if err := s.createMetadata(&req.Metadata); err != nil {
		return 0, fmt.Errorf("创建元数据失败: %v", err)
	}

	if len(req.SeriesData.Points) == 0 {
		return 0, errors.New("时序数据点不能为空")
	}

	// 为每个点添加标签和字段（tags内容从metadata添加，不包含uid，因为数据是设备采集的，与用户关系不大）
	for i := range req.SeriesData.Points {
		// 初始化tags和fields
		if req.SeriesData.Points[i].Tags == nil {
			req.SeriesData.Points[i].Tags = make(map[string]string)
		}
		if req.SeriesData.Points[i].Fields == nil {
			req.SeriesData.Points[i].Fields = make(map[string]interface{})
		}

		// 设置measurement（从metadata的extra_data获取，或使用默认值）
		if req.SeriesData.Points[i].Measurement == "" {
			if req.Metadata.ExtraData != nil {
				if m, ok := req.Metadata.ExtraData["measurement"].(string); ok && m != "" {
					req.SeriesData.Points[i].Measurement = m
				}
			}
			// 如果extra_data中没有，使用data_type作为默认值
			if req.SeriesData.Points[i].Measurement == "" {
				req.SeriesData.Points[i].Measurement = req.Metadata.DataType
			}
		}

		// tags包含dev_id(string)和quality_score(string，从metadata添加)
		req.SeriesData.Points[i].Tags["dev_id"] = fmt.Sprintf("%d", req.Metadata.DevID)

		// 优先使用point中fields的quality_score，否则使用metadata中的
		qualityScore := req.Metadata.QualityScore
		if qs, ok := req.SeriesData.Points[i].Fields["quality_score"].(float64); ok && qs > 0 {
			qualityScore = qs
		}
		if qualityScore > 0 {
			req.SeriesData.Points[i].Tags["quality_score"] = fmt.Sprintf("%.2f", qualityScore)
		}

		// 如果timestamp是0，使用metadata的timestamp
		if req.SeriesData.Points[i].Timestamp == 0 {
			if ts, err := utils.ConvertToInt64(req.Metadata.Timestamp); err == nil && ts > 0 {
				req.SeriesData.Points[i].Timestamp = ts
			} else {
				req.SeriesData.Points[i].Timestamp = time.Now().Unix()
			}
		}
	}

	if err := s.sensorDataRepo.CreateSeriesData(&req.SeriesData); err != nil {
		return 0, fmt.Errorf("写入InfluxDB失败: %v", err)
	}

	return req.Metadata.DataID, nil
}

// uploadFileData 上传文件数据
func (s *SensorDataService) uploadFileData(req *model.UploadSensorDataRequest) (*model.Metadata, error) {
	if err := s.createMetadata(&req.Metadata); err != nil {
		return nil, fmt.Errorf("创建元数据失败: %v", err)
	}

	// 确定bucket名称（如果为空，根据data_type或文件路径推断）
	bucketName := req.FileData.BucketName
	if bucketName == "" {
		// 优先使用data_type，否则根据文件路径推断
		if req.Metadata.DataType != "" && req.Metadata.DataType != model.DataTypeFileData {
			bucketName = req.Metadata.DataType
		} else {
			bucketName = getBucketNameByFilePath(req.FileData.FilePath)
		}
	}

	// 确定object key（如果为空，创建默认的dev_id/filename为key）
	objectKey := req.FileData.BucketKey
	if objectKey == "" {
		filename := filepath.Base(req.FileData.FilePath)
		objectKey = fmt.Sprintf("%d/%s", req.Metadata.DevID, filename)
	}

	// 确定content type
	contentType := getContentTypeByFilePath(req.FileData.FilePath)

	// 上传到MinIO
	if err := s.sensorDataRepo.FPutFile(bucketName, objectKey, req.FileData.FilePath, contentType); err != nil {
		return nil, fmt.Errorf("上传文件失败: %v", err)
	}

	// 更新元数据
	if req.Metadata.ExtraData == nil {
		req.Metadata.ExtraData = make(map[string]interface{})
	}
	req.Metadata.ExtraData["bucket_name"] = bucketName
	req.Metadata.ExtraData["bucket_key"] = objectKey

	return &req.Metadata, nil
}

// createMetadata 创建元数据
func (s *SensorDataService) createMetadata(metadata *model.Metadata) error {
	// 如果data_id未提供，生成新的
	if metadata.DataID == 0 {
		metadata.DataID = utils.GetDefaultSnowflake().Generate()
	}
	// 如果timestamp为空，使用当前时间
	if metadata.Timestamp == "" {
		metadata.Timestamp = fmt.Sprintf("%d", time.Now().Unix())
	}
	return s.metadataRepo.CreateMetadata(metadata)
}

// GetSeriesData 查询时序数据
func (s *SensorDataService) GetSeriesData(measurement string, devID, uid, dataID int64, startTime, endTime int64,
	tags map[string]string, fields []string, downSampleEvery string, aggregate string, limitPoints int) ([]model.Point, error) {
	return s.sensorDataRepo.QuerySeriesData(measurement, devID, uid, dataID, startTime, endTime, tags, fields, downSampleEvery, aggregate, limitPoints)
}

// GetSensorDataStatistic 获取时序数据统计信息
func (s *SensorDataService) GetSensorDataStatistic(devID int64, measurement string) (map[string]interface{}, error) {
	return s.sensorDataRepo.GetSeriesDataStatistics(measurement, devID)
}

// GetFileList 获取文件列表
// dataType: 数据类型，用于确定bucket名称（如image、video、audio等）
// devID: 设备ID，用于过滤文件（文件key格式为dev_id/filename）
// role: 用户角色（"admin"或普通用户），用于权限判断
// currentUID: 当前用户ID，普通用户只能查询有权限的设备数据
func (s *SensorDataService) GetFileList(page, pageSize int, dataType string, devID int64, role string, currentUID int64) ([]model.FileList, int64, error) {
	// 权限判断：普通用户需要检查设备权限，管理员不需要
	if role != "admin" {
		deviceUser, err := s.deviceUserRepo.GetDeviceUser(devID, currentUID)
		if err != nil {
			return nil, 0, errors.New("您没有权限访问该设备的数据")
		}
		// 检查是否有读权限
		if deviceUser.PermissionLevel != model.PermissionLevelRead &&
			deviceUser.PermissionLevel != model.PermissionLevelReadWrite {
			return nil, 0, errors.New("您没有读权限")
		}
		if !deviceUser.IsActive {
			return nil, 0, errors.New("设备绑定关系未激活")
		}
	}

	// 确定bucket名称（data_type就是bucket名称）
	bucketName := dataType
	if bucketName == "" {
		return nil, 0, errors.New("data_type不能为空")
	}

	// 从MinIO获取文件列表
	allFiles, err := s.sensorDataRepo.GetFileLists(bucketName)
	if err != nil {
		return nil, 0, fmt.Errorf("获取文件列表失败: %v", err)
	}

	// 根据dev_id过滤文件（文件key格式为dev_id/filename）
	filteredFiles := make([]model.FileList, 0)
	prefix := fmt.Sprintf("%d/", devID)
	for _, file := range allFiles {
		// 检查文件key是否以dev_id开头
		if strings.HasPrefix(file.BucketKey, prefix) {
			filteredFiles = append(filteredFiles, file)
		}
	}

	// 分页处理
	total := int64(len(filteredFiles))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(filteredFiles) {
		return []model.FileList{}, total, nil
	}
	if end > len(filteredFiles) {
		end = len(filteredFiles)
	}

	return filteredFiles[start:end], total, nil
}

// DownloadFile 下载文件
func (s *SensorDataService) DownloadFile(bucketName, objectKey string) (string, error) {
	if bucketName == "" || objectKey == "" {
		return "", errors.New("bucket_name和bucket_key不能为空")
	}
	return s.sensorDataRepo.DownloadFile(bucketName, objectKey, "")
}

// DeleteSeriesData 删除时序数据
func (s *SensorDataService) DeleteSeriesData(measurement string, devID, uid int64, startTime, endTime int64) error {
	return s.sensorDataRepo.DeleteSeriesData(measurement, devID, uid, startTime, endTime)
}

// DeleteFileData 删除文件数据
func (s *SensorDataService) DeleteFileData(bucketName, bucketKey string) error {
	if bucketName == "" || bucketKey == "" {
		return errors.New("bucket_name和bucket_key不能为空")
	}
	if err := s.sensorDataRepo.DeleteObject(bucketName, bucketKey); err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}
	// TODO: 从metadata表删除对应记录
	return nil
}

// getContentTypeByFilePath 根据文件路径获取contentType
func getContentTypeByFilePath(filePath string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))
	if contentType := utils.GetContentType(ext); contentType != "" {
		return contentType
	}
	return "application/octet-stream"
}

// getBucketNameByFilePath 根据文件路径推断bucket名称
func getBucketNameByFilePath(filePath string) string {
	contentType := getContentTypeByFilePath(filePath)
	if strings.HasPrefix(contentType, "image/") {
		return "image"
	}
	if strings.HasPrefix(contentType, "video/") {
		return "video"
	}
	if strings.HasPrefix(contentType, "audio/") {
		return "audio"
	}
	return "file"
}

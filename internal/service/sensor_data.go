package service

import (
	"backend/internal/model"
	"backend/internal/repo"
	"backend/pkg/utils"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SensorDataService struct {
	metadataRepo   *repo.MetadataRepository
	sensorDataRepo *repo.SensorDataRepository
	deviceUserRepo *repo.DeviceUserRepository
	deviceRepo     *repo.DeviceRepository
	// 用于临时存储上传信息的map（生产环境应该使用Redis）
	uploadSessions map[string]*UploadSession
	mu             sync.RWMutex
}

// UploadSession 上传会话信息
type UploadSession struct {
	UploadID    string
	DevID       int64
	BucketName  string
	BucketKey   string
	Filename    string
	ContentType string
	UID         int64
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

func NewSensorDataService() *SensorDataService {
	return &SensorDataService{
		metadataRepo:   repo.NewMetadataRepository(),
		sensorDataRepo: repo.NewSensorDataRepository(),
		deviceUserRepo: repo.NewDeviceUserRepository(),
		deviceRepo:     repo.NewDeviceRepository(),
		uploadSessions: make(map[string]*UploadSession),
	}
}

// GetPresignedPutURL 生成预签名PUT URL
func (s *SensorDataService) GetPresignedPutURL(devID int64, filename, bucketName, contentType string, uid int64) (map[string]any, error) {
	// 检查设备是否存在
	_, err := s.deviceRepo.GetDevice(devID)
	if err != nil {
		return nil, errors.New("设备不存在")
	}

	// 确定bucket名称（如果为空，根据文件类型推断）
	if bucketName == "" {
		bucketName = getBucketNameByFilePath(filename)
	}

	// 生成object key（格式: dev_id/YYYY/MM/DD/filename）
	now := time.Now()
	objectKey := fmt.Sprintf("%d/%d/%02d/%02d/%s", devID, now.Year(), now.Month(), now.Day(), filename)

	// 如果没有提供content_type，根据文件扩展名推断
	if contentType == "" {
		contentType = getContentTypeByFilePath(filename)
	}

	// 生成上传ID（使用MD5哈希）
	uploadID := generateUploadID(devID, objectKey, uid)

	// 生成预签名PUT URL（有效期15分钟）
	presignedURL, err := s.sensorDataRepo.PresignedPutObject(bucketName, objectKey, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("生成预签名URL失败: %v", err)
	}

	// 保存上传会话信息（生产环境应该使用Redis）
	session := &UploadSession{
		UploadID:    uploadID,
		DevID:       devID,
		BucketName:  bucketName,
		BucketKey:   objectKey,
		Filename:    filename,
		ContentType: contentType,
		UID:         uid,
		CreatedAt:   now,
		ExpiresAt:   now.Add(30 * time.Minute), // 会话有效期30分钟
	}

	s.mu.Lock()
	s.uploadSessions[uploadID] = session
	s.mu.Unlock()

	// 启动后台清理过期会话的goroutine
	go s.cleanExpiredSessions()

	return map[string]any{
		"upload_id":     uploadID,
		"upload_url":    presignedURL,
		"bucket_name":   bucketName,
		"bucket_key":    objectKey,
		"content_type":  contentType,
		"expires_in":    900,
		"upload_method": "PUT",
	}, nil
}

// cleanExpiredSessions 清理过期的上传会话
func (s *SensorDataService) cleanExpiredSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, session := range s.uploadSessions {
		if now.After(session.ExpiresAt) {
			delete(s.uploadSessions, id)
		}
	}
}

// generateUploadID 生成上传ID
func generateUploadID(devID int64, objectKey string, uid int64) string {
	data := fmt.Sprintf("%d-%s-%d-%d", devID, objectKey, uid, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// UploadSensorData 上传传感器数据（统一接口）
func (s *SensorDataService) UploadSensorData(req *model.UploadSensorDataRequest) (int64, error) {
	// 检查设备是否存在
	_, err := s.deviceRepo.GetDevice(req.Metadata.DevID)
	if err != nil {
		return 0, errors.New("设备不存在")
	}

	// 生成data_id
	if req.Metadata.DataID == 0 {
		req.Metadata.DataID = utils.GetDefaultSnowflake().Generate()
	}
	if req.Metadata.DataType == model.DataTypeSeries {
		if err := s.uploadSeriesData(req); err != nil {
			return 0, err
		}
		return req.Metadata.DataID, nil
	} else if req.Metadata.DataType == model.DataTypeFileData {
		if err := s.uploadFileData(req); err != nil {
			return 0, err
		}
		return req.Metadata.DataID, nil
	}
	return 0, errors.New("不支持的data_type")
}

// uploadSeriesData 上传时序数据
func (s *SensorDataService) uploadSeriesData(req *model.UploadSensorDataRequest) error {
	if len(req.SeriesData.Points) == 0 {
		return errors.New("时序数据点不能为空")
	}

	// 为每个点添加标签和字段（tags内容从metadata添加，不包含uid，因为数据是设备采集的，与用户关系不大）
	for i := range req.SeriesData.Points {
		// 初始化tags和fields
		if req.SeriesData.Points[i].Tags == nil {
			req.SeriesData.Points[i].Tags = make(map[string]string)
		}
		if req.SeriesData.Points[i].Fields == nil {
			req.SeriesData.Points[i].Fields = make(map[string]any)
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
		req.SeriesData.Points[i].Tags["data_id"] = fmt.Sprintf("%d", req.Metadata.DataID)

		// 优先使用point中fields的quality_score，否则使用metadata中的
		qualityScore := req.Metadata.QualityScore
		if qs, ok := req.SeriesData.Points[i].Fields["quality_score"].(float64); ok && qs > 0 {
			qualityScore = fmt.Sprintf("%.2f", qs)
		} else if qsStr, ok := req.SeriesData.Points[i].Fields["quality_score"].(string); ok && qsStr != "" {
			qualityScore = qsStr
		}
		if qualityScore != "" {
			req.SeriesData.Points[i].Tags["quality_score"] = qualityScore
		}

		// 验证Fields不能为空
		if len(req.SeriesData.Points[i].Fields) == 0 {
			return errors.New("时序数据点的fields不能为空，至少需要一个field")
		}

		// 如果timestamp是0，使用metadata的timestamp或当前时间
		// 为避免多个点timestamp相同导致覆盖，每个点递增1秒
		if req.SeriesData.Points[i].Timestamp == 0 {
			if !req.Metadata.Timestamp.IsZero() {
				// 使用metadata的时间 + i秒，确保每个点时间戳不同
				req.SeriesData.Points[i].Timestamp = req.Metadata.Timestamp.Unix() + int64(i)
			} else {
				// 使用当前时间 + i秒
				req.SeriesData.Points[i].Timestamp = time.Now().Unix() + int64(i)
			}
		}
	}

	// 先写入InfluxDB，成功后再创建元数据，保证原子性
	if err := s.sensorDataRepo.CreateSeriesData(&req.SeriesData); err != nil {
		return fmt.Errorf("写入InfluxDB失败: %v", err)
	}

	// InfluxDB写入成功，再创建元数据
	if err := s.createMetadata(&req.Metadata); err != nil {
		// 元数据创建失败，需要回滚InfluxDB中的数据
		// 注意：这里简化处理，实际生产环境可能需要更复杂的补偿机制
		return fmt.Errorf("创建元数据失败: %v（时序数据已写入InfluxDB）", err)
	}

	return nil
}

// uploadFileData 验证文件上传并创建元数据
func (s *SensorDataService) uploadFileData(req *model.UploadSensorDataRequest) error {
	uploadID := req.FileData.UploadID
	if uploadID == "" {
		return errors.New("file_data.upload_id不能为空")
	}

	// 获取上传会话信息
	s.mu.RLock()
	session, exists := s.uploadSessions[uploadID]
	s.mu.RUnlock()

	if !exists {
		return errors.New("无效的upload_id或上传会话已过期")
	}

	// 验证会话是否过期
	if time.Now().After(session.ExpiresAt) {
		s.mu.Lock()
		delete(s.uploadSessions, uploadID)
		s.mu.Unlock()
		return errors.New("上传会话已过期")
	}

	// 验证设备ID匹配
	if session.DevID != req.Metadata.DevID {
		return errors.New("设备ID不匹配")
	}

	// 使用会话中的bucket信息（如果客户端指定了bucket_name，则使用客户端的）
	bucketName := req.FileData.BucketName
	if bucketName == "" {
		bucketName = session.BucketName
	}

	// 使用会话中的bucket_key（如果客户端指定了bucket_key，则使用客户端的）
	bucketKey := req.FileData.BucketKey
	if bucketKey == "" {
		bucketKey = session.BucketKey
	}

	// 验证文件是否存在于MinIO（确认客户端上传成功）
	_, err := s.sensorDataRepo.GetObjectInfo(bucketName, bucketKey)
	if err != nil {
		return fmt.Errorf("文件未找到，请确认是否上传成功: %v", err)
	}

	// 更新元数据
	if req.Metadata.ExtraData == nil {
		req.Metadata.ExtraData = make(map[string]any)
	}
	req.Metadata.ExtraData["bucket_name"] = bucketName
	req.Metadata.ExtraData["bucket_key"] = bucketKey
	req.Metadata.ExtraData["filename"] = session.Filename
	req.Metadata.ExtraData["content_type"] = session.ContentType

	// 创建元数据
	if err := s.createMetadata(&req.Metadata); err != nil {
		return fmt.Errorf("创建元数据失败: %v", err)
	}

	// 删除上传会话
	s.mu.Lock()
	delete(s.uploadSessions, uploadID)
	s.mu.Unlock()

	return nil
}

// createMetadata 创建元数据
func (s *SensorDataService) createMetadata(metadata *model.Metadata) error {
	// 如果data_id未提供，生成新的
	if metadata.DataID == 0 {
		metadata.DataID = utils.GetDefaultSnowflake().Generate()
	}
	// 如果timestamp为空，使用当前时间
	if metadata.Timestamp.IsZero() {
		metadata.Timestamp = time.Now()
	}
	return s.metadataRepo.CreateMetadata(metadata)
}

// GetSeriesData 查询时序数据
func (s *SensorDataService) GetSeriesData(measurement string, devID int64, currentUID int64, startTime, endTime int64,
	tags map[string]string, fields map[string]any, downSampleInterval string, aggregate string, limitPoints int, role string) ([]model.Point, error) {

	// 权限判断：普通用户需要检查设备权限，管理员不需要
	if role != "admin" {
		deviceUser, err := s.deviceUserRepo.GetDeviceUser(devID, currentUID)
		if err != nil {
			return nil, errors.New("您没有权限访问该设备的数据")
		}
		// 检查是否有读权限
		if deviceUser.PermissionLevel != model.PermissionLevelRead &&
			deviceUser.PermissionLevel != model.PermissionLevelReadWrite {
			return nil, errors.New("您没有读权限")
		}
	}

	return s.sensorDataRepo.QuerySeriesData(measurement, devID, startTime, endTime, tags, fields, downSampleInterval, aggregate, limitPoints)
}

// GetSensorDataStatistic 获取时序数据统计信息
func (s *SensorDataService) GetSensorDataStatistic(devID int64, measurement string, currentUID int64, role string) (map[string]any, error) {
	// 权限判断：普通用户需要检查设备权限，管理员不需要
	if role != "admin" {
		deviceUser, err := s.deviceUserRepo.GetDeviceUser(devID, currentUID)
		if err != nil {
			return nil, errors.New("您没有权限访问该设备的数据")
		}
		// 检查是否有读权限
		if deviceUser.PermissionLevel != model.PermissionLevelRead &&
			deviceUser.PermissionLevel != model.PermissionLevelReadWrite {
			return nil, errors.New("您没有读权限")
		}
	}

	return s.sensorDataRepo.GetSeriesDataStatistics(measurement, devID)
}

// GetFileList 获取文件列表
// bucketName: bucket名称（如image、video、audio等）
// devID: 设备ID，用于过滤文件（文件key格式为dev_id/YYYY/MM/DD/filename）
// role: 用户角色（"admin"或普通用户），用于权限判断
// currentUID: 当前用户ID，普通用户只能查询有权限的设备数据
func (s *SensorDataService) GetFileList(page, pageSize int, bucketName string, devID int64, role string, currentUID int64) ([]model.FileList, int64, error) {
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
	}

	// 验证bucket名称
	if bucketName == "" {
		return nil, 0, errors.New("bucket_name不能为空")
	}

	// 从MinIO获取文件列表
	allFiles, err := s.sensorDataRepo.GetFileLists(bucketName)
	if err != nil {
		return nil, 0, fmt.Errorf("获取文件列表失败: %v", err)
	}

	// 根据dev_id过滤文件（文件key格式为dev_id/YYYY/MM/DD/filename 或 dev_id/filename）
	filteredFiles := make([]model.FileList, 0)
	prefix := fmt.Sprintf("%d/", devID)
	for _, file := range allFiles {
		// 检查文件key是否以 "dev_id/" 开头
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

// DeleteSeriesData 删除某设备在某时间范围内的时序数据元数据（软删除）
// 注意：InfluxDB v3 不支持删除数据，此操作只删除元数据
// InfluxDB 中的时序数据会通过保留策略自动过期删除
func (s *SensorDataService) DeleteSeriesData(devID int64, startTime, endTime *int64, currentUID int64, role string) error {
	// 权限判断：普通用户需要检查设备权限，管理员不需要
	if role != "admin" {
		deviceUser, err := s.deviceUserRepo.GetDeviceUser(devID, currentUID)
		if err != nil {
			return errors.New("您没有权限访问该设备的数据")
		}
		// 检查是否有写权限
		if deviceUser.PermissionLevel != model.PermissionLevelWrite &&
			deviceUser.PermissionLevel != model.PermissionLevelReadWrite {
			return errors.New("您没有写权限")
		}
	}

	// 删除元数据
	if err := s.metadataRepo.DeleteMetadataByDevIDAndTimeRange(devID, startTime, endTime); err != nil {
		return fmt.Errorf("删除元数据失败: %v", err)
	}

	// 注意：InfluxDB v3 不支持删除时序数据
	// 时序数据会保留在 InfluxDB 中，通过保留策略自动过期
	// 由于元数据已删除，这些数据不会再被查询到

	return nil
}

// DeleteFileData 删除文件数据
func (s *SensorDataService) DeleteFileData(bucketName, bucketKey string, currentUID int64, role string) error {
	if bucketName == "" || bucketKey == "" {
		return errors.New("bucket_name和bucket_key不能为空")
	}

	// 从bucket_key中提取dev_id（格式为：dev_id/filename）
	parts := strings.Split(bucketKey, "/")
	if len(parts) < 2 {
		return errors.New("bucket_key格式错误，应为dev_id/filename")
	}
	devID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return errors.New("无法从bucket_key中提取dev_id")
	}

	// 权限判断：普通用户需要检查设备权限，管理员不需要
	if role != "admin" {
		deviceUser, err := s.deviceUserRepo.GetDeviceUser(devID, currentUID)
		if err != nil {
			return errors.New("您没有权限访问该设备的数据")
		}
		// 检查是否有写权限
		if deviceUser.PermissionLevel != model.PermissionLevelWrite &&
			deviceUser.PermissionLevel != model.PermissionLevelReadWrite {
			return errors.New("您没有写权限")
		}
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

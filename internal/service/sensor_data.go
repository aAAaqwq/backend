package service

import (
	"backend/internal/model"
	"backend/internal/repo"
)

type SensorDataService struct {
	metadataRepo *repo.MetadataRepository
	sensorDataRepo *repo.SensorDataRepository
}

func NewSensorDataService() *SensorDataService {
	return &SensorDataService{metadataRepo: repo.NewMetadataRepository(), sensorDataRepo: repo.NewSensorDataRepository()}
}

// UploadMetadata 上传传感器元数据
func (s *SensorDataService) UploadMetadata(metadata *model.Metadata) error {
	return nil
}

// UploadSeriesData 上传传感器时序数据
func (s *SensorDataService) UploadSeriesData(metadata *model.Metadata) error {
	return nil
}

// UploadFileData 上传传感器文件数据
func (s *SensorDataService) UploadFileData(req *model.UploadSensorDataRequest) error {
	req.FileData.FilePath

	s.sensorDataRepo.FPutFile(req.BucketName, req.ObjectName, req.FilePath, req.ContentType)
	return nil
}

// GetFileList 获取传感器文件列表
func (s *SensorDataService) GetFileLists(bucketName string) ([]model.FileList, error) {
	return s.sensorDataRepo.GetFileLists(bucketName)
}

// DeleteSensorData 删除传感器数据
func (s *SensorDataService) DeleteSensorData(dataID int64, dataType string) error {
	return nil
}


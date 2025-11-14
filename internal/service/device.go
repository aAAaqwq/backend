package service

import (
	"backend/internal/model"
	"backend/internal/repo"
	"backend/pkg/utils"
)

type DeviceService struct {
	deviceRepo *repo.DeviceRepository
}

func NewDeviceService() *DeviceService {
	return &DeviceService{deviceRepo: repo.NewDeviceRepository()}
}

// CreateDevice 创建设备
func (s *DeviceService) CreateDevice(device *model.Device) (*model.Device, error) {
	// 生成设备ID
	device.DevID = utils.GetDefaultSnowflake().Generate()
	device.CreateAt = utils.GetCurrentTime()
	device.UpdateAt = utils.GetCurrentTime()

	// 设置默认值
	if device.ExtendedConfig == nil {
		device.ExtendedConfig = make(map[string]interface{})
	}

	err := s.deviceRepo.CreateDevice(device)
	if err != nil {
		return nil, err
	}

	return device, nil
}

// GetDevice 获取指定设备
func (s *DeviceService) GetDevice(devID int64) (*model.Device, error) {
	return s.deviceRepo.GetDevice(devID)
}

// GetDevices 获取设备列表
func (s *DeviceService) GetDevices(page, pageSize int, devType string, devStatus *int, keyword, sortBy, sortOrder string) ([]*model.Device, int64, error) {
	return s.deviceRepo.GetDevices(page, pageSize, devType, devStatus, keyword, sortBy, sortOrder)
}

// UpdateDevice 更新设备
func (s *DeviceService) UpdateDevice(device *model.Device) (*model.Device, error) {
	device.UpdateAt = utils.GetCurrentTime()
	err := s.deviceRepo.UpdateDevice(device)
	if err != nil {
		return nil, err
	}
	return device, nil
}

// DeleteDevice 删除设备
func (s *DeviceService) DeleteDevice(devID int64) error {
	return s.deviceRepo.DeleteDevice(devID)
}

// GetDeviceStatistics 获取设备统计信息
func (s *DeviceService) GetDeviceStatistics() (map[string]interface{}, error) {
	return s.deviceRepo.GetDeviceStatistics()
}

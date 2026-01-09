package service

import (
	"backend/internal/model"
	"backend/internal/repo"
	"backend/pkg/utils"
	"errors"
	"fmt"
)

const (
	MaxDeviceUsersCount = 3 // 一个设备最多绑定3个用户
)

type DeviceService struct {
	deviceRepo     *repo.DeviceRepository
	deviceUserRepo *repo.DeviceUserRepository
}

func NewDeviceService() *DeviceService {
	return &DeviceService{
		deviceRepo:     repo.NewDeviceRepository(),
		deviceUserRepo: repo.NewDeviceUserRepository(),
	}
}

// CreateDevice 创建设备
func (s *DeviceService) CreateDevice(device *model.Device, currentUID int64, role string) (*model.Device, error) {
	// 生成设备ID
	device.DevID = model.Int64ToID(utils.GetDefaultSnowflake().Generate())
	device.CreateAt = utils.GetCurrentTime()
	device.UpdateAt = utils.GetCurrentTime()

	// 设置默认值
	if device.ExtendedConfig == nil {
		device.ExtendedConfig = make(map[string]interface{})
	}

	// 如果DevName为空，设置默认名称 <dev_type>_<dev_id>
	if device.DevName == "" {
		device.DevName = fmt.Sprintf("%s_%d", device.DevType, device.DevID.Int64())
	}

	// // 设置默认设备状态为离线
	// if device.DevStatus == 0 {
	// 	device.DevStatus = model.DevStatusOffline
	// }

	err := s.deviceRepo.CreateDevice(device)
	if err != nil {
		return nil, err
	}

	return device, nil
}

// GetDevice 获取指定设备
func (s *DeviceService) GetDevice(devID int64, currentUID int64, role string) (*model.Device, error) {
	// 权限检查：普通用户需要有绑定关系，管理员可以查看所有
	if role != model.RoleAdmin {
		_, err := s.deviceUserRepo.GetDeviceUser(devID, currentUID)
		if err != nil {
			return nil, errors.New("您没有权限访问该设备")
		}
	}

	return s.deviceRepo.GetDevice(devID)
}

// GetDevices 获取设备列表
// 普通用户：只能查看自己绑定的设备
// 管理员：可以查看所有设备
func (s *DeviceService) GetDevices(page, pageSize int, devStatus *int, keyword, sortBy, sortOrder string, currentUID int64, role string) ([]*model.Device, int64, error) {
	// 管理员可以查看所有设备
	if role == model.RoleAdmin {
		return s.deviceRepo.GetDevices(page, pageSize, "", devStatus, keyword, sortBy, sortOrder)
	}

	// 普通用户：先获取绑定的设备ID列表
	userDevices, _, err := s.deviceUserRepo.GetUserDevices(currentUID, 1, 10000, "", devStatus, "", nil)
	if err != nil {
		return nil, 0, err
	}

	// 如果用户没有绑定任何设备，返回空列表
	if len(userDevices) == 0 {
		return []*model.Device{}, 0, nil
	}

	// 从用户绑定的设备中筛选
	devIDs := make([]int64, 0, len(userDevices))
	for _, dev := range userDevices {
		devIDs = append(devIDs, dev.DevID.Int64())
	}

	// 使用设备ID列表过滤
	return s.deviceRepo.GetDevicesByIDs(devIDs, page, pageSize, devStatus, keyword, sortBy, sortOrder)
}

// UpdateDevice 更新设备
// 普通用户：只能更新有（w或rw）权限的设备
// 管理员：可以更新所有设备
func (s *DeviceService) UpdateDevice(device *model.Device, currentUID int64, role string) (*model.Device, error) {
	// 权限检查
	if role != model.RoleAdmin {
		deviceUser, err := s.deviceUserRepo.GetDeviceUser(device.DevID.Int64(), currentUID)
		if err != nil {
			return nil, errors.New("您没有权限访问该设备")
		}
		// 检查写权限
		if deviceUser.PermissionLevel != model.PermissionLevelWrite &&
			deviceUser.PermissionLevel != model.PermissionLevelReadWrite {
			return nil, errors.New("您没有写权限")
		}
	}

	device.UpdateAt = utils.GetCurrentTime()
	err := s.deviceRepo.UpdateDevice(device)
	if err != nil {
		return nil, err
	}
	return device, nil
}

// DeleteDevice 删除设备
// 普通用户：只能删除有（w或rw）权限的设备
// 管理员：可以删除所有设备
func (s *DeviceService) DeleteDevice(devID int64, currentUID int64, role string) error {
	// 权限检查
	if role != model.RoleAdmin {
		deviceUser, err := s.deviceUserRepo.GetDeviceUser(devID, currentUID)
		if err != nil {
			return errors.New("您没有权限访问该设备")
		}
		// 检查写权限
		if deviceUser.PermissionLevel != model.PermissionLevelWrite &&
			deviceUser.PermissionLevel != model.PermissionLevelReadWrite {
			return errors.New("您没有写权限")
		}
	}

	return s.deviceRepo.DeleteDevice(devID)
}

// GetDeviceStatistics 获取设备统计信息
// 普通用户：只统计有权限的设备
// 管理员：统计所有设备
func (s *DeviceService) GetDeviceStatistics(currentUID int64, role string) (map[string]interface{}, error) {
	// 管理员统计所有设备
	if role == model.RoleAdmin {
		stats, err := s.deviceRepo.GetDeviceStatistics()
		if err != nil {
			return nil, err
		}

		// 计算设备种类数
		if byType, ok := stats["by_type"].(map[string]int64); ok {
			stats["type"] = len(byType)
		} else {
			stats["type"] = 0
		}

		return stats, nil
	}

	// 普通用户：只统计自己有权限的设备
	return s.deviceUserRepo.GetUserDeviceStatistics(currentUID)
}

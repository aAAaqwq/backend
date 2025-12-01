package service

import (
	"backend/internal/model"
	"backend/internal/repo"
	"backend/pkg/utils"
	"errors"
)

type DeviceUserService struct {
	deviceUserRepo *repo.DeviceUserRepository
	deviceRepo     *repo.DeviceRepository
	userRepo       *repo.UserRepository
}

func NewDeviceUserService() *DeviceUserService {
	return &DeviceUserService{
		deviceUserRepo: repo.NewDeviceUserRepository(),
		deviceRepo:     repo.NewDeviceRepository(),
		userRepo:       repo.NewUserRepository(),
	}
}

// BindDeviceUser 绑定用户到设备
func (s *DeviceUserService) BindDeviceUser(devID int64, req *model.DeviceUserBindingReq, currentUID int64, currentRole string) (*model.DeviceUser, error) {
	// 检查设备是否存在
	_, err := s.deviceRepo.GetDevice(devID)
	if err != nil {
		return nil, errors.New("设备不存在")
	}

	// 检查用户是否存在
	_, err = s.userRepo.GetUserByUID(req.UID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 权限检查：普通用户只能绑定自己，管理员可以绑定任意用户
	if currentRole != model.RoleAdmin && req.UID != currentUID {
		return nil, errors.New("普通用户只能将自己绑定到设备")
	}

	// 检查设备已绑定的用户数量（业务规则：一个设备最多绑定3个用户）
	boundUsers, err := s.deviceUserRepo.GetDeviceUsers(devID)
	if err != nil {
		return nil, err
	}
	if len(boundUsers) >= 3 {
		return nil, errors.New("该设备已绑定3个用户，无法继续绑定")
	}

	// 验证权限级别，如果未提供则默认为rw
	if req.PermissionLevel == "" {
		req.PermissionLevel = model.PermissionLevelReadWrite
	} else if req.PermissionLevel != model.PermissionLevelRead &&
		req.PermissionLevel != model.PermissionLevelWrite &&
		req.PermissionLevel != model.PermissionLevelReadWrite {
		return nil, errors.New("无效的权限级别，应为: r, w, rw")
	}

	deviceUser := &model.DeviceUser{
		UID:             req.UID,
		DevID:           devID,
		PermissionLevel: req.PermissionLevel,
		BindAt:          utils.GetCurrentTime(),
	}

	err = s.deviceUserRepo.BindDeviceUser(deviceUser)
	if err != nil {
		return nil, err
	}

	return deviceUser, nil
}

// GetDeviceUsers 获取设备的绑定用户列表
func (s *DeviceUserService) GetDeviceUsers(devID int64, currentUID int64, currentRole string) ([]*model.DeviceUserWithInfo, error) {
	// 权限检查：普通用户需要有读权限，管理员可以查看所有
	if currentRole != model.RoleAdmin {
		deviceUser, err := s.deviceUserRepo.GetDeviceUser(devID, currentUID)
		if err != nil {
			return nil, errors.New("您没有权限查看该设备的绑定用户")
		}
		// 检查是否有读权限
		if deviceUser.PermissionLevel != model.PermissionLevelRead &&
			deviceUser.PermissionLevel != model.PermissionLevelReadWrite {
			return nil, errors.New("您没有读权限")
		}
	}

	return s.deviceUserRepo.GetDeviceUsers(devID)
}

// UpdateDeviceUser 更新设备用户绑定关系
func (s *DeviceUserService) UpdateDeviceUser(devID, uid int64, req *model.DeviceUserUpdateReq, currentUID int64, currentRole string) (*model.DeviceUser, error) {
	// 获取现有绑定关系
	deviceUser, err := s.deviceUserRepo.GetDeviceUser(devID, uid)
	if err != nil {
		return nil, err
	}

	// 权限检查：普通用户只能更新自己的权限（需要有写权限），管理员可以更新任意用户
	if currentRole != model.RoleAdmin {
		if uid != currentUID {
			return nil, errors.New("普通用户只能更新自己的权限")
		}
		// 检查是否有写权限
		if deviceUser.PermissionLevel != model.PermissionLevelWrite &&
			deviceUser.PermissionLevel != model.PermissionLevelReadWrite {
			return nil, errors.New("您没有写权限")
		}
	}

	// 验证权限级别
	if req.PermissionLevel != model.PermissionLevelRead &&
		req.PermissionLevel != model.PermissionLevelWrite &&
		req.PermissionLevel != model.PermissionLevelReadWrite {
		return nil, errors.New("无效的权限级别，应为: r, w, rw")
	}

	deviceUser.PermissionLevel = req.PermissionLevel

	err = s.deviceUserRepo.UpdateDeviceUser(deviceUser)
	if err != nil {
		return nil, err
	}

	return deviceUser, nil
}

// UnbindDeviceUser 解绑用户设备
func (s *DeviceUserService) UnbindDeviceUser(devID, uid int64, currentUID int64, currentRole string) error {
	// 权限检查：普通用户只能解绑自己（需要有写权限），管理员可以解绑任意用户
	if currentRole != model.RoleAdmin {
		if uid != currentUID {
			return errors.New("普通用户只能解绑自己")
		}
		// 检查是否有写权限
		deviceUser, err := s.deviceUserRepo.GetDeviceUser(devID, uid)
		if err != nil {
			return errors.New("设备用户绑定关系不存在")
		}
		if deviceUser.PermissionLevel != model.PermissionLevelWrite &&
			deviceUser.PermissionLevel != model.PermissionLevelReadWrite {
			return errors.New("您没有写权限")
		}
	}

	return s.deviceUserRepo.UnbindDeviceUser(devID, uid)
}

// GetUserDevices 获取用户绑定的设备列表
func (s *DeviceUserService) GetUserDevices(uid int64, page, pageSize int, devType string, devStatus *int, permissionLevel string, isActive *bool) ([]*model.Device, int64, error) {
	return s.deviceUserRepo.GetUserDevices(uid, page, pageSize, devType, devStatus, permissionLevel, isActive)
}

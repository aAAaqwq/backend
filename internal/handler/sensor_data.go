package handler

import (
	"fmt"

	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/service"
	"backend/pkg/logger"
	"backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type SensorDataHandler struct {
	sensorDataService *service.SensorDataService
}

func NewSensorDataHandler() *SensorDataHandler {
	return &SensorDataHandler{sensorDataService: service.NewSensorDataService()}
}

// UploadSensorData 上传传感器数据（统一接口，根据data_type判断是时序数据还是文件数据）
func (h *SensorDataHandler) UploadSensorData(c *gin.Context) {
	req := &model.UploadSensorDataRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	// 参数验证
	if req.Metadata.DataType == "" {
		Error(c, CodeBadRequest, "metadata.data_type不能为空")
		return
	}
	if req.Metadata.DevID == 0 {
		Error(c, CodeBadRequest, "metadata.dev_id不能为空")
		return
	}
	if req.Metadata.UID == 0 {
		Error(c, CodeBadRequest, "metadata.uid不能为空")
		return
	}
	if req.Metadata.DataType == model.DataTypeSeries {
		if len(req.SeriesData.Points) == 0 {
			Error(c, CodeBadRequest, "时序数据点不能为空")
			return
		}
	} else if req.Metadata.DataType == model.DataTypeFileData {
		if req.FileData.FilePath == "" {
			Error(c, CodeBadRequest, "file_data.file_path不能为空")
			return
		}
	}

	var dataID int64
	var err error

	dataID, err = h.sensorDataService.UploadSensorData(req)
	if err != nil {
		logger.L().Error("上传传感器数据失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	SuccessWithCode(c, 201, "上传传感器数据成功", gin.H{"data_id": dataID})
}

// GetSeriesData 查询时序数据
func (h *SensorDataHandler) GetSeriesData(c *gin.Context) {
	measurement := c.Query("measurement")
	devIDStr := c.Query("dev_id")
	uidStr := c.Query("uid")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if measurement == "" || devIDStr == "" || uidStr == "" || startTimeStr == "" || endTimeStr == "" {
		Error(c, CodeBadRequest, "measurement、dev_id、uid、start_time、end_time为必填参数")
		return
	}

	devID, _ := utils.ConvertToInt64(devIDStr)
	uid, _ := utils.ConvertToInt64(uidStr)
	startTime, _ := utils.ConvertToInt64(startTimeStr)
	endTime, _ := utils.ConvertToInt64(endTimeStr)

	dataID, _ := utils.ConvertToInt64(c.Query("data_id"))
	fields := c.QueryArray("fileds")
	downSampleEvery := c.Query("down_sample_every")
	aggregate := c.Query("aggregate")
	limitPoints := 6000
	if lp, _ := utils.ConvertToInt64(c.Query("limit_points")); lp > 0 {
		limitPoints = int(lp)
	}

	// 构建tags（默认包含dev_id和data_id）
	tags := make(map[string]string)
	tags["dev_id"] = devIDStr
	if dataID > 0 {
		tags["data_id"] = fmt.Sprintf("%d", dataID)
	}

	points, err := h.sensorDataService.GetSeriesData(measurement, devID, uid, dataID, startTime, endTime, tags, fields, downSampleEvery, aggregate, limitPoints)
	if err != nil {
		logger.L().Error("查询时序数据失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "查询时序数据成功", gin.H{"points": points})
}

// GetSensorDataStatistic 时序数据统计
func (h *SensorDataHandler) GetSensorDataStatistic(c *gin.Context) {
	devID, _ := utils.ConvertToInt64(c.Query("dev_id"))
	measurement := c.Query("measurement")

	if devID == 0 || measurement == "" {
		Error(c, CodeBadRequest, "dev_id、measurement为必填参数")
		return
	}

	stats, err := h.sensorDataService.GetSensorDataStatistic(devID, measurement)
	if err != nil {
		logger.L().Error("获取统计信息失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "获取统计信息成功", stats)
}

// GetFileList 获取文件列表
func (h *SensorDataHandler) GetFileList(c *gin.Context) {
	// 参数验证
	page, _ := utils.ConvertToInt64(c.Query("page"))
	pageSize, _ := utils.ConvertToInt64(c.Query("page_size"))
	dataType := c.Query("data_type")
	devIDStr := c.Query("dev_id")

	if page <= 0 {
		Error(c, CodeBadRequest, "page必须大于0")
		return
	}
	if pageSize <= 0 {
		Error(c, CodeBadRequest, "page_size必须大于0")
		return
	}
	if dataType == "" {
		Error(c, CodeBadRequest, "data_type不能为空")
		return
	}
	if devIDStr == "" {
		Error(c, CodeBadRequest, "dev_id不能为空")
		return
	}

	devID, _ := utils.ConvertToInt64(devIDStr)
	if devID == 0 {
		Error(c, CodeBadRequest, "dev_id无效")
		return
	}

	// 从token中获取用户角色和ID
	role, _ := middleware.GetCurrentUserRole(c)
	currentUID, _ := middleware.GetCurrentUserID(c)

	fileList, total, err := h.sensorDataService.GetFileList(int(page), int(pageSize), dataType, devID, role, currentUID)
	if err != nil {
		logger.L().Error("获取文件列表失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	Success(c, "获取文件列表成功", gin.H{
		"items": fileList,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// DownloadFile 下载文件
func (h *SensorDataHandler) DownloadFile(c *gin.Context) {
	bucketName := c.Query("bucket_name")
	bucketKey := c.Query("bucket_key")

	if bucketName == "" || bucketKey == "" {
		Error(c, CodeBadRequest, "bucket_name和bucket_key不能为空")
		return
	}

	url, err := h.sensorDataService.DownloadFile(bucketName, bucketKey)
	if err != nil {
		logger.L().Error("获取下载URL失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "获取下载URL成功", gin.H{"download_url": url})
}

// DeleteSeriesData 删除时序数据
func (h *SensorDataHandler) DeleteSeriesData(c *gin.Context) {
	var req struct {
		Measurement string `json:"measurement" binding:"required"`
		DevID       int64  `json:"dev_id" binding:"required"`
		UID         int64  `json:"uid" binding:"required"`
		StartTime   string `json:"start_time" binding:"required"`
		EndTime     string `json:"end_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	startTime, _ := utils.ConvertToInt64(req.StartTime)
	endTime, _ := utils.ConvertToInt64(req.EndTime)

	if err := h.sensorDataService.DeleteSeriesData(req.Measurement, req.DevID, req.UID, startTime, endTime); err != nil {
		logger.L().Error("删除时序数据失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "删除时序数据成功", nil)
}

// DeleteFileData 删除文件数据
func (h *SensorDataHandler) DeleteFileData(c *gin.Context) {
	var req struct {
		BucketName string `json:"bucket_name" binding:"required"`
		BucketKey  string `json:"bucket_key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	if err := h.sensorDataService.DeleteFileData(req.BucketName, req.BucketKey); err != nil {
		logger.L().Error("删除文件数据失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "删除文件数据成功", nil)
}

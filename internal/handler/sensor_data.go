package handler

import (
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

// CreateSensorData 上传传感器数据
func (h *SensorDataHandler) CreateSensorData(c *gin.Context) {
	devIDStr := c.Param("dev_id")
	devID, err := utils.ConvertToInt64(devIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的设备ID")
		return
	}

	req := &model.Metadata{}
	if err := c.ShouldBindJSON(req); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	req.DevID = devID
	metadata, err := h.sensorDataService.CreateSensorData(req)
	if err != nil {
		logger.L().Error("上传传感器数据失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	SuccessWithCode(c, 201, "上传传感器数据成功", metadata)
}

// GetSensorData 查询传感器数据
func (h *SensorDataHandler) GetSensorData(c *gin.Context) {
	page, _ := c.GetQuery("page")
	pageSize, _ := c.GetQuery("page_size")
	dataType, _ := c.GetQuery("data_type")
	startTimeStr, _ := c.GetQuery("start_time")
	endTimeStr, _ := c.GetQuery("end_time")
	minQualityStr, _ := c.GetQuery("min_quality")
	maxQualityStr, _ := c.GetQuery("max_quality")
	keyword, _ := c.GetQuery("keyword")
	sortBy, _ := c.GetQuery("sort_by")
	sortOrder, _ := c.GetQuery("sort_order")

	pageInt, _ := utils.ConvertToInt64(page)
	if pageInt <= 0 {
		pageInt = 1
	}
	pageSizeInt, _ := utils.ConvertToInt64(pageSize)
	if pageSizeInt <= 0 {
		pageSizeInt = 10
	}

	var startTime, endTime *int64
	if !utils.IsEmpty(startTimeStr) {
		t, _ := utils.ConvertToInt64(startTimeStr)
		startTime = &t
	}
	if !utils.IsEmpty(endTimeStr) {
		t, _ := utils.ConvertToInt64(endTimeStr)
		endTime = &t
	}

	var minQuality, maxQuality *float64
	if !utils.IsEmpty(minQualityStr) {
		q, _ := utils.ConvertToFloat64(minQualityStr)
		minQuality = &q
	}
	if !utils.IsEmpty(maxQualityStr) {
		q, _ := utils.ConvertToFloat64(maxQualityStr)
		maxQuality = &q
	}

	dataList, total, err := h.sensorDataService.GetSensorData(int(pageInt), int(pageSizeInt), dataType, startTime, endTime, minQuality, maxQuality, keyword, sortBy, sortOrder)
	if err != nil {
		logger.L().Error("查询传感器数据失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	totalPages := int64((total + int64(pageSizeInt) - 1) / int64(pageSizeInt))
	if totalPages == 0 {
		totalPages = 1
	}

	Success(c, "查询传感器数据成功", gin.H{
		"items": dataList,
		"pagination": gin.H{
			"page":        pageInt,
			"page_size":   pageSizeInt,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetSensorDataStatistic 传感器数据统计
func (h *SensorDataHandler) GetSensorDataStatistic(c *gin.Context) {
	devIDStr := c.Param("dev_id")
	devID, err := utils.ConvertToInt64(devIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的设备ID")
		return
	}

	stats, err := h.sensorDataService.GetSensorDataStatistic(devID)
	if err != nil {
		logger.L().Error("获取传感器数据统计失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "获取传感器数据统计成功", stats)
}

// DeleteSensorData 删除传感器数据
func (h *SensorDataHandler) DeleteSensorData(c *gin.Context) {
	_ = c.Param("dev_id") // dev_id用于路径，但删除时主要使用data_id
	dataIDStr := c.Param("data_id")

	dataID, err := utils.ConvertToInt64(dataIDStr)
	if err != nil {
		Error(c, CodeBadRequest, "无效的数据ID")
		return
	}

	req := &model.Metadata{}
	if err := c.ShouldBindJSON(req); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	err = h.sensorDataService.DeleteSensorData(dataID, req.DataType)
	if err != nil {
		logger.L().Error("删除传感器数据失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "删除传感器数据成功", nil)
}

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

// UploadSensorData 上传传感器数据
func (h *SensorDataHandler) UploadSensorData(c *gin.Context) {
	var req model.UploadSensorDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	if req.Metadata.DataType == "" {
		Error(c, CodeBadRequest, "metadata.data_type不能为空")
		return
	}
	if req.Metadata.DevID == 0 {
		Error(c, CodeBadRequest, "metadata.dev_id不能为空")
		return
	}
	if req.Metadata.DataType == model.DataTypeSeries && len(req.SeriesData.Points) == 0 {
		Error(c, CodeBadRequest, "时序数据点不能为空")
		return
	}
	if req.Metadata.DataType == model.DataTypeFileData && req.FileData.UploadID == "" {
		Error(c, CodeBadRequest, "file_data.upload_id不能为空")
		return
	}

	dataID, err := h.sensorDataService.UploadSensorData(&req)
	if err != nil {
		logger.L().Error("上传传感器数据失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	SuccessWithCode(c, 201, "上传传感器数据成功", gin.H{"data_id": dataID})
}

// GetPresignedPutURL 获取预签名PUT URL
func (h *SensorDataHandler) GetPresignedPutURL(c *gin.Context) {
	var req model.GetPresignedPutURLReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	devID, err := utils.ConvertToInt64(req.DevID)
	if err != nil || devID == 0 {
		Error(c, CodeBadRequest, "dev_id无效")
		return
	}

	currentUID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		Error(c, CodeUnauthorized, "未认证")
		return
	}

	result, err := h.sensorDataService.GetPresignedPutURL(devID, req.Filename, req.BucketName, req.ContentType, currentUID)
	if err != nil {
		logger.L().Error("生成预签名URL失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "获取预签名PUT URL成功", result)
}

// GetSeriesData 查询时序数据
func (h *SensorDataHandler) GetSeriesData(c *gin.Context) {
	var req model.GetSeriesDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	currentUID, _ := middleware.GetCurrentUserID(c)
	role, _ := middleware.GetCurrentUserRole(c)

	if req.LimitPoints <= 0 {
		req.LimitPoints = 6000
	}

	if req.Tags == nil {
		req.Tags = make(map[string]string)
	}
	req.Tags["dev_id"] = fmt.Sprintf("%d", req.DevID)

	points, err := h.sensorDataService.GetSeriesData(req.Measurement, req.DevID, currentUID, req.StartTime, req.EndTime,
		req.Tags, req.Fields, req.DownSampleInterval, req.Aggregate, req.LimitPoints, role)
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

	currentUID, _ := middleware.GetCurrentUserID(c)
	role, _ := middleware.GetCurrentUserRole(c)

	stats, err := h.sensorDataService.GetSensorDataStatistic(devID, measurement, currentUID, role)
	if err != nil {
		logger.L().Error("获取统计信息失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "获取统计信息成功", stats)
}

// GetFileList 获取文件列表
func (h *SensorDataHandler) GetFileList(c *gin.Context) {
	page, _ := utils.ConvertToInt64(c.Query("page"))
	pageSize, _ := utils.ConvertToInt64(c.Query("page_size"))
	bucketName := c.Query("bucket_name")
	devID, _ := utils.ConvertToInt64(c.Query("dev_id"))

	if page <= 0 || pageSize <= 0 {
		Error(c, CodeBadRequest, "page和page_size必须大于0")
		return
	}
	if bucketName == "" || devID == 0 {
		Error(c, CodeBadRequest, "bucket_name和dev_id不能为空")
		return
	}

	role, _ := middleware.GetCurrentUserRole(c)
	currentUID, _ := middleware.GetCurrentUserID(c)

	fileList, total, err := h.sensorDataService.GetFileList(int(page), int(pageSize), bucketName, devID, role, currentUID)
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

// DeleteSeriesData 删除时序数据元数据
func (h *SensorDataHandler) DeleteSeriesData(c *gin.Context) {
	var req model.DeleteSeriesDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	if req.DevID == 0 {
		Error(c, CodeBadRequest, "dev_id不能为空")
		return
	}

	currentUID, _ := middleware.GetCurrentUserID(c)
	role, _ := middleware.GetCurrentUserRole(c)

	var startTime, endTime *int64
	if req.StartTime != "" {
		st, err := utils.ParseTime(req.StartTime)
		if err != nil {
			Error(c, CodeBadRequest, fmt.Sprintf("start_time格式错误: %v", err))
			return
		}
		timestamp := st.Unix()
		startTime = &timestamp
	}
	if req.EndTime != "" {
		et, err := utils.ParseTime(req.EndTime)
		if err != nil {
			Error(c, CodeBadRequest, fmt.Sprintf("end_time格式错误: %v", err))
			return
		}
		timestamp := et.Unix()
		endTime = &timestamp
	}

	if err := h.sensorDataService.DeleteSeriesData(req.DevID, startTime, endTime, currentUID, role); err != nil {
		logger.L().Error("删除时序数据元数据失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "删除时序数据元数据成功", nil)
}

// DeleteFileData 删除文件数据
func (h *SensorDataHandler) DeleteFileData(c *gin.Context) {
	var req model.DeleteFileDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, CodeBadRequest, err.Error())
		return
	}

	currentUID, _ := middleware.GetCurrentUserID(c)
	role, _ := middleware.GetCurrentUserRole(c)

	if err := h.sensorDataService.DeleteFileData(req.BucketName, req.BucketKey, currentUID, role); err != nil {
		logger.L().Error("删除文件数据失败", logger.WithError(err))
		Error(c, CodeInternalServerError, err.Error())
		return
	}

	Success(c, "删除文件数据成功", nil)
}

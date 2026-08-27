package handler

import (
	"fmt"
	"io"
	"strconv"

	"arlo-admin/internal/modules/file/dto"
	"arlo-admin/internal/modules/file/service"
	apperrors "arlo-admin/pkg/errors"
	"arlo-admin/pkg/middleware"
	"arlo-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// FileHandler 文件管理处理器
type FileHandler struct {
	svc *service.FileService
}

// NewFileHandler 创建 FileHandler
func NewFileHandler(svc *service.FileService) *FileHandler {
	return &FileHandler{svc: svc}
}

// Upload 上传文件
// @Summary      上传文件
// @Tags         File
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "文件"
// @Success      200   {object}  response.Response{data=dto.FileResponse}
// @Failure      400   {object}  response.Response
// @Router       /file/upload [post]
func (h *FileHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, apperrors.BadRequest, "请选择文件")
		return
	}

	isPublic := true // 默认公开；显式传 0/false 时为私有
	if v := c.PostForm("public"); v == "0" || v == "false" {
		isPublic = false
	}
	userID, username := middleware.GetCurrentUser(c)
	resp, err := h.svc.Upload(c.Request.Context(), file, userID, username, isPublic)
	if err != nil {
		response.Error(c, apperrors.Internal, err.Error())
		return
	}
	response.Success(c, resp)
}

// Serve 统一文件访问 GET /file/:accessKey
// 有有效 JWT 时可访问任意文件；未登录仅允许公开文件（登录页 Logo 等）
func (h *FileHandler) Serve(c *gin.Context) {
	accessKey := c.Param("accessKey")
	file, reader, err := h.svc.OpenByAccessKey(c.Request.Context(), accessKey)
	if err != nil {
		response.Error(c, apperrors.NotFound, err.Error())
		return
	}
	defer reader.Close()

	authed := middleware.TryJWTAuth(c)
	if !authed && file.IsPublic != 1 {
		response.Error(c, apperrors.Unauthorized, "私有文件需要登录后访问")
		return
	}

	inline := c.Query("inline") == "1" || (!authed && file.IsPublic == 1)
	if inline {
		c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, file.Name))
	} else {
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.Name))
	}
	c.Header("Content-Type", file.MimeType)
	c.Header("Content-Length", fmt.Sprintf("%d", file.Size))
	if file.IsPublic == 1 && !authed {
		c.Header("Cache-Control", "public, max-age=86400")
	} else {
		c.Header("Cache-Control", "private, max-age=300")
	}

	if _, err := io.Copy(c.Writer, reader); err != nil {
		c.Error(err)
	}
}

// SetPublic 设置文件公开/私有 PUT /api/v1/file/:id/public
func (h *FileHandler) SetPublic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}
	var req struct {
		Public *bool `json:"public" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误，需传 public: true/false")
		return
	}
	if err := h.svc.SetPublic(c.Request.Context(), id, *req.Public); err != nil {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}

// Delete 删除文件
func (h *FileHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, apperrors.BadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}

// List 文件列表
func (h *FileHandler) List(c *gin.Context) {
	var req dto.FileListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, apperrors.BadRequest, "参数错误")
		return
	}
	req.SetDefaults()

	uid, _ := middleware.GetCurrentUser(c)
	resp, err := h.svc.List(c.Request.Context(), &req, uid)
	if err != nil {
		response.Error(c, apperrors.Internal, "查询文件列表失败")
		return
	}
	response.Success(c, resp)
}

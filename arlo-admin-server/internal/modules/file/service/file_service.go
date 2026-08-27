package service

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"arlo-admin/internal/modules/file/dto"
	"arlo-admin/internal/modules/file/model"
	"arlo-admin/internal/modules/file/repository"
	"arlo-admin/internal/database"
	"arlo-admin/pkg/datascope"
	"arlo-admin/pkg/storage"
	"arlo-admin/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileService 文件服务
type FileService struct {
	repo        *repository.FileRepository
	driver      storage.Driver // 存储驱动（本地/OSS/...）
	maxSize     int64          // 最大上传大小(字节)
	allowedExts map[string]struct{}
}

// 未配置 allowedExts 时的默认白名单（不含可执行/脚本类）
var defaultAllowedExts = []string{
	"jpg", "jpeg", "png", "gif", "webp", "bmp", "ico", "svg",
	"mp3", "wav", "mp4", "webm", "mov", "avi",
	"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "txt", "csv", "md",
	"zip", "rar", "7z",
}

// NewFileService 创建 FileService
func NewFileService(repo *repository.FileRepository, driver storage.Driver, maxSize int64, allowedExts []string) *FileService {
	exts := allowedExts
	if len(exts) == 0 {
		exts = defaultAllowedExts
	}
	set := make(map[string]struct{}, len(exts))
	for _, e := range exts {
		e = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(e), "."))
		if e != "" {
			set[e] = struct{}{}
		}
	}
	return &FileService{repo: repo, driver: driver, maxSize: maxSize, allowedExts: set}
}

// Upload 上传文件；isPublic=true 时标记为公开可匿名访问
func (s *FileService) Upload(ctx context.Context, fileHeader *multipart.FileHeader, uploaderID uint64, uploader string, isPublic bool) (*dto.FileResponse, error) {
	// 1. 校验扩展名白名单（防可执行/脚本）
	if err := s.validateExt(fileHeader.Filename); err != nil {
		return nil, err
	}

	// 2. 校验文件大小
	if s.maxSize > 0 && fileHeader.Size > s.maxSize {
		return nil, fmt.Errorf("文件大小超过限制 (%dMB)", s.maxSize>>20)
	}

	// 3. 计算 MD5
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}

	hash := md5.New()
	if _, err := io.Copy(hash, src); err != nil {
		src.Close()
		return nil, fmt.Errorf("计算MD5失败: %w", err)
	}
	src.Close()
	md5Str := fmt.Sprintf("%x", hash.Sum(nil))

	// 4. MD5 去重：相同文件不重复存储，直接复用已有记录
	if existing, err := s.repo.FindByMD5(ctx, md5Str); err == nil {
		if isPublic && existing.IsPublic != 1 {
			_ = s.repo.MarkPublic(ctx, existing.ID)
			existing.IsPublic = 1
		}
		resp := toResponse(existing)
		return &resp, nil
	}

	// 5. 自动检测文件分类
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = detectMimeByExt(fileHeader.Filename)
	}
	category := detectCategory(mimeType)

	// 6. 生成存储 key（不含驱动前缀，统一的相对路径格式）
	dateDir := time.Now().Format("2006/01/02")
	ext := filepath.Ext(fileHeader.Filename)
	savedName := fmt.Sprintf("%s_%d%s", md5Str[:16], time.Now().UnixNano(), ext)
	storageKey := filepath.Join(dateDir, savedName)

	// 7. 写入存储
	src2, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("重新打开文件失败: %w", err)
	}
	defer src2.Close()

	if err := s.driver.Save(ctx, storageKey, src2); err != nil {
		return nil, fmt.Errorf("存储文件失败: %w", err)
	}

	var pub int8
	if isPublic {
		pub = 1
	}

	// 8. 写入数据库记录
	file := &model.SysFile{
		AccessKey:  newAccessKey(),
		Name:       fileHeader.Filename,
		Path:       storageKey, // 存储相对路径（driver 自行拼接前缀）
		Size:       fileHeader.Size,
		MimeType:   mimeType,
		Category:   category,
		IsPublic:   pub,
		MD5:        md5Str,
		UploaderID: uploaderID,
		Uploader:   uploader,
	}

	if err := s.repo.Create(ctx, file); err != nil {
		// 入DB失败，回滚存储
		s.driver.Remove(ctx, storageKey)
		return nil, fmt.Errorf("保存文件记录失败: %w", err)
	}

	resp := toResponse(file)
	return &resp, nil
}

func (s *FileService) validateExt(filename string) error {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	if ext == "" {
		return errors.New("不允许上传无扩展名的文件")
	}
	if _, ok := s.allowedExts[ext]; !ok {
		return fmt.Errorf("不允许上传 .%s 类型文件", ext)
	}
	return nil
}

func newAccessKey() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

// OpenByAccessKey 按访问钥打开文件流（调用方自行判断公开/鉴权）
func (s *FileService) OpenByAccessKey(ctx context.Context, accessKey string) (*model.SysFile, io.ReadCloser, error) {
	accessKey = strings.TrimSpace(accessKey)
	if accessKey == "" {
		return nil, nil, fmt.Errorf("文件不存在")
	}
	file, err := s.repo.FindByAccessKey(ctx, accessKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("文件不存在")
		}
		return nil, nil, err
	}
	reader, err := s.driver.Open(ctx, file.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("文件已被删除")
	}
	return file, reader, nil
}

// MarkPublic 将文件标记为公开
func (s *FileService) MarkPublic(ctx context.Context, id uint64) error {
	return s.SetPublic(ctx, id, true)
}

// SetPublic 设置文件公开/私有
func (s *FileService) SetPublic(ctx context.Context, id uint64, isPublic bool) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("文件不存在")
		}
		return err
	}
	return s.repo.SetPublic(ctx, id, isPublic)
}

// Delete 删除文件（DB软删除 + 存储删除）
func (s *FileService) Delete(ctx context.Context, id uint64) error {
	file, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("文件不存在")
		}
		return err
	}

	// 软删除 DB 记录
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// 删除存储文件（忽略错误，不阻塞DB操作）
	_ = s.driver.Remove(ctx, file.Path)

	return nil
}

// List 分页查询文件（按数据权限过滤）
func (s *FileService) List(ctx context.Context, req *dto.FileListQuery, currentUserID uint64) (*dto.FileListResponse, error) {
	req.SetDefaults()
	scope, _ := datascope.BuildFromDB(ctx, database.DB, currentUserID)
	files, total, err := s.repo.List(ctx, req.Name, req.MimeType, req.Category, req.IsPublic, scope, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	list := make([]dto.FileResponse, 0, len(files))
	for i := range files {
		list = append(list, toResponse(&files[i]))
	}

	return &dto.FileListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// ─── 分类检测 ──────────────────────────────────────────────

// detectCategory 根据 MIME 类型自动判断文件分类
func detectCategory(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	case strings.HasPrefix(mimeType, "application/"),
		strings.HasPrefix(mimeType, "text/"):
		return "document"
	default:
		return "other"
	}
}

// detectMimeByExt 根据文件扩展名补全 Content-Type（浏览器可能不传）
func detectMimeByExt(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".mp4":
		return "video/mp4"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".flac":
		return "audio/flac"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".txt":
		return "text/plain"
	case ".csv":
		return "text/csv"
	default:
		return "application/octet-stream"
	}
}

// toResponse model → dto（url 按 access_key 动态生成，不落库）
func toResponse(f *model.SysFile) dto.FileResponse {
	return dto.FileResponse{
		ID:         f.ID,
		AccessKey:  f.AccessKey,
		Name:       f.Name,
		URL:        fmt.Sprintf("/api/v1/file/%s", f.AccessKey),
		Size:       f.Size,
		MimeType:   f.MimeType,
		Category:   f.Category,
		IsPublic:   f.IsPublic,
		MD5:        f.MD5,
		UploaderID: f.UploaderID,
		Uploader:   f.Uploader,
		CreatedAt:  utils.FormatTime(f.CreatedAt),
	}
}

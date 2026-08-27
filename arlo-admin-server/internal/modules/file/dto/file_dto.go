package dto

// FileResponse 文件信息
type FileResponse struct {
	ID         uint64 `json:"id" example:"1"`                                                          // 文件ID（管理用）
	AccessKey  string `json:"accessKey" example:"a1b2c3d4e5f6789012345678abcdef01"`                    // 访问钥
	Name       string `json:"name" example:"avatar.png"`                                               // 原始文件名
	URL        string `json:"url" example:"/api/v1/file/a1b2c3d4e5f6789012345678abcdef01"`              // 统一访问地址
	Size       int64  `json:"size" example:"102400"`                                                   // 文件大小(字节)
	MimeType   string `json:"mimeType" example:"image/png"`                                            // MIME类型
	Category   string `json:"category" example:"image"`                                                // 文件分类: image/video/audio/document/other
	IsPublic   int8   `json:"isPublic" example:"0"`                                                    // 是否公开
	MD5        string `json:"md5" example:"d41d8cd98f00b204e9800998ecf8427e"`                          // MD5校验值
	UploaderID uint64 `json:"uploaderId" example:"1"`                                                  // 上传者ID
	Uploader   string `json:"uploader" example:"admin"`                                                // 上传者
	CreatedAt  string `json:"createdAt"`                                                               // 上传时间
}

// FileListQuery 文件列表查询
type FileListQuery struct {
	Page     int    `form:"page" example:"1"`       // 页码
	PageSize int    `form:"pageSize" example:"10"`  // 每页条数
	Name     string `form:"name" example:"avatar"`  // 文件名（模糊搜索）
	MimeType string `form:"mimeType" example:"image"` // MIME类型（模糊搜索）
	Category string `form:"category" example:"image"` // 文件分类: image/video/audio/document/other
	IsPublic *int8  `form:"isPublic" example:"1"`   // 是否公开: 0私有 1公开
}

// FileListResponse 文件列表
type FileListResponse struct {
	List     []FileResponse `json:"list"`
	Total    int64          `json:"total" example:"100"`
	Page     int            `json:"page" example:"1"`
	PageSize int            `json:"pageSize" example:"10"`
}

// SetDefaults 设置分页默认值
func (q *FileListQuery) SetDefaults() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 10
	}
}

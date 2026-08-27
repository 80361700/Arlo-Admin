package service

import (
	"context"
	"fmt"

	"arlo-admin/internal/modules/system/dto"
	"arlo-admin/internal/modules/system/model"
	"arlo-admin/internal/modules/system/repository"

	perrors "arlo-admin/pkg/errors"

	"gorm.io/gorm"
)

// DictService 字典管理服务
type DictService struct {
	dictRepo *repository.DictRepository
}

func NewDictService(dictRepo *repository.DictRepository) *DictService {
	return &DictService{dictRepo: dictRepo}
}

// --- 字典类型 ---

func (s *DictService) ListDictTypes(ctx context.Context, req *dto.DictTypeListRequest) (*dto.PageResponse, error) {
	dts, total, err := s.dictRepo.ListDictTypes(ctx, req.Name, req.Code, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	list := make([]dto.DictTypeResponse, 0, len(dts))
	for _, dt := range dts {
		list = append(list, dto.DictTypeResponse{
			ID:        dt.ID,
			Name:      dt.Name,
			Code:      dt.Code,
			Status:    dt.Status,
			Remark:    dt.Remark,
			CreatedAt: dt.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &dto.PageResponse{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *DictService) GetDictType(ctx context.Context, id uint64) (*dto.DictTypeResponse, error) {
	dt, err := s.dictRepo.FindDictTypeByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, perrors.New(perrors.NotFound, "字典类型不存在")
		}
		return nil, err
	}
	return &dto.DictTypeResponse{
		ID:        dt.ID,
		Name:      dt.Name,
		Code:      dt.Code,
		Status:    dt.Status,
		Remark:    dt.Remark,
		CreatedAt: dt.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *DictService) GetAllDictTypes(ctx context.Context) ([]dto.DictTypeResponse, error) {
	dts, err := s.dictRepo.FindAllDictTypes(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]dto.DictTypeResponse, 0, len(dts))
	for _, dt := range dts {
		list = append(list, dto.DictTypeResponse{
			ID:   dt.ID,
			Name: dt.Name,
			Code: dt.Code,
		})
	}
	return list, nil
}

func (s *DictService) CreateDictType(ctx context.Context, req *dto.CreateDictTypeRequest) error {
	exists, err := s.dictRepo.ExistsDictTypeByCode(ctx, req.Code, 0)
	if err != nil {
		return err
	}
	if exists {
		return perrors.New(perrors.ErrDictTypeExists, fmt.Sprintf("字典编码 %s 已存在", req.Code))
	}
	dt := &model.DictType{
		Name:   req.Name,
		Code:   req.Code,
		Status: req.Status,
		Remark: req.Remark,
	}
	if dt.Status == 0 {
		dt.Status = 1
	}
	return s.dictRepo.CreateDictType(ctx, dt)
}

func (s *DictService) UpdateDictType(ctx context.Context, req *dto.UpdateDictTypeRequest) error {
	dt, err := s.dictRepo.FindDictTypeByID(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.ErrDictTypeExists, "字典类型不存在")
		}
		return err
	}
	exists, err := s.dictRepo.ExistsDictTypeByCode(ctx, req.Code, req.ID)
	if err != nil {
		return err
	}
	if exists {
		return perrors.New(perrors.ErrDictTypeExists, fmt.Sprintf("字典编码 %s 已存在", req.Code))
	}
	dt.Name = req.Name
	dt.Code = req.Code
	dt.Status = req.Status
	dt.Remark = req.Remark
	return s.dictRepo.UpdateDictType(ctx, dt)
}

func (s *DictService) DeleteDictType(ctx context.Context, id uint64) error {
	_, err := s.dictRepo.FindDictTypeByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.ErrDictTypeExists, "字典类型不存在")
		}
		return err
	}
	return s.dictRepo.DeleteDictType(ctx, id)
}

// --- 字典数据 ---

func (s *DictService) ListDictDatas(ctx context.Context, req *dto.DictDataListRequest) (*dto.PageResponse, error) {
	dds, total, err := s.dictRepo.ListDictDatas(ctx, req.DictTypeID, req.Label, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	list := make([]dto.DictDataResponse, 0, len(dds))
	for _, dd := range dds {
		list = append(list, dto.DictDataResponse{
			ID:         dd.ID,
			DictTypeID: dd.DictTypeID,
			Label:      dd.Label,
			Value:      dd.Value,
			Sort:       dd.Sort,
			IsDefault:  dd.IsDefault,
			Status:     dd.Status,
			Remark:     dd.Remark,
			CreatedAt:  dd.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &dto.PageResponse{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

func (s *DictService) GetDictDatasByTypeID(ctx context.Context, dictTypeID uint64) ([]dto.DictDataResponse, error) {
	dds, err := s.dictRepo.FindDictDataByTypeID(ctx, dictTypeID)
	if err != nil {
		return nil, err
	}
	return toDictDataList(dds), nil
}

// GetDictDatasByCode 按字典编码获取启用中的字典项（业务下拉用）
func (s *DictService) GetDictDatasByCode(ctx context.Context, code string) ([]dto.DictDataResponse, error) {
	dt, err := s.dictRepo.FindDictTypeByCode(ctx, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, perrors.New(perrors.NotFound, "字典类型不存在")
		}
		return nil, err
	}
	if dt.Status != 1 {
		return []dto.DictDataResponse{}, nil
	}
	dds, err := s.dictRepo.FindEnabledDictDataByTypeID(ctx, dt.ID)
	if err != nil {
		return nil, err
	}
	return toDictDataList(dds), nil
}

func toDictDataList(dds []model.DictData) []dto.DictDataResponse {
	list := make([]dto.DictDataResponse, 0, len(dds))
	for _, dd := range dds {
		list = append(list, dto.DictDataResponse{
			ID:         dd.ID,
			DictTypeID: dd.DictTypeID,
			Label:      dd.Label,
			Value:      dd.Value,
			Sort:       dd.Sort,
			IsDefault:  dd.IsDefault,
			Status:     dd.Status,
			Remark:     dd.Remark,
		})
	}
	return list
}

func (s *DictService) GetDictData(ctx context.Context, id uint64) (*dto.DictDataResponse, error) {
	dd, err := s.dictRepo.FindDictDataByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, perrors.New(perrors.NotFound, "字典数据不存在")
		}
		return nil, err
	}
	return &dto.DictDataResponse{
		ID:         dd.ID,
		DictTypeID: dd.DictTypeID,
		Label:      dd.Label,
		Value:      dd.Value,
		Sort:       dd.Sort,
		IsDefault:  dd.IsDefault,
		Status:     dd.Status,
		Remark:     dd.Remark,
		CreatedAt:  dd.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *DictService) CreateDictData(ctx context.Context, req *dto.CreateDictDataRequest) error {
	dd := &model.DictData{
		DictTypeID: req.DictTypeID,
		Label:      req.Label,
		Value:      req.Value,
		Sort:       req.Sort,
		IsDefault:  req.IsDefault,
		Status:     req.Status,
		Remark:     req.Remark,
	}
	if dd.Status == 0 {
		dd.Status = 1
	}
	return s.dictRepo.CreateDictData(ctx, dd)
}

func (s *DictService) UpdateDictData(ctx context.Context, req *dto.UpdateDictDataRequest) error {
	dd, err := s.dictRepo.FindDictDataByID(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.NotFound, "字典数据不存在")
		}
		return err
	}
	dd.DictTypeID = req.DictTypeID
	dd.Label = req.Label
	dd.Value = req.Value
	dd.Sort = req.Sort
	dd.IsDefault = req.IsDefault
	dd.Status = req.Status
	dd.Remark = req.Remark
	return s.dictRepo.UpdateDictData(ctx, dd)
}

func (s *DictService) DeleteDictData(ctx context.Context, id uint64) error {
	_, err := s.dictRepo.FindDictDataByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return perrors.New(perrors.NotFound, "字典数据不存在")
		}
		return err
	}
	return s.dictRepo.DeleteDictData(ctx, id)
}

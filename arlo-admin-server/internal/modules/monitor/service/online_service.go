package service

import (
	"context"
	"sort"

	"arlo-admin/internal/modules/monitor/dto"
	"arlo-admin/pkg/onlinesession"
)

type OnlineService struct{}

func NewOnlineService() *OnlineService {
	return &OnlineService{}
}

func (s *OnlineService) List(ctx context.Context, req *dto.OnlineQuery) (*dto.OnlineListResponse, error) {
	page, pageSize := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	sessions, err := onlinesession.List(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(sessions, func(i, j int) bool {
		return sessions[i].LoginAt > sessions[j].LoginAt
	})

	total := int64(len(sessions))
	start := (page - 1) * pageSize
	if start > len(sessions) {
		start = len(sessions)
	}
	end := start + pageSize
	if end > len(sessions) {
		end = len(sessions)
	}

	list := make([]dto.OnlineSessionItem, 0, end-start)
	for _, sess := range sessions[start:end] {
		list = append(list, dto.OnlineSessionItem{
			UserID:     sess.UserID,
			Username:   sess.Username,
			RefreshJTI: sess.RefreshJTI,
			IP:         sess.IP,
			Browser:    sess.Browser,
			OS:         sess.OS,
			LoginAt:    sess.LoginAt,
		})
	}

	return &dto.OnlineListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *OnlineService) Kick(ctx context.Context, req *dto.KickRequest) error {
	if req.SessionID != "" {
		return onlinesession.KickSession(ctx, req.UserID, req.SessionID)
	}
	return onlinesession.KickUser(ctx, req.UserID)
}

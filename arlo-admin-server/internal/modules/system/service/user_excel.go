package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"arlo-admin/internal/domain/model"
	configsvc "arlo-admin/internal/modules/sysconfig/service"
	"arlo-admin/internal/modules/system/dto"
	"arlo-admin/pkg/excel"
	"arlo-admin/pkg/security"

	"golang.org/x/crypto/bcrypt"
)

var userExportHeaders = []string{"用户名", "昵称", "手机号", "邮箱", "性别", "部门", "岗位", "角色", "状态", "备注", "创建时间"}

// ExportUsers 按筛选条件导出用户（最多 10000 条）
func (s *UserService) ExportUsers(ctx context.Context, req *dto.UserListRequest, currentUserID uint64) ([]byte, error) {
	scope, _ := s.buildDataScope(ctx, currentUserID)
	page, pageSize := 1, 10000
	users, _, err := s.userRepo.List(ctx, req.Username, req.Nickname, req.Phone, req.Status, req.DeptID, scope, page, pageSize)
	if err != nil {
		return nil, err
	}

	userIDs := make([]uint64, len(users))
	deptIDs := make([]uint64, 0, len(users))
	deptIDSet := make(map[uint64]bool)
	for i, u := range users {
		userIDs[i] = u.ID
		if u.DeptID > 0 && !deptIDSet[u.DeptID] {
			deptIDSet[u.DeptID] = true
			deptIDs = append(deptIDs, u.DeptID)
		}
	}
	roleMap, _ := s.userRepo.FindRolesByUserIDs(ctx, userIDs)
	postMap, _ := s.userRepo.FindPostsByUserIDs(ctx, userIDs)
	deptNameMap, _ := s.deptRepo.FindByIDs(ctx, deptIDs)

	rows := make([][]interface{}, 0, len(users))
	for _, u := range users {
		gender := "未知"
		switch u.Gender {
		case 1:
			gender = "男"
		case 2:
			gender = "女"
		}
		status := "禁用"
		if u.Status == 1 {
			status = "启用"
		}
		roles := roleMap[u.ID]
		roleNames := make([]string, len(roles))
		for i, r := range roles {
			roleNames[i] = r.Name
		}
		posts := postMap[u.ID]
		postNames := make([]string, len(posts))
		for i, p := range posts {
			postNames[i] = p.Name
		}
		rows = append(rows, []interface{}{
			u.Username, u.Nickname, u.Phone, u.Email, gender,
			deptNameMap[u.DeptID],
			strings.Join(postNames, ","),
			strings.Join(roleNames, ","),
			status, u.Remark,
			u.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return excel.Write(excel.Sheet{Name: "用户", Headers: userExportHeaders, Rows: rows})
}

// ImportTemplate 用户导入模板
func (s *UserService) ImportTemplate() ([]byte, error) {
	rows := [][]interface{}{
		{"zhangsan", "张三", "13800138000", "zhangsan@example.com", "男", "技术部", "开发工程师", "普通用户", "启用", "示例行可删", ""},
	}
	headers := []string{
		"用户名*", "昵称*", "手机号", "邮箱", "性别(未知/男/女)",
		"部门", "岗位(多个用逗号分隔)", "角色(多个用逗号分隔)",
		"状态(启用/禁用)", "备注", "初始密码(空则用系统默认)",
	}
	return excel.Write(excel.Sheet{Name: "用户导入", Headers: headers, Rows: rows})
}

// ImportUsers 批量导入用户；返回成功数与错误明细
func (s *UserService) ImportUsers(ctx context.Context, data []byte) (ok int, errs []string, err error) {
	_, rows, err := excel.ReadRows(data)
	if err != nil {
		return 0, nil, err
	}
	initPwd := "admin123"
	minLen, complex := 6, false
	if s.configSvc != nil {
		if v := s.configSvc.GetString(ctx, configsvc.KeyInitPwd, ""); v != "" {
			initPwd = v
		}
		minLen = s.configSvc.PwdMinLength(ctx)
		complex = s.configSvc.PwdRequireComplexity(ctx)
	}

	deptNameMap, e := s.buildDeptNameMap(ctx)
	if e != nil {
		return 0, nil, e
	}
	postNameMap, e := s.buildPostNameMap(ctx)
	if e != nil {
		return 0, nil, e
	}
	roleNameMap, e := s.buildRoleNameMap(ctx)
	if e != nil {
		return 0, nil, e
	}

	for i, row := range rows {
		line := i + 2
		username := strings.TrimSpace(excel.Cell(row, 0))
		nickname := strings.TrimSpace(excel.Cell(row, 1))
		if username == "" && nickname == "" {
			continue
		}
		if username == "" || nickname == "" {
			errs = append(errs, fmt.Sprintf("第%d行: 用户名和昵称必填", line))
			continue
		}
		exists, e := s.userRepo.ExistsByUsername(ctx, username, 0)
		if e != nil {
			errs = append(errs, fmt.Sprintf("第%d行: %v", line, e))
			continue
		}
		if exists {
			errs = append(errs, fmt.Sprintf("第%d行: 用户名 %s 已存在", line, username))
			continue
		}
		pwd := strings.TrimSpace(excel.Cell(row, 10))
		if pwd == "" {
			pwd = initPwd
		}
		if e := security.ValidatePassword(pwd, minLen, complex); e != nil {
			errs = append(errs, fmt.Sprintf("第%d行: %v", line, e))
			continue
		}
		hashed, e := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if e != nil {
			errs = append(errs, fmt.Sprintf("第%d行: 密码加密失败", line))
			continue
		}

		deptName := strings.TrimSpace(excel.Cell(row, 5))
		var deptID uint64
		if deptName != "" {
			id, ok := deptNameMap[deptName]
			if !ok {
				errs = append(errs, fmt.Sprintf("第%d行: 部门「%s」不存在", line, deptName))
				continue
			}
			deptID = id
		}

		postIDs, missing := resolveNames(excel.Cell(row, 6), postNameMap)
		if len(missing) > 0 {
			errs = append(errs, fmt.Sprintf("第%d行: 岗位「%s」不存在", line, strings.Join(missing, "、")))
			continue
		}
		roleIDs, missing := resolveNames(excel.Cell(row, 7), roleNameMap)
		if len(missing) > 0 {
			errs = append(errs, fmt.Sprintf("第%d行: 角色「%s」不存在", line, strings.Join(missing, "、")))
			continue
		}

		gender := parseGender(excel.Cell(row, 4))
		status := parseStatus(excel.Cell(row, 8))
		now := time.Now()
		user := &model.User{
			Username:      username,
			Password:      string(hashed),
			Nickname:      nickname,
			Phone:         strings.TrimSpace(excel.Cell(row, 2)),
			Email:         strings.TrimSpace(excel.Cell(row, 3)),
			Gender:        gender,
			DeptID:        deptID,
			Status:        status,
			Remark:        strings.TrimSpace(excel.Cell(row, 9)),
			PwdUpdatedAt:  &now,
			MustChangePwd: 1,
		}
		if e := s.userRepo.Create(ctx, user); e != nil {
			errs = append(errs, fmt.Sprintf("第%d行: 创建失败 %v", line, e))
			continue
		}
		if len(roleIDs) > 0 {
			_ = s.userRepo.AssignRoles(ctx, user.ID, roleIDs)
		}
		if len(postIDs) > 0 {
			_ = s.userRepo.AssignPosts(ctx, user.ID, postIDs)
		}
		ok++
	}
	if ok > 0 {
		s.reloadPolicies(ctx)
	}
	return ok, errs, nil
}

func (s *UserService) buildDeptNameMap(ctx context.Context) (map[string]uint64, error) {
	depts, err := s.deptRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]uint64, len(depts))
	for _, d := range depts {
		m[d.Name] = d.ID
	}
	return m, nil
}

func (s *UserService) buildPostNameMap(ctx context.Context) (map[string]uint64, error) {
	posts, err := s.postRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]uint64, len(posts))
	for _, p := range posts {
		m[p.Name] = p.ID
	}
	return m, nil
}

func (s *UserService) buildRoleNameMap(ctx context.Context) (map[string]uint64, error) {
	roles, err := s.roleRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]uint64, len(roles))
	for _, r := range roles {
		m[r.Name] = r.ID
	}
	return m, nil
}

// resolveNames 将逗号/顿号分隔的中文名解析为 ID；返回找不到的名称列表
func resolveNames(raw string, nameMap map[string]uint64) (ids []uint64, missing []string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	raw = strings.ReplaceAll(raw, "，", ",")
	raw = strings.ReplaceAll(raw, "、", ",")
	parts := strings.Split(raw, ",")
	seen := make(map[uint64]bool)
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		id, ok := nameMap[name]
		if !ok {
			missing = append(missing, name)
			continue
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids, missing
}

func parseGender(s string) int8 {
	switch strings.TrimSpace(s) {
	case "男", "1":
		return 1
	case "女", "2":
		return 2
	default:
		return 0
	}
}

func parseStatus(s string) int8 {
	switch strings.TrimSpace(s) {
	case "禁用", "0":
		return 0
	default:
		return 1
	}
}

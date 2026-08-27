package casbin

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"arlo-admin/internal/domain/repository"

	casbinSDK "github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

// Enforcer Casbin 权限执行器封装
type Enforcer struct {
	enforcer *casbinSDK.Enforcer
	roleRepo *repository.RoleRepository
	mu       sync.RWMutex
}

// NewEnforcer 创建 Casbin 执行器
// modelPath: rbac_model.conf 文件路径
func NewEnforcer(modelPath string, roleRepo *repository.RoleRepository) (*Enforcer, error) {
	m, err := model.NewModelFromFile(modelPath)
	if err != nil {
		return nil, fmt.Errorf("加载 Casbin 模型失败: %w", err)
	}

	e, err := casbinSDK.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("创建 Casbin Enforcer 失败: %w", err)
	}

	return &Enforcer{
		enforcer: e,
		roleRepo: roleRepo,
	}, nil
}

// LoadPolicies 从数据库加载策略到 Casbin
// 加载两类策略:
//  1. p 策略 (policy): 角色 → URL路径 → 允许的HTTP方法
//  2. g 策略 (grouping): 用户ID → 角色
func (e *Enforcer) LoadPolicies(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 清除现有策略
	e.enforcer.ClearPolicy()

	// 加载 p 策略: 角色 → API路径 → HTTP方法
	roleMenus, err := e.roleRepo.FindAllRoleMenus(ctx)
	if err != nil {
		return fmt.Errorf("加载角色菜单失败: %w", err)
	}

	seen := make(map[string]bool)
	addPolicy := func(roleCode, apiPath string) error {
		key := roleCode + ":" + apiPath
		if seen[key] {
			return nil
		}
		seen[key] = true
		if _, err := e.enforcer.AddPolicy(roleCode, apiPath, "(GET|POST|PUT|DELETE)"); err != nil {
			return fmt.Errorf("添加策略失败 [%s, %s]: %w", roleCode, apiPath, err)
		}
		return nil
	}

	for _, rm := range roleMenus {
		for _, apiPath := range resolveAPIPaths(rm.MenuPath) {
			// 同时登记「精确路径」与「子路径 /*」：
			// keyMatch2 下 /system/menu/* 不能匹配 PUT /system/menu（无额外段），
			// 会导致能进页面/能看列表，但新增/编辑保存被拒。
			paths := expandAPIPaths(apiPath)
			for _, p := range paths {
				if err := addPolicy(rm.RoleCode, p); err != nil {
					return err
				}
			}
		}
	}

	// 超级管理员：始终放行全部 API（避免新接口未配菜单时也被拦）
	roles, err := e.roleRepo.FindAllRoles(ctx)
	if err != nil {
		return fmt.Errorf("加载角色列表失败: %w", err)
	}
	for _, role := range roles {
		if role.Code == "super_admin" {
			if err := addPolicy(role.Code, "/api/v1/*"); err != nil {
				return fmt.Errorf("添加超管策略失败: %w", err)
			}
		}
	}

	// 加载 g 策略: 用户 → 角色
	userRoles, err := e.roleRepo.FindAllUserRoles(ctx)
	if err != nil {
		return fmt.Errorf("加载用户角色失败: %w", err)
	}

	for _, ur := range userRoles {
		userIDStr := formatUserID(ur.UserID)
		_, err := e.enforcer.AddGroupingPolicy(userIDStr, ur.RoleCode)
		if err != nil {
			return fmt.Errorf("添加角色分配失败 [%s, %s]: %w", userIDStr, ur.RoleCode, err)
		}
	}

	return nil
}

// resolveAPIPaths 将前端菜单 path 映射为后端 API 路径列表
// 多数菜单 path 与 API 一致（如 /system/user → /api/v1/system/user/*）
// 「我的消息」单独列出接口，避免 /* 覆盖 /message/notice/*
func resolveAPIPaths(menuPath string) []string {
	switch menuPath {
	case "/message/my":
		return []string{
			"/api/v1/message/list",
			"/api/v1/message/send",
			"/api/v1/message/read-all",
			"/api/v1/message/unread-count",
			"/api/v1/message/:id",
			"/api/v1/message/:id/read",
		}
	case "/system/file":
		return []string{"/api/v1/file"}
	case "/system/sysconfig", "/system/config":
		return []string{"/api/v1/sysconfig"}
	case "/system/dict":
		// 覆盖 /dict/type/* 与 /dict/data/*
		return []string{"/api/v1/system/dict"}
	case "/system/member":
		return []string{"/api/v1/system/member"}
	case "/log/operation":
		return []string{"/api/v1/log/operation"}
	case "/log/login":
		return []string{"/api/v1/log/login"}
	case "/monitor/online":
		return []string{"/api/v1/monitor/online"}
	case "/monitor/server":
		return []string{"/api/v1/monitor/server"}
	case "/monitor/job":
		return []string{"/api/v1/monitor/job"}
	default:
		return []string{"/api/v1" + menuPath}
	}
}

// expandAPIPaths 生成 Casbin 策略路径：精确路径 + 可选子路径通配
func expandAPIPaths(apiPath string) []string {
	if strings.Contains(apiPath, "*") || strings.Contains(apiPath, ":") || isExactAPIPath(apiPath) {
		return []string{apiPath}
	}
	return []string{apiPath, apiPath + "/*"}
}

// isExactAPIPath 标记不应再追加 /* 的精确策略路径
func isExactAPIPath(apiPath string) bool {
	switch apiPath {
	case "/api/v1/message/list",
		"/api/v1/message/send",
		"/api/v1/message/read-all",
		"/api/v1/message/unread-count",
		"/api/v1/monitor/server":
		return true
	default:
		return false
	}
}

// Enforce 权限校验
// sub: 用户ID 字符串
// obj: 请求路径
// act: HTTP 方法
func (e *Enforcer) Enforce(sub, obj, act string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.Enforce(sub, obj, act)
}

// ReloadPolicies 重新加载策略（角色/菜单变更后调用）
func (e *Enforcer) ReloadPolicies(ctx context.Context) error {
	return e.LoadPolicies(ctx)
}

// formatUserID 将 uint64 的用户 ID 格式化为字符串
func formatUserID(id uint64) string {
	return fmt.Sprintf("%d", id)
}

package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	domainmodel "arlo-admin/internal/domain/model"
	"arlo-admin/internal/domain/repository"
	"arlo-admin/internal/modules/auth/dto"
	"arlo-admin/internal/modules/log/model"
	logrepo "arlo-admin/internal/modules/log/repository"
	configsvc "arlo-admin/internal/modules/sysconfig/service"
	"arlo-admin/pkg/captcha"
	jwtpkg "arlo-admin/pkg/jwt"
	"arlo-admin/pkg/onlinesession"
	"arlo-admin/pkg/security"
	"arlo-admin/pkg/tokenblacklist"
	"arlo-admin/pkg/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound     = errors.New("用户不存在")
	ErrPasswordWrong    = errors.New("密码错误")
	ErrUserDisabled     = errors.New("用户已被禁用")
	ErrRefreshTokenType = errors.New("刷新令牌类型错误")
	ErrCaptchaInvalid   = errors.New("验证码错误")
	ErrOldPasswordWrong = errors.New("原密码错误")
	ErrAccountLocked    = errors.New("账号已锁定")
	ErrPasswordWeak     = errors.New("密码不符合安全策略")
)

// AuthService 认证服务 — 只负责业务逻辑，数据访问委托给 Repository
type AuthService struct {
	userRepo     *repository.UserRepository
	roleRepo     *repository.RoleRepository
	menuRepo     *repository.MenuRepository
	loginLogRepo *logrepo.LoginLogRepository
	configSvc    *configsvc.ConfigService
}

// NewAuthService 创建 AuthService
func NewAuthService(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	menuRepo *repository.MenuRepository,
	loginLogRepo *logrepo.LoginLogRepository,
	configSvc *configsvc.ConfigService,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		menuRepo:     menuRepo,
		loginLogRepo: loginLogRepo,
		configSvc:    configSvc,
	}
}

// Login 用户登录
// 1. 查询用户（委托 Repository）
// 2. 验证密码
// 3. 检查状态
// 4. 生成 access + refresh token
// 5. 更新最后登录时间（委托 Repository）
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	req.Username = strings.TrimSpace(req.Username)
	// 1. 按系统配置决定是否校验验证码
	if s.configSvc == nil || s.configSvc.IsCaptchaEnabled(ctx) {
		if !captcha.Verify(req.CaptchaID, req.CaptchaCode) {
			return nil, ErrCaptchaInvalid
		}
	}

	// 账号锁定检查（Redis）
	if locked, ttl := security.IsLocked(ctx, req.Username); locked {
		_ = s.loginLogRepo.Create(ctx, &model.LoginLog{
			Username:  req.Username,
			IP:        req.IP,
			Browser:   req.Browser,
			OS:        req.OS,
			Status:    0,
			Msg:       fmt.Sprintf("账号锁定中，约 %d 分钟后重试", int(ttl.Minutes())+1),
			CreatedAt: time.Now(),
		})
		return nil, fmt.Errorf("%w: 约 %d 分钟后重试", ErrAccountLocked, int(ttl.Minutes())+1)
	}

	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.recordPwdFail(ctx, req, "用户不存在")
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		msg := s.recordPwdFail(ctx, req, "密码错误")
		if msg != "" && strings.Contains(msg, "锁定") {
			return nil, fmt.Errorf("%w: %s", ErrAccountLocked, msg)
		}
		if msg != "" {
			return nil, fmt.Errorf("%w: %s", ErrPasswordWrong, msg)
		}
		return nil, ErrPasswordWrong
	}

	// 检查用户状态
	if !user.IsEnabled() {
		_ = s.loginLogRepo.Create(ctx, &model.LoginLog{
			Username:  req.Username,
			IP:        req.IP,
			Browser:   req.Browser,
			OS:        req.OS,
			Status:    0,
			Msg:       "用户已被禁用",
			CreatedAt: time.Now(),
		})
		return nil, ErrUserDisabled
	}

	security.ClearLoginFail(ctx, req.Username)

	// 生成 Token
	accessToken, expiresIn, err := jwtpkg.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwtpkg.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, err
	}

	// 记录登录成功日志
	_ = s.loginLogRepo.Create(ctx, &model.LoginLog{
		Username:  req.Username,
		IP:        req.IP,
		Browser:   req.Browser,
		OS:        req.OS,
		Status:    1,
		Msg:       "登录成功",
		CreatedAt: time.Now(),
	})

	// 登记在线会话（Redis 不可用时静默跳过）
	accessClaims, _ := jwtpkg.ParseToken(accessToken)
	refreshClaims, _ := jwtpkg.ParseToken(refreshToken)
	if accessClaims != nil && refreshClaims != nil {
		onlinesession.Register(ctx, onlinesession.Session{
			UserID:     user.ID,
			Username:   user.Username,
			AccessJTI:  accessClaims.ID,
			RefreshJTI: refreshClaims.ID,
			IP:         req.IP,
			Browser:    req.Browser,
			OS:         req.OS,
			LoginAt:    utils.FormatTime(time.Now()),
		})
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// RefreshToken 刷新访问令牌
// 使用 refresh token 换取新的 access token
func (s *AuthService) RefreshToken(ctx context.Context, req *dto.RefreshRequest) (*dto.RefreshResponse, error) {
	claims, err := jwtpkg.ParseToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// 必须是 refresh 类型的 token 才能用来刷新
	if claims.Subject != "refresh" {
		return nil, ErrRefreshTokenType
	}

	// 已登出的 refresh 不可再续期
	if tokenblacklist.IsBlacklisted(ctx, req.RefreshToken, claims) {
		return nil, jwtpkg.ErrTokenInvalid
	}

	// 强制下线后不可续期
	if claims.IssuedAt != nil && onlinesession.IsKicked(ctx, claims.UserID, claims.IssuedAt.Time) {
		return nil, jwtpkg.ErrTokenInvalid
	}

	// 生成新的 access token
	accessToken, expiresIn, err := jwtpkg.GenerateAccessToken(claims.UserID, claims.Username)
	if err != nil {
		return nil, err
	}

	if accessClaims, err := jwtpkg.ParseToken(accessToken); err == nil {
		onlinesession.TouchAccess(ctx, claims.UserID, claims.ID, accessClaims.ID)
	}

	return &dto.RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	}, nil
}

// Logout 登出：将 access / refresh 写入 Redis 黑名单直至各自过期
func (s *AuthService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	if accessToken != "" {
		if claims, err := jwtpkg.ParseToken(accessToken); err == nil {
			_ = tokenblacklist.Add(ctx, accessToken, claims)
			onlinesession.UnregisterByAccess(ctx, claims.UserID, claims.ID)
		}
	}
	if refreshToken != "" {
		if claims, err := jwtpkg.ParseToken(refreshToken); err == nil {
			// 仅接受 refresh 类型；误传 access 也不重复处理（上面已加）
			if claims.Subject == "" || claims.Subject == "refresh" {
				_ = tokenblacklist.Add(ctx, refreshToken, claims)
				onlinesession.UnregisterByRefresh(ctx, claims.UserID, claims.ID)
			}
		}
	}
	return nil
}

// GetUserInfo 获取当前登录用户信息
// 包含用户基本信息和角色/权限列表
func (s *AuthService) GetUserInfo(ctx context.Context, userID uint64) (*dto.UserInfoResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// 查询用户角色
	roles, err := s.roleRepo.FindRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, 0, len(roles))
	roleIDs := make([]uint64, 0, len(roles))
	for _, r := range roles {
		roleNames = append(roleNames, r.Name)
		roleIDs = append(roleIDs, r.ID)
	}

	// 查询角色权限
	permissions, err := s.menuRepo.FindPermissionsByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	// 计算数据权限范围（取最宽松: 最小值）
	dataScope := int8(1)
	if len(roles) > 0 {
		dataScope = 5 // 初始最严格
		for _, r := range roles {
			ds := r.DataScope
			if ds == 0 {
				ds = 1 // 兼容旧数据
			}
			if ds < dataScope {
				dataScope = ds
			}
		}
	}

	// 自定义数据权限时收集可见部门ID
	var deptIDs []uint64
	if dataScope == 2 {
		for _, r := range roles {
			if r.DataScope == 2 {
				ids, _ := s.roleRepo.FindRoleDeptIDs(ctx, r.ID)
				deptIDs = append(deptIDs, ids...)
			}
		}
	}

	postNames, _ := s.userRepo.FindPostNamesByUserID(ctx, userID)
	deptName := s.userRepo.FindDeptName(ctx, user.DeptID)

	mustChange := user.MustChangePwd == 1
	pwdExpired := s.isPwdExpired(ctx, user.PwdUpdatedAt)

	return &dto.UserInfoResponse{
		ID:            user.ID,
		Username:      user.Username,
		Nickname:      user.Nickname,
		Avatar:        user.Avatar,
		Email:         user.Email,
		Phone:         user.Phone,
		Gender:        user.Gender,
		DeptID:        user.DeptID,
		DeptName:      deptName,
		Status:        user.Status,
		Remark:        user.Remark,
		RoleNames:     roleNames,
		PostNames:     postNames,
		Permissions:   permissions,
		DataScope:     dataScope,
		DeptIDs:       deptIDs,
		MustChangePwd: mustChange,
		PwdExpired:    pwdExpired,
	}, nil
}

// UpdateProfile 更新当前用户个人资料
func (s *AuthService) UpdateProfile(ctx context.Context, userID uint64, req *dto.UpdateProfileRequest) error {
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return s.userRepo.UpdateProfile(ctx, userID, req.Nickname, req.Gender, req.Phone, req.Email, req.Remark, req.Avatar)
}

// ChangePassword 修改当前用户密码（需校验原密码）
func (s *AuthService) ChangePassword(ctx context.Context, userID uint64, req *dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return ErrOldPasswordWrong
	}
	if err := s.validateNewPassword(ctx, req.NewPassword); err != nil {
		return err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(ctx, userID, string(hashed), true)
}

func (s *AuthService) recordPwdFail(ctx context.Context, req *dto.LoginRequest, baseMsg string) string {
	maxRetry := 5
	lockMin := 30
	if s.configSvc != nil {
		maxRetry = s.configSvc.LoginMaxRetry(ctx)
		lockMin = s.configSvc.LoginLockMinutes(ctx)
	}
	locked, tip := security.RecordLoginFail(ctx, req.Username, maxRetry, lockMin)
	msg := baseMsg
	if tip != "" {
		msg = tip
	} else if locked {
		msg = tip
	}
	_ = s.loginLogRepo.Create(ctx, &model.LoginLog{
		Username:  req.Username,
		IP:        req.IP,
		Browser:   req.Browser,
		OS:        req.OS,
		Status:    0,
		Msg:       msg,
		CreatedAt: time.Now(),
	})
	return tip
}

func (s *AuthService) validateNewPassword(ctx context.Context, password string) error {
	minLen := 6
	complex := false
	if s.configSvc != nil {
		minLen = s.configSvc.PwdMinLength(ctx)
		complex = s.configSvc.PwdRequireComplexity(ctx)
	}
	if err := security.ValidatePassword(password, minLen, complex); err != nil {
		return fmt.Errorf("%w: %s", ErrPasswordWeak, err.Error())
	}
	return nil
}

func (s *AuthService) isPwdExpired(ctx context.Context, pwdUpdatedAt *time.Time) bool {
	days := 0
	if s.configSvc != nil {
		days = s.configSvc.PwdExpireDays(ctx)
	}
	if days <= 0 {
		return false
	}
	if pwdUpdatedAt == nil {
		return true
	}
	return time.Since(*pwdUpdatedAt) > time.Duration(days)*24*time.Hour
}

// GetUserMenus 获取当前用户可见菜单树（按角色过滤，仅目录/菜单）
func (s *AuthService) GetUserMenus(ctx context.Context, userID uint64) ([]*dto.MenuTreeNode, error) {
	roles, err := s.roleRepo.FindRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	roleIDs := make([]uint64, 0, len(roles))
	for _, r := range roles {
		roleIDs = append(roleIDs, r.ID)
	}

	menus, err := s.menuRepo.FindMenusByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	// 补全父级目录：角色只勾了子菜单时，侧边栏仍需父级才能成树
	allMenus, err := s.menuRepo.FindAllVisible(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[uint64]*domainmodel.Menu, len(allMenus))
	for _, m := range allMenus {
		byID[m.ID] = m
	}
	selected := make(map[uint64]*domainmodel.Menu)
	for _, m := range menus {
		selected[m.ID] = m
	}
	for _, m := range menus {
		pid := m.ParentID
		for pid > 0 {
			if _, ok := selected[pid]; ok {
				break
			}
			parent, ok := byID[pid]
			if !ok {
				break
			}
			selected[pid] = parent
			pid = parent.ParentID
		}
	}
	merged := make([]*domainmodel.Menu, 0, len(selected))
	for _, m := range selected {
		merged = append(merged, m)
	}

	// 侧边栏/路由只需目录和菜单，按钮权限走 permissions
	filtered := make([]*domainmodel.Menu, 0, len(merged))
	for _, m := range merged {
		if m.Type == 1 || m.Type == 2 {
			filtered = append(filtered, m)
		}
	}
	// map 合并会打乱顺序，建树前按 sort、id 排回配置顺序
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].Sort != filtered[j].Sort {
			return filtered[i].Sort < filtered[j].Sort
		}
		return filtered[i].ID < filtered[j].ID
	})
	return buildMenuTree(filtered, 0), nil
}

func buildMenuTree(menus []*domainmodel.Menu, parentID uint64) []*dto.MenuTreeNode {
	tree := make([]*dto.MenuTreeNode, 0)
	for _, m := range menus {
		if m.ParentID != parentID {
			continue
		}
		node := &dto.MenuTreeNode{
			ID:         m.ID,
			ParentID:   m.ParentID,
			Name:       m.Name,
			Type:       m.Type,
			Path:       m.Path,
			Component:  m.Component,
			Icon:       m.Icon,
			Sort:       m.Sort,
			Permission: m.Permission,
			Visible:    m.Visible,
			KeepAlive:  m.KeepAlive,
		}
		children := buildMenuTree(menus, m.ID)
		if len(children) > 0 {
			node.Children = children
		}
		tree = append(tree, node)
	}
	sort.SliceStable(tree, func(i, j int) bool {
		if tree[i].Sort != tree[j].Sort {
			return tree[i].Sort < tree[j].Sort
		}
		return tree[i].ID < tree[j].ID
	})
	return tree
}

// GenerateCaptcha 生成图形验证码
func (s *AuthService) GenerateCaptcha() (*dto.CaptchaResponse, error) {
	id, b64s, err := captcha.Generate()
	if err != nil {
		return nil, err
	}
	return &dto.CaptchaResponse{
		CaptchaID: id,
		Captcha:   b64s,
	}, nil
}

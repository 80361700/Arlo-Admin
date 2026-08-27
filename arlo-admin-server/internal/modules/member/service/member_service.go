package service

import (
	"context"
	"errors"
	"strings"

	"arlo-admin/internal/modules/member/dto"
	"arlo-admin/internal/modules/member/model"
	"arlo-admin/internal/modules/member/repository"
	configsvc "arlo-admin/internal/modules/sysconfig/service"
	"arlo-admin/pkg/jwt"
	"arlo-admin/pkg/security"
	"arlo-admin/pkg/verify"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type MemberService struct {
	memberRepo *repository.MemberRepository
	configSvc  *configsvc.ConfigService
}

func NewMemberService(memberRepo *repository.MemberRepository, configSvc *configsvc.ConfigService) *MemberService {
	return &MemberService{memberRepo: memberRepo, configSvc: configSvc}
}

// SendCode 发送手机验证码
func (s *MemberService) SendCode(ctx context.Context, phone string) error {
	return verify.SendCode(ctx, phone)
}

// Login 手机号 + 验证码登录（首次登录自动注册）
func (s *MemberService) Login(ctx context.Context, phone string, code string) (*dto.LoginResponse, error) {
	// 1. 校验验证码
	if !verify.VerifyCode(ctx, phone, code) {
		return nil, errors.New("验证码错误")
	}

	// 2. 查找或创建会员
	member, err := s.memberRepo.FindByPhone(ctx, phone)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		member = &model.Member{
			Phone:    phone,
			Nickname: "用户" + phone[7:],
			Source:   "h5",
			Status:   1,
		}
		if err := s.memberRepo.Create(ctx, member); err != nil {
			return nil, errors.New("注册失败: " + err.Error())
		}
	} else if err != nil {
		return nil, errors.New("查询用户失败: " + err.Error())
	}

	// 3. 检查状态
	if !member.IsEnabled() {
		return nil, errors.New("账号已被禁用")
	}

	// 4. 生成 Token
	accessToken, expiresIn, err := jwt.GenerateMemberAccessToken(member.ID, member.Phone)
	if err != nil {
		return nil, errors.New("生成token失败: " + err.Error())
	}
	refreshToken, err := jwt.GenerateMemberRefreshToken(member.ID, member.Phone)
	if err != nil {
		return nil, errors.New("生成refreshToken失败: " + err.Error())
	}

	// 5. 更新最后登录时间
	_ = s.memberRepo.UpdateLastLogin(ctx, member.ID)

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// RefreshToken 刷新 Token
func (s *MemberService) RefreshToken(tokenString string) (*dto.LoginResponse, error) {
	claims, err := jwt.ParseMemberToken(tokenString)
	if err != nil {
		return nil, errors.New("refresh token无效或已过期")
	}
	if claims.Subject != "" && claims.Subject != "member-refresh" {
		return nil, errors.New("非法的刷新令牌类型")
	}

	member, err := s.memberRepo.FindByID(context.Background(), claims.MemberID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	if !member.IsEnabled() {
		return nil, errors.New("账号已被禁用")
	}

	accessToken, expiresIn, err := jwt.GenerateMemberAccessToken(member.ID, member.Phone)
	if err != nil {
		return nil, err
	}
	newRefreshToken, err := jwt.GenerateMemberRefreshToken(member.ID, member.Phone)
	if err != nil {
		return nil, err
	}

	if err := s.memberRepo.UpdateLastLogin(context.Background(), member.ID); err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// GetInfo 获取会员信息
func (s *MemberService) GetInfo(ctx context.Context, memberID uint64) (*dto.MemberInfoResponse, error) {
	member, err := s.memberRepo.FindByID(ctx, memberID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	return &dto.MemberInfoResponse{
		ID:        member.ID,
		Phone:     member.Phone,
		Nickname:  member.Nickname,
		Avatar:    member.Avatar,
		Gender:    member.Gender,
		Source:    member.Source,
		Status:    member.Status,
		CreatedAt: member.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateProfile 更新个人资料
func (s *MemberService) UpdateProfile(ctx context.Context, memberID uint64, req *dto.UpdateProfileRequest) error {
	member, err := s.memberRepo.FindByID(ctx, memberID)
	if err != nil {
		return errors.New("用户不存在")
	}

	member.Nickname = req.Nickname
	member.Avatar = req.Avatar
	member.Gender = req.Gender

	return s.memberRepo.Update(ctx, member)
}

// List 管理员分页查询会员列表
func (s *MemberService) List(ctx context.Context, req *dto.PageRequest) ([]dto.MemberItem, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	members, total, err := s.memberRepo.List(ctx, req.Page, req.PageSize, req.Phone, req.Nickname, req.Source, req.Status)
	if err != nil {
		return nil, 0, err
	}

	items := make([]dto.MemberItem, 0, len(members))
	for _, m := range members {
		items = append(items, toMemberItem(&m))
	}

	return items, total, nil
}

// GetDetail 管理员查看会员详情
func (s *MemberService) GetDetail(ctx context.Context, id uint64) (*dto.MemberDetailResponse, error) {
	m, err := s.memberRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("会员不存在")
		}
		return nil, err
	}
	lastLogin := ""
	if m.LastLogin != nil {
		lastLogin = m.LastLogin.Format("2006-01-02 15:04:05")
	}
	return &dto.MemberDetailResponse{
		ID:        m.ID,
		Phone:     m.Phone,
		Nickname:  m.Nickname,
		Avatar:    m.Avatar,
		Gender:    m.Gender,
		Openid:    m.Openid,
		Unionid:   m.Unionid,
		MpOpenid:  m.MpOpenid,
		Source:    m.Source,
		Status:    m.Status,
		LastLogin: lastLogin,
		CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: m.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// AdminCreate 管理员手动录入会员
func (s *MemberService) AdminCreate(ctx context.Context, req *dto.AdminCreateMemberRequest) error {
	exists, err := s.memberRepo.ExistsByPhone(ctx, req.Phone, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("手机号已存在")
	}
	pwd := strings.TrimSpace(req.Password)
	if pwd == "" {
		pwd = "123456"
		if s.configSvc != nil {
			if v := s.configSvc.GetString(ctx, configsvc.KeyInitPwd, ""); v != "" {
				pwd = v
			}
		}
	}
	if err := s.validatePassword(ctx, pwd); err != nil {
		return err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	nickname := req.Nickname
	if nickname == "" {
		nickname = "用户" + req.Phone[7:]
	}
	member := &model.Member{
		Phone:    req.Phone,
		Password: string(hashed),
		Nickname: nickname,
		Avatar:   req.Avatar,
		Gender:   req.Gender,
		Source:   req.Source,
		Status:   req.Status,
	}
	return s.memberRepo.Create(ctx, member)
}

// AdminUpdate 管理员更新会员资料
func (s *MemberService) AdminUpdate(ctx context.Context, req *dto.AdminUpdateMemberRequest) error {
	m, err := s.memberRepo.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("会员不存在")
		}
		return err
	}
	m.Nickname = req.Nickname
	m.Avatar = req.Avatar
	m.Gender = req.Gender
	m.Source = req.Source
	m.Status = req.Status
	return s.memberRepo.Update(ctx, m)
}

// UpdatePassword 管理员重置会员密码
func (s *MemberService) UpdatePassword(ctx context.Context, req *dto.UpdateMemberPasswordRequest) error {
	_, err := s.memberRepo.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("会员不存在")
		}
		return err
	}
	if err := s.validatePassword(ctx, req.Password); err != nil {
		return err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	return s.memberRepo.UpdatePassword(ctx, req.ID, string(hashed))
}

func (s *MemberService) validatePassword(ctx context.Context, password string) error {
	minLen, complex := 6, false
	if s.configSvc != nil {
		minLen = s.configSvc.PwdMinLength(ctx)
		complex = s.configSvc.PwdRequireComplexity(ctx)
	}
	if err := security.ValidatePassword(password, minLen, complex); err != nil {
		return err
	}
	return nil
}

// UpdateStatus 管理员更新会员状态
func (s *MemberService) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	_, err := s.memberRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("会员不存在")
		}
		return err
	}
	return s.memberRepo.UpdateStatus(ctx, id, status)
}

// Delete 管理员软删除会员
func (s *MemberService) Delete(ctx context.Context, id uint64) error {
	_, err := s.memberRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("会员不存在")
		}
		return err
	}
	return s.memberRepo.Delete(ctx, id)
}

func toMemberItem(m *model.Member) dto.MemberItem {
	lastLogin := ""
	if m.LastLogin != nil {
		lastLogin = m.LastLogin.Format("2006-01-02 15:04:05")
	}
	return dto.MemberItem{
		ID:        m.ID,
		Phone:     m.Phone,
		Nickname:  m.Nickname,
		Avatar:    m.Avatar,
		Gender:    m.Gender,
		Source:    m.Source,
		Status:    m.Status,
		LastLogin: lastLogin,
		CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

package engine

import (
	"fmt"
)

// UserRef 解析出的审批人
type UserRef struct {
	ID   uint64
	Name string
}

// OrgStore 组织数据访问（由外部注入）
type OrgStore interface {
	GetUser(id uint64) (name string, deptID uint64, ok bool)
	ListUserIDsByRole(roleID uint64) ([]uint64, error)
	ListUserIDsByDept(deptID uint64) ([]uint64, error)
	GetDeptLeader(deptID uint64) (leaderID uint64, ok bool)
	GetDeptParent(deptID uint64) (parentID uint64, ok bool)
	UserRoleIDs(userID uint64) ([]uint64, error)
	UserInDept(userID, deptID uint64) bool
}

// ResolveAssignees 按节点 setType 解析审批人
// selected: 发起人自选时前端传入的用户列表
func ResolveAssignees(node *Node, initiatorID uint64, selected []UserRef, org OrgStore) ([]UserRef, error) {
	if node == nil {
		return nil, fmt.Errorf("节点为空")
	}
	setType := node.SetType
	if setType == 0 {
		setType = 1
	}
	var users []UserRef
	var miss []uint64
	var err error
	switch setType {
	case 1: // 指定成员
		users, miss = fromAssigneeListDetailed(node.NodeAssigneeList, org)
		if len(users) == 0 {
			return nil, emptyDesignatedErr(node, miss)
		}
	case 2: // 主管
		users, err = resolveManagers(initiatorID, node.ExamineLevel, false, 0, org)
	case 3: // 角色（按节点配置解析）
		users, err = fromRoles(node.NodeAssigneeList, org)
	case 4: // 发起人自选（selectMode=3 为自选角色）
		if len(selected) > 0 {
			if err = validateAgainstCandidates(node, selected); err != nil {
				return nil, err
			}
			if node.SelectMode == 3 {
				roleList := make([]Assignee, 0, len(selected))
				for _, u := range selected {
					roleList = append(roleList, Assignee{ID: u.ID, Name: u.Name})
				}
				users, err = fromRoles(roleList, org)
			} else {
				users = selected
			}
		} else {
			// 无自选结果时不把候选人名单当成审批人（设计器 nodeCandidate 仅是可选范围）
			users, _ = fromAssigneeListDetailed(node.NodeAssigneeList, org)
			if len(users) == 0 {
				return nil, fmt.Errorf("节点[%s]为发起人自选，但未传入所选审批人（子流程自动启动时无法自选，请改为指定成员/角色/主管/发起人自己）", node.NodeName)
			}
		}
	case 5: // 发起人自己
		name, _, ok := org.GetUser(initiatorID)
		if !ok {
			return nil, fmt.Errorf("发起人不存在")
		}
		users = []UserRef{{ID: initiatorID, Name: name}}
	case 6: // 连续多级主管
		endLevel := directorEndLevel(node)
		if node.DirectorMode == 0 {
			endLevel = 0 // 直到最上层
		}
		users, err = resolveManagers(initiatorID, 0, true, endLevel, org)
	default:
		users, miss = fromAssigneeListDetailed(node.NodeAssigneeList, org)
		if len(users) == 0 {
			return nil, emptyDesignatedErr(node, miss)
		}
	}
	if err != nil {
		return nil, err
	}

	// 审批人与发起人同一人
	beforeSelf := uniqueUsers(users)
	users = applyApproveSelf(beforeSelf, initiatorID, node.ApproveSelf, org)
	users = uniqueUsers(users)
	if len(users) == 0 {
		// 配置了人但因「与发起人同一人自动跳过」清空 → 允许空（引擎自动通过）
		if len(beforeSelf) > 0 && node.ApproveSelf == 1 {
			return nil, nil
		}
		if len(beforeSelf) > 0 && (node.ApproveSelf == 2 || node.ApproveSelf == 3) {
			return nil, fmt.Errorf("节点[%s]审批人与发起人相同，转交上级/部门负责人失败（未找到可转交的主管）", node.NodeName)
		}
		return nil, fmt.Errorf("节点[%s]未解析到审批人", node.NodeName)
	}
	return users, nil
}

// UseClaimPool 角色节点且关闭「全员参与」→ 认领池
func UseClaimPool(node *Node) bool {
	if node == nil || node.GroupStrategy != 0 {
		return false
	}
	if node.SetType == 3 {
		return true
	}
	// 发起人自选角色
	return node.SetType == 4 && node.SelectMode == 3
}

// ResolveClaimRoles 认领池：不展开用户，按角色挂槽
func ResolveClaimRoles(node *Node, selected []UserRef) ([]UserRef, error) {
	if node == nil {
		return nil, fmt.Errorf("节点为空")
	}
	var roles []Assignee
	if node.SetType == 4 && node.SelectMode == 3 {
		if len(selected) > 0 {
			if err := validateAgainstCandidates(node, selected); err != nil {
				return nil, err
			}
			for _, u := range selected {
				roles = append(roles, Assignee{ID: u.ID, Name: u.Name})
			}
		} else {
			roles = node.NodeAssigneeList
		}
	} else {
		roles = node.NodeAssigneeList
	}
	var out []UserRef
	for _, a := range roles {
		id := AssigneeIDUint(a)
		if id == 0 {
			continue
		}
		name := a.Name
		if name == "" {
			name = fmt.Sprintf("角色#%d", id)
		}
		out = append(out, UserRef{ID: id, Name: name})
	}
	out = uniqueUsers(out)
	if len(out) == 0 {
		return nil, fmt.Errorf("节点[%s]未配置可认领角色", node.NodeName)
	}
	return out, nil
}

// ResolveCC 抄送人
func ResolveCC(node *Node, selected []UserRef, org OrgStore) []UserRef {
	users := fromAssigneeList(node.NodeAssigneeList, org)
	if node.AllowSelection && len(selected) > 0 {
		users = append(users, selected...)
	}
	return uniqueUsers(users)
}

func fromAssigneeList(list []Assignee, org OrgStore) []UserRef {
	users, _ := fromAssigneeListDetailed(list, org)
	return users
}

// fromAssigneeListDetailed 解析指定成员；missing 为配置了 id 但用户表查不到的项
func fromAssigneeListDetailed(list []Assignee, org OrgStore) (users []UserRef, missing []uint64) {
	for _, a := range list {
		id := AssigneeIDUint(a)
		if id == 0 {
			continue
		}
		name, _, ok := org.GetUser(id)
		if !ok {
			// 用户不存在/已删：不挂幽灵审批人，避免待办无人可办
			missing = append(missing, id)
			continue
		}
		users = append(users, UserRef{ID: id, Name: name})
	}
	return users, missing
}

func emptyDesignatedErr(node *Node, missing []uint64) error {
	name := "未命名"
	if node != nil && node.NodeName != "" {
		name = node.NodeName
	}
	if len(missing) > 0 {
		return fmt.Errorf("节点[%s]指定成员用户不存在或已删除（id=%v）", name, missing)
	}
	n := 0
	if node != nil {
		n = len(node.NodeAssigneeList)
	}
	if n == 0 {
		return fmt.Errorf("节点[%s]为指定成员，但未配置人员", name)
	}
	return fmt.Errorf("节点[%s]指定成员无法解析（人员 id 无效）", name)
}

func fromRoles(list []Assignee, org OrgStore) ([]UserRef, error) {
	var out []UserRef
	for _, a := range list {
		roleID := AssigneeIDUint(a)
		if roleID == 0 {
			continue
		}
		ids, err := org.ListUserIDsByRole(roleID)
		if err != nil {
			return nil, err
		}
		for _, uid := range ids {
			name, _, ok := org.GetUser(uid)
			if !ok {
				continue
			}
			out = append(out, UserRef{ID: uid, Name: name})
		}
	}
	return out, nil
}

// resolveManagers level: 第 N 级主管（从直接上级算 1）；chain=true 时收集多级
func resolveManagers(initiatorID uint64, level int, chain bool, endLevel int, org OrgStore) ([]UserRef, error) {
	_, deptID, ok := org.GetUser(initiatorID)
	if !ok || deptID == 0 {
		return nil, fmt.Errorf("发起人部门不存在，无法解析主管")
	}
	if level <= 0 {
		level = 1
	}
	var out []UserRef
	curDept := deptID
	step := 0
	for curDept > 0 {
		leaderID, ok := org.GetDeptLeader(curDept)
		parentID, _ := org.GetDeptParent(curDept)
		step++
		if ok && leaderID > 0 && leaderID != initiatorID {
			name, _, uok := org.GetUser(leaderID)
			if uok {
				if chain {
					out = append(out, UserRef{ID: leaderID, Name: name})
					if endLevel > 0 && step >= endLevel {
						break
					}
				} else if step == level {
					return []UserRef{{ID: leaderID, Name: name}}, nil
				}
			}
		}
		if !chain && step >= level {
			break
		}
		if parentID == 0 || parentID == curDept {
			break
		}
		curDept = parentID
		if !chain && step > 50 {
			break
		}
		if chain && step > 50 {
			break
		}
	}
	if !chain {
		return nil, fmt.Errorf("未找到第%d级主管", level)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("未找到连续主管")
	}
	return out, nil
}

func applyApproveSelf(users []UserRef, initiatorID uint64, mode int, org OrgStore) []UserRef {
	hasSelf := false
	for _, u := range users {
		if u.ID == initiatorID {
			hasSelf = true
			break
		}
	}
	if !hasSelf {
		return users
	}
	switch mode {
	case 1: // 自动跳过：去掉自己
		var out []UserRef
		for _, u := range users {
			if u.ID != initiatorID {
				out = append(out, u)
			}
		}
		return out
	case 2: // 转交直接上级（第 1 级主管）
		var out []UserRef
		for _, u := range users {
			if u.ID != initiatorID {
				out = append(out, u)
			}
		}
		mgrs, err := resolveManagers(initiatorID, 1, false, 0, org)
		if err == nil {
			out = append(out, mgrs...)
		}
		return out
	case 3: // 转交部门负责人（所属部门链路最上层负责人）
		var out []UserRef
		for _, u := range users {
			if u.ID != initiatorID {
				out = append(out, u)
			}
		}
		if head, err := resolveTopDeptLeader(initiatorID, org); err == nil {
			out = append(out, head)
		}
		return out
	default: // 0 由发起人自己批
		return users
	}
}

// directorEndLevel 连续主管自定义终点：优先 directorLevel，兼容误写入 examineLevel 的旧数据
func directorEndLevel(node *Node) int {
	if node.DirectorLevel > 0 {
		return node.DirectorLevel
	}
	if node.ExamineLevel > 0 {
		return node.ExamineLevel
	}
	return 1
}

// ValidateNodeCandidates 校验自选人员是否在候选范围内（对外暴露给发起校验）
func ValidateNodeCandidates(node *Node, selected []UserRef) error {
	return validateAgainstCandidates(node, selected)
}

// validateAgainstCandidates 发起人自选须落在设计器候选名单内（名单为空则不限制）
func validateAgainstCandidates(node *Node, selected []UserRef) error {
	if node == nil || node.NodeCandidate == nil || len(node.NodeCandidate.Assignees) == 0 {
		return nil
	}
	allow := map[uint64]struct{}{}
	for _, a := range node.NodeCandidate.Assignees {
		if id := AssigneeIDUint(a); id > 0 {
			allow[id] = struct{}{}
		}
	}
	if len(allow) == 0 {
		return nil
	}
	for _, u := range selected {
		if _, ok := allow[u.ID]; !ok {
			return fmt.Errorf("节点[%s]所选人不在候选范围内", node.NodeName)
		}
	}
	return nil
}

// resolveTopDeptLeader 取发起人部门链路上最上层非本人负责人
func resolveTopDeptLeader(initiatorID uint64, org OrgStore) (UserRef, error) {
	_, deptID, ok := org.GetUser(initiatorID)
	if !ok || deptID == 0 {
		return UserRef{}, fmt.Errorf("发起人部门不存在，无法解析部门负责人")
	}
	var last UserRef
	curDept := deptID
	for i := 0; i < 50 && curDept > 0; i++ {
		leaderID, lok := org.GetDeptLeader(curDept)
		parentID, _ := org.GetDeptParent(curDept)
		if lok && leaderID > 0 && leaderID != initiatorID {
			name, _, uok := org.GetUser(leaderID)
			if uok {
				last = UserRef{ID: leaderID, Name: name}
			}
		}
		if parentID == 0 || parentID == curDept {
			break
		}
		curDept = parentID
	}
	if last.ID == 0 {
		return UserRef{}, fmt.Errorf("未找到部门负责人")
	}
	return last, nil
}

func uniqueUsers(in []UserRef) []UserRef {
	seen := map[uint64]struct{}{}
	var out []UserRef
	for _, u := range in {
		if u.ID == 0 {
			continue
		}
		if _, ok := seen[u.ID]; ok {
			continue
		}
		seen[u.ID] = struct{}{}
		out = append(out, u)
	}
	return out
}

// BuildBelongChecker 发起人条件
func BuildBelongChecker(org OrgStore, initiatorID, initiatorDept uint64, roleIDs []uint64) func(kind string, ids []uint64) bool {
	roleSet := map[uint64]struct{}{}
	for _, r := range roleIDs {
		roleSet[r] = struct{}{}
	}
	return func(kind string, ids []uint64) bool {
		switch kind {
		case "user":
			for _, id := range ids {
				if id == initiatorID {
					return true
				}
			}
		case "role":
			for _, id := range ids {
				if _, ok := roleSet[id]; ok {
					return true
				}
			}
		case "dept":
			for _, id := range ids {
				if id == initiatorDept || org.UserInDept(initiatorID, id) {
					return true
				}
			}
		}
		return false
	}
}

package engine

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// EvalContext 条件求值上下文
type EvalContext struct {
	FormData      map[string]interface{}
	InitiatorID   uint64
	InitiatorDept uint64
	// 发起人角色 ID 列表
	InitiatorRoles []uint64
	// BelongChecker 判断发起人是否属于 dept/role/user 集合
	BelongChecker func(kind string, ids []uint64) bool
}

// MatchConditionGroups 组间 OR，组内 AND
func MatchConditionGroups(groups [][]ConditionItem, ctx *EvalContext) bool {
	if len(groups) == 0 {
		return true // 空条件视为默认分支可命中（由调用方决定是否当默认）
	}
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}
		ok := true
		for _, item := range group {
			if !matchItem(item, ctx) {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func matchItem(item ConditionItem, ctx *EvalContext) bool {
	typ := strings.ToLower(strings.TrimSpace(item.Type))
	if typ == "" || typ == "custom" {
		typ = "form"
	}
	switch typ {
	case "initiator":
		return matchInitiator(item, ctx)
	default:
		return matchForm(item, ctx)
	}
}

func matchInitiator(item ConditionItem, ctx *EvalContext) bool {
	ids := make([]uint64, 0, len(item.ValueList))
	kinds := map[string][]uint64{}
	for _, m := range item.ValueList {
		id := OrgMemberIDUint(m)
		if id == 0 {
			continue
		}
		kind := strings.ToLower(m.Kind)
		if kind == "" {
			kind = "user"
		}
		kinds[kind] = append(kinds[kind], id)
		ids = append(ids, id)
	}
	belong := false
	if ctx.BelongChecker != nil {
		for kind, list := range kinds {
			if ctx.BelongChecker(kind, list) {
				belong = true
				break
			}
		}
	} else {
		// 兜底：仅用户 ID
		for _, id := range kinds["user"] {
			if id == ctx.InitiatorID {
				belong = true
				break
			}
		}
	}
	op := item.Operator
	switch op {
	case "belong", "include":
		return belong
	case "notbelong", "notinclude":
		return !belong
	default:
		return belong
	}
}

func matchForm(item ConditionItem, ctx *EvalContext) bool {
	field := item.Field
	if field == "" {
		return false
	}
	actual := lookupForm(ctx.FormData, field)
	expected := item.Value
	op := item.Operator
	switch op {
	case "==", "=", "eq":
		return compareEqual(actual, expected)
	case "!=", "ne":
		return !compareEqual(actual, expected)
	case ">":
		return compareNumber(actual, expected) > 0
	case ">=":
		return compareNumber(actual, expected) >= 0
	case "<":
		return compareNumber(actual, expected) < 0
	case "<=":
		return compareNumber(actual, expected) <= 0
	case "include":
		return strings.Contains(fmt.Sprint(actual), fmt.Sprint(expected))
	case "notinclude":
		return !strings.Contains(fmt.Sprint(actual), fmt.Sprint(expected))
	default:
		return compareEqual(actual, expected)
	}
}

func lookupForm(data map[string]interface{}, field string) interface{} {
	if data == nil {
		return nil
	}
	if v, ok := data[field]; ok {
		return v
	}
	// 兼容嵌套
	parts := strings.Split(field, ".")
	var cur interface{} = data
	for _, p := range parts {
		m, ok := cur.(map[string]interface{})
		if !ok {
			return nil
		}
		cur, ok = m[p]
		if !ok {
			return nil
		}
	}
	return cur
}

func compareEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// 数字宽松比较
	af, aOk := toFloat(a)
	bf, bOk := toFloat(b)
	if aOk && bOk {
		return af == bf
	}
	return fmt.Sprint(a) == fmt.Sprint(b)
}

func compareNumber(a, b interface{}) int {
	af, _ := toFloat(a)
	bf, _ := toFloat(b)
	if af > bf {
		return 1
	}
	if af < bf {
		return -1
	}
	return 0
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint64:
		return float64(n), true
	case jsonNumber:
		f, err := n.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int64:
			return float64(rv.Int()), true
		case reflect.Float64:
			return rv.Float(), true
		}
		return 0, false
	}
}

// 避免直接依赖 encoding/json.Number 在接口断言中的麻烦
type jsonNumber interface {
	Float64() (float64, error)
}

// PickExclusiveBranch 按优先级选第一条命中；若都无则取「默认条件」（名称含默认或无条件列表）
func PickExclusiveBranch(branches []*Node, ctx *EvalContext) *Node {
	if len(branches) == 0 {
		return nil
	}
	sorted := sortByPriority(branches)
	var fallback *Node
	for _, b := range sorted {
		if isDefaultBranch(b) {
			fallback = b
			continue
		}
		if MatchConditionGroups(b.ConditionList, ctx) {
			return b
		}
	}
	if fallback != nil {
		return fallback
	}
	// 无显式默认则取最后一条
	return sorted[len(sorted)-1]
}

// PickInclusiveBranches 所有命中的分支；若无命中则走默认
func PickInclusiveBranches(branches []*Node, ctx *EvalContext) []*Node {
	sorted := sortByPriority(branches)
	var hit []*Node
	var fallback *Node
	for _, b := range sorted {
		if isDefaultBranch(b) {
			fallback = b
			continue
		}
		if MatchConditionGroups(b.ConditionList, ctx) {
			hit = append(hit, b)
		}
	}
	if len(hit) > 0 {
		return hit
	}
	if fallback != nil {
		return []*Node{fallback}
	}
	if len(sorted) > 0 {
		return []*Node{sorted[len(sorted)-1]}
	}
	return nil
}

// PickRouteTarget 路由：命中条件的目标 nodeKey
func PickRouteTarget(routes []*RouteNode, ctx *EvalContext) string {
	if len(routes) == 0 {
		return ""
	}
	// 按 priority
	cp := make([]*RouteNode, len(routes))
	copy(cp, routes)
	for i := 0; i < len(cp); i++ {
		for j := i + 1; j < len(cp); j++ {
			if cp[j].PriorityLevel < cp[i].PriorityLevel {
				cp[i], cp[j] = cp[j], cp[i]
			}
		}
	}
	for _, r := range cp {
		if len(r.ConditionList) == 0 || MatchConditionGroups(r.ConditionList, ctx) {
			return r.NodeKey
		}
	}
	return ""
}

func isDefaultBranch(b *Node) bool {
	if b == nil {
		return false
	}
	if strings.Contains(b.NodeName, "默认") {
		return true
	}
	if strings.HasSuffix(b.NodeKey, "default") {
		return true
	}
	return len(b.ConditionList) == 0
}

func sortByPriority(branches []*Node) []*Node {
	out := make([]*Node, 0, len(branches))
	for _, b := range branches {
		if b != nil {
			out = append(out, b)
		}
	}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].PriorityLevel < out[i].PriorityLevel {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

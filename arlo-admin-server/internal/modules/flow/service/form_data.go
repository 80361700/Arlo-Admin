package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"arlo-admin/internal/modules/flow/engine"
	"arlo-admin/internal/modules/flow/model"
)

// applyFormConfigRead 读详情时脱敏：opera=2 隐藏字段从返回的 formData 中移除
func applyFormConfigRead(formData map[string]interface{}, formConfig []map[string]interface{}) map[string]interface{} {
	out := cloneFormMap(formData)
	if out == nil || len(formConfig) == 0 {
		return out
	}
	for _, c := range formConfig {
		if formConfigOpera(c) != 2 {
			continue
		}
		id := formConfigFieldID(c)
		if id != "" {
			delete(out, id)
		}
	}
	return out
}

// parseActionURLFormID 解析节点 actionUrl："id:name" 或纯数字 id
func parseActionURLFormID(actionURL string) uint64 {
	actionURL = strings.TrimSpace(actionURL)
	if actionURL == "" {
		return 0
	}
	idStr := actionURL
	if i := strings.Index(actionURL, ":"); i > 0 {
		idStr = actionURL[:i]
	}
	var id uint64
	fmt.Sscanf(idStr, "%d", &id)
	return id
}

// mergeEpicFormSchemas 将子表单字段追加到主表单 schema（同 field 不重复）
func mergeEpicFormSchemas(main, sub map[string]interface{}) map[string]interface{} {
	if !formMapHasWidgets(sub) {
		return main
	}
	if !formMapHasWidgets(main) {
		return sub
	}
	out := cloneFormMapDeep(main)
	mainRoot := epicSchemaRoot(out)
	subRoot := epicSchemaRoot(sub)
	if mainRoot == nil || subRoot == nil {
		return main
	}
	mainChildren, _ := mainRoot["children"].([]interface{})
	subChildren, _ := subRoot["children"].([]interface{})
	exist := map[string]struct{}{}
	for _, c := range mainChildren {
		m, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if f, _ := m["field"].(string); f != "" {
			exist[f] = struct{}{}
		}
	}
	for _, c := range subChildren {
		m, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		f, _ := m["field"].(string)
		if f != "" {
			if _, ok := exist[f]; ok {
				continue
			}
			exist[f] = struct{}{}
		}
		mainChildren = append(mainChildren, cloneFormMapDeep(m))
	}
	mainRoot["children"] = mainChildren
	return out
}

func epicSchemaRoot(schema map[string]interface{}) map[string]interface{} {
	if schema == nil {
		return nil
	}
	raw, ok := schema["schemas"].([]interface{})
	if !ok || len(raw) == 0 {
		return nil
	}
	root, _ := raw[0].(map[string]interface{})
	return root
}

func cloneFormMapDeep(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return cloneFormMap(m)
	}
	var out map[string]interface{}
	if json.Unmarshal(b, &out) != nil {
		return cloneFormMap(m)
	}
	return out
}

// appendMissingFormConfig 为 schema 中未出现在 formConfig 的字段补默认可编辑权限
func appendMissingFormConfig(formConfig []map[string]interface{}, schema map[string]interface{}, defaultOpera int) []map[string]interface{} {
	exist := map[string]struct{}{}
	for _, c := range formConfig {
		if id := formConfigFieldID(c); id != "" {
			exist[id] = struct{}{}
		}
	}
	root := epicSchemaRoot(schema)
	if root == nil {
		return formConfig
	}
	children, _ := root["children"].([]interface{})
	out := append([]map[string]interface{}{}, formConfig...)
	for _, c := range children {
		m, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		field, _ := m["field"].(string)
		if field == "" {
			continue
		}
		if _, ok := exist[field]; ok {
			continue
		}
		label, _ := m["label"].(string)
		out = append(out, map[string]interface{}{
			"id":    field,
			"label": label,
			"opera": defaultOpera,
		})
		exist[field] = struct{}{}
	}
	return out
}

// applyFormConfigWrite 按节点 formConfig 合并提交：opera=1 可写，其余忽略提交值
func applyFormConfigWrite(existing, submitted map[string]interface{}, formConfig []map[string]interface{}) map[string]interface{} {
	if submitted == nil {
		return cloneFormMap(existing)
	}
	out := cloneFormMap(existing)
	if out == nil {
		out = map[string]interface{}{}
	}
	if len(formConfig) == 0 {
		for k, v := range submitted {
			out[k] = v
		}
		return out
	}
	writable := map[string]struct{}{}
	for _, c := range formConfig {
		id := formConfigFieldID(c)
		if id == "" {
			continue
		}
		if formConfigOpera(c) == 1 {
			writable[id] = struct{}{}
		}
	}
	for k, v := range submitted {
		if _, ok := writable[k]; ok {
			out[k] = v
		}
	}
	return out
}

func formConfigFieldID(c map[string]interface{}) string {
	if c == nil {
		return ""
	}
	if v, ok := c["id"]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if v, ok := c["field"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func formConfigOpera(c map[string]interface{}) int {
	if c == nil {
		return 1
	}
	switch v := c["opera"].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case string:
		if v == "0" {
			return 0
		}
		if v == "2" {
			return 2
		}
		return 1
	default:
		return 1
	}
}

func cloneFormMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// mergeChildFormIntoParent 子流程结束时把子实例表单合并回父实例：
// 仅合并子表单 schema 中存在的字段，避免子实例整包覆盖父单无关键。
func mergeChildFormIntoParent(parent, child *model.FlowInstance) bool {
	if parent == nil || child == nil {
		return false
	}
	childFD := unmarshalMap(child.FormData)
	if len(childFD) == 0 {
		return false
	}
	fields := collectEpicFieldIDs(unmarshalMap(child.ProcessForm))
	if len(fields) == 0 {
		// 子表无字段定义时不回写，防止误覆盖
		return false
	}
	parentFD := unmarshalMap(parent.FormData)
	if parentFD == nil {
		parentFD = map[string]interface{}{}
	}
	changed := false
	for k, v := range childFD {
		if _, ok := fields[k]; !ok {
			continue
		}
		parentFD[k] = v
		changed = true
	}
	if !changed {
		return false
	}
	b, err := json.Marshal(parentFD)
	if err != nil {
		return false
	}
	parent.FormData = string(b)
	return true
}

// collectEpicFieldIDs 收集 epic schema 中带 field 的控件 id
func collectEpicFieldIDs(schema map[string]interface{}) map[string]struct{} {
	ids := map[string]struct{}{}
	walkEpicWidgets(schema, func(w map[string]interface{}) {
		if f, _ := w["field"].(string); f != "" {
			ids[f] = struct{}{}
		}
	})
	return ids
}

// filterFormDataBySchema 只保留 schema 中存在的表单字段
func filterFormDataBySchema(data, schema map[string]interface{}) map[string]interface{} {
	fields := collectEpicFieldIDs(schema)
	if len(fields) == 0 || len(data) == 0 {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(fields))
	for k, v := range data {
		if _, ok := fields[k]; ok {
			out[k] = v
		}
	}
	return out
}

func nodeFormConfig(modelJSON, nodeKey string) []map[string]interface{} {
	mc, err := engine.ParseModel(modelJSON)
	if err != nil || mc == nil || mc.NodeConfig == nil {
		return nil
	}
	if nodeKey == "" {
		return formConfigOf(mc.NodeConfig)
	}
	n := engine.FindNode(mc.NodeConfig, nodeKey)
	if n == nil {
		return nil
	}
	return formConfigOf(n)
}

func startFormConfig(modelJSON string) []map[string]interface{} {
	mc, err := engine.ParseModel(modelJSON)
	if err != nil || mc == nil {
		return nil
	}
	return formConfigOf(mc.NodeConfig)
}

// sanitizeInstanceForm 合并实例已有数据与提交数据（按任务节点 formConfig + 子表单缺省权限）
// allowWriteWithoutConfig：发起人改单等场景，无配置/全只读时允许整表提交
// 审批节点未配置可编辑字段时，忽略提交中的表单改动（防止子流程继承父表单后被篡改）
func (s *FlowService) sanitizeInstanceForm(ctx context.Context, inst *model.FlowInstance, nodeKey string, submitted map[string]interface{}, allowWriteWithoutConfig bool) map[string]interface{} {
	if submitted == nil || inst == nil {
		return submitted
	}
	existing := unmarshalMap(inst.FormData)
	cfg := nodeFormConfig(inst.ModelContent, nodeKey)
	var node *engine.Node
	if mc, err := engine.ParseModel(inst.ModelContent); err == nil && mc != nil {
		if nodeKey != "" {
			node = engine.FindNode(mc.NodeConfig, nodeKey)
		}
		if node == nil {
			node = mc.NodeConfig
		}
	}
	schema := s.resolveNodeFormSchema(ctx, inst, node)
	// 子表单新字段若设计器未写入 formConfig：有 actionUrl 时默认可写，否则只读
	defaultOpera := 0
	if node != nil && strings.TrimSpace(node.ActionURL) != "" {
		defaultOpera = 1
	}
	cfg = appendMissingFormConfig(cfg, schema, defaultOpera)
	if len(cfg) == 0 || !formConfigHasWritable(cfg) {
		if allowWriteWithoutConfig {
			return applyFormConfigWrite(existing, submitted, nil)
		}
		return cloneFormMap(existing)
	}
	return applyFormConfigWrite(existing, submitted, cfg)
}

func formConfigHasWritable(formConfig []map[string]interface{}) bool {
	for _, c := range formConfig {
		if formConfigOpera(c) == 1 {
			return true
		}
	}
	return false
}

// walkEpicWidgets 深度遍历 epic schema 中带 field 的控件
func walkEpicWidgets(schema map[string]interface{}, fn func(widget map[string]interface{})) {
	if schema == nil || fn == nil {
		return
	}
	var walk func(node interface{})
	walk = func(node interface{}) {
		switch n := node.(type) {
		case map[string]interface{}:
			if field, _ := n["field"].(string); field != "" {
				fn(n)
			}
			if children, ok := n["children"].([]interface{}); ok {
				for _, c := range children {
					walk(c)
				}
			}
			// 部分布局把子节点放在 columns / list
			if cols, ok := n["columns"].([]interface{}); ok {
				for _, c := range cols {
					walk(c)
				}
			}
			if list, ok := n["list"].([]interface{}); ok {
				for _, c := range list {
					walk(c)
				}
			}
		case []interface{}:
			for _, c := range n {
				walk(c)
			}
		}
	}
	if root := epicSchemaRoot(schema); root != nil {
		walk(root)
		return
	}
	walk(schema)
}

func epicWidgetRequired(w map[string]interface{}) (required bool, message string) {
	if w == nil {
		return false, ""
	}
	if v, ok := w["required"]; ok {
		switch t := v.(type) {
		case bool:
			if t {
				return true, ""
			}
		case string:
			if t == "true" || t == "1" {
				return true, ""
			}
		}
	}
	rules, ok := w["rules"].([]interface{})
	if !ok {
		return false, ""
	}
	for _, r := range rules {
		m, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		switch t := m["required"].(type) {
		case bool:
			if !t {
				continue
			}
		case string:
			if t != "true" && t != "1" {
				continue
			}
		default:
			continue
		}
		msg, _ := m["message"].(string)
		return true, strings.TrimSpace(msg)
	}
	return false, ""
}

func isFormValueEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t) == ""
	case []interface{}:
		return len(t) == 0
	case map[string]interface{}:
		return len(t) == 0
	case float64, float32, int, int64, int32, uint, uint64, bool:
		return false
	default:
		// json 数字偶发为 json.Number
		if s, ok := v.(fmt.Stringer); ok {
			return strings.TrimSpace(s.String()) == ""
		}
		b, err := json.Marshal(v)
		if err != nil {
			return true
		}
		s := strings.TrimSpace(string(b))
		return s == "" || s == "null" || s == "[]" || s == "{}"
	}
}

// validateRequiredFields 校验表单必填。
// launchOrRevise=true：校验所有非隐藏(opera≠2)必填；false：仅校验本节点可写(opera=1)必填。
func validateRequiredFields(schema map[string]interface{}, formData map[string]interface{}, formConfig []map[string]interface{}, launchOrRevise bool) error {
	if !formMapHasWidgets(schema) {
		return nil
	}
	operaOf := map[string]int{}
	for _, c := range formConfig {
		id := formConfigFieldID(c)
		if id == "" {
			continue
		}
		operaOf[id] = formConfigOpera(c)
	}
	var firstErr error
	walkEpicWidgets(schema, func(w map[string]interface{}) {
		if firstErr != nil {
			return
		}
		field, _ := w["field"].(string)
		if field == "" {
			return
		}
		req, msg := epicWidgetRequired(w)
		if !req {
			return
		}
		opera, has := operaOf[field]
		if !has {
			if !launchOrRevise {
				return
			}
			opera = 1
		}
		if opera == 2 {
			return
		}
		if !launchOrRevise && opera != 1 {
			return
		}
		val := interface{}(nil)
		if formData != nil {
			val = formData[field]
		}
		if !isFormValueEmpty(val) {
			return
		}
		if msg == "" {
			label, _ := w["label"].(string)
			if label != "" {
				msg = fmt.Sprintf("「%s」为必填项", label)
			} else {
				msg = fmt.Sprintf("字段 %s 为必填项", field)
			}
		}
		firstErr = fmt.Errorf("%s", msg)
	})
	return firstErr
}

// validateInstanceNodeForm 按节点 schema + formConfig 校验必填（审批仅可写字段；发起/改单为可见必填）
func (s *FlowService) validateInstanceNodeForm(ctx context.Context, inst *model.FlowInstance, nodeKey string, formData map[string]interface{}, launchOrRevise bool) error {
	if inst == nil {
		return nil
	}
	var node *engine.Node
	if mc, err := engine.ParseModel(inst.ModelContent); err == nil && mc != nil {
		if nodeKey != "" {
			node = engine.FindNode(mc.NodeConfig, nodeKey)
		}
		if node == nil {
			node = mc.NodeConfig
		}
	}
	schema := s.resolveNodeFormSchema(ctx, inst, node)
	cfg := nodeFormConfig(inst.ModelContent, nodeKey)
	defaultOpera := 0
	if launchOrRevise {
		defaultOpera = 1
	} else if node != nil && strings.TrimSpace(node.ActionURL) != "" {
		defaultOpera = 1
	}
	cfg = appendMissingFormConfig(cfg, schema, defaultOpera)
	return validateRequiredFields(schema, formData, cfg, launchOrRevise)
}

// nodeHasUnfilledWritableRequired 当前节点存在可写必填且值为空（批量同意前置提示用）
func (s *FlowService) nodeHasUnfilledWritableRequired(ctx context.Context, inst *model.FlowInstance, nodeKey string) bool {
	if inst == nil {
		return false
	}
	err := s.validateInstanceNodeForm(ctx, inst, nodeKey, unmarshalMap(inst.FormData), false)
	return err != nil
}

// formMapHasWidgets 判断 epic PageSchema 是否含可展示字段
func formMapHasWidgets(m map[string]interface{}) bool {
	if m == nil || len(m) == 0 {
		return false
	}
	raw, ok := m["schemas"]
	if !ok {
		return false
	}
	schemas, ok := raw.([]interface{})
	if !ok || len(schemas) == 0 {
		return false
	}
	root, ok := schemas[0].(map[string]interface{})
	if !ok {
		return false
	}
	children, _ := root["children"].([]interface{})
	return len(children) > 0
}

// resolveProcessFormJSON 解析流程定义表单：业务流程优先绑定表单库（系统表单无 schema 时返回空对象）
func (s *FlowService) resolveProcessFormJSON(ctx context.Context, p *model.FlowProcess) string {
	if p == nil {
		return ""
	}
	if p.ProcessType == "business" {
		if fid, ok := parseBindFormID(p.ProcessSetting); ok && fid > 0 {
			if f, e := s.repo.GetForm(ctx, fid); e == nil && f != nil {
				if f.FormType == 2 {
					return `{}`
				}
				if strings.TrimSpace(f.FormSchema) != "" {
					return f.FormSchema
				}
			}
		}
	}
	return p.ProcessForm
}

// resolveBoundFormRender 解析业务流程绑定表单的渲染方式
func (s *FlowService) resolveBoundFormRender(ctx context.Context, p *model.FlowProcess) (renderType, component string) {
	renderType = "designer"
	if p == nil || p.ProcessType != "business" {
		return renderType, ""
	}
	fid, ok := parseBindFormID(p.ProcessSetting)
	if !ok || fid == 0 {
		return renderType, ""
	}
	f, err := s.repo.GetForm(ctx, fid)
	if err != nil || f == nil {
		return renderType, ""
	}
	if f.FormType == 2 && strings.TrimSpace(f.PcURL) != "" {
		return "vue", normalizeFormPcURL(f.PcURL)
	}
	return renderType, ""
}

func normalizeFormPcURL(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimPrefix(s, "src/views/")
	s = strings.TrimPrefix(s, "views/")
	s = strings.TrimSuffix(s, ".vue")
	return s
}

// resolveInstanceFormSchema 实例表单 schema：快照 → 流程定义 → 父实例（子流程常共用父表单）
func (s *FlowService) resolveInstanceFormSchema(ctx context.Context, inst *model.FlowInstance) map[string]interface{} {
	if inst == nil {
		return map[string]interface{}{}
	}
	schema := unmarshalMap(inst.ProcessForm)
	if formMapHasWidgets(schema) {
		return schema
	}
	if p, err := s.repo.GetProcess(ctx, inst.ProcessID); err == nil && p != nil {
		schema = unmarshalMap(s.resolveProcessFormJSON(ctx, p))
		if formMapHasWidgets(schema) {
			return schema
		}
	}
	// 子流程实例：定义侧可能无独立表单，回退父流程表单
	seen := map[uint64]struct{}{inst.ID: {}}
	parentID := inst.ParentInstanceID
	for parentID > 0 {
		if _, ok := seen[parentID]; ok {
			break
		}
		seen[parentID] = struct{}{}
		parent, err := s.repo.GetInstance(ctx, parentID)
		if err != nil || parent == nil {
			break
		}
		schema = unmarshalMap(parent.ProcessForm)
		if formMapHasWidgets(schema) {
			return schema
		}
		if p, e := s.repo.GetProcess(ctx, parent.ProcessID); e == nil && p != nil {
			schema = unmarshalMap(s.resolveProcessFormJSON(ctx, p))
			if formMapHasWidgets(schema) {
				return schema
			}
		}
		parentID = parent.ParentInstanceID
	}
	return schema
}

// resolveNodeFormSchema 主表单 + 指定节点 actionUrl 子表单合并
func (s *FlowService) resolveNodeFormSchema(ctx context.Context, inst *model.FlowInstance, node *engine.Node) map[string]interface{} {
	base := s.resolveInstanceFormSchema(ctx, inst)
	if node == nil {
		return base
	}
	return s.mergeActionURLForm(ctx, base, node.ActionURL)
}

// resolveDisplayFormSchema 详情展示：主表单 + 指定节点的子表单（已办补录 / 当前可编辑时的办理子表）。
func (s *FlowService) resolveDisplayFormSchema(ctx context.Context, inst *model.FlowInstance, root *engine.Node, includeNodeKeys map[string]struct{}) map[string]interface{} {
	base := s.resolveInstanceFormSchema(ctx, inst)
	if root == nil {
		return base
	}
	seen := map[uint64]struct{}{}
	var walk func(n *engine.Node)
	walk = func(n *engine.Node) {
		if n == nil {
			return
		}
		_, allow := includeNodeKeys[n.NodeKey]
		if includeNodeKeys == nil || allow {
			if fid := parseActionURLFormID(n.ActionURL); fid > 0 {
				if _, ok := seen[fid]; !ok {
					seen[fid] = struct{}{}
					base = s.mergeActionURLForm(ctx, base, n.ActionURL)
				}
			}
		}
		walk(n.ChildNode)
		for _, c := range n.ConditionNodes {
			walk(c)
			walk(c.ChildNode)
		}
		for _, c := range n.ParallelNodes {
			walk(c)
			walk(c.ChildNode)
		}
		for _, c := range n.InclusiveNodes {
			walk(c)
			walk(c.ChildNode)
		}
	}
	walk(root)
	return base
}

// collectDisplayFormNodeKeys 发起 + 历史已办节点；可办理时再拼当前节点。
// 只读查看时：若当前活动节点已有人办过，或实例 formData 已含该节点子表字段，仍拼入（依次审批未完结时无 his_task，否则子表数据有值无 schema）。
func (s *FlowService) collectDisplayFormNodeKeys(ctx context.Context, workInst *model.FlowInstance, root *engine.Node, curTaskNodeKey string, includeCurrent bool) map[string]struct{} {
	keys := map[string]struct{}{}
	if root != nil && root.NodeKey != "" {
		keys[root.NodeKey] = struct{}{}
	}
	if workInst != nil {
		if his, err := s.repo.ListHisTasks(ctx, workInst.ID); err == nil {
			for _, h := range his {
				if k := strings.TrimSpace(h.NodeKey); k != "" {
					keys[k] = struct{}{}
				}
			}
		}
		if includeCurrent {
			if k := strings.TrimSpace(workInst.CurrentNodeKey); k != "" {
				keys[k] = struct{}{}
			}
			if actives, err := s.repo.ListActiveTasksByInstance(ctx, workInst.ID); err == nil {
				for _, t := range actives {
					if k := strings.TrimSpace(t.NodeKey); k != "" {
						keys[k] = struct{}{}
					}
				}
			}
			if k := strings.TrimSpace(curTaskNodeKey); k != "" {
				keys[k] = struct{}{}
			}
		} else {
			// 只读：补拼「已有办理痕迹 / 已有子表数据」的活动节点
			formData := unmarshalMap(workInst.FormData)
			addActive := func(nodeKey string) {
				k := strings.TrimSpace(nodeKey)
				if k == "" {
					return
				}
				if _, ok := keys[k]; ok {
					return
				}
				var node *engine.Node
				if root != nil {
					node = engine.FindNode(root, k)
				}
				if node == nil || strings.TrimSpace(node.ActionURL) == "" {
					return
				}
				sub := s.mergeActionURLForm(ctx, map[string]interface{}{}, node.ActionURL)
				if formDataIntersectsSchema(formData, sub) {
					keys[k] = struct{}{}
				}
			}
			if actives, err := s.repo.ListActiveTasksByInstance(ctx, workInst.ID); err == nil {
				for _, t := range actives {
					acted := false
					if actors, e := s.repo.ListTaskActors(ctx, t.ID); e == nil {
						for _, a := range actors {
							if a.ActorState != model.ActorPending {
								acted = true
								break
							}
						}
					}
					if acted {
						if k := strings.TrimSpace(t.NodeKey); k != "" {
							keys[k] = struct{}{}
						}
					} else {
						addActive(t.NodeKey)
					}
				}
			}
			addActive(workInst.CurrentNodeKey)
			addActive(curTaskNodeKey)
		}
	} else if includeCurrent {
		if k := strings.TrimSpace(curTaskNodeKey); k != "" {
			keys[k] = struct{}{}
		}
	}
	return keys
}

// formDataIntersectsSchema 实例数据是否已含该 schema 中的任一字段（用于判断子表是否已填过）
func formDataIntersectsSchema(formData, schema map[string]interface{}) bool {
	if len(formData) == 0 || !formMapHasWidgets(schema) {
		return false
	}
	hit := false
	walkEpicWidgets(schema, func(w map[string]interface{}) {
		if hit {
			return
		}
		field, _ := w["field"].(string)
		if field == "" {
			return
		}
		if _, ok := formData[field]; ok {
			hit = true
		}
	})
	return hit
}

func (s *FlowService) mergeActionURLForm(ctx context.Context, base map[string]interface{}, actionURL string) map[string]interface{} {
	fid := parseActionURLFormID(actionURL)
	if fid == 0 {
		return base
	}
	f, err := s.repo.GetForm(ctx, fid)
	// 子表单只合并设计表单 schema；系统表单无字段可并
	if err != nil || f == nil || f.FormType == 2 || strings.TrimSpace(f.FormSchema) == "" {
		return base
	}
	return mergeEpicFormSchemas(base, unmarshalMap(f.FormSchema))
}

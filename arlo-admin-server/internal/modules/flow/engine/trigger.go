package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"arlo-admin/internal/modules/flow/model"
)

const triggerHTTPTimeout = 15 * time.Second

// invokeHTTPTrigger 同步 HTTP 触发：extendConfig.trigger 为 http(s) URL，或 args 内含 url。
// 未配置地址时跳过（兼容旧占位节点）。
func invokeHTTPTrigger(ctx context.Context, inst *model.FlowInstance, node *Node, formData map[string]interface{}) (string, error) {
	if node == nil {
		return "跳过", nil
	}
	urlStr, method, args := resolveTriggerHTTP(node)
	if urlStr == "" {
		return "未配置触发地址，已跳过", nil
	}
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return "", fmt.Errorf("触发地址须为 http/https：%s", urlStr)
	}
	if method == "" {
		method = http.MethodPost
	}
	body := map[string]interface{}{
		"instanceId":  inst.ID,
		"processId":   inst.ProcessID,
		"processKey":  inst.ProcessKey,
		"processName": inst.ProcessName,
		"nodeKey":     node.NodeKey,
		"nodeName":    node.NodeName,
		"formData":    formData,
		"args":        args,
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, method, urlStr, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: triggerHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("触发器请求失败: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("触发器返回 HTTP %d", resp.StatusCode)
	}
	return fmt.Sprintf("HTTP %s %s → %d", method, urlStr, resp.StatusCode), nil
}

func resolveTriggerHTTP(node *Node) (urlStr, method string, args interface{}) {
	if node == nil || node.ExtendConfig == nil {
		return "", "", nil
	}
	ext := node.ExtendConfig
	if v, ok := ext["trigger"].(string); ok {
		v = strings.TrimSpace(v)
		if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
			urlStr = v
		}
	}
	if v, ok := ext["url"].(string); ok {
		v = strings.TrimSpace(v)
		if v != "" {
			urlStr = v
		}
	}
	if v, ok := ext["method"].(string); ok {
		method = strings.ToUpper(strings.TrimSpace(v))
	}
	if raw, ok := ext["args"]; ok {
		switch t := raw.(type) {
		case string:
			s := strings.TrimSpace(t)
			if s != "" {
				var parsed interface{}
				if json.Unmarshal([]byte(s), &parsed) == nil {
					args = parsed
					if urlStr == "" {
						if m, ok := parsed.(map[string]interface{}); ok {
							if u, ok := m["url"].(string); ok {
								urlStr = strings.TrimSpace(u)
							}
						}
					}
				} else {
					args = s
				}
			}
		default:
			args = raw
			if urlStr == "" {
				if m, ok := raw.(map[string]interface{}); ok {
					if u, ok := m["url"].(string); ok {
						urlStr = strings.TrimSpace(u)
					}
				}
			}
		}
	}
	return urlStr, method, args
}

func triggerPayloadKind(payload string) string {
	var meta struct {
		Kind string `json:"kind"`
	}
	_ = json.Unmarshal([]byte(payload), &meta)
	return meta.Kind
}

func nodeFromTriggerPayload(payload string) *Node {
	var raw map[string]interface{}
	if json.Unmarshal([]byte(payload), &raw) != nil {
		return nil
	}
	ext, _ := raw["extendConfig"].(map[string]interface{})
	n := &Node{
		NodeKey:      fmt.Sprint(raw["nodeKey"]),
		NodeName:     fmt.Sprint(raw["nodeName"]),
		Type:         NodeTrigger,
		ExtendConfig: ext,
	}
	if n.NodeKey == "" || n.NodeKey == "<nil>" {
		n.NodeKey = ""
	}
	return n
}

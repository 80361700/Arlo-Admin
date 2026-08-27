package job

import (
	"context"
	"sync"
)

// HandlerFunc 任务处理器；params 为 JSON 字符串；返回 result 摘要写入执行日志
type HandlerFunc func(ctx context.Context, params string) (result string, err error)

type handlerMeta struct {
	Name        string
	Description string
	Fn          HandlerFunc
}

var (
	registryMu sync.RWMutex
	registry   = map[string]handlerMeta{}
)

// Register 注册处理器（进程启动时调用）
func Register(code, name, description string, fn HandlerFunc) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[code] = handlerMeta{Name: name, Description: description, Fn: fn}
}

// GetHandler 获取处理器
func GetHandler(code string) (HandlerFunc, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	h, ok := registry[code]
	return h.Fn, ok
}

// HandlerInfo 对外展示
type HandlerInfo struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListHandlers 已注册处理器列表
func ListHandlers() []HandlerInfo {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]HandlerInfo, 0, len(registry))
	for code, h := range registry {
		out = append(out, HandlerInfo{Code: code, Name: h.Name, Description: h.Description})
	}
	return out
}

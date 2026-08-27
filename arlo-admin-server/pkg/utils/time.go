package utils

import "time"

// FormatTime 将 time.Time 格式化为 2006-01-02 15:04:05 字符串
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatPtrTime 将 *time.Time 格式化为 2006-01-02 15:04:05 字符串，nil 返回空字符串
func FormatPtrTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

package job

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CronSpec 标准 5 段：分 时 日 月 周（0=周日）
type CronSpec struct {
	minute  fieldSet
	hour    fieldSet
	day     fieldSet
	month   fieldSet
	weekday fieldSet
	raw     string
}

type fieldSet struct {
	any bool
	set map[int]struct{}
}

// ParseCron 解析 "分 时 日 月 周"，支持 *、n、n-m、*/n、逗号列表
func ParseCron(expr string) (*CronSpec, error) {
	fields := strings.Fields(strings.TrimSpace(expr))
	if len(fields) != 5 {
		return nil, fmt.Errorf("cron 需为 5 段（分 时 日 月 周），当前: %q", expr)
	}
	min, err := parseField(fields[0], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("分: %w", err)
	}
	hour, err := parseField(fields[1], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("时: %w", err)
	}
	day, err := parseField(fields[2], 1, 31)
	if err != nil {
		return nil, fmt.Errorf("日: %w", err)
	}
	month, err := parseField(fields[3], 1, 12)
	if err != nil {
		return nil, fmt.Errorf("月: %w", err)
	}
	weekday, err := parseField(fields[4], 0, 6)
	if err != nil {
		return nil, fmt.Errorf("周: %w", err)
	}
	return &CronSpec{minute: min, hour: hour, day: day, month: month, weekday: weekday, raw: strings.TrimSpace(expr)}, nil
}

func parseField(raw string, min, max int) (fieldSet, error) {
	raw = strings.TrimSpace(raw)
	if raw == "*" {
		return fieldSet{any: true}, nil
	}
	out := fieldSet{set: map[int]struct{}{}}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "*/") {
			step, err := strconv.Atoi(part[2:])
			if err != nil || step <= 0 {
				return fieldSet{}, fmt.Errorf("非法步进 %q", part)
			}
			for v := min; v <= max; v += step {
				out.set[v] = struct{}{}
			}
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			a, err1 := strconv.Atoi(bounds[0])
			b, err2 := strconv.Atoi(bounds[1])
			if err1 != nil || err2 != nil || a > b || a < min || b > max {
				return fieldSet{}, fmt.Errorf("非法区间 %q", part)
			}
			for v := a; v <= b; v++ {
				out.set[v] = struct{}{}
			}
			continue
		}
		v, err := strconv.Atoi(part)
		if err != nil || v < min || v > max {
			return fieldSet{}, fmt.Errorf("非法值 %q", part)
		}
		out.set[v] = struct{}{}
	}
	if !out.any && len(out.set) == 0 {
		return fieldSet{}, fmt.Errorf("空字段")
	}
	return out, nil
}

func (f fieldSet) match(v int) bool {
	if f.any {
		return true
	}
	_, ok := f.set[v]
	return ok
}

// Matches 判断 t（截断到分钟）是否命中
func (c *CronSpec) Matches(t time.Time) bool {
	t = t.Truncate(time.Minute)
	wd := int(t.Weekday()) // Sunday=0
	return c.minute.match(t.Minute()) &&
		c.hour.match(t.Hour()) &&
		c.day.match(t.Day()) &&
		c.month.match(int(t.Month())) &&
		c.weekday.match(wd)
}

// Next 返回 now 之后下一次触发时间（分钟精度，最多向后扫 366 天）
func (c *CronSpec) Next(now time.Time) time.Time {
	t := now.Truncate(time.Minute).Add(time.Minute)
	limit := t.Add(366 * 24 * time.Hour)
	for !t.After(limit) {
		if c.Matches(t) {
			return t
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}
}

func (c *CronSpec) String() string {
	if c == nil {
		return ""
	}
	return c.raw
}

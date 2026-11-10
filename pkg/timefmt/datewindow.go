package timefmt

import (
	"fmt"
	"time"
)

// DateWindowError 表示 since/until 不是合法的 YYYY-MM-DD 日期。
// 解析失败必须显式化：非法日期应让调用方拿到错误，而不是静默退化成「不过滤」。
type DateWindowError struct {
	Param string // 出错的参数名（since / until）
	Value string // 原始入参
	Err   error  // 底层 *time.ParseError
}

func (e *DateWindowError) Error() string {
	return fmt.Sprintf("参数 %s=%q 不是合法日期（期望 YYYY-MM-DD）", e.Param, e.Value)
}

func (e *DateWindowError) Unwrap() error { return e.Err }

// ParseDateWindow 解析统计口径的日期窗口（YYYY-MM-DD，本地时区）：
//
//	since 取当日零点；until 取「当日末尾」（+24h-1ns），窗口对当日提交闭区间包含。
//	空串表示该侧不限。任一侧非空但解析失败时返回 *DateWindowError。
//
// /api/v1/stats 的 commits / analyze / export/csv 三个端点共用本函数，
// 保证同一组 query 参数在三处得到同一个窗口。
func ParseDateWindow(since, until string) (sinceTime, untilTime time.Time, err error) {
	if since != "" {
		if sinceTime, err = time.ParseInLocation(LayoutDate, since, time.Local); err != nil {
			return time.Time{}, time.Time{}, &DateWindowError{Param: "since", Value: since, Err: err}
		}
	}
	if until != "" {
		if untilTime, err = time.ParseInLocation(LayoutDate, until, time.Local); err != nil {
			return time.Time{}, time.Time{}, &DateWindowError{Param: "until", Value: until, Err: err}
		}
		untilTime = untilTime.Add(24*time.Hour - time.Nanosecond)
	}
	return sinceTime, untilTime, nil
}

// InDateWindow 报告提交时间 t 是否落在窗口内；零值边界表示该侧不限。
// 与统计侧（calculateStatsFast）的过滤判据同源：早于 since 或晚于 until 即排除。
func InDateWindow(t, sinceTime, untilTime time.Time) bool {
	if !sinceTime.IsZero() && t.Before(sinceTime) {
		return false
	}
	if !untilTime.IsZero() && t.After(untilTime) {
		return false
	}
	return true
}

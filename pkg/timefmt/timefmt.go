// Package timefmt 集中管理项目内散落的时间布局常量，避免各处重复定义字面量。
package timefmt

const (
	// LayoutDate 日期部分：2006-01-02
	LayoutDate = "2006-01-02"

	// LayoutDateTime 日志与内部展示用的日期时间：2006-01-02 15:04:05
	LayoutDateTime = "2006-01-02 15:04:05"

	// LayoutAPITime 项目历史沿用的 ISO 风格布局。注意末尾 Z 是字面量而非时区占位：
	// Format 时本地时间也会被输出成 "Z" 结尾；如需真实 RFC3339 时区请用 time.RFC3339。
	LayoutAPITime = "2006-01-02T15:04:05Z"
)

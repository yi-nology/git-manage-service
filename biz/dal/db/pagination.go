package db

import "github.com/oekit/pkg/pagination"

// Paginate 返回溢出安全的 SQL offset/limit。
//
// 口径：page<=0 → 1；pageSize<=0 → 20；pageSize>500 → 钳到 500（OverMaxClamp）。
// offset 用 oekit pagination 的乘前判溢出实现——page/pageSize 来自外部 query
// 参数，手写 (page-1)*pageSize 在 page 为巨值时会 int 溢出回绕成正数，
// 返回垃圾偏移（oekit#146 同类缺陷，此处从源头杜绝）。
func Paginate(page, pageSize int) (offset, limit int) {
	page, pageSize = pagination.Policy{}.Normalize(page, pageSize)
	return pagination.Policy{}.Offset(page, pageSize), pagination.Policy{}.Limit(page, pageSize)
}

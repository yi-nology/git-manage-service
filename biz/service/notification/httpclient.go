package notification

import (
	"net/http"
	"time"
)

// notifyClient 是通知外发共享的 HTTP 客户端，带 10s 超时。
//
// 通知是尽力而为的旁路，目标端点不响应时不应无限等待：原先 4 处走
// http.Post（DefaultClient 无超时）+ 1 处 &http.Client{}，挂死会拖住
// 调用方（同步路径上就是 API 请求不返回）。
//
// 放行面不设限：通知目标由管理员显式配置，公网 webhook（钉钉/飞书/企微）
// 与内网自建端点都是合法形态；是否收窄为「仅公网」属产品决策（防管理员
// 配置被诱导指向内网），见 issue 记录。
var notifyClient = &http.Client{Timeout: 10 * time.Second}

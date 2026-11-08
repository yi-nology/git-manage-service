package middleware

import (
	"context"
	"crypto/subtle"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// BearerTokenAuth 静态 token 鉴权门（issue #14）：配置了 server.auth_token
// 时，除放行名单（健康检查/CORS 预检）外的所有请求必须携带
// `Authorization: Bearer <token>`。比较走 crypto/subtle 常量时间——
// 时序侧信道不因「静态 token 是临时方案」而豁免。
//
// token 为空时中间件不挂载（调用方负责回落只绑回环 + 启动告警，
// 见 cmd/server/main.go）。
func BearerTokenAuth(token string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if string(c.Method()) == "OPTIONS" {
			c.Next(ctx)
			return
		}
		path := string(c.Path())
		for _, p := range authExemptPaths {
			if path == p {
				c.Next(ctx)
				return
			}
		}
		h := string(c.GetHeader("Authorization"))
		got, ok := strings.CutPrefix(h, "Bearer ")
		if !ok || subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			c.SetStatusCode(401)
			c.SetBodyString(`{"code":401,"message":"未授权：缺少或错误的 Bearer token"}`)
			c.Abort()
			return
		}
		c.Next(ctx)
	}
}

// authExemptPaths 鉴权放行面：健康/就绪探针（K8s 探针不带凭据）。
var authExemptPaths = []string{
	"/healthz",
	"/readyz",
	"/ping",
}

// RegisterAuth 按配置决定是否挂鉴权门；返回值告知调用方绑定面的收口形态
// （true = 已有 token 保护，可绑外卡；false = 无 token，必须绑回环）。
func RegisterAuth(h *server.Hertz, token string) bool {
	if strings.TrimSpace(token) == "" {
		return false
	}
	h.Use(BearerTokenAuth(strings.TrimSpace(token)))
	return true
}

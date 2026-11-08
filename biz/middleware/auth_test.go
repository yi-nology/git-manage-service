package middleware

import (
	"context"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// runAuth 跑一遍中间件，返回（是否拒绝、是否放行）。
// hertz 单元测试无下游链：放行 = 未 Abort（拒绝路径必然 Abort）。
func runAuth(t *testing.T, handler app.HandlerFunc, headers map[string]string, path, method string) (aborted bool, passed bool) {
	t.Helper()
	c := app.NewContext(0)
	c.Request.SetMethod(method)
	c.Request.SetRequestURI(path)
	for k, v := range headers {
		c.Request.Header.Set(k, v)
	}
	handler(context.Background(), c)
	aborted = c.IsAborted()
	passed = !aborted
	return aborted, passed
}

// issue #14：静态 Bearer 鉴权门——正确 token 放行、错误/缺失/伪造拒绝、
// 放行面（探针/OPTIONS）不拦。
func TestBearerTokenAuth(t *testing.T) {
	const tok = "s3cret-token"
	h := BearerTokenAuth(tok)

	if _, ok := runAuth(t, h, map[string]string{"Authorization": "Bearer " + tok}, "/api/v1/repo/list", "GET"); !ok {
		t.Fatal("正确 token 必须放行")
	}
	if aborted, _ := runAuth(t, h, nil, "/api/v1/repo/list", "GET"); !aborted {
		t.Fatal("缺失 token 必须拒绝")
	}
	if aborted, _ := runAuth(t, h, map[string]string{"Authorization": "Bearer wrong"}, "/api/v1/repo/delete", "POST"); !aborted {
		t.Fatal("错误 token 必须拒绝")
	}
	if aborted, _ := runAuth(t, h, map[string]string{"Authorization": "Bearer " + tok + "-extra"}, "/api/v1/x", "GET"); !aborted {
		t.Fatal("token 后缀伪装必须拒绝（CutPrefix 全等比较）")
	}
	for _, p := range authExemptPaths {
		if _, ok := runAuth(t, h, nil, p, "GET"); !ok {
			t.Fatalf("探针 %s 必须放行", p)
		}
	}
	if _, ok := runAuth(t, h, nil, "/api/v1/repo/list", "OPTIONS"); !ok {
		t.Fatal("OPTIONS 预检必须放行")
	}
}

// RegisterAuth 空白 token 不挂门且报告回环形态（调用方据此收口绑定面）。
func TestRegisterAuthEmptyTokenReportsLoopback(t *testing.T) {
	h := server.Default()
	if RegisterAuth(h, "  ") {
		t.Fatal("空白 token 必须报告回环形态")
	}
	if !RegisterAuth(h, "tok") {
		t.Fatal("非空 token 必须报告已保护")
	}
}

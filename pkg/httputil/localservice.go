// Package httputil 提供访问用户自托管服务（Ollama 等）的 HTTP 工具。
//
// 这类服务的 base_url 由 API 调用方提供，服务端会主动发起请求；不校验的话
// 构成 SSRF：可被用来探测内网、打云 metadata（169.254.169.254）等。判定引擎
// 为 oekit urlguard：本服务的「只许本机/内网」策略经 WithCheckIPFunc/
// WithCheckHostFunc 注入，成为所有判定点（URL 层 IP 字面量、拨号层预检、
// 建连后端点复核）的单一判定源，且永不跳过——构造态豁免只作用于 urlguard
// 内置阻断层；内置的 exotic IPv4 字面量（八进制/十进制整型等宽松解析形态）、
// NAT64/Teredo/6to4 内嵌地址、IPv6 zone、约定域后缀防护全部保持激活。
package httputil

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/oekit/pkg/urlguard"
)

// allowedLocalCIDRs 是「本机/内网」策略的网段表达。不含链路本地——
// 169.254.0.0/16 含云 metadata、fe80::/10 同理，两者不在豁免清单里，
// urlguard 内置阻断直接拒绝（与 AllowedLocalIP 语义一致）。
var allowedLocalCIDRs = []string{
	"127.0.0.0/8",    // loopback
	"10.0.0.0/8",     // RFC1918
	"172.16.0.0/12",  // RFC1918
	"192.168.0.0/16", // RFC1918
	"::1/128",        // IPv6 loopback
	"fc00::/7",       // IPv6 ULA
}

// guard 把本服务的 private-only 策略注入 urlguard 的所有判定点。
// allowCIDRs 清单是编译期常量，文法错误只可能是编码失误，构造期直接暴露。
var guard = sync.OnceValue(func() *urlguard.Guard {
	g, err := urlguard.New(
		// 豁免 urlguard 内置的私网/loopback/ULA 阻断：本服务目标就是内网服务。
		urlguard.WithAllowCIDRs(allowedLocalCIDRs...),
		// localhost/local/internal/localdomain 是自托管服务的常见主机名，
		// 显式豁免内置的约定域后缀阻断；豁免后仍要过 host 策略（解析校验）。
		urlguard.WithAllowDomains("localhost", "local", "internal", "localdomain"),
		// 我们的策略成为单一判定源：每个 IP 判定点（URL 字面量/拨号预检/
		// 端点复核）都要过 AllowedLocalIP，公网地址在此被拒。
		urlguard.WithCheckIPFunc(func(ip net.IP) error {
			if !AllowedLocalIP(ip) {
				return fmt.Errorf("address %s is not a local/private address", ip)
			}
			return nil
		}),
		// 主机名判定点：解析出的所有 IP 都必须是本机/内网。
		urlguard.WithCheckHostFunc(func(host string) error {
			ips, err := net.LookupIP(host)
			if err != nil {
				return fmt.Errorf("resolve host %s: %w", host, err)
			}
			for _, rip := range ips {
				if !AllowedLocalIP(rip) {
					return fmt.Errorf("host %s resolves to non-local address %s", host, rip)
				}
			}
			return nil
		}),
	)
	if err != nil {
		panic(fmt.Sprintf("httputil: invalid urlguard guard: %v", err))
	}
	return g
})

// AllowedLocalIP 报告 ip 是否属于本机或私有网络（RFC1918 / IPv6 ULA / 环回）。
// link-local（含云 metadata）与公网地址一律拒绝。
func AllowedLocalIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || isIPv6ULA(ip)
}

func isIPv6ULA(ip net.IP) bool {
	if len(ip) != net.IPv6len {
		return false
	}
	return ip[0]&0xfe == 0xfc
}

// ValidateLocalServiceURL 校验 raw 是一个指向本机/内网服务的 http(s) URL：
// scheme 白名单、exotic IPv4 字面量、约定域后缀、IP 字面量与主机名策略
// （解析校验）全部由 guard.CheckURL 承载。
func ValidateLocalServiceURL(raw string) error {
	return guard().CheckURL(raw)
}

// NewLocalServiceClient 返回一个只允许连接本机/内网地址的 HTTP 客户端。
// guard.DialFunc 在拨号前预检（解析出的每个 IP 都过策略）、建连后按实际
// 端点复核；内层保留「以校验过的 IP 实际拨号」的 IP 钉扎，DNS rebinding
// 无效。重定向目标在拨号层同样受检。
func NewLocalServiceClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	dial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		ip := net.ParseIP(host)
		if ip == nil {
			ips, err := net.LookupIP(host)
			if err != nil {
				return nil, fmt.Errorf("resolve %s: %w", host, err)
			}
			for _, candidate := range ips {
				if AllowedLocalIP(candidate) {
					ip = candidate
					break
				}
			}
			if ip == nil {
				return nil, fmt.Errorf("host %s has no local/private address", host)
			}
		} else if !AllowedLocalIP(ip) {
			return nil, fmt.Errorf("address %s is not a local/private address", host)
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
	}

	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext:           guard().DialFunc(dial),
			ResponseHeaderTimeout: 30 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("stopped after 5 redirects")
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return fmt.Errorf("redirect to non-http scheme %q blocked", req.URL.Scheme)
			}
			return nil
		},
	}
}

// Package httputil 提供访问用户自托管服务（Ollama 等）的 HTTP 工具。
//
// 这类服务的 base_url 由 API 调用方提供，服务端会主动发起请求；不校验的话
// 构成 SSRF：可被用来探测内网、打云 metadata（169.254.169.254）等。此处的
// 客户端在 dial 时校验目标 IP，只允许本机与内网地址，同时杜绝 DNS rebinding
// 与重定向绕过。
package httputil

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

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

// ValidateLocalServiceURL 校验 raw 是一个指向本机/内网服务的 http(s) URL。
func ValidateLocalServiceURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("only http/https is allowed, got %q", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("url has no host")
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if !AllowedLocalIP(ip) {
			return fmt.Errorf("host %s is not a local/private address", host)
		}
		return nil
	}
	// 主机名：解析出的所有 IP 都必须合规，防止解析到公网地址。
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
}

// NewLocalServiceClient 返回一个只允许连接本机/内网地址的 HTTP 客户端。
// 校验发生在 dial 时（对每一次连接、包括重定向后的连接生效），并以校验过的
// IP 实际拨号，URL 主机名不会被二次解析，DNS rebinding 无效。
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
			DialContext:           dial,
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

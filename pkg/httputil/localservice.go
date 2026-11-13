// Package httputil 提供访问用户自托管服务（Ollama 等）的 HTTP 工具。
//
// 这类服务的 base_url 由 API 调用方提供，服务端会主动发起请求；不校验的话
// 构成 SSRF：可被用来探测内网、打云 metadata（169.254.169.254）等。
//
// 判定引擎为 oagentkit guardx/egress（tier0 零第三方依赖）：本服务的
// 「只许本机/内网」语义即 egress 的 SelfService 放行面（私网+回环放行，
// link-local/保留段/exotic IPv4 无条件拒绝），写面与拨号层从同一个
// Policy 派生，结构上不会分叉。
//
// 纵深三层：
//  1. 写面 CheckResolving——域名解析后逐 IP 过放行面，fail-closed，
//     消除「解析到内网的域名」（169.254.169.254.nip.io）与 DNS 重绑定的第一跳；
//  2. 拨号层守卫（Policy.NewClient）——拨号地址字面量判定，覆盖 HTTPS 的
//     DialTLS 路径，与写面同一 Policy；
//  3. 本包的 pinning dial——解析后只从已检结果里选本地地址拨号，闭合
//     「检查后换解析」的 TOCTOU。
package httputil

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/oagentkit/guardx/egress"
)

// AllowedLocalIP 报告 ip 是否属于本机或私有网络（RFC1918 / IPv6 ULA / 环回）。
// link-local（含云 metadata）与公网地址一律拒绝——与 egress 的 addrPolicy 在
// SelfService 放行面下的判定一致；本函数供 pinning dial 挑选拨号地址。
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
//
// 两层组合（方向互补）：
//  1. egress CheckResolving——域名解析后逐 IP 过放行面，fail-closed，负责
//     exotic IPv4、link-local、保留段、NAT64/Teredo 内嵌还原；
//  2. 白名单层——egress 是**黑名单**模型（防 SSRF 打内网，公网恒放行），
//     本服务要的是**白名单**语义：目标必须落在本机/内网，公网一律拒绝
//     （base_url 只能是自托管服务，连公网即越界）。
//
// 用 CheckResolving 而非 Check：字面量判定对「解析到内网的域名」
// （169.254.169.254.nip.io）零防护。
func ValidateLocalServiceURL(ctx context.Context, raw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	addrs, err := egress.SelfService.CheckResolving(ctx, raw, nil)
	if err != nil {
		return err
	}
	for _, a := range addrs {
		// egress 已在前面挡掉 link-local/保留段/非回环私网外的禁段；
		// 这里补白名单：公网地址一律拒（含 NAT64/Teredo 还原后落在公网的形态）。
		if !a.IsLoopback() && !a.IsPrivate() {
			return fmt.Errorf("urlguard: host resolves to non-local address %s", a)
		}
	}
	return nil
}

// NewLocalServiceClient 返回一个只允许连接本机/内网地址的 HTTP 客户端。
//
// egress.MustClient 提供拨号层字面量守卫（含 HTTPS DialTLS 路径），
// CheckRedirect 逐跳复检 + 继承 net/http 默认 10 跳上限；内层 pinning dial
// 在字面量放行后按解析结果挑本地地址拨号——两层同源于 SelfService 放行面。
func NewLocalServiceClient(timeout time.Duration) *http.Client {
	base := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext:           pinningDialContext,
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}
	c := egress.SelfService.MustClient(base)
	c.CheckRedirect = egress.SelfService.CheckRedirect
	return c
}

// pinningDialContext 解析 host 后只从本地/私网地址里挑一个拨号——域名形态
// 在 egress 拨号层（字面量判定）会被放行，本层负责把「解析到公网/元数据」
// 的结果挡掉，并按选定 IP 实际拨号，使权威 DNS 在检查与拨号之间换答案
// （rebinding）失效。
func pinningDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	if ip == nil {
		ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
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
	d := &net.Dialer{Timeout: 10 * time.Second}
	return d.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
}

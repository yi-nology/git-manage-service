package httputil

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestAllowedLocalIP(t *testing.T) {
	cases := []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", true},
		{"127.8.8.8", true},
		{"::1", true},
		{"10.0.0.5", true},
		{"172.16.1.1", true},
		{"192.168.1.10", true},
		{"fd00::1234", true},       // IPv6 ULA
		{"169.254.169.254", false}, // 云 metadata
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"2001:db8::1", false},
		{"0.0.0.0", false},
	}
	for _, c := range cases {
		if got := AllowedLocalIP(net.ParseIP(c.ip)); got != c.want {
			t.Errorf("AllowedLocalIP(%s) = %v, want %v", c.ip, got, c.want)
		}
	}
	if AllowedLocalIP(nil) {
		t.Error("AllowedLocalIP(nil) should be false")
	}
}

func TestValidateLocalServiceURL(t *testing.T) {
	// 不需要 DNS 的用例：IP 字面量。
	if err := ValidateLocalServiceURL("http://127.0.0.1:11434"); err != nil {
		t.Errorf("loopback url rejected: %v", err)
	}
	if err := ValidateLocalServiceURL("http://192.168.1.5:11434/"); err != nil {
		t.Errorf("private url rejected: %v", err)
	}
	if err := ValidateLocalServiceURL("http://169.254.169.254/latest/meta-data"); err == nil {
		t.Error("metadata endpoint must be rejected")
	}
	if err := ValidateLocalServiceURL("http://8.8.8.8/x"); err == nil {
		t.Error("public ip must be rejected")
	}
	if err := ValidateLocalServiceURL("file:///etc/passwd"); err == nil {
		t.Error("non-http scheme must be rejected")
	}
	if err := ValidateLocalServiceURL("http://"); err == nil {
		t.Error("empty host must be rejected")
	}
}

func TestNewLocalServiceClientDial(t *testing.T) {
	client := NewLocalServiceClient(time.Second)
	tr, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport type %T", client.Transport)
	}

	// 公网地址在 dial 层拒绝，不发起真实外连。
	if _, err := tr.DialContext(context.Background(), "tcp", "8.8.8.8:80"); err == nil {
		t.Error("dial to public address must be rejected")
	}
	// metadata 地址同样拒绝。
	if _, err := tr.DialContext(context.Background(), "tcp", "169.254.169.254:80"); err == nil {
		t.Error("dial to metadata address must be rejected")
	}

	// 本机地址正常连通。
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			conn.Close()
		}
	}()
	if _, err := tr.DialContext(context.Background(), "tcp", ln.Addr().String()); err != nil {
		t.Errorf("dial to loopback listener failed: %v", err)
	}
}

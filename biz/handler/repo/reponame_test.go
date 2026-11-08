package repo

import (
	"strings"
	"testing"
)

// issue #20：repoNameFromURL 的路径逃逸消毒——不合格名回退时间戳名，
// 目录逃逸从根上闭合（父目录守卫拦不住 ".."/"." 形态）。
func TestRepoNameFromURLSanitizes(t *testing.T) {
	timestamp := func(name string) bool {
		return strings.HasPrefix(name, "repo-") && len(name) == len("repo-20060102150405")
	}
	cases := []struct {
		in   string
		want string // "" = 时间戳名
	}{
		{"http://x/user/repo/..", ""},
		{"http://x/user/repo/.", ""},
		{"http://x/user/repo/?", ""},
		{"http://x/user/repo/", ""}, // 空段
		{"http://x/user/repo.git", "repo"},
		{"http://x/user/repo?query=x", ""}, // 含 '?' 的段名整体回退（目录名不该带查询串）
	}
	for _, c := range cases {
		got := repoNameFromURL(c.in)
		if c.want == "" {
			if !timestamp(got) {
				t.Errorf("%q: 期望时间戳名, got %q", c.in, got)
			}
		} else if got != c.want {
			t.Errorf("%q: got %q, want %q", c.in, got, c.want)
		}
	}
	// 段内 ".." 被取段自然中和（最后一段才是名字），不回退也不逃逸
	if got := repoNameFromURL("http://x/a/b/../evil"); got != "evil" {
		t.Errorf("段内 ..: got %q, want evil", got)
	}
	if got := repoNameFromURL("http://x/user/rep\ttab"); !timestamp(got) {
		t.Errorf("控制字符: got %q", got)
	}
	// 正常名不受影响
	if got := repoNameFromURL("git@github.com:owner/my-repo.git"); got != "my-repo" {
		t.Errorf("scp 形态: got %q", got)
	}
	if got := repoNameFromURL("https://example.com/x/My_Project.git"); got != "My_Project" {
		t.Errorf("下划线: got %q", got)
	}
}

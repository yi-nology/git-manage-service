package stats

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/timefmt"
)

// 端点一致性测试（#24）：同一组 since/until query 参数在
//   - /api/v1/stats/commits（GitService.GetCommits + ParseCommits）
//   - /api/v1/stats/analyze 与 /api/v1/stats/export/csv（共用 GetStats）
// 两条链路上必须得到同一个时间窗：commits 端点只返回窗口内提交且全部落在窗口内，
// analyze/export 的作者集合与窗口内提交的作者集合一致。
//
// fixture 的 author date 固定在 2024 年（committer date 保持为当下）：若过滤误按
// committer date 生效（如把 --since/--until 直接下推给 git log），2024 窗口将一条
// 都取不到，测试随之失败。

type windowFixtureCommit struct {
	msg    string
	date   string // 本地时区日期时间（git 按 local 解读无偏移的时间串）
	author string
	email  string
	file   string
}

var windowFixture = []windowFixtureCommit{
	{"commit-a", "2024-01-10T10:00:00", "甲", "a@example.com", "a.go"},
	{"commit-b", "2024-01-20T10:00:00", "乙", "b@example.com", "b.py"},
	{"commit-c", "2024-01-30T10:00:00", "甲", "a@example.com", "c.js"},
}

func buildWindowRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	run("init")
	for _, fc := range windowFixture {
		if err := os.WriteFile(filepath.Join(dir, fc.file), []byte("// "+fc.msg+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		run("add", fc.file)
		run("-c", "user.name="+fc.author, "-c", "user.email="+fc.email,
			"commit", "--date", fc.date, "-m", fc.msg)
	}
	return dir
}

// waitForStatsReady 轮询 GetStats 直至异步计算完成。
func waitForStatsReady(t *testing.T, svc *StatsService, dir, since, until string) *api.StatsResponse {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		data, status, err, _ := svc.GetStats(dir, "", since, until)
		if status != StatusProcessing {
			if err != nil {
				t.Fatalf("GetStats(%q, %q): %v", since, until, err)
			}
			if status != StatusReady {
				t.Fatalf("GetStats(%q, %q) 状态 = %v, 期望 %v", since, until, status, StatusReady)
			}
			return data
		}
		if time.Now().After(deadline) {
			t.Fatalf("GetStats(%q, %q) 计算超时", since, until)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestStatsEndpointsDateWindowConsistency(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git 不可用")
	}
	dir := buildWindowRepo(t)

	tests := []struct {
		name        string
		since       string
		until       string
		wantMsgs    []string // git log 新提交在前；无窗口提交时为 nil
		wantAuthors []string // 排序后的作者邮箱
	}{
		{
			name:        "无窗口返回全量",
			wantMsgs:    []string{"commit-c", "commit-b", "commit-a"},
			wantAuthors: []string{"a@example.com", "b@example.com"},
		},
		{
			name:        "仅since且当日提交包含",
			since:       "2024-01-10",
			wantMsgs:    []string{"commit-c", "commit-b", "commit-a"},
			wantAuthors: []string{"a@example.com", "b@example.com"},
		},
		{
			name:        "仅since过滤早期提交",
			since:       "2024-01-15",
			wantMsgs:    []string{"commit-c", "commit-b"},
			wantAuthors: []string{"a@example.com", "b@example.com"},
		},
		{
			name:        "仅until当日末尾包含当日提交",
			until:       "2024-01-20",
			wantMsgs:    []string{"commit-b", "commit-a"},
			wantAuthors: []string{"a@example.com", "b@example.com"},
		},
		{
			name:        "单日闭区间窗口",
			since:       "2024-01-20",
			until:       "2024-01-20",
			wantMsgs:    []string{"commit-b"},
			wantAuthors: []string{"b@example.com"},
		},
		{
			name:        "窗口只覆盖第一天",
			since:       "2024-01-01",
			until:       "2024-01-19",
			wantMsgs:    []string{"commit-a"},
			wantAuthors: []string{"a@example.com"},
		},
		{
			name:        "未来窗口必须为空",
			since:       "2999-01-01",
			until:       "2999-12-31",
			wantMsgs:    nil,
			wantAuthors: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &StatsService{Git: git.NewGitService()}

			// —— /api/v1/stats/commits 链路 ——
			raw, err := svc.Git.GetCommits(dir, "", tt.since, tt.until)
			if err != nil {
				t.Fatalf("GetCommits(%q, %q): %v", tt.since, tt.until, err)
			}
			commits := svc.ParseCommits(raw)
			var gotMsgs []string
			for _, cm := range commits {
				gotMsgs = append(gotMsgs, cm.Message)
			}
			if !slices.Equal(gotMsgs, tt.wantMsgs) {
				t.Errorf("commits 端点返回 %v, 期望 %v", gotMsgs, tt.wantMsgs)
			}

			// #24 要求的核心断言：窗口内返回的提交必须全部落在窗口内，
			// 否则说明参数没有生效。
			sinceTime, untilTime, err := timefmt.ParseDateWindow(tt.since, tt.until)
			if err != nil {
				t.Fatalf("ParseDateWindow(%q, %q): %v", tt.since, tt.until, err)
			}
			for _, cm := range commits {
				if !timefmt.InDateWindow(cm.Date, sinceTime, untilTime) {
					t.Errorf("提交 %s (%v) 落在窗口外却被返回", cm.Message, cm.Date)
				}
			}

			// —— /api/v1/stats/analyze 与 /export/csv 链路（共用 GetStats）——
			data := waitForStatsReady(t, svc, dir, tt.since, tt.until)
			var gotAuthors []string
			for _, a := range data.Authors {
				gotAuthors = append(gotAuthors, a.Email)
			}
			slices.Sort(gotAuthors)
			if !slices.Equal(gotAuthors, tt.wantAuthors) {
				t.Errorf("analyze/export 作者集合 = %v, 期望 %v（与 commits 端点同窗不一致）", gotAuthors, tt.wantAuthors)
			}
		})
	}
}

// TestStatsDateWindowInvalidDateExplicit 非法日期必须显式报错（#24 的同族静默失败面）：
// 此前 stats 侧忽略 ParseInLocation 错误、静默退化成「不过滤」。
func TestStatsDateWindowInvalidDateExplicit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git 不可用")
	}
	dir := buildWindowRepo(t)
	svc := &StatsService{Git: git.NewGitService()}

	// /stats/commits：GetCommits 显式报错
	_, err := svc.Git.GetCommits(dir, "", "2024-13-99", "")
	var dwe *timefmt.DateWindowError
	if !errors.As(err, &dwe) {
		t.Fatalf("GetCommits 非法 since 期望 *DateWindowError, 得到 %v", err)
	}

	// /analyze（/export/csv 同源）：GetStats 同步 fail-fast，不进入 processing
	_, status, err, _ := svc.GetStats(dir, "", "", "not-a-date")
	if !errors.As(err, &dwe) {
		t.Fatalf("GetStats 非法 until 期望 *DateWindowError, 得到 %v", err)
	}
	if status != StatusFailed {
		t.Errorf("GetStats 非法 until 状态 = %v, 期望 %v", status, StatusFailed)
	}

	// 非法日期不产生 processing 残留：随后同一仓库的合法窗口可正常完成
	data := waitForStatsReady(t, svc, dir, "2024-01-01", "2024-12-31")
	if len(data.Authors) != 2 {
		t.Errorf("非法日期之后的合法调用 Authors = %d, 期望 2", len(data.Authors))
	}
}

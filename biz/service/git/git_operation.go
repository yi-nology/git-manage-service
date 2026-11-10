package git

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/pkg/timefmt"
	"github.com/yi-nology/go-git-platform/gitbackend"
)

type channelWriter struct {
	ch chan string
}

func (w *channelWriter) Write(p []byte) (n int, err error) {
	if w.ch != nil {
		w.ch <- string(p)
	}
	return len(p), nil
}

func (s *GitService) IsGitRepo(path string) bool {
	// An empty path would make `git -C "" ...` resolve to the process working
	// directory, falsely treating it (e.g. the server's own repo) as the target.
	if path == "" {
		return false
	}
	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func (s *GitService) Fetch(path, remote string, progress io.Writer, skipTLS ...bool) error {
	insecure := len(skipTLS) > 0 && skipTLS[0]

	_, err := s.backend.Fetch(context.Background(), gitbackend.FetchOptions{
		RepoPath:        path,
		Remote:          remote,
		Tags:            true,
		InsecureSkipTLS: insecure,
		Progress:        progress,
	})
	return err
}

func (s *GitService) Clone(remoteURL, localPath, authType, authKey, authSecret string, skipTLS ...bool) error {
	return s.CloneWithProgress(remoteURL, localPath, authType, authKey, authSecret, nil, skipTLS...)
}

func (s *GitService) CloneWithProgress(remoteURL, localPath, authType, authKey, authSecret string, progressChan chan string, skipTLS ...bool) error {
	auth := s.buildSDKAuth(authType, authKey, authSecret)

	var progress io.Writer
	if progressChan != nil {
		progress = &channelWriter{ch: progressChan}
	}

	insecure := len(skipTLS) > 0 && skipTLS[0]

	return s.backend.Clone(context.Background(), gitbackend.CloneOptions{
		URL:             remoteURL,
		Path:            localPath,
		Auth:            auth,
		Progress:        progress,
		InsecureSkipTLS: insecure,
	})
}

func (s *GitService) CloneWithAuthMethod(remoteURL, localPath string, auth gitbackend.AuthConfig, progressChan chan string, skipTLS ...bool) error {
	var progress io.Writer
	if progressChan != nil {
		progress = &channelWriter{ch: progressChan}
	}

	insecure := len(skipTLS) > 0 && skipTLS[0]

	return s.backend.Clone(context.Background(), gitbackend.CloneOptions{
		URL:             remoteURL,
		Path:            localPath,
		Auth:            auth,
		Progress:        progress,
		InsecureSkipTLS: insecure,
	})
}

func (s *GitService) GetCommitHash(path, remote, branch string) (string, error) {
	refName := fmt.Sprintf("refs/remotes/%s/%s", remote, branch)
	hash, err := s.backend.RevParse(context.Background(), path, refName)
	if err != nil {
		return "", fmt.Errorf("remote branch %s/%s not found: %v", remote, branch, err)
	}
	return hash, nil
}

func (s *GitService) IsAncestor(path, ancestor, descendant string) (bool, error) {
	return s.backend.IsAncestor(context.Background(), path, ancestor, descendant)
}

func (s *GitService) GetBranches(path string) ([]string, error) {
	return s.backend.ListLocalBranches(context.Background(), path)
}

// GetCommits 输出提交列表（%H|%aN|%aE|%ad|%s，author date）。
// since/until 为 YYYY-MM-DD 本地日期窗口，口径与统计侧（calculateStatsFast）一致：
// since 取当日零点、until 取当日末尾（见 timefmt.ParseDateWindow）。
//
// 时间过滤在 Go 侧按 author date（%ad）进行：git log --since/--until 过滤的是
// committer date，与本服务的统计口径（%ad/%at = author date）不一致，rebase /
// amend 过的提交两个日期会分叉，故不用 --since/--until 直接下推。
func (s *GitService) GetCommits(path, branch, since, until string) (string, error) {
	sinceTime, untilTime, err := timefmt.ParseDateWindow(since, until)
	if err != nil {
		return "", err
	}

	args := []string{"-C", path, "log", "--pretty=format:%H|%aN|%aE|%ad|%s", "--date=format:%Y-%m-%d %H:%M:%S %z"}
	if branch != "" {
		args = append(args, branch)
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %v", err)
	}
	return filterCommitsByWindow(string(output), sinceTime, untilTime), nil
}

// filterCommitsByWindow 逐行按提交时间（author date）过滤 git log 输出。
// 无窗口时原样返回；有窗口时解析不了提交时间的行一律丢弃（fail-closed，
// 无法证明落在窗口内的行不应混入结果）。
func filterCommitsByWindow(raw string, sinceTime, untilTime time.Time) string {
	if sinceTime.IsZero() && untilTime.IsZero() {
		return raw
	}
	lines := strings.Split(raw, "\n")
	lines = slices.DeleteFunc(lines, func(line string) bool {
		date, ok := commitAuthorDate(line)
		if !ok {
			return true
		}
		return !timefmt.InDateWindow(date, sinceTime, untilTime)
	})
	return strings.Join(lines, "\n")
}

// commitAuthorDate 从 GetCommits 输出的一行（hash|name|email|date|message）
// 中解析 author date；解析失败返回 ok=false。
func commitAuthorDate(line string) (time.Time, bool) {
	parts := strings.SplitN(line, "|", 5)
	if len(parts) < 5 {
		return time.Time{}, false
	}
	t, err := time.Parse(timefmt.LayoutGitLog, parts[3])
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func (s *GitService) ResolveRevision(path, rev string) (string, error) {
	return s.backend.RevParse(context.Background(), path, rev)
}

func (s *GitService) GetHeadBranch(path string) (string, error) {
	return s.backend.GetCurrentBranch(context.Background(), path)
}

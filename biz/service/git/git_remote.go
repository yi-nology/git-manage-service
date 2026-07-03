package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/pkg/logger"
	"github.com/yi-nology/git-platform-sdk/gitbackend"
)

// GetRemotes 获取所有远程仓库名称
func (s *GitService) GetRemotes(path string) ([]string, error) {
	return s.backend.GetRemotes(context.Background(), path)
}

// GetRemoteURL 获取远程仓库 URL
func (s *GitService) GetRemoteURL(path, remoteName string) (string, error) {
	return s.backend.GetRemoteURL(context.Background(), path, remoteName)
}

// AddRemote 添加远程仓库
func (s *GitService) AddRemote(path, name, url string, isMirror bool) error {
	logger.Info("Adding remote", logrus.Fields{
		"path":     path,
		"name":     name,
		"url":      url,
		"isMirror": isMirror,
	})

	err := s.backend.AddRemote(context.Background(), path, name, url)
	if err != nil {
		logger.ErrorWithErr("Failed to add remote", err, logrus.Fields{"name": name})
		return err
	}

	logger.Info("Remote added successfully", logrus.Fields{"name": name})
	return nil
}

// RemoveRemote 删除远程仓库
func (s *GitService) RemoveRemote(path, name string) error {
	logger.Info("Removing remote", logrus.Fields{"path": path, "name": name})

	err := s.backend.RemoveRemote(context.Background(), path, name)
	if err != nil {
		logger.ErrorWithErr("Failed to remove remote", err, logrus.Fields{"name": name})
		return err
	}

	logger.Info("Remote removed successfully", logrus.Fields{"name": name})
	return nil
}

// SetRemotePushURL 设置远程仓库的推送 URL
func (s *GitService) SetRemotePushURL(path, name, url string) error {
	logger.Info("Setting remote push URL", logrus.Fields{
		"path": path,
		"name": name,
		"url":  url,
	})

	key := fmt.Sprintf("remote.%s.url", name)
	if err := s.backend.SetConfig(context.Background(), path, key, url); err != nil {
		logger.ErrorWithErr("Failed to set remote URL", err, logrus.Fields{"name": name})
		return err
	}

	logger.Info("Remote URL set successfully", logrus.Fields{"name": name})
	return nil
}

// GetRepoConfig 获取仓库配置信息
func (s *GitService) GetRepoConfig(path string) (*domain.GitRepoConfig, error) {
	out, _, err := s.backend.RunRaw(context.Background(), path, []string{"config", "--local", "--list"})
	if err != nil {
		return nil, fmt.Errorf("failed to read repo config: %w", err)
	}

	repoConfig := &domain.GitRepoConfig{
		Remotes:  []domain.GitRemote{},
		Branches: []domain.GitBranch{},
	}

	remoteMap := make(map[string]*domain.GitRemote)
	branchMap := make(map[string]*domain.GitBranch)

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := line[:idx]
		val := line[idx+1:]

		parts := strings.Split(key, ".")
		// remote.<name>.url / remote.<name>.mirror / remote.<name>.fetch / remote.<name>.pushurl
		if len(parts) >= 3 && parts[0] == "remote" {
			name := parts[1]
			prop := strings.Join(parts[2:], ".")
			r := remoteMap[name]
			if r == nil {
				r = &domain.GitRemote{Name: name, FetchSpecs: []string{}, PushSpecs: []string{}}
				remoteMap[name] = r
			}
			switch prop {
			case "url":
				if r.FetchURL == "" {
					r.FetchURL = val
				}
				r.PushURL = val
				r.FetchSpecs = append(r.FetchSpecs, val)
			case "pushurl":
				r.PushURL = val
			case "fetch":
				r.FetchSpecs = append(r.FetchSpecs, val)
			case "mirror":
				if val == "true" {
					r.IsMirror = true
				}
			}
		}
		// branch.<name>.remote / branch.<name>.merge
		if len(parts) >= 3 && parts[0] == "branch" {
			name := parts[1]
			prop := strings.Join(parts[2:], ".")
			b := branchMap[name]
			if b == nil {
				b = &domain.GitBranch{Name: name}
				branchMap[name] = b
			}
			switch prop {
			case "remote":
				b.Remote = val
			case "merge":
				b.Merge = val
			}
		}
	}

	for _, r := range remoteMap {
		repoConfig.Remotes = append(repoConfig.Remotes, *r)
	}
	for _, b := range branchMap {
		if b.Remote != "" && b.Merge != "" {
			shortRef := b.Merge
			shortRef = strings.TrimPrefix(shortRef, "refs/heads/")
			b.UpstreamRef = fmt.Sprintf("%s/%s", b.Remote, shortRef)
		}
		repoConfig.Branches = append(repoConfig.Branches, *b)
	}

	logger.Debug("Repo config retrieved", logrus.Fields{
		"path":     path,
		"remotes":  len(repoConfig.Remotes),
		"branches": len(repoConfig.Branches),
	})
	return repoConfig, nil
}

// ListRemoteBranches 获取指定远程的所有分支名（基于本地 remote-tracking refs）
func (s *GitService) ListRemoteBranches(path, remoteName string) ([]string, error) {
	branches, err := s.backend.ListRemoteBranches(context.Background(), path, remoteName)
	if err != nil {
		return nil, err
	}

	logger.Debug("Remote branches listed", logrus.Fields{"path": path, "remote": remoteName, "count": len(branches)})
	return branches, nil
}

// TestRemoteConnection 测试远程连接
func (s *GitService) TestRemoteConnection(url string, skipTLS ...bool) error {
	logger.Info("Testing remote connection", logrus.Fields{"url": url})
	err := s.backend.TestConnection(context.Background(), url, s.detectSDKAuth(url))
	if err != nil {
		logger.ErrorWithErr("Remote connection test failed", err, logrus.Fields{"url": url})
		return err
	}
	logger.Info("Remote connection test successful", logrus.Fields{"url": url})
	return nil
}

// detectSDKAuth builds a gitbackend.AuthConfig by auto-detecting SSH keys.
func (s *GitService) detectSDKAuth(urlStr string) gitbackend.AuthConfig {
	return s.detectSSHAuth(urlStr)
}

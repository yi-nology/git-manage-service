package git

import (
	"context"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/pkg/logger"
)

// globalGitConfig runs `git config --global <key>` and returns the trimmed value.
func globalGitConfig(key string) string {
	out, err := exec.Command("git", "config", "--global", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// setGlobalGitConfig runs `git config --global <key> <value>`.
func setGlobalGitConfig(key, value string) error {
	return exec.Command("git", "config", "--global", key, value).Run()
}

// GetGitUser 获取仓库的 git 用户配置
func (s *GitService) GetGitUser(path string) (string, string, error) {
	logger.Debug("Getting git user", logrus.Fields{"path": path})

	var name, email string

	// 1. 尝试本地配置
	localName, err := s.backend.GetConfig(context.Background(), path, "user.name")
	if err == nil {
		name = localName
	}
	localEmail, err := s.backend.GetConfig(context.Background(), path, "user.email")
	if err == nil {
		email = localEmail
	}

	if name != "" && email != "" {
		return name, email, nil
	}

	// 2. 尝试全局配置 (~/.gitconfig)
	if name == "" {
		name = globalGitConfig("user.name")
	}
	if email == "" {
		email = globalGitConfig("user.email")
	}

	logger.Debug("Git user retrieved", logrus.Fields{
		"path":  path,
		"name":  name,
		"email": email,
	})
	return name, email, nil
}

// GetGlobalGitUser 获取全局 git 用户配置
func (s *GitService) GetGlobalGitUser() (string, string, error) {
	logger.Debug("Getting global git user")

	name := globalGitConfig("user.name")
	email := globalGitConfig("user.email")

	logger.Debug("Global git user retrieved", logrus.Fields{
		"name":  name,
		"email": email,
	})
	return name, email, nil
}

// SetGlobalGitUser 设置全局 git 用户配置
func (s *GitService) SetGlobalGitUser(name, email string) error {
	logger.Info("Setting global git user", logrus.Fields{
		"name":  name,
		"email": email,
	})

	if err := setGlobalGitConfig("user.name", name); err != nil {
		logger.ErrorWithErr("Failed to set global user.name", err, nil)
		return err
	}
	if err := setGlobalGitConfig("user.email", email); err != nil {
		logger.ErrorWithErr("Failed to set global user.email", err, nil)
		return err
	}

	logger.Info("Global git user set successfully", logrus.Fields{
		"name":  name,
		"email": email,
	})
	return nil
}

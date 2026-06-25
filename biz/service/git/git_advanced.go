package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AuthorInfo 作者信息
type AuthorInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetAuthors 获取仓库的所有提交作者列表
func (s *GitService) GetAuthors(path string) ([]AuthorInfo, error) {
	// 使用 SDK 获取最近的提交来提取作者
	commits, err := s.backend.GetCommitsBetween(context.Background(), path, "", "HEAD")
	if err != nil {
		return nil, err
	}

	// 使用 map 去重
	authorMap := make(map[string]AuthorInfo)
	for _, c := range commits {
		key := c.Author + "|"
		if _, exists := authorMap[key]; !exists {
			authorMap[key] = AuthorInfo{
				Name:  c.Author,
				Email: "",
			}
		}
	}

	// 转换为切片
	authors := make([]AuthorInfo, 0, len(authorMap))
	for _, author := range authorMap {
		authors = append(authors, author)
	}

	return authors, nil
}

// CherryPick 执行cherry-pick操作
func (s *GitService) CherryPick(path, commitHash string, noCommit bool) (string, []string, error) {
	if noCommit {
		// 使用 SDK 的 CherryPick（不提交）
		err := s.backend.CherryPick(context.Background(), path, commitHash)
		if err != nil {
			// 检查是否是冲突
			if strings.Contains(err.Error(), "conflict") {
				conflicts := s.getConflictFiles(path)
				return "", conflicts, fmt.Errorf("cherry-pick conflict")
			}
			return "", nil, err
		}
		return "", nil, nil
	}

	// 使用 SDK 的 CherryPick
	err := s.backend.CherryPick(context.Background(), path, commitHash)
	if err != nil {
		// 检查是否是冲突
		if strings.Contains(err.Error(), "conflict") {
			conflicts := s.getConflictFiles(path)
			return "", conflicts, fmt.Errorf("cherry-pick conflict")
		}
		return "", nil, err
	}

	// 获取新的commit hash
	newHash, _ := s.backend.RevParse(context.Background(), path, "HEAD")
	return newHash, nil, nil
}

// CherryPickAbort 中止cherry-pick
func (s *GitService) CherryPickAbort(path string) error {
	// SDK 没有 CherryPickAbort，使用 RunCommand
	_, err := s.RunCommand(path, "cherry-pick", "--abort")
	return err
}

// Rebase 执行rebase操作
func (s *GitService) Rebase(path, upstream, onto string) (bool, []string, error) {
	if onto != "" {
		// SDK 不支持 --onto，使用 RunCommand
		output, err := s.RunCommand(path, "rebase", "--onto", onto, upstream)
		if err != nil {
			if strings.Contains(output, "conflict") || strings.Contains(output, "CONFLICT") {
				conflicts := s.getConflictFiles(path)
				return false, conflicts, nil
			}
			return false, nil, err
		}
		return true, nil, nil
	}

	// 使用 SDK 的 Rebase
	err := s.backend.Rebase(context.Background(), path, upstream)
	if err != nil {
		if strings.Contains(err.Error(), "conflict") {
			conflicts := s.getConflictFiles(path)
			return false, conflicts, nil
		}
		return false, nil, err
	}
	return true, nil, nil
}

// RebaseAbort 中止rebase
func (s *GitService) RebaseAbort(path string) error {
	return s.backend.RebaseAbort(context.Background(), path)
}

// RebaseContinue 继续rebase
func (s *GitService) RebaseContinue(path string) (bool, []string, error) {
	err := s.backend.RebaseContinue(context.Background(), path)
	if err != nil {
		if strings.Contains(err.Error(), "conflict") {
			conflicts := s.getConflictFiles(path)
			return false, conflicts, nil
		}
		return false, nil, err
	}
	return true, nil, nil
}

// RebaseSkip 跳过当前commit
func (s *GitService) RebaseSkip(path string) error {
	_, err := s.RunCommand(path, "rebase", "--skip")
	return err
}

// IsRebaseInProgress 检查是否有进行中的rebase
func (s *GitService) IsRebaseInProgress(path string) bool {
	// 检查 .git/rebase-merge 或 .git/rebase-apply 目录
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(filepath.Join(gitDir, "rebase-merge")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(gitDir, "rebase-apply")); err == nil {
		return true
	}
	return false
}

// getConflictFiles 获取冲突文件列表
func (s *GitService) getConflictFiles(path string) []string {
	output, err := s.RunCommand(path, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil
	}
	if output == "" {
		return nil
	}
	return strings.Split(strings.TrimSpace(output), "\n")
}

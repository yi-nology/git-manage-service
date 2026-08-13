package git

import (
	"context"
	"fmt"
	"time"

	"github.com/yi-nology/git-platform-sdk/gitbackend"
)

type CommitInfo struct {
	Author         string    `json:"author"`
	AuthorEmail    string    `json:"authorEmail"`
	Committer      string    `json:"committer"`
	CommitterEmail string    `json:"committerEmail"`
	Message        string    `json:"message"`
	CommitTime     time.Time `json:"commitTime"`
}

// GetCommitInfo 获取提交信息
func (s *GitService) GetCommitInfo(repoPath, hashStr string) (*CommitInfo, error) {
	commits, err := s.backend.GetCommitsBetween(context.Background(), repoPath, hashStr, hashStr)
	if err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, nil
	}

	commit := commits[0]
	return &CommitInfo{
		Author:      commit.Author,
		AuthorEmail: "",
		Message:     commit.Message,
		CommitTime:  time.Time{},
	}, nil
}

// GetRecentCommits 获取最近的提交历史
func (s *GitService) GetRecentCommits(repoPath string, limit int) ([]string, error) {
	commits, err := s.backend.GetCommitsBetween(context.Background(), repoPath, fmt.Sprintf("HEAD~%d", limit), "HEAD")
	if err != nil {
		return nil, err
	}

	var hashes []string
	for _, commit := range commits {
		hashes = append(hashes, commit.Hash)
	}
	return hashes, nil
}

// GetCommitDiffSimple 获取提交的diff（简化版，用于分析）
func (s *GitService) GetCommitDiffSimple(repoPath, hashStr string) (string, error) {
	diff, err := s.backend.Diff(context.Background(), repoPath, gitbackend.DiffOptions{
		From: hashStr + "~1",
		To:   hashStr,
	})
	if err != nil {
		return "", err
	}

	return diff, nil
}

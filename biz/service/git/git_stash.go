package git

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/pkg/logger"
)

// StashEntry Stash 条目
type StashEntry struct {
	Index   int    `json:"index"`
	Ref     string `json:"ref"`
	Message string `json:"message"`
	Branch  string `json:"branch"`
	Date    string `json:"date"`
}

// StashList 列出所有 stash
func (s *GitService) StashList(path string) ([]StashEntry, error) {
	logger.Debug("Listing stash entries", logrus.Fields{"path": path})

	entries, err := s.backend.StashList(context.Background(), path)
	if err != nil {
		logger.ErrorWithErr("Failed to list stash", err, logrus.Fields{"path": path})
		return nil, err
	}

	result := make([]StashEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, StashEntry{
			Index:   entry.Index,
			Ref:     fmt.Sprintf("stash@{%d}", entry.Index),
			Message: entry.Message,
		})
	}

	logger.Debug("Stash entries retrieved", logrus.Fields{"path": path, "count": len(result)})
	return result, nil
}

// StashSave 保存当前更改到 stash
func (s *GitService) StashSave(path, message string, includeUntracked bool) error {
	logger.Info("Saving stash", logrus.Fields{
		"path":              path,
		"message":           message,
		"include_untracked": includeUntracked,
	})

	// 检查是否有可暂存的更改
	status, err := s.backend.GetStatus(context.Background(), path)
	if err != nil {
		logger.ErrorWithErr("Failed to check status", err, logrus.Fields{"path": path})
		return err
	}

	hasChanges := !status.IsClean
	if !hasChanges {
		logger.Warn("No changes to stash", logrus.Fields{"path": path})
		return fmt.Errorf("no changes to stash")
	}

	err = s.backend.StashSave(context.Background(), path, message)
	if err != nil {
		logger.ErrorWithErr("Failed to save stash", err, logrus.Fields{"path": path})
		return err
	}

	logger.Info("Stash saved successfully", logrus.Fields{"path": path})
	return nil
}

// StashApply 应用 stash（不删除）
func (s *GitService) StashApply(path string, index int) error {
	ref := fmt.Sprintf("stash@{%d}", index)
	logger.Info("Applying stash", logrus.Fields{"path": path, "ref": ref})

	err := s.backend.StashApply(context.Background(), path, index)
	if err != nil {
		logger.ErrorWithErr("Failed to apply stash", err, logrus.Fields{"ref": ref})
		return err
	}

	logger.Info("Stash applied successfully", logrus.Fields{"ref": ref})
	return nil
}

// StashPop 弹出 stash（应用并删除）
func (s *GitService) StashPop(path string, index int) error {
	ref := fmt.Sprintf("stash@{%d}", index)
	logger.Info("Popping stash", logrus.Fields{"path": path, "ref": ref})

	err := s.backend.StashPop(context.Background(), path, index)
	if err != nil {
		logger.ErrorWithErr("Failed to pop stash", err, logrus.Fields{"ref": ref})
		return err
	}

	logger.Info("Stash popped successfully", logrus.Fields{"ref": ref})
	return nil
}

// StashDrop 删除指定 stash
func (s *GitService) StashDrop(path string, index int) error {
	ref := fmt.Sprintf("stash@{%d}", index)
	logger.Info("Dropping stash", logrus.Fields{"path": path, "ref": ref})

	err := s.backend.StashDrop(context.Background(), path, index)
	if err != nil {
		logger.ErrorWithErr("Failed to drop stash", err, logrus.Fields{"ref": ref})
		return err
	}

	logger.Info("Stash dropped successfully", logrus.Fields{"ref": ref})
	return nil
}

// StashClear 清空所有 stash
func (s *GitService) StashClear(path string) error {
	logger.Info("Clearing all stash entries", logrus.Fields{"path": path})

	err := s.backend.StashClear(context.Background(), path)
	if err != nil {
		logger.ErrorWithErr("Failed to clear stash", err, logrus.Fields{"path": path})
		return err
	}

	logger.Info("All stash entries cleared", logrus.Fields{"path": path})
	return nil
}

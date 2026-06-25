package git

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/pkg/logger"
	"github.com/yi-nology/git-platform-sdk/gitbackend"
)

// TagInfo Tag 信息结构
type TagInfo struct {
	Name    string    `json:"name"`
	Hash    string    `json:"hash"`
	Message string    `json:"message"`
	Tagger  string    `json:"tagger"`
	Date    time.Time `json:"date"`
}

// CreateTag 创建标签
func (s *GitService) CreateTag(path, tagName, ref, message, authorName, authorEmail string) error {
	logger.Info("Creating tag", logrus.Fields{
		"path":    path,
		"tag":     tagName,
		"ref":     ref,
		"message": message,
	})

	err := s.backend.CreateTag(context.Background(), path, tagName, ref)
	if err != nil {
		logger.ErrorWithErr("Failed to create tag", err, logrus.Fields{"tag": tagName})
		return err
	}

	logger.Info("Tag created successfully", logrus.Fields{"tag": tagName})
	return nil
}

// PushTag 推送标签到远程
func (s *GitService) PushTag(path, remoteName, tagName, authType, authKey, authSecret string, skipTLS ...bool) error {
	logger.Info("Pushing tag", logrus.Fields{
		"path":   path,
		"remote": remoteName,
		"tag":    tagName,
	})

	auth := s.buildSDKAuth(authType, authKey, authSecret)

	err := s.backend.PushTag(context.Background(), path, remoteName, tagName, auth)
	if err != nil {
		logger.ErrorWithErr("Failed to push tag", err, logrus.Fields{"tag": tagName})
		return err
	}

	logger.Info("Tag pushed successfully", logrus.Fields{"tag": tagName, "remote": remoteName})
	return nil
}

// DeleteTag 删除本地标签
func (s *GitService) DeleteTag(path, tagName string) error {
	logger.Info("Deleting tag", logrus.Fields{"path": path, "tag": tagName})

	err := s.backend.DeleteTag(context.Background(), path, tagName)
	if err != nil {
		logger.ErrorWithErr("Failed to delete tag", err, logrus.Fields{"tag": tagName})
		return err
	}

	logger.Info("Tag deleted successfully", logrus.Fields{"tag": tagName})
	return nil
}

// DeleteRemoteTag 删除远程标签
func (s *GitService) DeleteRemoteTag(path, remoteName, tagName, authType, authKey, authSecret string, skipTLS ...bool) error {
	logger.Info("Deleting remote tag", logrus.Fields{
		"path":   path,
		"remote": remoteName,
		"tag":    tagName,
	})

	auth := s.buildSDKAuth(authType, authKey, authSecret)

	// Push empty refspec to delete remote tag
	_, err := s.backend.Push(context.Background(), gitbackend.PushOptions{
		RepoPath: path,
		Remote:   remoteName,
		RefSpecs: []string{":refs/tags/" + tagName},
		Auth:     auth,
	})
	if err != nil {
		logger.ErrorWithErr("Failed to delete remote tag", err, logrus.Fields{"tag": tagName})
		return err
	}

	logger.Info("Remote tag deleted successfully", logrus.Fields{"tag": tagName, "remote": remoteName})
	return nil
}

// GetTags 获取所有标签名
func (s *GitService) GetTags(path string) ([]string, error) {
	tags, err := s.backend.GetTagList(context.Background(), path)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, tag := range tags {
		names = append(names, tag.Name)
	}

	logger.Debug("Tags retrieved", logrus.Fields{"path": path, "count": len(names)})
	return names, nil
}

// GetTagList 获取详细的标签列表
func (s *GitService) GetTagList(path string) ([]TagInfo, error) {
	tags, err := s.backend.GetTagList(context.Background(), path)
	if err != nil {
		return nil, err
	}

	var result []TagInfo
	for _, tag := range tags {
		result = append(result, TagInfo{
			Name:    tag.Name,
			Hash:    tag.Hash,
			Message: tag.Message,
			Tagger:  tag.Author,
			Date:    time.Time{},
		})
	}

	logger.Debug("Tag list retrieved", logrus.Fields{"path": path, "count": len(result)})
	return result, nil
}

// biz/service/git/git_file.go - Git文件浏览服务

package git

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

// TreeEntry 目录树条目
type TreeEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"` // "file" 或 "dir"
	Size int64  `json:"size"`
	Mode string `json:"mode"`
	Hash string `json:"hash"`
}

// BlobContent 文件内容
type BlobContent struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"` // "utf-8" 或 "base64"
	Size     int64  `json:"size"`
	IsBinary bool   `json:"is_binary"`
	MimeType string `json:"mime_type"`
}

// FileCommit 文件提交记录
type FileCommit struct {
	Hash      string `json:"hash"`
	ShortHash string `json:"short_hash"`
	Message   string `json:"message"`
	Author    string `json:"author"`
	Date      string `json:"date"`
}

// GetTree 获取目录树
func (s *GitService) GetTree(repoPath, ref, dirPath string, recursive bool) ([]TreeEntry, error) {
	sdkEntries, err := s.backend.GetTree(context.Background(), repoPath, ref, dirPath, recursive)
	if err != nil {
		return nil, err
	}

	var entries []TreeEntry
	for _, e := range sdkEntries {
		entries = append(entries, TreeEntry{
			Name: e.Name,
			Path: e.Path,
			Type: string(e.Type),
			Mode: e.Mode,
			Hash: e.Hash,
			Size: e.Size,
		})
	}
	return entries, nil
}

func (s *GitService) GetWorktree(repoPath, dirPath string) ([]TreeEntry, error) {
	fullPath := filepath.Join(repoPath, dirPath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	skipDirs := map[string]bool{".git": true}

	var result []TreeEntry
	for _, e := range entries {
		name := e.Name()
		if skipDirs[name] {
			continue
		}

		entryPath := name
		if dirPath != "" {
			entryPath = filepath.Join(dirPath, name)
		}
		entryPath = filepath.ToSlash(entryPath)

		if e.IsDir() {
			result = append(result, TreeEntry{
				Name: name,
				Path: entryPath,
				Type: "dir",
				Mode: "40000",
			})
		} else {
			info, err := e.Info()
			if err != nil {
				continue
			}
			result = append(result, TreeEntry{
				Name: name,
				Path: entryPath,
				Type: "file",
				Size: info.Size(),
				Mode: "100644",
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Type != result[j].Type {
			return result[i].Type == "dir"
		}
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func (s *GitService) GetBlob(repoPath, ref, filePath string) (*BlobContent, error) {
	sdkBlob, err := s.backend.GetBlob(context.Background(), repoPath, ref, filePath)
	if err != nil {
		return nil, err
	}

	return &BlobContent{
		Content:  sdkBlob.Content,
		Encoding: string(sdkBlob.Encoding),
		Size:     sdkBlob.Size,
		IsBinary: sdkBlob.IsBinary,
		MimeType: getMimeType(filePath),
	}, nil
}

func (s *GitService) GetWorktreeBlob(repoPath, filePath string) (*BlobContent, error) {
	fullPath := filepath.Join(repoPath, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	isBinary := !utf8.Valid(content) || containsNullByte(content)

	result := &BlobContent{
		Size:     int64(len(content)),
		IsBinary: isBinary,
		MimeType: getMimeType(filePath),
	}

	if isBinary {
		result.Content = base64.StdEncoding.EncodeToString(content)
		result.Encoding = "base64"
	} else {
		result.Content = string(content)
		result.Encoding = "utf-8"
	}

	return result, nil
}

// GetFileHistory 获取文件的提交历史
func (s *GitService) GetFileHistory(repoPath, ref, filePath string, limit int) ([]FileCommit, error) {
	// 使用 SDK 获取文件历史
	commits, err := s.backend.GetFileHistory(context.Background(), repoPath, filePath, limit)
	if err != nil {
		return nil, err
	}

	var result []FileCommit
	for _, c := range commits {
		shortHash := c.Hash
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}
		result = append(result, FileCommit{
			Hash:      c.Hash,
			ShortHash: shortHash,
			Message:   c.Message,
			Author:    c.Author,
			Date:      c.Date,
		})
	}

	return result, nil
}

// containsNullByte 检查是否包含空字节
func containsNullByte(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

// getMimeType 根据文件扩展名获取MIME类型
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeTypes := map[string]string{
		".go":   "text/x-go",
		".js":   "application/javascript",
		".ts":   "application/typescript",
		".py":   "text/x-python",
		".java": "text/x-java",
		".c":    "text/x-c",
		".cpp":  "text/x-c++",
		".h":    "text/x-c",
		".rs":   "text/x-rust",
		".rb":   "text/x-ruby",
		".php":  "text/x-php",
		".sh":   "text/x-shellscript",
		".md":   "text/markdown",
		".json": "application/json",
		".xml":  "application/xml",
		".yaml": "text/yaml",
		".yml":  "text/yaml",
		".html": "text/html",
		".css":  "text/css",
		".sql":  "text/x-sql",
		".txt":  "text/plain",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
	}
	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

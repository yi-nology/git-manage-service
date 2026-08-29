package spec

import (
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

// SpecEntry 是目录扫描产出的原始条目。ModTime 保留 time.Time，
// 不在这里做格式化；调用方自行决定序列化布局（通常用 timefmt.LayoutAPITime）。
type SpecEntry struct {
	Name    string
	Path    string // 相对仓库根目录的路径
	IsDir   bool
	Size    int64
	ModTime time.Time
}

// SpecWalkOptions 控制 WalkSpecEntries 在两个历史实现之间的行为差异点。
type SpecWalkOptions struct {
	// IncludeDirs 为 true 时同时收集目录条目（spec 树构建需要）；
	// 为 false 时仅收集普通 *.spec 文件。
	IncludeDirs bool
	// SkipAnyGitPath 为 true 时按"路径包含 .git"判断跳过（含 .github 等子串命中的路径）；
	// 为 false 时仅跳过名为 ".git" 的目录。
	SkipAnyGitPath bool
}

// WalkSpecEntries 用 filepath.WalkDir 遍历仓库目录、跳过 .git、收集 spec 条目。
// 返回的条目按 WalkDir 的字典序排列；出错时返回已收集的部分结果与错误。
func WalkSpecEntries(repoPath string, opts SpecWalkOptions) ([]SpecEntry, error) {
	var entries []SpecEntry

	err := filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if opts.SkipAnyGitPath {
			if strings.Contains(path, ".git") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		} else if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		var want bool
		if opts.IncludeDirs {
			want = strings.HasSuffix(d.Name(), ".spec") || d.IsDir()
		} else {
			want = !d.IsDir() && strings.HasSuffix(d.Name(), ".spec")
		}
		if !want {
			return nil
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		relPath, _ := filepath.Rel(repoPath, path)
		entries = append(entries, SpecEntry{
			Name:    d.Name(),
			Path:    relPath,
			IsDir:   d.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})

		return nil
	})

	return entries, err
}

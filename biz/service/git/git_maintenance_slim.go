package git

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
)

func validateSlimPath(p string) error {
	if p == "" || len(p) > 4096 {
		return fmt.Errorf("invalid path %q", p)
	}
	if filepath.IsAbs(p) || strings.HasPrefix(p, "-") {
		return fmt.Errorf("path must be repo-relative and must not start with '-': %q", p)
	}
	if strings.ContainsAny(p, "'\"`$;&|<>(){}[]!\\*?\n\r") {
		return fmt.Errorf("path contains forbidden characters: %q", p)
	}
	for _, seg := range strings.Split(filepath.ToSlash(p), "/") {
		if seg == ".." {
			return fmt.Errorf("path must not contain '..': %q", p)
		}
	}
	return nil
}

func (s *MaintenanceService) SlimHistory(repoPath string, paths []string, addGitignore bool, taskID string) error {
	for _, p := range paths {
		if err := validateSlimPath(p); err != nil {
			return err
		}
	}
	tm := GlobalTaskManager
	dao := db.NewMaintenanceDAO()
	appendLog := func(msg string) {
		tm.AppendLog(taskID, msg)
		record, _ := dao.FindByTaskID(taskID)
		if record != nil {
			t, ok := tm.GetTask(taskID)
			if ok {
				logJSON, _ := json.Marshal(t.Progress)
				dao.UpdateProgress(taskID, string(logJSON))
			}
		}
	}
	appendLog("开始仓库瘦身...")
	if addGitignore {
		appendLog("更新 .gitignore...")
		if err := appendToGitignore(repoPath, paths); err != nil {
			appendLog("警告: 更新 .gitignore 失败: " + err.Error())
		}
	}

	needCleanup := false
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = repoPath
	if statusOut, statusErr := statusCmd.CombinedOutput(); statusErr == nil && strings.TrimSpace(string(statusOut)) != "" {
		needCleanup = true
		appendLog("检测到未提交变更，临时提交...")
		addCmd := exec.Command("git", "add", "-A")
		addCmd.Dir = repoPath
		addCmd.Run()
		commitCmd := exec.Command("git", "commit", "-m", "chore: temp commit for repo slim")
		commitCmd.Dir = repoPath
		commitCmd.Run()
	}

	appendLog("执行 filter-branch 清除历史文件...")
	// 每个路径单独单引号包裹：filter-branch 内部会 eval filter 脚本，
	// 不加引号时路径中的元字符会被当作命令执行。
	quoted := make([]string, len(paths))
	for i, p := range paths {
		quoted[i] = "'" + strings.ReplaceAll(p, "'", "'\\''") + "'"
	}
	indexFilter := "git rm --cached --ignore-unmatch -- " + strings.Join(quoted, " ")
	cmdStr := "git filter-branch --force --index-filter '" + strings.ReplaceAll(indexFilter, "'", "'\\''") + "' --prune-empty -- --all"
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "FILTER_BRANCH_SQUELCH_WARNING=1", "GIT_ASKPASS=", "GIT_TERMINAL_PROMPT=0")
	cmd.Stdin = nil
	output, err := cmd.CombinedOutput()
	if err != nil {
		if needCleanup {
			resetCmd := exec.Command("git", "reset", "--soft", "HEAD~1")
			resetCmd.Dir = repoPath
			resetCmd.Run()
		}
		return fmt.Errorf("filter-branch failed: %w, output: %s", err, string(output))
	}
	appendLog("filter-branch 完成")

	if needCleanup {
		appendLog("撤销临时提交...")
		resetCmd := exec.Command("git", "reset", "--soft", "HEAD~1")
		resetCmd.Dir = repoPath
		resetCmd.Run()
	}
	appendLog("更新指向旧 commit 的 tags...")
	cmd = exec.Command("git", "tag", "-l")
	cmd.Dir = repoPath
	tagOutput, _ := cmd.CombinedOutput()
	for _, tag := range strings.Split(strings.TrimSpace(string(tagOutput)), "\n") {
		if tag == "" {
			continue
		}
		cmd = exec.Command("git", "rev-parse", tag+"^{}")
		cmd.Dir = repoPath
		commitHash, err := cmd.CombinedOutput()
		if err != nil {
			continue
		}
		commit := strings.TrimSpace(string(commitHash))
		cmd = exec.Command("git", "ls-tree", "-r", commit)
		cmd.Dir = repoPath
		treeOut, err := cmd.CombinedOutput()
		if err != nil {
			continue
		}
		needsUpdate := false
		for _, p := range paths {
			if strings.Contains(string(treeOut), "\t"+p+"\n") || strings.Contains(string(treeOut), "\t"+p+" ") {
				needsUpdate = true
				break
			}
		}
		if needsUpdate {
			tagCmd := exec.Command("git", "tag", "-f", tag, "HEAD")
			tagCmd.Dir = repoPath
			tagCmd.Run()
			appendLog("更新 tag: " + tag)
		}
	}
	appendLog("清理 backup refs...")
	cmd = exec.Command("git", "for-each-ref", "--format=%(refname)", "refs/original/")
	cmd.Dir = repoPath
	refsOutput, _ := cmd.CombinedOutput()
	for _, ref := range strings.Split(strings.TrimSpace(string(refsOutput)), "\n") {
		if ref != "" {
			delCmd := exec.Command("git", "update-ref", "-d", ref)
			delCmd.Dir = repoPath
			delCmd.Run()
		}
	}
	appendLog("清理 reflog...")
	cmd = exec.Command("git", "reflog", "expire", "--expire=now", "--all")
	cmd.Dir = repoPath
	if reflogOut, reflogErr := cmd.CombinedOutput(); reflogErr != nil {
		appendLog("reflog expire 警告: " + string(reflogOut))
	}
	appendLog("执行 gc --prune=now ...")
	cmd = exec.Command("git", "gc", "--prune=now", "--force")
	cmd.Dir = repoPath
	gcOutput, gcErr := cmd.CombinedOutput()
	if gcErr != nil {
		appendLog("gc 警告: " + string(gcOutput))
	} else {
		appendLog("gc 完成")
	}
	appendLog("执行 prune 清理不可达对象...")
	cmd = exec.Command("git", "prune", "--expire=now")
	cmd.Dir = repoPath
	if pruneOut, pruneErr := cmd.CombinedOutput(); pruneErr != nil {
		appendLog("prune 警告: " + string(pruneOut))
	} else {
		appendLog("prune 完成")
	}

	appendLog("验证清理结果...")
	if stillExist := s.verifyPathsRemoved(repoPath, paths); len(stillExist) > 0 {
		appendLog("检测到残留文件，执行二次清理: " + strings.Join(stillExist, ", "))
		reflogRetry := exec.Command("git", "reflog", "expire", "--expire=now", "--all")
		reflogRetry.Dir = repoPath
		reflogRetry.Run()
		gcRetry := exec.Command("git", "gc", "--prune=now", "--force")
		gcRetry.Dir = repoPath
		gcRetry.Run()
		pruneRetry := exec.Command("git", "prune", "--expire=now")
		pruneRetry.Dir = repoPath
		pruneRetry.Run()
		if retryExist := s.verifyPathsRemoved(repoPath, paths); len(retryExist) > 0 {
			appendLog("警告: 以下文件仍在对象库中: " + strings.Join(retryExist, ", "))
		} else {
			appendLog("二次清理成功")
		}
	} else {
		appendLog("验证通过，目标文件已从历史中移除")
	}

	afterSnap := s.TakeSnapshot(repoPath)
	afterJSON, _ := json.Marshal(afterSnap)
	dao.UpdateStatus(taskID, "success", "", string(afterJSON))
	appendLog("仓库瘦身完成！")
	return nil
}

func (s *MaintenanceService) verifyPathsRemoved(repoPath string, paths []string) []string {
	var stillExist []string
	batchOutput, err := runGitCmd(repoPath, "cat-file", "--batch-check", "--batch-all-objects")
	if err != nil {
		return stillExist
	}
	blobSizes := make(map[string]int64)
	for _, line := range strings.Split(string(batchOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == "blob" {
			if size, err := strconv.ParseInt(fields[2], 10, 64); err == nil {
				blobSizes[fields[0]] = size
			}
		}
	}
	revOutput, err := runGitCmd(repoPath, "rev-list", "--objects", "--all")
	if err != nil {
		return stillExist
	}
	for _, line := range strings.Split(string(revOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 {
			if _, ok := blobSizes[fields[0]]; ok {
				for _, p := range paths {
					if fields[1] == p {
						stillExist = append(stillExist, p)
						break
					}
				}
			}
		}
	}
	reflogOutput, err := runGitCmd(repoPath, "reflog", "--format=%H")
	if err != nil || strings.TrimSpace(string(reflogOutput)) == "" {
		return stillExist
	}
	for _, commitSHA := range strings.Split(strings.TrimSpace(string(reflogOutput)), "\n") {
		commitSHA = strings.TrimSpace(commitSHA)
		if commitSHA == "" {
			continue
		}
		treeOutput, err := runGitCmd(repoPath, "ls-tree", "-r", commitSHA)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(treeOutput), "\n") {
			parts := strings.SplitN(line, "\t", 2)
			if len(parts) < 2 {
				continue
			}
			meta := strings.Fields(parts[0])
			if len(meta) < 3 {
				continue
			}
			blobSHA := meta[2]
			filePath := parts[1]
			if _, ok := blobSizes[blobSHA]; ok {
				for _, p := range paths {
					if filePath == p {
						stillExist = append(stillExist, p)
						break
					}
				}
			}
		}
	}
	seen := make(map[string]bool)
	var unique []string
	for _, p := range stillExist {
		if !seen[p] {
			seen[p] = true
			unique = append(unique, p)
		}
	}
	return unique
}
func (s *MaintenanceService) SlimHistoryByPrefix(repoPath string, prefixes []string, addGitignore bool, taskID string) error {
	tm := GlobalTaskManager
	dao := db.NewMaintenanceDAO()
	appendLog := func(msg string) {
		tm.AppendLog(taskID, msg)
		record, _ := dao.FindByTaskID(taskID)
		if record != nil {
			t, ok := tm.GetTask(taskID)
			if ok {
				logJSON, _ := json.Marshal(t.Progress)
				dao.UpdateProgress(taskID, string(logJSON))
			}
		}
	}

	appendLog("扫描匹配前缀的文件...")
	files, err := s.FindByPrefix(repoPath, prefixes)
	if err != nil {
		return fmt.Errorf("scan prefix files failed: %w", err)
	}
	if len(files) == 0 {
		appendLog("未找到匹配的文件")
		return nil
	}

	paths := make([]string, 0, len(files))
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	appendLog(fmt.Sprintf("找到 %d 个匹配文件，总大小 %s", len(paths), formatSize(sumSizeBytes(files))))

	return s.SlimHistory(repoPath, paths, addGitignore, taskID)
}

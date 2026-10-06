package git

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func (s *MaintenanceService) GarbageCollect(repoPath string, taskID string) error {
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
	appendLog("开始垃圾回收...")
	appendLog("清理 reflog...")
	exec.Command("git", "reflog", "expire", "--expire=now", "--all").Run()
	appendLog("执行 git gc --aggressive --prune=now ...")
	cmd := exec.Command("git", "gc", "--aggressive", "--prune=now")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git gc failed: %w, output: %s", err, string(output))
	}
	afterSnap := s.TakeSnapshot(repoPath)
	afterJSON, _ := json.Marshal(afterSnap)
	dao.UpdateStatus(taskID, "success", "", string(afterJSON))
	appendLog("垃圾回收完成！")
	return nil
}

func (s *MaintenanceService) AddToGitignore(repoPath string, paths []string) error {
	return appendToGitignore(repoPath, paths)
}

func CreateMaintenanceRecord(repoID uint, opType string, repoPath string) (*po.MaintenanceRecord, error) {
	svc := NewMaintenanceService()
	beforeSnap := svc.TakeSnapshot(repoPath)
	beforeJSON, _ := json.Marshal(beforeSnap)
	now := time.Now()
	record := &po.MaintenanceRecord{
		RepoID:         repoID,
		Type:           opType,
		Status:         "pending",
		TriggerBy:      "manual",
		SnapshotBefore: string(beforeJSON),
		StartedAt:      &now,
	}
	if err := db.NewMaintenanceDAO().Create(record); err != nil {
		return nil, err
	}
	return record, nil
}

func appendToGitignore(repoPath string, paths []string) error {
	gitignorePath := filepath.Join(repoPath, ".gitignore")
	var existing []byte
	existing, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(existing)
	var newLines []string
	for _, p := range paths {
		line := "/" + p
		if strings.Contains(content, line) {
			continue
		}
		newLines = append(newLines, line)
	}
	if len(newLines) == 0 {
		return nil
	}
	suffix := "\n"
	if len(existing) == 0 || !strings.HasSuffix(content, "\n") {
		suffix = "\n"
	}
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(suffix + "# Auto-added by repo slim\n")
	for _, line := range newLines {
		f.WriteString(line + "\n")
	}
	return nil
}

package git

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	conf "github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-platform-sdk/gitbackend"
)

type GitService struct {
	backend gitbackend.GitBackend
}

func NewGitService() *GitService {
	backend, err := gitbackend.NewGitBackend(gitbackend.Options{Type: ""})
	if err != nil {
		log.Printf("[WARN] Failed to create git backend: %v, falling back to native", err)
		backend, _ = gitbackend.NewGitBackend(gitbackend.Options{Type: "native"})
	}
	return &GitService{backend: backend}
}

// RunCommand executes a raw git command.
// Deprecated: Ideally use go-git methods. However, kept for operations not fully supported by go-git (e.g. Merge logic, Config branch description).
func (s *GitService) RunCommand(dir string, args ...string) (string, error) {
	cmdStr := strings.Join(args, " ")
	if conf.DebugMode {
		log.Printf("[DEBUG] Executing in %s: git %s", dir, cmdStr)
	} else {
		log.Printf("[INFO] Executing git command: %s", cmdStr)
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	// Prevent password prompts and force English output
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[ERROR] Git command failed: %s, output: %s", err, string(out))
		return string(out), fmt.Errorf("git command failed: %s, output: %s", err, string(out))
	}
	output := strings.TrimSpace(string(out))
	if conf.DebugMode {
		log.Printf("[DEBUG] Git command output: %s", output)
	}
	return output, nil
}

// FetchWithDBKey 使用数据库 SSH密钥进行 fetch
func (s *GitService) FetchWithDBKey(path, remoteURL, privateKey, passphrase string, progress io.Writer, refSpecs ...string) error {
	auth, err := s.dbKeyAuth(privateKey, passphrase)
	if err != nil {
		return err
	}

	_, err = s.backend.Fetch(context.Background(), gitbackend.FetchOptions{
		RepoPath: path,
		Remote:   remoteURL,
		Tags:     true,
		Auth:     auth,
		Progress: progress,
	})
	return err
}

// CloneWithDBKey 使用数据库 SSH密钥进行克隆
func (s *GitService) CloneWithDBKey(remoteURL, localPath, privateKey, passphrase string, progressChan chan string) error {
	auth, err := s.dbKeyAuth(privateKey, passphrase)
	if err != nil {
		return err
	}

	var progress io.Writer
	if progressChan != nil {
		progress = &channelWriter{ch: progressChan}
	}

	return s.backend.Clone(context.Background(), gitbackend.CloneOptions{
		URL:      remoteURL,
		Path:     localPath,
		Auth:     auth,
		Progress: progress,
	})
}

// PushWithDBKey 使用数据库 SSH密钥进行 push
func (s *GitService) PushWithDBKey(path, targetRemoteURL, sourceHash, targetBranch, privateKey, passphrase string, options []string, progress io.Writer) error {
	auth, err := s.dbKeyAuth(privateKey, passphrase)
	if err != nil {
		return err
	}

	force, mirror, _ := parsePushOptions(options)
	refSpec := fmt.Sprintf("%s:refs/heads/%s", sourceHash, targetBranch)

	_, err = s.backend.Push(context.Background(), gitbackend.PushOptions{
		RepoPath: path,
		Remote:   targetRemoteURL,
		RefSpecs: []string{refSpec},
		Force:    force,
		Mirror:   mirror,
		Auth:     auth,
		Progress: progress,
	})
	return err
}

func (s *GitService) AddAndCommit(repoPath string, filePath string, message string) error {
	cmd := exec.Command("git", "add", "-A", "--", filePath)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "nothing to commit") {
			return nil
		}
		return fmt.Errorf("git commit failed: %w, output: %s", err, string(output))
	}
	return nil
}

func (s *GitService) RemoveAndCommit(repoPath string, filePath string, message string) error {
	_, statErr := os.Stat(filepath.Join(repoPath, filePath))
	var stageCmd *exec.Cmd
	if os.IsNotExist(statErr) {
		stageCmd = exec.Command("git", "add", "-A", "--", filePath)
	} else {
		stageCmd = exec.Command("git", "rm", "-f", "--", filePath)
	}
	stageCmd.Dir = repoPath
	if output, err := stageCmd.CombinedOutput(); err != nil {
		if !os.IsNotExist(statErr) && !strings.Contains(string(output), "did not match any files") {
			return fmt.Errorf("git stage remove failed: %w, output: %s", err, string(output))
		}
	}

	diffCmd := exec.Command("git", "diff", "--cached", "--quiet")
	diffCmd.Dir = repoPath
	if err := diffCmd.Run(); err == nil {
		return nil
	}

	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = repoPath
	output, err := commitCmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "nothing to commit") {
			return nil
		}
		return fmt.Errorf("git commit failed: %w, output: %s", err, string(output))
	}
	return nil
}

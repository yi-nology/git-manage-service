package git

import (
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

func (s *GitService) Backend() gitbackend.GitBackend {
	return s.backend
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

// FetchWithDBKey 使用数据库 SSH密钥进行 fetch（使用原生 git 命令，更可靠）
func (s *GitService) FetchWithDBKey(path, remoteURL, privateKey, passphrase string, progress io.Writer, refSpecs ...string) error {
	log.Printf("[INFO] Starting git fetch with DB key for repository: %s", path)
	helper := NewSSHKeyHelper()

	// 处理私钥内容
	log.Printf("[DEBUG] Processing private key")
	keyContent, err := helper.ProcessPrivateKey(privateKey, passphrase)
	if err != nil {
		log.Printf("[ERROR] Failed to process private key: %v", err)
		return fmt.Errorf("failed to process private key: %v", err)
	}

	// 创建临时密钥文件
	log.Printf("[DEBUG] Creating temporary key file")
	tmpFile, err := helper.CreateTempKeyFile(keyContent)
	if err != nil {
		log.Printf("[ERROR] Failed to create temporary key file: %v", err)
		return fmt.Errorf("failed to create temporary key file: %v", err)
	}
	defer helper.CleanupTempFile(tmpFile)

	// 构建 SSH 命令
	sshCmd := helper.BuildSSHCommand(tmpFile)

	args := []string{"fetch", remoteURL}
	args = append(args, refSpecs...)
	cmdStr := strings.Join(args, " ")

	log.Printf("[INFO] Executing git fetch: %s", cmdStr)
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	if progress != nil {
		cmd.Stdout = progress
		cmd.Stderr = progress
		if err := cmd.Run(); err != nil {
			log.Printf("[ERROR] Git fetch failed: %v", err)
			return fmt.Errorf("git fetch failed: %v", err)
		}
	} else {
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("[ERROR] Git fetch failed: %v, output: %s", err, string(output))
			return fmt.Errorf("git fetch failed: %v, output: %s", err, string(output))
		}
		if conf.DebugMode {
			log.Printf("[DEBUG] Git fetch output: %s", string(output))
		}
	}

	log.Printf("[INFO] Git fetch completed successfully")
	return nil
}

// CloneWithDBKey 使用数据库 SSH密钥进行克隆（使用原生 git 命令，更可靠）
func (s *GitService) CloneWithDBKey(remoteURL, localPath, privateKey, passphrase string, progressChan chan string) error {
	log.Printf("[INFO] Starting git clone with DB key: %s -> %s", remoteURL, localPath)
	helper := NewSSHKeyHelper()

	// 处理私钥内容
	log.Printf("[DEBUG] Processing private key")
	keyContent, err := helper.ProcessPrivateKey(privateKey, passphrase)
	if err != nil {
		log.Printf("[ERROR] Failed to process private key: %v", err)
		return fmt.Errorf("failed to process private key: %v", err)
	}

	// 创建临时密钥文件
	log.Printf("[DEBUG] Creating temporary key file")
	tmpFile, err := helper.CreateTempKeyFile(keyContent)
	if err != nil {
		log.Printf("[ERROR] Failed to create temporary key file: %v", err)
		return fmt.Errorf("failed to create temporary key file: %v", err)
	}
	defer helper.CleanupTempFile(tmpFile)

	// 构建 SSH 命令
	sshCmd := helper.BuildSSHCommand(tmpFile)

	cmd := exec.Command("git", "clone", remoteURL, localPath)
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	if progressChan != nil {
		cmd.Stdout = &channelWriter{ch: progressChan}
		cmd.Stderr = &channelWriter{ch: progressChan}
	}

	log.Printf("[INFO] Executing git clone command")
	if err := cmd.Run(); err != nil {
		log.Printf("[ERROR] Git clone failed: %v", err)
		return fmt.Errorf("git clone failed: %v", err)
	}

	log.Printf("[INFO] Git clone completed successfully: %s", localPath)
	return nil
}

// PushWithDBKey 使用数据库 SSH密钥进行 push（使用原生 git 命令，更可靠）
func (s *GitService) PushWithDBKey(path, targetRemoteURL, sourceHash, targetBranch, privateKey, passphrase string, options []string, progress io.Writer) error {
	log.Printf("[INFO] Starting git push with DB key for repository: %s", path)
	helper := NewSSHKeyHelper()

	// 处理私钥内容
	log.Printf("[DEBUG] Processing private key")
	keyContent, err := helper.ProcessPrivateKey(privateKey, passphrase)
	if err != nil {
		log.Printf("[ERROR] Failed to process private key: %v", err)
		return fmt.Errorf("failed to process private key: %v", err)
	}

	// 创建临时密钥文件
	log.Printf("[DEBUG] Creating temporary key file")
	tmpFile, err := helper.CreateTempKeyFile(keyContent)
	if err != nil {
		log.Printf("[ERROR] Failed to create temporary key file: %v", err)
		return fmt.Errorf("failed to create temporary key file: %v", err)
	}
	defer helper.CleanupTempFile(tmpFile)

	// 构建 SSH 命令
	sshCmd := helper.BuildSSHCommand(tmpFile)

	refSpec := fmt.Sprintf("%s:refs/heads/%s", sourceHash, targetBranch)
	args := []string{"push", targetRemoteURL, refSpec}
	args = append(args, options...)
	cmdStr := strings.Join(args, " ")

	log.Printf("[INFO] Executing git push: %s", cmdStr)
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	if progress != nil {
		cmd.Stdout = progress
		cmd.Stderr = progress
		if err := cmd.Run(); err != nil {
			log.Printf("[ERROR] Git push failed: %v", err)
			return fmt.Errorf("git push failed: %v", err)
		}
	} else {
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("[ERROR] Git push failed: %v, output: %s", err, string(output))
			return fmt.Errorf("git push failed: %v, output: %s", err, string(output))
		}
		if conf.DebugMode {
			log.Printf("[DEBUG] Git push output: %s", string(output))
		}
	}

	log.Printf("[INFO] Git push completed successfully")
	return nil
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

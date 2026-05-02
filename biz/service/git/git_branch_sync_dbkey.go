package git

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"strings"

	ssh2 "golang.org/x/crypto/ssh"
)

func (s *GitService) PushBranchWithDBKey(path, remote, branch, privateKey, passphrase string) error {
	tmpFile, err := os.CreateTemp("", "git_ssh_key_*")
	if err != nil {
		return fmt.Errorf("failed to create temp key file: %v", err)
	}
	tmpKeyPath := tmpFile.Name()
	defer os.Remove(tmpKeyPath)

	keyContent := privateKey
	if !strings.HasSuffix(keyContent, "\n") {
		keyContent += "\n"
	}

	if passphrase != "" {
		rawKey, err := ssh2.ParseRawPrivateKeyWithPassphrase([]byte(keyContent), []byte(passphrase))
		if err != nil {
			return fmt.Errorf("failed to parse encrypted private key: %v", err)
		}

		pemBytes, err := x509.MarshalPKCS8PrivateKey(rawKey)
		if err != nil {
			return fmt.Errorf("failed to marshal private key: %v", err)
		}

		pemBlock := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: pemBytes,
		}
		keyContent = string(pem.EncodeToMemory(pemBlock))
	}

	if _, err := tmpFile.WriteString(keyContent); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write key file: %v", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpKeyPath, 0600); err != nil {
		return fmt.Errorf("failed to set key file permissions: %v", err)
	}

	urlCmd := exec.Command("git", "remote", "get-url", remote)
	urlCmd.Dir = path
	urlOutput, err := urlCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %v", err)
	}
	remoteURL := strings.TrimSpace(string(urlOutput))

	sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o IdentitiesOnly=yes", tmpKeyPath)

	refSpec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)
	cmd := exec.Command("git", "push", remoteURL, refSpec)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %v, output: %s", err, string(output))
	}

	return nil
}

func (s *GitService) PullBranchWithDBKey(path, remote, branch, privateKey, passphrase string) error {
	tmpFile, err := os.CreateTemp("", "git_ssh_key_*")
	if err != nil {
		return fmt.Errorf("failed to create temp key file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	keyContent := privateKey
	if !strings.HasSuffix(keyContent, "\n") {
		keyContent += "\n"
	}

	if passphrase != "" {
		rawKey, err := ssh2.ParseRawPrivateKeyWithPassphrase([]byte(keyContent), []byte(passphrase))
		if err != nil {
			return fmt.Errorf("failed to parse encrypted private key: %v", err)
		}

		pemBytes, err := x509.MarshalPKCS8PrivateKey(rawKey)
		if err != nil {
			return fmt.Errorf("failed to marshal private key: %v", err)
		}

		pemBlock := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: pemBytes,
		}
		keyContent = string(pem.EncodeToMemory(pemBlock))
	}

	if _, err := tmpFile.WriteString(keyContent); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write key file: %v", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0600); err != nil {
		return fmt.Errorf("failed to set key file permissions: %v", err)
	}

	sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null", tmpFile.Name())

	cmd := exec.Command("git", "pull", remote, branch)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %v, output: %s", err, string(output))
	}

	return nil
}

func (s *GitService) FetchAllWithDBKey(path, privateKey, passphrase string) error {
	tmpFile, err := os.CreateTemp("", "git_ssh_key_*")
	if err != nil {
		return fmt.Errorf("failed to create temp key file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	keyContent := privateKey
	if !strings.HasSuffix(keyContent, "\n") {
		keyContent += "\n"
	}

	if passphrase != "" {
		rawKey, err := ssh2.ParseRawPrivateKeyWithPassphrase([]byte(keyContent), []byte(passphrase))
		if err != nil {
			return fmt.Errorf("failed to parse encrypted private key: %v", err)
		}

		pemBytes, err := x509.MarshalPKCS8PrivateKey(rawKey)
		if err != nil {
			return fmt.Errorf("failed to marshal private key: %v", err)
		}

		pemBlock := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: pemBytes,
		}
		keyContent = string(pem.EncodeToMemory(pemBlock))
	}

	if _, err := tmpFile.WriteString(keyContent); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write key file: %v", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0600); err != nil {
		return fmt.Errorf("failed to set key file permissions: %v", err)
	}

	sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o IdentitiesOnly=yes", tmpFile.Name())

	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch --all failed: %v, output: %s", err, string(output))
	}

	return nil
}

func (s *GitService) FetchBranchWithDBKey(path, remote, branch, privateKey, passphrase string) error {
	tmpFile, err := os.CreateTemp("", "git_ssh_key_*")
	if err != nil {
		return fmt.Errorf("failed to create temp key file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	keyContent := privateKey
	if !strings.HasSuffix(keyContent, "\n") {
		keyContent += "\n"
	}

	if passphrase != "" {
		rawKey, err := ssh2.ParseRawPrivateKeyWithPassphrase([]byte(keyContent), []byte(passphrase))
		if err != nil {
			return fmt.Errorf("failed to parse encrypted private key: %v", err)
		}

		pemBytes, err := x509.MarshalPKCS8PrivateKey(rawKey)
		if err != nil {
			return fmt.Errorf("failed to marshal private key: %v", err)
		}

		pemBlock := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: pemBytes,
		}
		keyContent = string(pem.EncodeToMemory(pemBlock))
	}

	if _, err := tmpFile.WriteString(keyContent); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write key file: %v", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0600); err != nil {
		return fmt.Errorf("failed to set key file permissions: %v", err)
	}

	sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null", tmpFile.Name())

	cmd := exec.Command("git", "fetch", remote, branch)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch failed: %v, output: %s", err, string(output))
	}

	return nil
}

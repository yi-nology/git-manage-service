package codereview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/pkg/logger"
)

type BaseCLIService struct {
	ExecPath string
	Config   map[string]interface{}
}

func (s *BaseCLIService) runCommand(ctx context.Context, args []string, dir string) (string, int, error) {
	start := time.Now()
	cmd := exec.CommandContext(ctx, s.ExecPath, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := int(time.Since(start).Seconds())
	output := stdout.String()

	if err != nil {
		return output, duration, fmt.Errorf("CLI execution failed: %w, stderr: %s", err, stderr.String())
	}

	return output, duration, nil
}

type ClaudeCLIService struct {
	BaseCLIService
}

func NewClaudeCLIService(config map[string]interface{}) *ClaudeCLIService {
	execPath := "claude"
	if path, ok := config["exec_path"].(string); ok && path != "" {
		execPath = path
	}
	return &ClaudeCLIService{BaseCLIService{ExecPath: execPath, Config: config}}
}

func (s *ClaudeCLIService) ReviewCode(ctx context.Context, req *CLIReviewRequest) (*CLIReviewResult, error) {
	args := []string{"--print"}
	if req.CustomPrompt != "" {
		args = append(args, "-p", req.CustomPrompt)
	}

	output, duration, err := s.runCommand(ctx, args, req.RepoPath)
	if err != nil {
		logger.ErrorWithErr("Claude CLI review failed", err, logrus.Fields{"duration": duration})
		return nil, err
	}

	return s.parseOutput(output, duration)
}

func (s *ClaudeCLIService) ValidateInstallation() error {
	cmd := exec.Command(s.ExecPath, "--version")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("claude CLI not found: %w (%s)", err, string(output))
	}
	return nil
}

func (s *ClaudeCLIService) GetVersion() (string, error) {
	cmd := exec.Command(s.ExecPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (s *ClaudeCLIService) parseOutput(output string, duration int) (*CLIReviewResult, error) {
	result := &CLIReviewResult{Content: output, Duration: duration}

	var parsed struct {
		Findings []CLIReviewIssue `json:"findings"`
		Summary  string           `json:"summary"`
		Score    int              `json:"score"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err == nil {
		result.Issues = parsed.Findings
		result.Summary = parsed.Summary
		result.Score = parsed.Score
	}

	return result, nil
}

type OpenCodeCLIService struct {
	BaseCLIService
}

func NewOpenCodeCLIService(config map[string]interface{}) *OpenCodeCLIService {
	execPath := "opencode"
	if path, ok := config["exec_path"].(string); ok && path != "" {
		execPath = path
	}
	return &OpenCodeCLIService{BaseCLIService{ExecPath: execPath, Config: config}}
}

func (s *OpenCodeCLIService) ReviewCode(ctx context.Context, req *CLIReviewRequest) (*CLIReviewResult, error) {
	args := []string{"-p", req.CustomPrompt}
	if req.CustomPrompt == "" {
		args = []string{"-p", "Review this code change and report issues in JSON format"}
	}

	output, duration, err := s.runCommand(ctx, args, req.RepoPath)
	if err != nil {
		return nil, err
	}

	result := &CLIReviewResult{Content: output, Duration: duration}
	var parsed struct {
		Findings []CLIReviewIssue `json:"findings"`
		Summary  string           `json:"summary"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err == nil {
		result.Issues = parsed.Findings
		result.Summary = parsed.Summary
	}
	return result, nil
}

func (s *OpenCodeCLIService) ValidateInstallation() error {
	cmd := exec.Command(s.ExecPath, "--version")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("opencode CLI not found: %w (%s)", err, string(output))
	}
	return nil
}

func (s *OpenCodeCLIService) GetVersion() (string, error) {
	cmd := exec.Command(s.ExecPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

type QoderCLIService struct {
	BaseCLIService
}

func NewQoderCLIService(config map[string]interface{}) *QoderCLIService {
	execPath := "qodercli"
	if path, ok := config["exec_path"].(string); ok && path != "" {
		execPath = path
	}
	return &QoderCLIService{BaseCLIService{ExecPath: execPath, Config: config}}
}

func (s *QoderCLIService) ReviewCode(ctx context.Context, req *CLIReviewRequest) (*CLIReviewResult, error) {
	args := []string{"review"}
	if req.CommitRange != "" {
		args = append(args, "--range", req.CommitRange)
	}
	if req.CustomPrompt != "" {
		args = append(args, "--prompt", req.CustomPrompt)
	}

	output, duration, err := s.runCommand(ctx, args, req.RepoPath)
	if err != nil {
		return nil, err
	}

	result := &CLIReviewResult{Content: output, Duration: duration}
	var parsed struct {
		Findings []CLIReviewIssue `json:"findings"`
		Summary  string           `json:"summary"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err == nil {
		result.Issues = parsed.Findings
		result.Summary = parsed.Summary
	}
	return result, nil
}

func (s *QoderCLIService) ValidateInstallation() error {
	cmd := exec.Command(s.ExecPath, "--version")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("qoder CLI not found: %w (%s)", err, string(output))
	}
	return nil
}

func (s *QoderCLIService) GetVersion() (string, error) {
	cmd := exec.Command(s.ExecPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

type CodexCLIService struct {
	BaseCLIService
}

func NewCodexCLIService(config map[string]interface{}) *CodexCLIService {
	execPath := "codex"
	if path, ok := config["exec_path"].(string); ok && path != "" {
		execPath = path
	}
	return &CodexCLIService{BaseCLIService{ExecPath: execPath, Config: config}}
}

func (s *CodexCLIService) ReviewCode(ctx context.Context, req *CLIReviewRequest) (*CLIReviewResult, error) {
	prompt := req.CustomPrompt
	if prompt == "" {
		prompt = "Review this code change and report issues in JSON format"
	}
	args := []string{"-q", "--approval-mode", "full-auto", prompt}

	output, duration, err := s.runCommand(ctx, args, req.RepoPath)
	if err != nil {
		return nil, err
	}

	result := &CLIReviewResult{Content: output, Duration: duration}
	var parsed struct {
		Findings []CLIReviewIssue `json:"findings"`
		Summary  string           `json:"summary"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err == nil {
		result.Issues = parsed.Findings
		result.Summary = parsed.Summary
	}
	return result, nil
}

func (s *CodexCLIService) ValidateInstallation() error {
	cmd := exec.Command(s.ExecPath, "--version")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("codex CLI not found: %w (%s)", err, string(output))
	}
	return nil
}

func (s *CodexCLIService) GetVersion() (string, error) {
	cmd := exec.Command(s.ExecPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

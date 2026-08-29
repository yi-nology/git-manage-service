package stats

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/pkg/timefmt"
)

type BlameLineInfo struct {
	Author    string
	Email     string
	Timestamp int64
}

func (lc *LineCounter) isGitRepo(repoPath string) bool {
	gitDir := filepath.Join(repoPath, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

func (lc *LineCounter) getGitBlameInfo(repoPath, filePath, branch string) (map[int]*BlameLineInfo, error) {
	args := []string{"blame", "--line-porcelain"}
	if branch != "" {
		args = append(args, branch, "--")
	}

	relPath, err := filepath.Rel(repoPath, filePath)
	if err != nil {
		relPath = filePath
	}
	args = append(args, relPath)

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return lc.parseBlameOutput(string(output))
}

func (lc *LineCounter) parseBlameOutput(output string) (map[int]*BlameLineInfo, error) {
	result := make(map[int]*BlameLineInfo)
	lines := strings.Split(output, "\n")

	var currentLine int
	var currentInfo *BlameLineInfo

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if len(line) >= 40 && isHexString(line[:40]) {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				lineNum, _ := strconv.Atoi(parts[2])
				currentLine = lineNum
				currentInfo = &BlameLineInfo{}
			}
			continue
		}

		if currentInfo == nil {
			continue
		}

		if strings.HasPrefix(line, "author ") {
			currentInfo.Author = strings.TrimPrefix(line, "author ")
		} else if strings.HasPrefix(line, "author-mail ") {
			email := strings.TrimPrefix(line, "author-mail ")
			email = strings.Trim(email, "<>")
			currentInfo.Email = email
		} else if strings.HasPrefix(line, "author-time ") {
			timestamp, _ := strconv.ParseInt(strings.TrimPrefix(line, "author-time "), 10, 64)
			currentInfo.Timestamp = timestamp
		} else if strings.HasPrefix(line, "\t") {
			if currentLine > 0 && currentInfo != nil {
				result[currentLine] = currentInfo
			}
			currentInfo = nil
		}
	}

	return result, nil
}

func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func (lc *LineCounter) shouldCountLine(info *BlameLineInfo, config CountConfig) bool {
	if info == nil {
		return true
	}

	if config.Author != "" {
		authorMatch := strings.Contains(strings.ToLower(info.Author), strings.ToLower(config.Author)) ||
			strings.Contains(strings.ToLower(info.Email), strings.ToLower(config.Author))
		if !authorMatch {
			return false
		}
	}

	if config.Since != "" {
		sinceTime, err := time.Parse(timefmt.LayoutDate, config.Since)
		if err == nil && info.Timestamp < sinceTime.Unix() {
			return false
		}
	}

	if config.Until != "" {
		untilTime, err := time.Parse(timefmt.LayoutDate, config.Until)
		if err == nil {
			untilTime = untilTime.Add(24 * time.Hour)
			if info.Timestamp >= untilTime.Unix() {
				return false
			}
		}
	}

	return true
}

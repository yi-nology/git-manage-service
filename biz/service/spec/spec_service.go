package spec

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/pkg/timefmt"
)

type SpecService struct{}

func NewSpecService() *SpecService {
	return &SpecService{}
}

func (s *SpecService) ListSpecFiles(repoPath string) ([]SpecFileInfo, error) {
	entries, err := WalkSpecEntries(repoPath, SpecWalkOptions{
		IncludeDirs:    false,
		SkipAnyGitPath: true,
	})
	if err != nil {
		return nil, err
	}

	var files []SpecFileInfo
	for _, e := range entries {
		files = append(files, SpecFileInfo{
			Name:    e.Name,
			Path:    e.Path,
			IsDir:   e.IsDir,
			Size:    e.Size,
			ModTime: e.ModTime.Format(timefmt.LayoutAPITime),
		})
	}

	return files, nil
}

// resolveInRepo joins repoPath with a user-supplied relative path and rejects
// any result that escapes the repo directory (path traversal, e.g.
// "../../etc/passwd"). All user-supplied spec paths must go through this.
func resolveInRepo(repoPath, relPath string) (string, error) {
	fullPath := filepath.Join(repoPath, relPath)
	// filepath.Join cleans "..", so verify the result is still inside repoPath.
	if !strings.HasPrefix(fullPath, repoPath+string(os.PathSeparator)) && fullPath != repoPath {
		return "", fmt.Errorf("path %q escapes repository directory", relPath)
	}
	return fullPath, nil
}

func (s *SpecService) GetSpecContent(repoPath, specPath string) (string, error) {
	fullPath, err := resolveInRepo(repoPath, specPath)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read spec file: %v", err)
	}
	return string(content), nil
}

func (s *SpecService) SaveSpecContent(repoPath, specPath, content, commitMessage string) error {
	fullPath, err := resolveInRepo(repoPath, specPath)
	if err != nil {
		return err
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write spec file: %v", err)
	}

	return nil
}

func (s *SpecService) CreateSpecFile(repoPath, dirPath, fileName string) (string, error) {
	return s.CreateSpecFileWithContent(repoPath, dirPath, fileName, "")
}

func (s *SpecService) CreateSpecFileWithContent(repoPath, dirPath, fileName, content string) (string, error) {
	fullDir, err := resolveInRepo(repoPath, dirPath)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	fullPath, err := resolveInRepo(fullDir, fileName)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(fullPath); err == nil {
		return "", fmt.Errorf("file already exists: %s", fileName)
	}

	if content == "" {
		content = s.GetSpecTemplate()
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to create spec file: %v", err)
	}

	relPath, _ := filepath.Rel(repoPath, fullPath)
	return relPath, nil
}

func (s *SpecService) DeleteSpecFile(repoPath, specPath string) error {
	fullPath, err := resolveInRepo(repoPath, specPath)
	if err != nil {
		return err
	}
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete spec file: %v", err)
	}
	return nil
}

func (s *SpecService) GetSpecTemplate() string {
	tmpl := `Name:
Version:
Release:        1%{?dist}
Summary:

License:
URL:
Source0:

BuildRequires:
Requires:

%description


%prep
%setup -q

%build

%install
rm -rf $RPM_BUILD_ROOT

%clean
rm -rf $RPM_BUILD_ROOT

%files
%doc

%changelog
* $(date +"%a %b %d %Y") Your Name <your.email@example.com> - VERSION-1
- Initial package
`
	// Replace the shell date placeholder with the actual date (Go, not bash).
	tmpl = strings.ReplaceAll(tmpl, "$(date +\"%a %b %d %Y\")", time.Now().Format("Mon Jan 02 2006"))
	return tmpl
}

type SpecFileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size,omitempty"`
	ModTime string `json:"mod_time,omitempty"`
}

type SpecValidationResult struct {
	Valid    bool              `json:"valid"`
	Issues   []SpecIssue       `json:"issues"`
	Warnings []SpecIssue       `json:"warnings"`
	Stats    map[string]string `json:"stats"`
}

type SpecIssue struct {
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Rule     string `json:"rule"`
	RuleDesc string `json:"rule_desc"`
	QuickFix string `json:"quick_fix,omitempty"`
}

type SpecRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Pattern     string `json:"pattern"`
	Enabled     bool   `json:"enabled"`
	Category    string `json:"category"`
	AutoFix     bool   `json:"auto_fix"`
}

func (s *SpecService) ReadSpecLines(content string) []string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

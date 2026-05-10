package codereview

import (
	"fmt"
	"os/exec"
	"strings"
)

type ScannedCLI struct {
	CLIType     string `json:"cliType"`
	Name        string `json:"name"`
	ExecPath    string `json:"execPath"`
	Version     string `json:"version"`
	IsInstalled bool   `json:"isInstalled"`
}

var knownCLIs = []struct {
	cliType  string
	name     string
	binNames []string
}{
	{"claude_cli", "Claude Code", []string{"claude"}},
	{"opencode_cli", "OpenCode", []string{"opencode"}},
	{"qoder_cli", "Qoder CLI", []string{"qodercli"}},
	{"codex_cli", "Codex CLI", []string{"codex", "openai-codex"}},
}

func ScanInstalledCLIs() []ScannedCLI {
	results := make([]ScannedCLI, 0, len(knownCLIs))
	for _, cli := range knownCLIs {
		scanned := ScannedCLI{
			CLIType: cli.cliType,
			Name:    cli.name,
		}
		for _, bin := range cli.binNames {
			path, err := exec.LookPath(bin)
			if err == nil {
				scanned.ExecPath = path
				scanned.IsInstalled = true
				version, _ := detectVersion(path)
				scanned.Version = version
				break
			}
		}
		results = append(results, scanned)
	}
	return results
}

func detectVersion(execPath string) (string, error) {
	for _, arg := range []string{"--version", "-v", "version"} {
		out, err := exec.Command(execPath, arg).Output()
		if err == nil {
			v := strings.TrimSpace(string(out))
			if v != "" {
				if len(v) > 100 {
					v = v[:100]
				}
				return v, nil
			}
		}
	}
	return "", fmt.Errorf("version unknown")
}

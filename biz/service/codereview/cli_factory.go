package codereview

import "fmt"

func NewCLIService(cliType string, config map[string]interface{}) (CLIService, error) {
	switch cliType {
	case "claude", "claude_cli":
		return NewClaudeCLIService(config), nil
	case "opencode", "opencode_cli":
		return NewOpenCodeCLIService(config), nil
	case "qoder", "qoder_cli":
		return NewQoderCLIService(config), nil
	case "codex", "codex_cli":
		return NewCodexCLIService(config), nil
	default:
		return nil, fmt.Errorf("unsupported CLI type: %s", cliType)
	}
}

func GetSupportedCLITypes() []string {
	return []string{"claude", "opencode", "qoder", "codex"}
}
